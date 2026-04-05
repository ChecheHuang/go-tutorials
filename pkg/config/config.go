// Package config 提供應用程式的設定管理功能
// 使用環境變數來設定伺服器、資料庫與 JWT 相關參數
package config

import (
	"os"
	"time"
)

// Config 定義應用程式的所有設定
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig 定義伺服器相關設定
type ServerConfig struct {
	Port string // 伺服器監聽的埠號
	Mode string // Gin 模式：debug / release / test
}

// DatabaseConfig 定義資料庫相關設定
type DatabaseConfig struct {
	DSN string // 資料來源名稱（SQLite 檔案路徑）
}

// JWTConfig 定義 JWT 認證相關設定
type JWTConfig struct {
	Secret     string        // JWT 簽名密鑰
	Expiration time.Duration // Token 過期時間
}

// Load 從環境變數載入設定，若未設定則使用預設值
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DB_DSN", "blog.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "my-secret-key-change-in-production"),
			Expiration: parseDuration(getEnv("JWT_EXPIRATION", "24h")),
		},
	}
}

// getEnv 取得環境變數，若不存在則回傳預設值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// parseDuration 解析時間長度字串，解析失敗時回傳 24 小時
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}
