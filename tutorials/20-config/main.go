// ==========================================================================
// 第二十課：Config 管理（Configuration Management）
// ==========================================================================
//
// 什麼是 Config 管理？
//   程式在不同環境（開發/測試/正式）需要不同的設定值：
//     開發：db = sqlite, log_level = debug, port = 8080
//     正式：db = postgres://prod-host..., log_level = warn, port = 80
//
//   把這些設定值「硬寫」在程式碼裡是不好的：
//     ❌ db, _ := gorm.Open(sqlite.Open("blog.db"), ...)
//     問題：換環境就要改程式碼、密碼會被 commit 進 git！
//
//   正確做法：設定從「外部」來，程式只讀取它
//
// 設定的來源（優先順序從高到低）：
//   1. 環境變數（最高優先，K8s/Docker 用這個）
//   2. .env 檔案（本地開發用）
//   3. config.yaml 檔案（預設值）
//   4. 程式碼中的預設值（最低優先）
//
// 為什麼用 viper？
//   viper 是 Go 最流行的設定管理套件，支援：
//   - 讀取 .env、YAML、JSON、TOML 等格式
//   - 自動讀取環境變數
//   - 支援預設值
//   - 熱重載（設定檔變更時自動更新）
//
// 執行方式：go run ./tutorials/20-config
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"fmt"     // 格式化輸出
	"os"      // 環境變數操作
	"strings" // strings.NewReplacer（給 viper.SetEnvKeyReplacer 用）
	"time"    // 時間類型（用在 Config struct）

	"github.com/spf13/viper" // Config 管理套件
)

// ==========================================================================
// 1. Config Struct（把所有設定集中在一個地方）
// ==========================================================================
//
// 把設定用 struct 管理有幾個好處：
//   - 類型安全（Port 是 int，不是 string）
//   - IDE 自動補全
//   - 編譯時就能發現拼錯的欄位名稱
//   - 方便傳遞給需要設定的元件（依賴注入）

// Config 應用程式的所有設定
type Config struct {
	App      AppConfig      // 應用程式基本設定
	Database DatabaseConfig // 資料庫設定
	JWT      JWTConfig      // JWT 設定
	Log      LogConfig      // 日誌設定
	Redis    RedisConfig    // Redis 設定
}

// AppConfig 應用程式基本設定
type AppConfig struct {
	Name    string        // 應用程式名稱
	Port    int           // 監聽埠號
	Env     string        // 環境：development、production、test
	Timeout time.Duration // 請求超時時間
}

// DatabaseConfig 資料庫設定
type DatabaseConfig struct {
	Driver   string // 驅動：sqlite、postgres、mysql
	DSN      string // 連線字串（Data Source Name）
	MaxConns int    // 最大連線數
	Debug    bool   // 是否顯示 SQL 日誌
}

// JWTConfig JWT 設定
type JWTConfig struct {
	Secret     string        // 簽名密鑰
	Expiration time.Duration // Token 有效期
}

// LogConfig 日誌設定
type LogConfig struct {
	Level  string // 最低日誌層級：debug, info, warn, error
	Format string // 格式：json, console
}

// RedisConfig Redis 設定
type RedisConfig struct {
	URL string // 連線 URL：redis://localhost:6379
	DB  int    // 使用的資料庫編號（0-15）
}

// ==========================================================================
// 2. LoadConfig — 讀取設定（支援多種來源）
// ==========================================================================

// LoadConfig 載入所有設定（viper 版本）
func LoadConfig() (*Config, error) {
	v := viper.New() // 建立新的 viper 實例（比用全域 viper 更好測試）

	// ──── 設定預設值（最低優先）────
	// 這些值是「萬一什麼都沒設定時」的 fallback
	v.SetDefault("app.name", "go-tutorials")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.env", "development")
	v.SetDefault("app.timeout", "30s")

	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "./blog.db")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("database.debug", false)

	v.SetDefault("jwt.secret", "change-me-in-production") // 記住：正式環境必須換！
	v.SetDefault("jwt.expiration", "24h")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	v.SetDefault("redis.url", "redis://localhost:6379")
	v.SetDefault("redis.db", 0)

	// ──── 讀取 config.yaml（如果存在）────
	v.SetConfigName("config")     // 設定檔名稱（不含副檔名）
	v.SetConfigType("yaml")       // 設定檔格式
	v.AddConfigPath(".")          // 從當前目錄找
	v.AddConfigPath("./config")   // 從 config/ 子目錄找
	v.AddConfigPath("$HOME/.app") // 從 Home 目錄找

	if err := v.ReadInConfig(); err != nil { // 嘗試讀取設定檔
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok { // 如果不是「找不到檔案」的錯誤
			return nil, fmt.Errorf("讀取設定檔失敗: %w", err) // 真正的錯誤（格式錯誤等）
		}
		// 找不到設定檔是正常的（用環境變數或預設值）
		fmt.Println("ℹ️  找不到 config.yaml，使用環境變數和預設值")
	} else {
		fmt.Printf("✅ 已載入設定檔: %s\n", v.ConfigFileUsed())
	}

	// ──── 自動讀取環境變數（最高優先）────
	// 環境變數名稱規則：前綴_欄位_子欄位（全大寫、用底線分隔）
	// 例如：APP_PORT=9090 對應 app.port
	//       DATABASE_DSN=postgres://... 對應 database.dsn
	v.SetEnvPrefix("APP") // 環境變數前綴（避免和系統變數衝突）
	v.AutomaticEnv()      // 自動讀取所有符合前綴的環境變數

	// 把巢狀設定的點（.）替換成底線（_），方便環境變數讀取
	// APP_DATABASE_DSN → database.dsn
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ──── 解析到 Config struct ────
	var config Config
	if err := v.Unmarshal(&config); err != nil { // 把 viper 的值解析到 struct
		return nil, fmt.Errorf("解析設定失敗: %w", err)
	}

	return &config, nil
}

// ==========================================================================
// 3. 不用 viper 的輕量版本（只用 os.Getenv）
// ==========================================================================
//
// 如果不想引入 viper 這個相依套件，可以只用標準庫：
// 適合小型專案或設定值不多的情況

// getEnvOrDefault 讀取環境變數，如果不存在則回傳預設值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" { // 如果環境變數有設定
		return value // 回傳環境變數的值
	}
	return defaultValue // 回傳預設值
}

// LoadConfigSimple 不用 viper 的簡單版（只讀環境變數）
func LoadConfigSimple() *Config {
	return &Config{
		App: AppConfig{
			Name: getEnvOrDefault("APP_NAME", "go-tutorials"), // 讀取 APP_NAME
			Port: func() int { // 讀取 APP_PORT（需要轉型）
				port := 8080
				if p := os.Getenv("APP_PORT"); p != "" {
					fmt.Sscanf(p, "%d", &port) // 字串轉整數
				}
				return port
			}(),
			Env: getEnvOrDefault("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Driver: getEnvOrDefault("DB_DRIVER", "sqlite"),
			DSN:    getEnvOrDefault("DB_DSN", "./blog.db"),
		},
		JWT: JWTConfig{
			Secret: getEnvOrDefault("JWT_SECRET", "change-me-in-production"),
		},
		Log: LogConfig{
			Level:  getEnvOrDefault("LOG_LEVEL", "info"),
			Format: getEnvOrDefault("LOG_FORMAT", "json"),
		},
		Redis: RedisConfig{
			URL: getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
		},
	}
}

// ==========================================================================
// 示範函式
// ==========================================================================

func demonstrateSimpleConfig() { // 示範簡單版 Config
	fmt.Println("=== 1. 輕量版 Config（只用 os.Getenv）===\n")

	// 模擬設定環境變數（正式環境是在 Docker/K8s 設定的）
	os.Setenv("APP_PORT", "9090")                       // 覆蓋埠號
	os.Setenv("DB_DSN", "postgres://user:pass@db/prod") // 正式環境 DB
	os.Setenv("LOG_LEVEL", "warn")                      // 正式環境只看 WARN 以上

	config := LoadConfigSimple() // 載入設定

	fmt.Printf("應用程式名稱: %s\n", config.App.Name)
	fmt.Printf("監聽埠號:     %d\n", config.App.Port) // 應該是 9090
	fmt.Printf("環境:         %s\n", config.App.Env)
	fmt.Printf("資料庫:       %s\n", config.Database.DSN) // 應該是 postgres://...
	fmt.Printf("日誌層級:     %s\n", config.Log.Level)     // 應該是 warn
	fmt.Printf("JWT 密鑰:     %s\n", config.JWT.Secret)

	// 清除示範用的環境變數
	os.Unsetenv("APP_PORT")
	os.Unsetenv("DB_DSN")
	os.Unsetenv("LOG_LEVEL")
}

func demonstrateViperConfig() { // 示範 viper 版 Config
	fmt.Println("\n=== 2. viper Config（支援 YAML + 環境變數）===\n")

	// 模擬正式環境的環境變數（通常由 Docker/K8s 設定）
	os.Setenv("APP_APP_PORT", "443")    // viper 前綴是 APP，所以 app.port → APP_APP_PORT
	os.Setenv("APP_LOG_LEVEL", "error") // app.log.level → APP_LOG_LEVEL

	v := viper.New()
	v.SetDefault("app.name", "go-tutorials")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.env", "development")
	v.SetDefault("database.dsn", "./blog.db")
	v.SetDefault("log.level", "info")
	v.SetDefault("jwt.secret", "dev-secret")

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	fmt.Printf("app.name  = %s\n", v.GetString("app.name"))
	fmt.Printf("app.port  = %d (環境變數覆蓋: %d)\n", 8080, v.GetInt("app.port"))
	fmt.Printf("log.level = %s (環境變數覆蓋: %s)\n", "info", v.GetString("log.level"))
	fmt.Printf("db.dsn    = %s\n", v.GetString("database.dsn"))

	os.Unsetenv("APP_APP_PORT")
	os.Unsetenv("APP_LOG_LEVEL")
}

func demonstrateConfigInProject() { // 示範在實際專案中的使用
	fmt.Println("\n=== 3. 在專案中如何使用 Config ===\n")

	// 讀取設定
	config := LoadConfigSimple()

	fmt.Println("// 在 main() 中：")
	fmt.Println("config, err := LoadConfig()")
	fmt.Println("if err != nil { log.Fatal(err) }")
	fmt.Println()
	fmt.Println("// 把 Config 傳給需要的元件（依賴注入）：")
	fmt.Printf("db   := initDB(config.Database)    // DSN: %s\n", config.Database.DSN)
	fmt.Printf("log  := initLogger(config.Log)     // Level: %s\n", config.Log.Level)
	fmt.Printf("srv  := &http.Server{Addr: \":%d\"} // Port: %d\n", config.App.Port, config.App.Port)
	fmt.Println()
	fmt.Println("// 不同環境的設定方式：")
	fmt.Println()
	fmt.Println("  開發環境（.env 檔案）：")
	fmt.Println("    APP_PORT=8080")
	fmt.Println("    DB_DSN=./blog.db")
	fmt.Println("    LOG_LEVEL=debug")
	fmt.Println()
	fmt.Println("  正式環境（K8s Secret / ConfigMap）：")
	fmt.Println("    APP_PORT=80")
	fmt.Println("    DB_DSN=postgres://user:password@db-host:5432/blog")
	fmt.Println("    JWT_SECRET=very-long-random-secret-from-vault")
	fmt.Println("    LOG_LEVEL=warn")
}

func demonstrateSecurityBestPractices() { // 示範安全最佳實踐
	fmt.Println("\n=== 4. Config 安全最佳實踐 ===\n")

	fmt.Println("✅ 正確做法：")
	fmt.Println()
	fmt.Println("  1. 敏感資訊（密碼、密鑰）永遠用環境變數，不要寫在 config.yaml：")
	fmt.Println(`     ✅ JWT_SECRET 環境變數`)
	fmt.Println(`     ❌ jwt.secret: "mysecret" 在 config.yaml → 會被 git 追蹤！`)
	fmt.Println()
	fmt.Println("  2. 提供 .env.example（把格式記錄下來，但不包含實際值）：")
	fmt.Println("     APP_PORT=8080")
	fmt.Println("     DB_DSN=sqlite://./blog.db")
	fmt.Println("     JWT_SECRET=")
	fmt.Println()
	fmt.Println("  3. 把 .env 加入 .gitignore（確保不被 commit）：")
	fmt.Println("     echo '.env' >> .gitignore")
	fmt.Println()
	fmt.Println("  4. 正式環境用 Secret 管理工具：")
	fmt.Println("     K8s Secret / AWS Secrets Manager / HashiCorp Vault")
	fmt.Println()
	fmt.Println("  5. 啟動時驗證必要的設定是否存在：")
	fmt.Println(`     if config.JWT.Secret == "" {`)
	fmt.Println(`         log.Fatal("JWT_SECRET 環境變數是必填的")`)
	fmt.Println("     }")
}

func main() { // 程式進入點
	fmt.Println("==========================================")
	fmt.Println(" 第二十課：Config 管理")
	fmt.Println("==========================================")

	demonstrateSimpleConfig()          // 1. 輕量版 Config
	demonstrateViperConfig()           // 2. viper 版 Config
	demonstrateConfigInProject()       // 3. 在專案中的使用方式
	demonstrateSecurityBestPractices() // 4. 安全最佳實踐

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
