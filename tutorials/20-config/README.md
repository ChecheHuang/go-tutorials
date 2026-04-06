# 第二十課：Config 管理（Configuration Management）

> **一句話總結**：設定值絕不能硬寫在程式碼裡，要從環境變數讀取，讓同一份程式碼能在不同環境運行。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：Config 管理是每個後端專案都需要的基本功 |
| 🔴 資深工程師 | **必備**：多環境設定策略、Secret 管理、Config 驗證、熱重載 |

## 你會學到什麼？

- 為什麼設定不能硬寫在程式碼裡，以及正確的設定管理思維
- 三種設定方式比較：`os.Getenv`、viper、K8s Secrets
- 如何用 viper 讀取 YAML + 環境變數，並解析到 struct
- Config struct 的設計原則：類型安全、巢狀結構、mapstructure 標籤
- 設定的優先順序：環境變數 > .env > config.yaml > 程式碼預設值
- 不同環境（開發/測試/正式）的設定策略
- `.env` 檔案格式與 `.env.example` 慣例
- 啟動時的 Config 驗證（Fail Fast 原則）
- 安全最佳實踐：永遠不要把密碼 commit 進 git

## 執行方式

```bash
go run ./tutorials/20-config
```

## 生活比喻：Config = 遙控器

```
想像你的程式是一台電視，Config 就是遙控器：

  遙控器（Config）讓你：
    按「1」→ 開發頻道：sqlite、debug 模式、port 8080
    按「2」→ 測試頻道：postgres 測試庫、info 模式、port 8081
    按「3」→ 正式頻道：postgres 正式庫、warn 模式、port 443

  不同的環境 = 不同的頻道
  同一台電視 = 同一份程式碼
  遙控器     = Config（從外部控制行為）

  ❌ 硬寫在程式碼裡 = 把頻道焊死在電視上，換台就要拆開重焊
  ✅ 用 Config 管理 = 按一下遙控器就能切換
```

## 為什麼設定不能硬寫？

```go
// ❌ 危險：硬寫在程式碼裡
db, _ := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
jwtSecret := "my-super-secret-key"

// 問題 1：換環境要改程式碼、重新編譯、重新部署
// 問題 2：密碼會被 commit 進 git → 全世界都看得到
// 問題 3：不同開發者的本地環境不同，互相衝突

// ✅ 正確：從外部讀取
db, _ := gorm.Open(sqlite.Open(config.Database.DSN), &gorm.Config{})
jwtSecret := config.JWT.Secret
```

## 設定的優先順序

```
環境變數（最高）
    ↓ 覆蓋
.env 檔案
    ↓ 覆蓋
config.yaml
    ↓ 覆蓋
程式碼預設值（最低）
```

**為什麼這樣設計？**

| 來源 | 用途 | 誰設定的 |
|------|------|---------|
| 環境變數 | 正式環境覆蓋（K8s、Docker） | 維運工程師 / CI/CD |
| `.env` 檔案 | 本地開發方便 | 開發者自己 |
| `config.yaml` | 所有環境共用的預設值 | 團隊共識（可 commit） |
| 程式碼預設值 | 萬一什麼都沒設定的 fallback | 程式開發者 |

正式環境用環境變數覆蓋最安全，因為環境變數不會被 commit 進 git，而且 K8s/Docker 原生支援。

## 三種設定方式比較

| 比較點 | `os.Getenv` | `viper` | K8s Secrets |
|--------|-------------|---------|-------------|
| **複雜度** | 最簡單 | 中等 | 需要 K8s 環境 |
| **支援格式** | 只有環境變數 | YAML/JSON/TOML/ENV/環境變數 | 環境變數或掛載檔案 |
| **類型轉換** | 手動（全是 string） | 自動（Unmarshal 到 struct） | 手動 |
| **預設值** | 自己寫 helper | `v.SetDefault()` 內建 | 不適用 |
| **熱重載** | 不支援 | `v.WatchConfig()` 支援 | 支援（Volume 掛載） |
| **適用場景** | 小專案、設定值少 | 中大型專案 | 正式環境敏感資訊 |
| **外部依賴** | 無（標準庫） | `github.com/spf13/viper` | K8s 平台 |

**建議**：
- 5 個以下的設定 → `os.Getenv` 就夠了
- 專案有多個模組 → 用 `viper`
- 正式環境密碼/密鑰 → K8s Secrets 或 HashiCorp Vault

## 方式一：os.Getenv（輕量版）

最簡單的方式，只用 Go 標準庫：

```go
// getEnvOrDefault 讀取環境變數，如果不存在則回傳預設值
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// 使用範例
config := &Config{
    App: AppConfig{
        Name: getEnvOrDefault("APP_NAME", "go-tutorials"),
        Port: func() int {
            port := 8080
            if p := os.Getenv("APP_PORT"); p != "" {
                fmt.Sscanf(p, "%d", &port)
            }
            return port
        }(),
        Env: getEnvOrDefault("APP_ENV", "development"),
    },
    Database: DatabaseConfig{
        Driver: getEnvOrDefault("DB_DRIVER", "sqlite"),
        DSN:    getEnvOrDefault("DB_DSN", "./blog.db"),
    },
}
```

**缺點**：所有值都是 `string`，需要手動轉型（像 Port 要 `Sscanf`）。設定一多就會很亂。

## 方式二：viper（推薦）

### viper 是什麼？

viper 是 Go 生態系最流行的設定管理套件，由 Cobra CLI 的作者開發。它支援：

- 讀取多種格式：YAML、JSON、TOML、HCL、envfile
- 自動讀取環境變數
- 設定預設值
- 熱重載（設定檔變更時自動更新）
- 遠端設定（etcd、Consul）

### 安裝

```bash
go get github.com/spf13/viper
```

### 基本用法

```go
func LoadConfig() (*Config, error) {
    v := viper.New()

    // ── 1. 設定預設值（最低優先）──
    v.SetDefault("app.name", "go-tutorials")
    v.SetDefault("app.port", 8080)
    v.SetDefault("app.env", "development")
    v.SetDefault("database.driver", "sqlite")
    v.SetDefault("database.dsn", "./blog.db")
    v.SetDefault("jwt.secret", "change-me-in-production")
    v.SetDefault("log.level", "info")

    // ── 2. 讀取 config.yaml（如果存在）──
    v.SetConfigName("config")     // 檔名（不含副檔名）
    v.SetConfigType("yaml")       // 格式
    v.AddConfigPath(".")          // 搜尋路徑 1：當前目錄
    v.AddConfigPath("./config")   // 搜尋路徑 2：config/ 子目錄
    v.AddConfigPath("$HOME/.app") // 搜尋路徑 3：Home 目錄

    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("讀取設定檔失敗: %w", err)
        }
        // 找不到設定檔是正常的（用環境變數和預設值）
    }

    // ── 3. 自動讀取環境變數（最高優先）──
    v.SetEnvPrefix("APP")      // 前綴：APP_DATABASE_DSN → database.dsn
    v.AutomaticEnv()           // 自動讀取環境變數
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // ── 4. 解析到 struct ──
    var config Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("解析設定失敗: %w", err)
    }

    return &config, nil
}
```

### viper 設定檔搜尋順序

viper 會按照 `AddConfigPath` 的順序搜尋設定檔：

```go
v.AddConfigPath(".")          // 第 1 順位：當前目錄
v.AddConfigPath("./config")   // 第 2 順位：config/ 子目錄
v.AddConfigPath("$HOME/.app") // 第 3 順位：Home 目錄
```

找到第一個就停止。如果全部都找不到，`ReadInConfig()` 會回傳 `ConfigFileNotFoundError`。

### 環境變數前綴（EnvPrefix）

```go
v.SetEnvPrefix("APP")

// 設定這個前綴後，viper 只讀以 APP_ 開頭的環境變數：
// APP_APP_PORT  → app.port
// APP_DATABASE_DSN → database.dsn
// APP_LOG_LEVEL → log.level

// 為什麼要前綴？
// 避免和系統環境變數衝突（例如 PATH、HOME、USER）
```

### Key Replacer（鍵名轉換）

```go
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

// viper 內部用 "." 分隔巢狀設定：database.dsn
// 環境變數用 "_" 分隔：DATABASE_DSN
// Replacer 負責把 "." 轉成 "_"，讓兩邊對得上
```

## Config Struct 設計

### 為什麼用 struct？

```
直接用 viper.GetString("database.dsn")：
  ❌ 打錯字 → 編譯不報錯，runtime 才炸
  ❌ 沒有 IDE 自動補全
  ❌ 到處都要知道 key 名稱

用 struct：
  ✅ 類型安全（Port 是 int，不是 string）
  ✅ IDE 自動補全 config.Database.DSN
  ✅ 編譯時就能發現拼錯的欄位名稱
  ✅ 方便用依賴注入傳給各模組
```

### 巢狀結構設計

```go
type Config struct {
    App      AppConfig      // 應用程式基本設定
    Database DatabaseConfig // 資料庫設定
    JWT      JWTConfig      // JWT 認證設定
    Log      LogConfig      // 日誌設定
    Redis    RedisConfig    // Redis 快取設定
}

type AppConfig struct {
    Name    string        // 應用程式名稱
    Port    int           // 監聽埠號
    Env     string        // 環境：development / staging / production
    Timeout time.Duration // 請求超時時間
}

type DatabaseConfig struct {
    Driver   string // 驅動：sqlite / postgres / mysql
    DSN      string // 連線字串（Data Source Name）
    MaxConns int    // 最大連線數
    Debug    bool   // 是否顯示 SQL 日誌
}

type JWTConfig struct {
    Secret     string        // 簽名密鑰（必須從環境變數來！）
    Expiration time.Duration // Token 有效期
}

type LogConfig struct {
    Level  string // 最低日誌層級：debug / info / warn / error
    Format string // 格式：json / console
}

type RedisConfig struct {
    URL string // 連線 URL：redis://localhost:6379
    DB  int    // 使用的資料庫編號（0-15）
}
```

### mapstructure 標籤

viper 用 `mapstructure` 標籤（不是 `json`）來對應設定 key 和 struct 欄位：

```go
type DatabaseConfig struct {
    Driver   string `mapstructure:"driver"`
    DSN      string `mapstructure:"dsn"`
    MaxConns int    `mapstructure:"max_conns"` // YAML: database.max_conns
    Debug    bool   `mapstructure:"debug"`
}
```

如果欄位名和 key 名一樣（不區分大小寫），可以省略標籤。但如果 key 包含底線（`max_conns`），而欄位名是 CamelCase（`MaxConns`），就需要明確標注。

## config.yaml 範例

```yaml
app:
  name: blog-api
  port: 8080
  env: development
  timeout: 30s

database:
  driver: sqlite
  dsn: ./blog.db
  max_conns: 10
  debug: true

jwt:
  secret: dev-secret-change-me
  expiration: 24h

log:
  level: debug
  format: console

redis:
  url: redis://localhost:6379
  db: 0
```

**注意**：`config.yaml` 可以 commit 進 git（只放預設值和非敏感設定），但密碼和密鑰不要寫在裡面。

## .env 檔案

### 格式

`.env` 檔案是最簡單的 key=value 格式：

```bash
# .env（本地開發用，不要 commit 進 git！）
APP_PORT=8080
APP_ENV=development
DB_DSN=./blog.db
JWT_SECRET=my-local-dev-secret
LOG_LEVEL=debug
REDIS_URL=redis://localhost:6379
```

### .env.example 慣例

提供一個 `.env.example` 給團隊成員參考格式（這個可以 commit）：

```bash
# .env.example（複製成 .env 後填入實際值）
APP_PORT=8080
APP_ENV=development
DB_DSN=
JWT_SECRET=
LOG_LEVEL=info
REDIS_URL=redis://localhost:6379
```

新進成員只需要：
```bash
cp .env.example .env
# 填入自己的設定值
```

### godotenv 套件

Go 標準庫不會自動讀取 `.env` 檔案，需要用 `godotenv` 套件：

```go
import "github.com/joho/godotenv"

func init() {
    // 載入 .env 檔案（找不到不報錯）
    _ = godotenv.Load()
}
```

viper 也能讀 `.env`，只要設定 `v.SetConfigType("env")`。

### .gitignore 設定

```gitignore
# .gitignore
.env
*.env
!.env.example
```

## 環境策略：開發 / 測試 / 正式

| 設定項 | 開發（development） | 測試（staging） | 正式（production） |
|--------|-------------------|----------------|-------------------|
| **資料庫** | sqlite（本地檔案） | postgres（測試庫） | postgres（正式庫 + 連線池） |
| **日誌層級** | debug（看到所有細節） | info（一般資訊） | warn（只看警告和錯誤） |
| **日誌格式** | console（人類好讀） | json（機器好解析） | json（ELK 收集） |
| **Port** | 8080 | 8080 | 80 / 443 |
| **JWT 密鑰** | 隨便一個 dev secret | 測試用密鑰 | 從 Vault/K8s Secret 來 |
| **Debug 模式** | 開（顯示 SQL） | 關 | 關 |

```bash
# 開發環境（.env 檔案）
APP_ENV=development
DB_DSN=./blog.db
LOG_LEVEL=debug

# 正式環境（K8s ConfigMap + Secret）
APP_ENV=production
DB_DSN=postgres://user:password@db-host:5432/blog?sslmode=require
JWT_SECRET=very-long-random-secret-from-vault
LOG_LEVEL=warn
```

## Config 驗證（Fail Fast）

啟動時就檢查必要的設定是否存在，不要等到請求進來才爆炸：

```go
// ValidateConfig 驗證必要的設定是否存在
func ValidateConfig(cfg *Config) error {
    if cfg.App.Port <= 0 || cfg.App.Port > 65535 {
        return fmt.Errorf("無效的 port: %d", cfg.App.Port)
    }

    if cfg.Database.DSN == "" {
        return fmt.Errorf("DB_DSN 是必填的")
    }

    // 正式環境必須有 JWT 密鑰
    if cfg.App.Env == "production" {
        if cfg.JWT.Secret == "" || cfg.JWT.Secret == "change-me-in-production" {
            return fmt.Errorf("正式環境必須設定 JWT_SECRET")
        }
    }

    validEnvs := map[string]bool{
        "development": true, "staging": true, "production": true, "test": true,
    }
    if !validEnvs[cfg.App.Env] {
        return fmt.Errorf("無效的環境: %s（允許：development/staging/production/test）", cfg.App.Env)
    }

    return nil
}

// 在 main() 中使用
func main() {
    config, err := LoadConfig()
    if err != nil {
        log.Fatalf("載入設定失敗: %v", err)
    }

    if err := ValidateConfig(config); err != nil {
        log.Fatalf("設定驗證失敗: %v", err)  // 啟動就失敗，不要帶傷上陣
    }
}
```

**Fail Fast 原則**：寧可啟動時就失敗並告訴你哪裡設錯，也不要執行到一半才爆炸（那時候已經有用戶在用了）。

## 熱重載（Hot Reload）

viper 支援監控設定檔變更，搭配 `fsnotify` 套件自動重載：

```go
import "github.com/fsnotify/fsnotify"

v.WatchConfig()
v.OnConfigChange(func(e fsnotify.Event) {
    fmt.Printf("設定檔變更: %s\n", e.Name)

    // 重新解析到 struct
    var newConfig Config
    if err := v.Unmarshal(&newConfig); err != nil {
        log.Printf("重新解析設定失敗: %v", err)
        return
    }

    // 更新設定（注意 thread safety！）
    configMutex.Lock()
    currentConfig = &newConfig
    configMutex.Unlock()
})
```

**注意事項**：
- 不是所有設定都能熱重載（例如 Port 改了要重啟 HTTP Server）
- 適合熱重載的：日誌層級、功能開關（Feature Flag）
- 不適合熱重載的：Port、資料庫連線、JWT 密鑰
- 一定要注意 thread safety（用 Mutex 或 atomic）

## 安全最佳實踐

```
✅ 正確做法：

  1. 敏感資訊（密碼、密鑰）→ 永遠用環境變數
     JWT_SECRET、DB_PASSWORD、API_KEY 都不能寫在 config.yaml

  2. 提供 .env.example → 記錄格式，不包含實際值

  3. .env 加入 .gitignore → 確保不被 commit

  4. 正式環境用 Secret 管理工具：
     - K8s Secret（基本款）
     - HashiCorp Vault（進階款）
     - AWS Secrets Manager / GCP Secret Manager（雲端）

  5. 啟動時驗證 → 缺少必要設定就直接 Fatal

❌ 危險做法：

  1. 密碼寫在 config.yaml 然後 commit 進 git
  2. 在 README 或程式碼註解中放真實密碼
  3. 正式環境用預設的 "change-me-in-production"
  4. 把 .env 檔案推上 GitHub（即使後來刪了，git 歷史還在！）
```

## 在部落格專案中的對應

部落格專案的 `pkg/config/config.go` 就是用這一課教的模式：

```go
// pkg/config/config.go
package config

type Config struct {
    Server   ServerConfig    // 對應教學的 AppConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Log      LogConfig
    Redis    RedisConfig
}

func Load() *Config {
    v := viper.New()

    // 1. 設定檔搜尋
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")
    v.AddConfigPath("./config")

    // 2. 預設值
    v.SetDefault("server.port", "8080")
    v.SetDefault("jwt.expiration", "24h")

    // 3. 環境變數覆蓋
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // 4. 讀取設定檔
    v.ReadInConfig()

    // 5. 解析到 struct
    return &Config{
        Server: ServerConfig{
            Port: v.GetString("server.port"),
            Mode: v.GetString("server.mode"),
        },
        // ...
    }
}
```

在 `main.go` 中的使用方式：

```go
func main() {
    cfg := config.Load()

    // 依賴注入：把 Config 傳給需要的元件
    db := database.Init(cfg.Database)       // 資料庫用 DSN
    logger := logging.Init(cfg.Log)         // 日誌用 Level + Format
    server := http.Server{Addr: ":" + cfg.Server.Port} // 伺服器用 Port
}
```

## 程式碼速查

```go
// ── os.Getenv（標準庫）──
value := os.Getenv("APP_PORT")              // 讀環境變數（不存在回傳 ""）
os.Setenv("APP_PORT", "9090")               // 設定環境變數

// ── viper 基本操作 ──
v := viper.New()                             // 建立實例（推薦，比全域好測試）
v.SetDefault("app.port", 8080)               // 設定預設值
v.SetConfigName("config")                    // 設定檔名（不含副檔名）
v.SetConfigType("yaml")                      // 設定檔格式
v.AddConfigPath(".")                         // 搜尋路徑
v.ReadInConfig()                             // 讀取設定檔

// ── viper 環境變數 ──
v.SetEnvPrefix("APP")                        // 環境變數前綴
v.AutomaticEnv()                             // 自動讀取環境變數
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // key 轉換

// ── viper 取值 ──
v.GetString("app.name")                      // 取字串
v.GetInt("app.port")                         // 取整數
v.GetBool("database.debug")                  // 取布林值
v.GetDuration("app.timeout")                 // 取時間長度

// ── viper 解析到 struct ──
var config Config
v.Unmarshal(&config)                         // 把所有設定解析到 struct

// ── viper 熱重載 ──
v.WatchConfig()                              // 監控設定檔變更
v.OnConfigChange(func(e fsnotify.Event) {})  // 變更時的回呼
```

## 常見問題 FAQ

### Q: viper 的環境變數前綴是怎麼運作的？

設定 `v.SetEnvPrefix("APP")` 後，viper 會自動在 key 前面加上 `APP_` 來找環境變數：

```
config key     →  環境變數
app.port       →  APP_APP_PORT
database.dsn   →  APP_DATABASE_DSN
log.level      →  APP_LOG_LEVEL
```

注意 `app.port` 會變成 `APP_APP_PORT`（前綴 APP + key 的 APP + PORT），這有時候會讓人困惑。你可以選擇不用前綴，或者調整 key 的命名。

### Q: viper.New() 和直接用 viper.SetDefault() 有什麼差？

```go
// 全域 viper（所有地方共用一個實例）
viper.SetDefault("app.port", 8080)  // 用 viper 套件的全域函式

// 建立新實例（推薦：每個模組獨立，方便測試）
v := viper.New()
v.SetDefault("app.port", 8080)
```

建議用 `viper.New()` 建立獨立實例，避免不同模組互相干擾，也更容易寫測試。

### Q: config.yaml 找不到怎麼辦？

這是正常的！viper 找不到設定檔時會回傳 `ConfigFileNotFoundError`，你可以忽略它，改用環境變數和預設值：

```go
if err := v.ReadInConfig(); err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
        // 正常：用環境變數和預設值就好
    } else {
        // 不正常：設定檔格式錯誤等問題
        return nil, err
    }
}
```

### Q: 什麼時候該用 config.yaml，什麼時候該用環境變數？

```
config.yaml：非敏感的預設值（port、log level、app name）
環境變數：    敏感資訊（密碼、密鑰）+ 環境差異的值（DSN、環境名稱）
```

簡單原則：**如果這個值被別人看到會出問題，就用環境變數**。

### Q: 正式環境的 Secret 怎麼管理？

不要用 `.env` 檔案，用專業的 Secret 管理工具：

- **K8s Secret**：最基本，適合已經用 K8s 的團隊
- **HashiCorp Vault**：功能最完整，支援動態 Secret、自動輪替
- **AWS Secrets Manager / GCP Secret Manager**：雲端原生方案

這些工具會把 Secret 注入成環境變數，你的程式碼不需要改。

## 練習

1. **建立 config.yaml**：在 `tutorials/20-config/` 目錄下建立一個 `config.yaml`，設定 `app.port` 為 `3000`，執行程式看看是否讀取到
2. **環境變數覆蓋**：用環境變數覆蓋 config.yaml 的值，例如 `APP_APP_PORT=9999 go run ./tutorials/20-config`，確認環境變數的優先順序高於 config.yaml
3. **新增設定項**：在 Config struct 加入一個 `RateLimit` 設定（包含 `MaxRequests int` 和 `Window time.Duration`），設定預設值並測試
4. **Config 驗證**：實作一個 `ValidateConfig` 函式，檢查 Port 範圍（1-65535）、DSN 非空、正式環境必須有 JWT Secret，啟動時驗證
5. **.env 整合**：安裝 `godotenv` 套件，建立 `.env` 檔案設定 `DB_DSN=postgres://localhost/test`，驗證程式能正確讀取

## 下一課預告

**第二十一課：結構化日誌（Structured Logging）** —— 學習如何把 `fmt.Println` 升級成帶標籤的 JSON 日誌，讓你在幾百萬筆日誌中一秒找到問題。用 `slog`（Go 1.21 內建）和 `zap` 實作生產級日誌系統。
