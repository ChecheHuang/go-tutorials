# 第二十八課：Wire 依賴注入框架

> **一句話總結**：Wire 讓你只需要描述「這個元件需要什麼」，它自動生成組裝程式碼，讓大型專案的 DI 不再是惡夢。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解為什麼需要 DI 框架 |
| 🔴 資深工程師 | **重點**：Wire 在大型 Go 專案中幾乎是標配 |

## 執行方式

```bash
go run ./tutorials/28-wire
```

## Wire 的工作流程

```
你寫的 wire.go：
  func InitializeApp() (*App, error) {
      wire.Build(ProvideConfig, ProvideDB, ProvideRepo, ProvideUsecase, ...)
      return nil, nil
  }
          ↓ wire gen
Wire 生成的 wire_gen.go：
  func InitializeApp() (*App, error) {
      config := ProvideConfig()
      db, err := ProvideDB(config)
      ...
      return ProvideApp(config, db, ...), nil
  }
```

## 安裝 Wire

```bash
go install github.com/google/wire/cmd/wire@latest
wire gen ./your-package/
```

## Provider 範例

```go
// Provider = 建立元件的函式
// Wire 從參數類型分析依賴關係
func ProvideDB(cfg *Config) (*Database, error) { ... }       // 依賴 Config
func ProvideRepo(db *Database, log *Logger) *Repo { ... }    // 依賴 Database + Logger
func ProvideUsecase(repo *Repo) *Usecase { ... }             // 依賴 Repo
```

## 何時用 Wire？

| | 手動 DI | Wire |
|--|--------|------|
| 適合 | 小專案（< 10 個元件）| 大型專案（10+ 元件）|
| 可讀性 | 直觀 | 需要學習 |
| 型別安全 | 是 | 是（編譯時）|
| 維護性 | 手動調整順序 | 自動分析 |
