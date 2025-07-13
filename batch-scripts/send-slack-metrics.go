package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Metrics struct {
	DBSizeMB         float64 `json:"db_size_mb"`
	TableCount       int     `json:"table_count"`
	ActiveConnections int     `json:"active_connections"`
}

type SlackMessage struct {
	Text   string       `json:"text,omitempty"`
	Blocks []SlackBlock `json:"blocks,omitempty"`
}

type SlackBlock struct {
	Type   string                 `json:"type"`
	Text   *SlackText             `json:"text,omitempty"`
	Fields []SlackField           `json:"fields,omitempty"`
}

type SlackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackField struct {
	Type string `json:"type"`
	Text string `json:"text"`
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

func collectMetrics(db *sql.DB) (*Metrics, error) {
	metrics := &Metrics{}

	// データベースのサイズ
	var dbSize int64
	err := db.QueryRow("SELECT pg_database_size(current_database())").Scan(&dbSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	metrics.DBSizeMB = float64(dbSize) / 1024 / 1024

	// テーブル数
	err = db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public'
	`).Scan(&metrics.TableCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get table count: %w", err)
	}

	// アクティブな接続数
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pg_stat_activity 
		WHERE state = 'active'
	`).Scan(&metrics.ActiveConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}

	return metrics, nil
}

func formatMessage(metrics *Metrics) *SlackMessage {
	reportFormat := os.Getenv("REPORT_FORMAT")
	metricsType := os.Getenv("METRICS_TYPE")
	if metricsType == "" {
		metricsType = "database"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

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

func sendToSlack(message *SlackMessage) error {
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

func main() {
	log.Println("Connecting to database...")
	db, err := getDBConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Collecting metrics...")
	metrics, err := collectMetrics(db)
	if err != nil {
		log.Fatalf("Failed to collect metrics: %v", err)
	}

	log.Println("Formatting message...")
	message := formatMessage(metrics)

	log.Println("Sending to Slack...")
	err = sendToSlack(message)
	if err != nil {
		log.Fatalf("Failed to send to Slack: %v", err)
	}

	log.Println("Metrics sent successfully!")
}