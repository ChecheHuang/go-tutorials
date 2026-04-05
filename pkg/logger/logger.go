// Package logger 提供結構化日誌功能
// 基於 Go 1.21+ 的 log/slog，支援 JSON 與 Text 格式
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init 根據設定初始化全域 slog logger
// level: "debug", "info", "warn", "error"
// format: "json", "text"
func Init(level, format string) {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: logLevel}

	var handler slog.Handler
	if strings.ToLower(format) == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
