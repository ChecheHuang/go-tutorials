// ==========================================================================
// 第三十一課：Wire 依賴注入框架
// ==========================================================================
//
// 第二十四課學了手動依賴注入（DI）：
//   db → repo → usecase → handler → router
//   在 main() 裡一行一行手動 new
//
// 問題：專案越大，main() 越複雜
//   當有 20 個服務、每個服務有 3-4 個依賴時，main() 就很難維護
//
// Wire 是 Google 開發的「編譯時依賴注入」工具：
//   你只需要告訴 Wire「這個元件需要什麼」
//   Wire 自動生成組裝的程式碼（在編譯時，不是執行時）
//
// Wire vs 手動 DI vs 執行時 DI 框架：
//   手動 DI    → 你自己寫組裝邏輯（小專案，清楚易懂）
//   Wire       → 你定義依賴關係，Wire 生成組裝程式碼（大專案，編譯時檢查）
//   反射 DI    → 執行時用反射（如 Java Spring），慢且難除錯
//
//   Go 的哲學：選擇 Wire（編譯時 → 型別安全、快速）
//
// 本課的做法：
//   Wire 需要執行 `wire gen` 命令生成 wire_gen.go
//   為了讓課程可以直接 `go run`，我們：
//   1. 先寫 wire.go（給 Wire 工具用的），讓你看到 Wire 的語法
//   2. 再寫 wire_gen.go（模擬 Wire 生成的程式碼），讓你看到輸出結果
//   3. main.go 示範如何使用生成的程式碼
//
// 執行方式：go run ./tutorials/28-wire
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"fmt"  // 格式化輸出
	"log"  // 日誌輸出
	"time" // 時間

	"github.com/glebarez/sqlite" // SQLite 驅動
	"go.uber.org/zap"            // 結構化日誌
	"gorm.io/gorm"               // ORM
	"gorm.io/gorm/logger"        // GORM 日誌
)

// ==========================================================================
// 1. 定義應用程式元件
// ==========================================================================
//
// 在 Wire 的世界裡，每個元件都有一個「Provider」：
//   Provider = 建立這個元件的函式
//   Wire 分析所有 Provider，找出依賴關係，生成初始化程式碼

// Config 應用程式設定
type Config struct {
	Port    int    // 監聽埠號
	DBPath  string // 資料庫路徑
	LogMode string // 日誌模式：development 或 production
}

// ProvideConfig 建立 Config（Wire Provider）
// Wire 看到這個函式，就知道「要建立 Config，只需要呼叫 ProvideConfig()」
func ProvideConfig() *Config {
	return &Config{
		Port:    8080,
		DBPath:  "file::memory:?cache=shared",
		LogMode: "development",
	}
}

// Database 資料庫連線包裝
type Database struct {
	DB *gorm.DB // GORM 連線
}

// PostModel 用於示範的 Post 模型
type PostModel struct {
	ID      uint   `gorm:"primaryKey"`
	Title   string `gorm:"not null"`
	Content string
}

// ProvideDatabase 建立資料庫連線（Wire Provider）
// 注意：接收 *Config 作為參數 → Wire 知道需要先建立 Config
func ProvideDatabase(cfg *Config) (*Database, error) {
	db, err := gorm.Open(
		sqlite.Open(cfg.DBPath),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		return nil, fmt.Errorf("ProvideDatabase: %w", err)
	}
	if err := db.AutoMigrate(&PostModel{}); err != nil {
		return nil, fmt.Errorf("AutoMigrate: %w", err)
	}
	return &Database{DB: db}, nil
}

// Logger 日誌包裝
type Logger struct {
	Zap *zap.Logger
}

// ProvideLogger 建立日誌（Wire Provider）
func ProvideLogger(cfg *Config) (*Logger, error) {
	var zapLog *zap.Logger
	var err error
	if cfg.LogMode == "production" {
		zapLog, err = zap.NewProduction()
	} else {
		zapLog, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, fmt.Errorf("ProvideLogger: %w", err)
	}
	return &Logger{Zap: zapLog}, nil
}

// PostRepository 文章資料存取
type PostRepository struct {
	db     *Database // 注入資料庫
	logger *Logger   // 注入日誌
}

// ProvidePostRepository 建立 PostRepository（Wire Provider）
// 接收 *Database 和 *Logger → Wire 知道需要先建立這兩個
func ProvidePostRepository(db *Database, log *Logger) *PostRepository {
	return &PostRepository{db: db, logger: log}
}

func (r *PostRepository) Create(title, content string) (*PostModel, error) {
	post := PostModel{Title: title, Content: content}
	if err := r.db.DB.Create(&post).Error; err != nil {
		return nil, fmt.Errorf("PostRepository.Create: %w", err)
	}
	r.logger.Zap.Info("文章建立成功", zap.Uint("id", post.ID))
	return &post, nil
}

func (r *PostRepository) FindAll() ([]*PostModel, error) {
	var posts []*PostModel
	if err := r.db.DB.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// PostUsecase 文章業務邏輯
type PostUsecase struct {
	repo   *PostRepository // 注入 Repository
	logger *Logger         // 注入日誌
}

// ProvidePostUsecase 建立 PostUsecase（Wire Provider）
func ProvidePostUsecase(repo *PostRepository, log *Logger) *PostUsecase {
	return &PostUsecase{repo: repo, logger: log}
}

func (u *PostUsecase) CreatePost(title, content string) (*PostModel, error) {
	if title == "" {
		return nil, fmt.Errorf("標題不能為空")
	}
	return u.repo.Create(title, content)
}

func (u *PostUsecase) ListPosts() ([]*PostModel, error) {
	return u.repo.FindAll()
}

// PostHandler HTTP 處理器（這裡簡化，不實際啟動 HTTP）
type PostHandler struct {
	usecase *PostUsecase // 注入 Usecase
	logger  *Logger      // 注入日誌
}

// ProvidePostHandler 建立 PostHandler（Wire Provider）
func ProvidePostHandler(uc *PostUsecase, log *Logger) *PostHandler {
	return &PostHandler{usecase: uc, logger: log}
}

func (h *PostHandler) HandleCreate(title, content string) {
	post, err := h.usecase.CreatePost(title, content)
	if err != nil {
		h.logger.Zap.Error("建立文章失敗", zap.Error(err))
		return
	}
	h.logger.Zap.Info("Handler：文章建立", zap.Uint("id", post.ID), zap.String("title", post.Title))
}

func (h *PostHandler) HandleList() []*PostModel {
	posts, err := h.usecase.ListPosts()
	if err != nil {
		h.logger.Zap.Error("列出文章失敗", zap.Error(err))
		return nil
	}
	return posts
}

// App 整個應用程式（所有元件的集合）
type App struct {
	Config  *Config      // 設定
	DB      *Database    // 資料庫
	Logger  *Logger      // 日誌
	Handler *PostHandler // HTTP 處理器
}

// ProvideApp 建立 App（Wire Provider，最頂層）
// 接收所有元件，Wire 負責先把所有元件都 new 好
func ProvideApp(cfg *Config, db *Database, log *Logger, handler *PostHandler) *App {
	return &App{
		Config:  cfg,
		DB:      db,
		Logger:  log,
		Handler: handler,
	}
}

// ==========================================================================
// 2. wire.go — 告訴 Wire 如何組裝（真正的 Wire 語法）
// ==========================================================================
//
// 在真實專案中，你會建立一個 wire.go 檔案（這裡用程式碼說明）：
//
//   //go:build wireinject
//   // +build wireinject
//
//   package main
//
//   import "github.com/google/wire"
//
//   // InitializeApp 是「Wire 的入口點」
//   // Wire 看到這個函式，分析所有 Provider，生成 wire_gen.go
//   func InitializeApp() (*App, error) {
//       wire.Build(
//           ProvideConfig,
//           ProvideDatabase,
//           ProvideLogger,
//           ProvidePostRepository,
//           ProvidePostUsecase,
//           ProvidePostHandler,
//           ProvideApp,
//       )
//       return nil, nil  // Wire 會替換掉這個，不需要實際實作
//   }
//
// 然後執行：wire gen ./tutorials/28-wire/
// Wire 會生成 wire_gen.go，包含正確的初始化程式碼

// ==========================================================================
// 3. wire_gen.go — Wire 生成的程式碼（手動模擬）
// ==========================================================================
//
// 執行 wire gen 後，Wire 會生成類似這樣的程式碼：

// InitializeApp Wire 會生成這個函式的實作（這裡是手動模擬）
func InitializeApp() (*App, error) {
	// Wire 分析依賴關係，自動生成這個順序的初始化程式碼：

	// 1. 沒有依賴的先初始化
	config := ProvideConfig() // Config 不依賴任何東西

	// 2. 依賴 Config 的
	db, err := ProvideDatabase(config) // Database 依賴 Config
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Database: %w", err)
	}
	log, err := ProvideLogger(config) // Logger 依賴 Config
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Logger: %w", err)
	}

	// 3. 依賴 Database 和 Logger 的
	repo := ProvidePostRepository(db, log) // Repository 依賴 Database + Logger

	// 4. 依賴 Repository 的
	uc := ProvidePostUsecase(repo, log) // Usecase 依賴 Repository + Logger

	// 5. 依賴 Usecase 的
	handler := ProvidePostHandler(uc, log) // Handler 依賴 Usecase + Logger

	// 6. 最頂層
	app := ProvideApp(config, db, log, handler) // App 依賴所有東西

	return app, nil
}

// ==========================================================================
// 依賴圖（Wire 自動分析並按照這個順序初始化）
// ==========================================================================
//
//   Config ─────────────────────────────────▶ App
//      │                                        ↑
//      ├──▶ Database ──▶ PostRepository ──▶ PostUsecase ──▶ PostHandler
//      │                      ↑                 ↑                ↑
//      └──▶ Logger ──────────┴─────────────────┴────────────────┘
//
//   Wire 看到這張圖，生成正確的初始化順序：
//   Config → Database → Logger → PostRepository → PostUsecase → PostHandler → App

// ==========================================================================
// 4. Provider Sets（Wire 的進階功能：把相關的 Provider 打包）
// ==========================================================================
//
// 當 Provider 很多時，可以用 wire.NewSet 打包：
//
//   var InfraSet = wire.NewSet(
//       ProvideDatabase,
//       ProvideLogger,
//   )
//
//   var DomainSet = wire.NewSet(
//       ProvidePostRepository,
//       ProvidePostUsecase,
//   )
//
//   var HandlerSet = wire.NewSet(
//       ProvidePostHandler,
//   )
//
//   func InitializeApp() (*App, error) {
//       wire.Build(
//           ProvideConfig,
//           InfraSet,   // 替代 ProvideDatabase + ProvideLogger
//           DomainSet,  // 替代 ProvidePostRepository + ProvidePostUsecase
//           HandlerSet, // 替代 ProvidePostHandler
//           ProvideApp,
//       )
//       return nil, nil
//   }

// ==========================================================================
// 主程式
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================")
	fmt.Println(" 第二十八課：Wire 依賴注入框架")
	fmt.Println("==========================================")

	// ──── 用 Wire 生成的 InitializeApp 一行建立整個應用程式 ────
	start := time.Now()

	app, err := InitializeApp() // 一行搞定所有初始化！
	if err != nil {
		log.Fatalf("初始化失敗: %v", err)
	}
	defer app.Logger.Zap.Sync()

	app.Logger.Zap.Info("應用程式初始化完成",
		zap.Duration("init_time", time.Since(start)),
		zap.Int("port", app.Config.Port),
	)

	// ──── 使用應用程式 ────
	fmt.Println()
	fmt.Println("=== 示範：透過 Wire 注入的完整依賴鏈 ===")
	fmt.Println()

	// Handler → Usecase → Repository → DB，全部由 Wire 注入
	app.Handler.HandleCreate("Wire 教學文章", "這篇文章由 Wire 注入的 Handler 建立")
	app.Handler.HandleCreate("Go 依賴注入", "DI 讓程式碼更容易測試和維護")
	app.Handler.HandleCreate("", "這篇會失敗（空標題）") // 業務邏輯驗證

	posts := app.Handler.HandleList()
	fmt.Printf("\n資料庫中的文章（共 %d 篇）：\n", len(posts))
	for _, p := range posts {
		fmt.Printf("  [%d] %s\n", p.ID, p.Title)
	}

	// ──── 對比：手動 DI vs Wire ────
	fmt.Println()
	fmt.Println("=== 手動 DI vs Wire 的比較 ===")
	fmt.Println()
	fmt.Println("【手動 DI（第二十四課的做法）】")
	fmt.Println("  config := &Config{...}")
	fmt.Println("  db, err := ProvideDatabase(config)")
	fmt.Println("  log, err := ProvideLogger(config)")
	fmt.Println("  repo := ProvidePostRepository(db, log)")
	fmt.Println("  uc := ProvidePostUsecase(repo, log)")
	fmt.Println("  handler := ProvidePostHandler(uc, log)")
	fmt.Println("  app := ProvideApp(config, db, log, handler)")
	fmt.Println("  → 小專案：清楚易懂，推薦這種方式")
	fmt.Println("  → 大專案：20 個服務 → main() 幾百行")
	fmt.Println()
	fmt.Println("【Wire（本課的做法）】")
	fmt.Println("  // wire.go（你寫的）")
	fmt.Println("  func InitializeApp() (*App, error) {")
	fmt.Println("      wire.Build(ProvideConfig, ProvideDatabase, ...)")
	fmt.Println("      return nil, nil")
	fmt.Println("  }")
	fmt.Println("  ")
	fmt.Println("  // wire gen → 自動生成初始化程式碼")
	fmt.Println("  app, err := InitializeApp()  ← 一行搞定！")
	fmt.Println("  → 大專案：依賴自動分析，編譯時報錯，不用手動排順序")
	fmt.Println()
	fmt.Println("【使用 Wire 的時機】")
	fmt.Println("  ✅ 超過 10 個 Provider（元件）")
	fmt.Println("  ✅ 需要多個環境的不同實作（dev/prod 注入不同的 Repository）")
	fmt.Println("  ✅ 需要 ProviderSet 把相關元件打包")
	fmt.Println("  ❌ 小專案（手動 DI 更簡單）")

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
