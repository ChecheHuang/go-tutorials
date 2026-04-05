# 第二十課：Config 管理

> **一句話總結**：設定值絕不能硬寫在程式碼裡，要從環境變數讀取，讓同一份程式碼能在不同環境運行。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：Config 管理是每個後端專案都需要的 |
| 🔴 資深工程師 | **必備**：多環境設定策略、Secret 管理、Config 驗證 |

## 執行方式

```bash
go run ./tutorials/20-config
```

## 核心概念

```go
// 優先順序：環境變數 > .env 檔案 > config.yaml > 程式碼預設值
type Config struct {
    App      AppConfig
    Database DatabaseConfig
    JWT      JWTConfig
}

// 讀取（viper 一行解決）
v.AutomaticEnv()
v.Unmarshal(&config)
```

## 三種方法比較

| 方法 | 適用場景 | 套件 |
|------|---------|------|
| `os.Getenv` | 簡單、少量設定 | 標準庫 |
| `viper` | 大型專案、多格式 | `github.com/spf13/viper` |
| K8s Secret | 正式環境敏感資訊 | 平台工具 |

## 安全規則

```
✅ 敏感資訊 → 環境變數（JWT_SECRET、DB_PASSWORD）
✅ 提供 .env.example（格式說明，不含實際值）
✅ .env 加入 .gitignore
❌ 不要把密碼寫在 config.yaml 然後 commit 進 git！
```
