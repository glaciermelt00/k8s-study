package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Metrics struct {
	DBSizeMB          float64   `json:"db_size_mb"`
	TableCount        int       `json:"table_count"`
	ActiveConnections int       `json:"active_connections"`
	Timestamp         time.Time `json:"timestamp"`
}

type SlackMessage struct {
	Text   string       `json:"text,omitempty"`
	Blocks []SlackBlock `json:"blocks,omitempty"`
}

type SlackBlock struct {
	Type   string       `json:"type"`
	Text   *SlackText   `json:"text,omitempty"`
	Fields []SlackField `json:"fields,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackField struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type APIServer struct {
	db *sql.DB
}

func NewAPIServer() (*APIServer, error) {
	db, err := getDBConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &APIServer{db: db}, nil
}

func getDBConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
	return sql.Open("postgres", connStr)
}

func (s *APIServer) collectMetrics() (*Metrics, error) {
	metrics := &Metrics{
		Timestamp: time.Now(),
	}

	// データベースのサイズ
	var dbSize int64
	err := s.db.QueryRow("SELECT pg_database_size(current_database())").Scan(&dbSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	metrics.DBSizeMB = float64(dbSize) / 1024 / 1024

	// テーブル数
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public'
	`).Scan(&metrics.TableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}

	// アクティブな接続数
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM pg_stat_activity
		WHERE state = 'active'
	`).Scan(&metrics.ActiveConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}

	return metrics, nil
}

func (s *APIServer) formatMessage(metrics *Metrics) *SlackMessage {
	reportFormat := os.Getenv("REPORT_FORMAT")
	metricsType := os.Getenv("METRICS_TYPE")
	if metricsType == "" {
		metricsType = "database"
	}

	timestamp := metrics.Timestamp.Format("2006-01-02 15:04:05")

	if reportFormat == "detailed" {
		return &SlackMessage{
			Blocks: []SlackBlock{
				{
					Type: "header",
					Text: &SlackText{
						Type: "plain_text",
						Text: fmt.Sprintf("📊 %s Metrics Report", metricsType),
					},
				},
				{
					Type: "section",
					Fields: []SlackField{
						{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*Timestamp:*\n%s", timestamp),
						},
						{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*Database Size:*\n%.2f MB", metrics.DBSizeMB),
						},
						{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*Table Count:*\n%d", metrics.TableCount),
						},
						{
							Type: "mrkdwn",
							Text: fmt.Sprintf("*Active Connections:*\n%d", metrics.ActiveConnections),
						},
					},
				},
			},
		}
	}

	text := fmt.Sprintf(`📊 Database Metrics Report - %s
• Database Size: %.2f MB
• Table Count: %d
• Active Connections: %d`,
		timestamp,
		metrics.DBSizeMB,
		metrics.TableCount,
		metrics.ActiveConnections,
	)
	return &SlackMessage{Text: text}
}

func (s *APIServer) sendToSlack(message *SlackMessage) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is not set")
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send to Slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Printf("Failed to read response body: %v", readErr)
		} else {
			log.Printf("Slack response body: %s", string(body))
		}
		return fmt.Errorf("Slack returned status code: %d", resp.StatusCode)
	}

	return nil
}

// HTTP Handlers
func (s *APIServer) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.collectMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to collect metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *APIServer) handleSendSlack(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.collectMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to collect metrics: %v", err), http.StatusInternalServerError)
		return
	}

	message := s.formatMessage(metrics)
	if err := s.sendToSlack(message); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send to Slack: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "Metrics sent to Slack successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	// データベース接続を確認
	if err := s.db.Ping(); err != nil {
		http.Error(w, fmt.Sprintf("Database connection failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	response := map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	log.Println("Starting Slack Metrics API Server...")

	server, err := NewAPIServer()
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}
	defer server.db.Close()

	// ルートハンドラ
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "slack-metrics-api",
			"version": "1.0.0",
		})
	})

	// API エンドポイント
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		server.handleGetMetrics(w, r)
	})

	http.HandleFunc("/slack/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		server.handleSendSlack(w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		server.handleHealth(w, r)
	})

	// サーバー起動
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}