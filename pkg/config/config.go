// Package config 提供應用程式的設定管理功能
// 使用 Viper 支援 YAML 設定檔 + 環境變數覆蓋
// 教學對應：第 20 課（Config 管理）
package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 定義應用程式的所有設定
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	Redis    RedisConfig
}

// ServerConfig 定義伺服器相關設定
type ServerConfig struct {
	Port           string
	Mode           string
	AllowedOrigins []string // CORS 白名單，空陣列或包含 "*" 時允許所有來源
}

// DatabaseConfig 定義資料庫相關設定
type DatabaseConfig struct {
	DSN string
}

// JWTConfig 定義 JWT 認證相關設定
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// LogConfig 定義日誌相關設定
type LogConfig struct {
	Level  string // debug / info / warn / error
	Format string // json / text
}

// RedisConfig 定義 Redis 相關設定
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Enabled  bool
}

// Load 從 config.yaml 載入設定，環境變數可覆蓋
func Load() *Config {
	v := viper.New()

	// 設定檔名稱與路徑
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// 設定預設值
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.allowed_origins", []string{}) // 預設空陣列，debug 模式下允許所有來源
	v.SetDefault("database.dsn", "blog.db")
	v.SetDefault("jwt.secret", "my-secret-key-change-in-production")
	v.SetDefault("jwt.expiration", "24h")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.enabled", false)

	// 環境變數覆蓋：SERVER_PORT、DB_DSN、JWT_SECRET 等
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 讀取設定檔（找不到也沒關係，用預設值 + 環境變數）
	if err := v.ReadInConfig(); err != nil {
		log.Printf("[設定] 未找到 config.yaml，使用預設值與環境變數")
	}

	expiration, err := time.ParseDuration(v.GetString("jwt.expiration"))
	if err != nil {
		expiration = 24 * time.Hour
	}

	jwtSecret := v.GetString("jwt.secret")
	mode := v.GetString("server.mode")

	// 非 debug 模式下，JWT secret 必須足夠強
	if mode != "debug" && len(jwtSecret) < 32 {
		log.Fatalf("[設定] JWT secret 長度不足：生產環境至少需要 32 字元（目前 %d 字元）。請設定環境變數 JWT_SECRET 或修改 config.yaml", len(jwtSecret))
	}
	if jwtSecret == "my-secret-key-change-in-production" && mode != "debug" {
		log.Fatalf("[設定] 偵測到預設 JWT secret，禁止在生產環境使用。請設定環境變數 JWT_SECRET")
	}

	return &Config{
		Server: ServerConfig{
			Port:           v.GetString("server.port"),
			Mode:           mode,
			AllowedOrigins: v.GetStringSlice("server.allowed_origins"),
		},
		Database: DatabaseConfig{
			DSN: v.GetString("database.dsn"),
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			Expiration: expiration,
		},
		Log: LogConfig{
			Level:  v.GetString("log.level"),
			Format: v.GetString("log.format"),
		},
		Redis: RedisConfig{
			Addr:     v.GetString("redis.addr"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
			Enabled:  v.GetBool("redis.enabled"),
		},
	}
}
