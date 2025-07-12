package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "正常系：すべての環境変数が設定されている",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_SSLMODE":  "disable",
				"PORT":        "8080",
			},
			wantErr: false,
		},
		{
			name: "異常系：DB_PASSWORDが未設定",
			envVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"DB_USER": "testuser",
				"DB_NAME": "testdb",
			},
			wantErr: true,
			errMsg:  "DB_PASSWORD environment variable is required",
		},
		{
			name: "異常系：無効なDB_PORT",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "invalid",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
			},
			wantErr: true,
			errMsg:  "invalid DB_PORT: invalid",
		},
		{
			name: "異常系：範囲外のDB_PORT",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "70000",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
			},
			wantErr: true,
			errMsg:  "invalid DB_PORT: 70000",
		},
		{
			name: "異常系：無効なDB_SSLMODE",
			envVars: map[string]string{
				"DB_HOST":     "localhost",
				"DB_PORT":     "5432",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_SSLMODE":  "invalid-mode",
			},
			wantErr: true,
			errMsg:  "invalid DB_SSLMODE: invalid-mode",
		},
		{
			name: "正常系：デフォルト値の使用",
			envVars: map[string]string{
				"DB_PASSWORD": "testpass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数をクリア
			os.Clearenv()
			
			// テスト用の環境変数を設定
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			
			// 設定を読み込み
			cfg, err := Load()
			
			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// エラーメッセージのチェック
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Load() error = %v, want %v", err.Error(), tt.errMsg)
			}
			
			// 正常系の場合、設定が正しく読み込まれているか確認
			if err == nil && cfg == nil {
				t.Error("Load() returned nil config without error")
			}
		})
	}
}