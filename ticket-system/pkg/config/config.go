package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Log      LogConfig
	GRPC     GRPCConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type LogConfig struct {
	Level  string
	Format string
}

type GRPCConfig struct {
	Port string
}

func Load() *Config {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.dsn", "ticket.db")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("grpc.port", "9090")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Printf("[設定] 未找到 config.yaml，使用預設值")
	}

	return &Config{
		Server:   ServerConfig{Port: v.GetString("server.port"), Mode: v.GetString("server.mode")},
		Database: DatabaseConfig{DSN: v.GetString("database.dsn")},
		Redis:    RedisConfig{Addr: v.GetString("redis.addr"), Password: v.GetString("redis.password"), DB: v.GetInt("redis.db")},
		Log:      LogConfig{Level: v.GetString("log.level"), Format: v.GetString("log.format")},
		GRPC:     GRPCConfig{Port: v.GetString("grpc.port")},
	}
}
