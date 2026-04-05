// ==========================================================================
// 第二十四課：Clean Architecture 進階 + 依賴注入
// ==========================================================================
//
// 第十課學了 Clean Architecture 的概念（用一個簡單示範）
// 這一課把前面所有技術整合起來，做一個「接近正式等級」的 API 伺服器：
//
//   Gin（HTTP）+ GORM（資料庫）+ zap（日誌）+ JWT（認證）
//
// 這一課最重要的兩個概念：
//
// 1. 依賴注入（Dependency Injection，DI）
//    不要讓元件自己建立依賴，而是從外部「注入」進來
//    這樣元件可以獨立測試，可以輕易替換實作
//
//    ❌ 錯誤做法（元件自己建立依賴）：
//      type ArticleUsecase struct{}
//      func (u *ArticleUsecase) GetArticle(id uint) (*Article, error) {
//          db, _ := gorm.Open(...)  ← 在這裡建立 DB 連線！
//          // 問題：怎麼測試？怎麼換成 Redis？
//      }
//
//    ✅ 正確做法（依賴從外部注入）：
//      type ArticleUsecase struct {
//          repo ArticleRepository  ← 只持有介面，不管實作
//      }
//      func NewArticleUsecase(repo ArticleRepository) *ArticleUsecase {
//          return &ArticleUsecase{repo: repo}  ← 從外部傳進來
//      }
//
// 2. 介面（Interface）讓層與層之間解耦
//    Usecase 只知道「我需要一個能存取文章的東西」
//    不管那個「東西」是 SQLite、PostgreSQL 還是記憶體
//
// 整個系統的組裝發生在 main()：
//   db → GormArticleRepository → ArticleUsecase → ArticleHandler → Router
//   這條線就是「依賴注入鏈」
//
// 目錄結構（真實專案會拆成多個檔案，這裡合在一起方便學習）：
//
//   domain/     → Article struct、ArticleRepository 介面、ArticleUsecase 介面
//   repository/ → GormArticleRepository（實作 ArticleRepository）
//   usecase/    → ArticleUsecase（業務邏輯）
//   handler/    → ArticleHandler（HTTP 處理器）
//   middleware/ → JWT、Logging 中介層
//   main.go     → 組裝所有元件（DI Container）
//
// 執行方式：go run ./tutorials/24-clean-arch-advanced
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"context"       // 用於 Graceful Shutdown 的超時控制
	"errors"        // 標準錯誤處理
	"fmt"           // 格式化輸出
	"log"           // 日誌輸出（初始化前用）
	"net/http"      // HTTP 伺服器 + 狀態碼
	"os"            // 環境變數 + 系統信號
	"os/signal"     // 系統信號接收（SIGTERM、SIGINT）
	"runtime"       // 取得系統資訊（goroutine 數量）
	"strconv"       // 字串轉整數
	"syscall"       // 系統呼叫常數（SIGTERM）
	"time"          // 時間處理

	"github.com/gin-gonic/gin"           // HTTP 框架
	"github.com/glebarez/sqlite"         // 純 Go SQLite 驅動
	"github.com/golang-jwt/jwt/v5"       // JWT 套件
	"go.uber.org/zap"                    // 結構化日誌
	"gorm.io/gorm"                       // ORM 框架
	"gorm.io/gorm/logger"                // GORM 日誌設定
)

// ==========================================================================
// ██████╗  ██████╗ ███╗   ███╗ █████╗ ██╗███╗   ██╗
// ██╔══██╗██╔═══██╗████╗ ████║██╔══██╗██║████╗  ██║
// ██║  ██║██║   ██║██╔████╔██║███████║██║██╔██╗ ██║
// ██║  ██║██║   ██║██║╚██╔╝██║██╔══██║██║██║╚██╗██║
// ██████╔╝╚██████╔╝██║ ╚═╝ ██║██║  ██║██║██║ ╚████║
// ╚═════╝  ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝
//
// 領域層（Domain Layer）
// ————————————————————————
// 這是整個系統的「核心」
// 定義：業務實體（Entity）、Repository 介面、Usecase 介面
// 規則：這一層不依賴任何外部套件（不能 import gin、gorm、redis）
// ==========================================================================

// Article 文章實體（業務核心概念）
// 只包含業務相關欄位，不含任何框架標籤
type Article struct {
	ID        uint      // 唯一識別碼
	Title     string    // 標題
	Content   string    // 內容
	AuthorID  uint      // 作者 ID
	Status    string    // 狀態：draft（草稿）、published（已發布）
	CreatedAt time.Time // 建立時間
	UpdatedAt time.Time // 更新時間
}

// CreateArticleInput 建立文章的輸入資料
type CreateArticleInput struct {
	Title    string // 標題（必填）
	Content  string // 內容（必填）
	AuthorID uint   // 作者 ID（必填）
}

// UpdateArticleInput 更新文章的輸入資料（指標表示「可選」）
type UpdateArticleInput struct {
	Title   *string // 標題（可選，nil = 不更新）
	Content *string // 內容（可選，nil = 不更新）
	Status  *string // 狀態（可選，nil = 不更新）
}

// 領域錯誤定義（Usecase 回傳這些，Handler 根據這些決定 HTTP 狀態碼）
var (
	ErrArticleNotFound    = errors.New("文章不存在")       // 404
	ErrArticleForbidden   = errors.New("沒有權限操作此文章") // 403
	ErrInvalidInput       = errors.New("輸入資料不合法")    // 400
	ErrTitleRequired      = errors.New("標題不能為空")      // 400
	ErrContentRequired    = errors.New("內容不能為空")      // 400
)

// ArticleRepository 文章資料存取介面（定義在 Domain 層！）
// Repository 的「契約」：Usecase 透過這個介面存取資料，不管底層是什麼
type ArticleRepository interface {
	FindByID(id uint) (*Article, error)                        // 根據 ID 查詢
	FindByAuthor(authorID uint) ([]*Article, error)            // 查詢作者的所有文章
	FindPublished(page, pageSize int) ([]*Article, int64, error) // 查詢已發布的文章（分頁）
	Create(input CreateArticleInput) (*Article, error)         // 建立文章
	Update(id uint, input UpdateArticleInput) (*Article, error) // 更新文章
	Delete(id uint) error                                      // 刪除文章
}

// ArticleUsecase 文章業務邏輯介面
// Handler 透過這個介面呼叫業務邏輯，不管背後的實作
type ArticleUsecase interface {
	GetArticle(id uint) (*Article, error)                               // 取得單篇文章
	ListPublished(page, pageSize int) ([]*Article, int64, error)        // 取得已發布文章列表
	CreateArticle(authorID uint, input CreateArticleInput) (*Article, error) // 建立文章
	UpdateArticle(id, authorID uint, input UpdateArticleInput) (*Article, error) // 更新文章
	DeleteArticle(id, authorID uint) error                              // 刪除文章
}

// ==========================================================================
// ██████╗ ███████╗██████╗  ██████╗
// ██╔══██╗██╔════╝██╔══██╗██╔═══██╗
// ██████╔╝█████╗  ██████╔╝██║   ██║
// ██╔══██╗██╔══╝  ██╔═══╝ ██║   ██║
// ██║  ██║███████╗██║     ╚██████╔╝
// ╚═╝  ╚═╝╚══════╝╚═╝      ╚═════╝
//
// Repository 層（資料存取層）
// ————————————————————————————
// 實作 ArticleRepository 介面
// 這裡才有 gorm 的細節，Domain 層完全不知道
// ==========================================================================

// ArticleModel GORM 用的資料庫模型（跟 Domain 的 Article 分開，避免污染）
// 為什麼要分開？因為 GORM tag 是基礎設施細節，不應該出現在 Domain 層
type ArticleModel struct {
	ID        uint           `gorm:"primaryKey"`              // 主鍵
	Title     string         `gorm:"size:200;not null;index"` // 標題（加索引）
	Content   string         `gorm:"type:text"`               // 內容
	AuthorID  uint           `gorm:"not null;index"`          // 作者 ID（加索引）
	Status    string         `gorm:"size:20;default:'draft';index"` // 狀態（加索引）
	CreatedAt time.Time      `gorm:"index"`                   // 建立時間
	UpdatedAt time.Time                                       // 更新時間
	DeletedAt gorm.DeletedAt `gorm:"index"`                   // 軟刪除（GORM 內建）
}

// TableName 讓 GORM 知道這個模型對應的表名
func (ArticleModel) TableName() string { return "articles" } // 表名：articles

// toArticle 把 GORM Model 轉換成 Domain Entity
func (m *ArticleModel) toArticle() *Article { // 轉換
	return &Article{
		ID:        m.ID,
		Title:     m.Title,
		Content:   m.Content,
		AuthorID:  m.AuthorID,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// GormArticleRepository 用 GORM 實作 ArticleRepository
type GormArticleRepository struct {
	db     *gorm.DB    // GORM 資料庫連線（注入進來的）
	logger *zap.Logger // 日誌（注入進來的）
}

// NewGormArticleRepository 建立 GormArticleRepository（依賴注入用）
func NewGormArticleRepository(db *gorm.DB, logger *zap.Logger) ArticleRepository {
	return &GormArticleRepository{db: db, logger: logger} // 回傳介面，不是具體型別！
}

// FindByID 根據 ID 查詢文章
func (r *GormArticleRepository) FindByID(id uint) (*Article, error) {
	var model ArticleModel                                // 建立空的 Model
	result := r.db.First(&model, id)                     // 查詢（自動加 WHERE id = ?）
	if errors.Is(result.Error, gorm.ErrRecordNotFound) { // 如果找不到
		return nil, ErrArticleNotFound                   // 回傳領域錯誤
	}
	if result.Error != nil { // 其他 DB 錯誤
		r.logger.Error("查詢文章失敗", zap.Uint("id", id), zap.Error(result.Error))
		return nil, result.Error // 回傳原始錯誤
	}
	return model.toArticle(), nil // 轉換成 Domain Entity 後回傳
}

// FindByAuthor 查詢作者的所有文章
func (r *GormArticleRepository) FindByAuthor(authorID uint) ([]*Article, error) {
	var models []ArticleModel                           // 建立空的 Model slice
	result := r.db.Where("author_id = ?", authorID).   // 條件：作者 ID
		Order("created_at desc").                        // 最新的排前面
		Find(&models)                                    // 執行查詢
	if result.Error != nil {                             // 如果查詢失敗
		return nil, result.Error                         // 回傳錯誤
	}
	articles := make([]*Article, len(models))            // 建立結果 slice
	for i, m := range models {                           // 遍歷並轉換
		articles[i] = m.toArticle()
	}
	return articles, nil // 回傳結果
}

// FindPublished 查詢已發布的文章（分頁）
func (r *GormArticleRepository) FindPublished(page, pageSize int) ([]*Article, int64, error) {
	var models []ArticleModel // 建立空的 Model slice
	var total int64           // 儲存總數

	query := r.db.Where("status = ?", "published") // 基礎條件

	query.Model(&ArticleModel{}).Count(&total)      // 先計算總數（不限分頁）

	result := query.
		Order("created_at desc").                   // 最新的排前面
		Offset((page - 1) * pageSize).              // 跳過幾筆
		Limit(pageSize).                             // 每頁幾筆
		Find(&models)                               // 執行查詢

	if result.Error != nil { // 如果查詢失敗
		return nil, 0, result.Error // 回傳錯誤
	}

	articles := make([]*Article, len(models)) // 建立結果 slice
	for i, m := range models {                // 遍歷並轉換
		articles[i] = m.toArticle()
	}
	return articles, total, nil // 回傳結果、總數
}

// Create 建立文章
func (r *GormArticleRepository) Create(input CreateArticleInput) (*Article, error) {
	model := ArticleModel{       // 建立 Model
		Title:    input.Title,   // 設定標題
		Content:  input.Content, // 設定內容
		AuthorID: input.AuthorID, // 設定作者
		Status:   "draft",       // 預設草稿
	}
	result := r.db.Create(&model) // 執行 INSERT
	if result.Error != nil {       // 如果失敗
		return nil, result.Error   // 回傳錯誤
	}
	return model.toArticle(), nil // 回傳新建的文章
}

// Update 更新文章
func (r *GormArticleRepository) Update(id uint, input UpdateArticleInput) (*Article, error) {
	var model ArticleModel                                // 建立空的 Model
	if err := r.db.First(&model, id).Error; err != nil { // 先查詢
		if errors.Is(err, gorm.ErrRecordNotFound) {      // 找不到
			return nil, ErrArticleNotFound               // 回傳領域錯誤
		}
		return nil, err // 其他錯誤
	}

	// 只更新有提供的欄位（指標不為 nil 才更新）
	updates := map[string]any{} // 用 map 儲存要更新的欄位
	if input.Title != nil {     // 如果有提供標題
		updates["title"] = *input.Title // 加入更新 map
	}
	if input.Content != nil { // 如果有提供內容
		updates["content"] = *input.Content
	}
	if input.Status != nil { // 如果有提供狀態
		updates["status"] = *input.Status
	}

	if len(updates) > 0 { // 只有有東西要更新才執行
		r.db.Model(&model).Updates(updates) // 執行 UPDATE
	}

	return model.toArticle(), nil // 回傳更新後的文章
}

// Delete 軟刪除文章（GORM 的 DeletedAt 讓資料還在，只是標記為已刪除）
func (r *GormArticleRepository) Delete(id uint) error {
	result := r.db.Delete(&ArticleModel{}, id) // 軟刪除（設定 deleted_at）
	if result.RowsAffected == 0 {              // 沒有刪除任何東西
		return ErrArticleNotFound              // 回傳領域錯誤
	}
	return result.Error // 回傳錯誤（nil 代表成功）
}

// ==========================================================================
// ██╗   ██╗███████╗███████╗ ██████╗ █████╗ ███████╗███████╗
// ██║   ██║██╔════╝██╔════╝██╔════╝██╔══██╗██╔════╝██╔════╝
// ██║   ██║███████╗█████╗  ██║     ███████║███████╗█████╗
// ██║   ██║╚════██║██╔══╝  ██║     ██╔══██║╚════██║██╔══╝
// ╚██████╔╝███████║███████╗╚██████╗██║  ██║███████║███████╗
//  ╚═════╝ ╚══════╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚══════╝
//
// Usecase 層（業務邏輯層）
// ————————————————————————
// 只包含業務規則，不知道 HTTP、資料庫、日誌的細節
// 所有依賴都透過介面注入
// ==========================================================================

// articleUsecase 實作 ArticleUsecase 介面
type articleUsecase struct {
	repo   ArticleRepository // Repository 介面（注入進來）
	logger *zap.Logger       // 日誌（注入進來）
}

// NewArticleUsecase 建立 ArticleUsecase（依賴注入用）
func NewArticleUsecase(repo ArticleRepository, logger *zap.Logger) ArticleUsecase {
	return &articleUsecase{repo: repo, logger: logger} // 回傳介面
}

// GetArticle 取得單篇文章（業務規則：任何人都可以看已發布的文章）
func (u *articleUsecase) GetArticle(id uint) (*Article, error) {
	u.logger.Info("取得文章", zap.Uint("id", id)) // 記錄日誌
	return u.repo.FindByID(id)                   // 委託給 Repository
}

// ListPublished 取得已發布文章列表
func (u *articleUsecase) ListPublished(page, pageSize int) ([]*Article, int64, error) {
	if page < 1 {     // 頁數不能小於 1
		page = 1       // 自動修正
	}
	if pageSize < 1 || pageSize > 100 { // 每頁大小限制在 1-100
		pageSize = 10  // 預設 10
	}
	return u.repo.FindPublished(page, pageSize) // 委託給 Repository
}

// CreateArticle 建立文章（業務規則：標題和內容不能為空）
func (u *articleUsecase) CreateArticle(authorID uint, input CreateArticleInput) (*Article, error) {
	// 業務驗證（這不是 HTTP 驗證，是業務規則）
	if input.Title == "" {   // 標題不能為空
		return nil, ErrTitleRequired // 回傳業務錯誤
	}
	if input.Content == "" { // 內容不能為空
		return nil, ErrContentRequired // 回傳業務錯誤
	}

	input.AuthorID = authorID // 確保 AuthorID 是登入的使用者

	article, err := u.repo.Create(input) // 呼叫 Repository 建立
	if err != nil {                       // 如果失敗
		u.logger.Error("建立文章失敗",     // 記錄錯誤日誌
			zap.Uint("author_id", authorID),
			zap.Error(err),
		)
		return nil, err // 回傳錯誤
	}

	u.logger.Info("文章建立成功",          // 記錄成功日誌
		zap.Uint("article_id", article.ID),
		zap.Uint("author_id", authorID),
	)
	return article, nil // 回傳新文章
}

// UpdateArticle 更新文章（業務規則：只有作者才能更新自己的文章）
func (u *articleUsecase) UpdateArticle(id, authorID uint, input UpdateArticleInput) (*Article, error) {
	existing, err := u.repo.FindByID(id) // 先查詢文章是否存在
	if err != nil {                       // 如果查詢失敗
		return nil, err                  // 回傳錯誤（可能是 ErrArticleNotFound）
	}

	// 業務規則：只有作者才能更新
	if existing.AuthorID != authorID { // 如果不是作者
		return nil, ErrArticleForbidden // 回傳 403 錯誤
	}

	return u.repo.Update(id, input) // 呼叫 Repository 更新
}

// DeleteArticle 刪除文章（業務規則：只有作者才能刪除自己的文章）
func (u *articleUsecase) DeleteArticle(id, authorID uint) error {
	existing, err := u.repo.FindByID(id) // 先查詢文章是否存在
	if err != nil {                       // 如果查詢失敗
		return err // 回傳錯誤
	}

	// 業務規則：只有作者才能刪除
	if existing.AuthorID != authorID { // 如果不是作者
		return ErrArticleForbidden // 回傳 403 錯誤
	}

	return u.repo.Delete(id) // 呼叫 Repository 刪除
}

// ==========================================================================
// ██╗  ██╗ █████╗ ███╗   ██╗██████╗ ██╗     ███████╗██████╗
// ██║  ██║██╔══██╗████╗  ██║██╔══██╗██║     ██╔════╝██╔══██╗
// ███████║███████║██╔██╗ ██║██║  ██║██║     █████╗  ██████╔╝
// ██╔══██║██╔══██║██║╚██╗██║██║  ██║██║     ██╔══╝  ██╔══██╗
// ██║  ██║██║  ██║██║ ╚████║██████╔╝███████╗███████╗██║  ██║
// ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝
//
// Handler 層（展示層）
// ————————————————————
// 負責處理 HTTP 請求和回應
// 不包含業務邏輯，只做：解析請求 → 呼叫 Usecase → 格式化回應
// ==========================================================================

// ArticleHandler HTTP 處理器（持有 Usecase 介面）
type ArticleHandler struct {
	usecase ArticleUsecase // Usecase 介面（注入進來）
	logger  *zap.Logger    // 日誌（注入進來）
}

// NewArticleHandler 建立 ArticleHandler（依賴注入用）
func NewArticleHandler(usecase ArticleUsecase, logger *zap.Logger) *ArticleHandler {
	return &ArticleHandler{usecase: usecase, logger: logger}
}

// RegisterRoutes 把路由註冊到 Gin Router
func (h *ArticleHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 公開路由（不需要 JWT）
	router.GET("/articles", h.ListArticles)       // 取得文章列表
	router.GET("/articles/:id", h.GetArticle)     // 取得單篇文章

	// 需要認證的路由（需要 JWT）
	auth := router.Group("/")            // 建立子路由組
	auth.Use(authMiddleware)             // 套用 JWT 中介層
	auth.POST("/articles", h.CreateArticle)           // 建立文章
	auth.PUT("/articles/:id", h.UpdateArticle)         // 更新文章
	auth.DELETE("/articles/:id", h.DeleteArticle)      // 刪除文章
}

// articleResponse 文章的 JSON 回應格式
type articleResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  uint      `json:"author_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// toResponse 把 Domain Entity 轉換成 JSON 回應格式
func toResponse(a *Article) articleResponse {
	return articleResponse{
		ID:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		AuthorID:  a.AuthorID,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
	}
}

// errorResponse 錯誤的 JSON 回應格式
type errorResponse struct {
	Error string `json:"error"` // 錯誤訊息
}

// handleError 把 Domain 錯誤轉換成對應的 HTTP 狀態碼
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrArticleNotFound):   // 找不到 → 404
		c.JSON(http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrArticleForbidden):  // 沒有權限 → 403
		c.JSON(http.StatusForbidden, errorResponse{Error: err.Error()})
	case errors.Is(err, ErrTitleRequired),     // 輸入錯誤 → 400
		errors.Is(err, ErrContentRequired),
		errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	default:                                   // 其他錯誤 → 500
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "伺服器內部錯誤"})
	}
}

// ListArticles GET /articles?page=1&page_size=10
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))           // 取得頁碼
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10")) // 取得每頁大小

	articles, total, err := h.usecase.ListPublished(page, pageSize) // 呼叫 Usecase
	if err != nil {                                                   // 如果失敗
		handleError(c, err)                                          // 轉換錯誤
		return
	}

	responses := make([]articleResponse, len(articles)) // 建立回應 slice
	for i, a := range articles {                        // 遍歷並轉換
		responses[i] = toResponse(a)
	}

	c.JSON(http.StatusOK, gin.H{ // 回傳 JSON
		"data":      responses,    // 文章列表
		"total":     total,        // 總數
		"page":      page,         // 當前頁碼
		"page_size": pageSize,     // 每頁大小
	})
}

// GetArticle GET /articles/:id
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64) // 解析路徑參數
	if err != nil {                                       // 如果 id 不是數字
		c.JSON(http.StatusBadRequest, errorResponse{Error: "無效的文章 ID"})
		return
	}

	article, err := h.usecase.GetArticle(uint(id)) // 呼叫 Usecase
	if err != nil {                                  // 如果失敗
		handleError(c, err)                         // 轉換錯誤
		return
	}

	c.JSON(http.StatusOK, toResponse(article)) // 回傳 JSON
}

// createArticleRequest 建立文章的請求 body
type createArticleRequest struct {
	Title   string `json:"title" binding:"required"`   // 必填
	Content string `json:"content" binding:"required"` // 必填
}

// CreateArticle POST /articles（需要 JWT）
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req createArticleRequest                  // 解析請求 body
	if err := c.ShouldBindJSON(&req); err != nil { // 如果格式不對
		c.JSON(http.StatusBadRequest, errorResponse{Error: "請提供 title 和 content"})
		return
	}

	authorID := c.GetUint("user_id") // 從 JWT 中介層取得使用者 ID

	article, err := h.usecase.CreateArticle(authorID, CreateArticleInput{
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil { // 如果失敗
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(article)) // 201 Created
}

// updateArticleRequest 更新文章的請求 body
type updateArticleRequest struct {
	Title   *string `json:"title"`   // 可選
	Content *string `json:"content"` // 可選
	Status  *string `json:"status"`  // 可選（draft 或 published）
}

// UpdateArticle PUT /articles/:id（需要 JWT）
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64) // 解析路徑參數
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "無效的文章 ID"})
		return
	}

	var req updateArticleRequest                          // 解析請求 body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "請求格式錯誤"})
		return
	}

	authorID := c.GetUint("user_id")                    // 從 JWT 取得使用者 ID

	article, err := h.usecase.UpdateArticle(uint(id), authorID, UpdateArticleInput{
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(article)) // 200 OK
}

// DeleteArticle DELETE /articles/:id（需要 JWT）
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64) // 解析路徑參數
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "無效的文章 ID"})
		return
	}

	authorID := c.GetUint("user_id")           // 從 JWT 取得使用者 ID

	if err := h.usecase.DeleteArticle(uint(id), authorID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章已刪除"}) // 200 OK
}

// ==========================================================================
// ███╗   ███╗██╗██████╗ ██████╗ ██╗     ███████╗██╗    ██╗ █████╗ ██████╗ ███████╗
// ████╗ ████║██║██╔══██╗██╔══██╗██║     ██╔════╝██║    ██║██╔══██╗██╔══██╗██╔════╝
// ██╔████╔██║██║██║  ██║██║  ██║██║     █████╗  ██║ █╗ ██║███████║██████╔╝█████╗
// ██║╚██╔╝██║██║██║  ██║██║  ██║██║     ██╔══╝  ██║███╗██║██╔══██║██╔══██╗██╔══╝
// ██║ ╚═╝ ██║██║██████╔╝██████╔╝███████╗███████╗╚███╔███╔╝██║  ██║██║  ██║███████╗
// ╚═╝     ╚═╝╚═╝╚═════╝ ╚═════╝ ╚══════╝╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
//
// Middleware 層（中介層）
// ==========================================================================

// jwtSecret JWT 簽名密鑰（正式環境從環境變數讀取）
var jwtSecret = []byte(func() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s // 從環境變數讀取
	}
	return "dev-secret-change-in-production" // 開發用預設值
}())

// Claims JWT Payload 結構
type Claims struct {
	UserID   uint   `json:"user_id"`  // 使用者 ID
	Username string `json:"username"` // 使用者名稱
	jwt.RegisteredClaims               // 標準欄位（exp、iat 等）
}

// GenerateToken 產生 JWT Token（測試和登入 API 用）
func GenerateToken(userID uint, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 小時過期
			IssuedAt:  jwt.NewNumericDate(time.Now()),                     // 發行時間
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 建立 token
	return token.SignedString(jwtSecret)                        // 簽名並回傳
}

// JWTMiddleware JWT 認證中介層
func JWTMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization") // 取得 Authorization header
		if authHeader == "" {                       // 如果沒有提供
			c.JSON(http.StatusUnauthorized, errorResponse{Error: "缺少 Authorization header"})
			c.Abort() // 中止後續處理
			return
		}

		// 格式應該是 "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, errorResponse{Error: "Authorization 格式應為 Bearer <token>"})
			c.Abort()
			return
		}

		tokenStr := authHeader[7:] // 取出 token 部分（去掉 "Bearer "）

		claims := &Claims{}   // 準備接收 Claims
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			return jwtSecret, nil // 回傳密鑰（用於驗證簽名）
		})

		if err != nil || !token.Valid { // 如果解析失敗或 token 無效
			logger.Warn("JWT 驗證失敗", zap.Error(err))
			c.JSON(http.StatusUnauthorized, errorResponse{Error: "Token 無效或已過期"})
			c.Abort()
			return
		}

		// 把 user_id 存入 Gin context，Handler 可以用 c.GetUint("user_id") 取得
		c.Set("user_id", claims.UserID)      // 使用者 ID
		c.Set("username", claims.Username)   // 使用者名稱
		c.Next()                             // 繼續下一個 handler
	}
}

// LoggingMiddleware 請求日誌中介層（每個請求自動記錄）
func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // 記錄開始時間

		c.Next() // 先執行後續 handler（洋蔥模型）

		// handler 執行完成後記錄日誌
		duration := time.Since(start)                    // 計算耗時
		status := c.Writer.Status()                      // 取得狀態碼
		logger.Info("HTTP 請求",                          // 記錄請求日誌
			zap.String("method", c.Request.Method),     // HTTP 方法
			zap.String("path", c.Request.URL.Path),     // 請求路徑
			zap.Int("status", status),                  // 狀態碼
			zap.Duration("duration", duration),          // 耗時
			zap.String("ip", c.ClientIP()),             // 客戶端 IP
		)
	}
}

// ==========================================================================
// ███╗   ███╗ █████╗ ██╗███╗   ██╗
// ████╗ ████║██╔══██╗██║████╗  ██║
// ██╔████╔██║███████║██║██╔██╗ ██║
// ██║╚██╔╝██║██╔══██║██║██║╚██╗██║
// ██║ ╚═╝ ██║██║  ██║██║██║ ╚████║
// ╚═╝     ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝
//
// main() — 依賴注入的組裝點
// ——————————————————————————
// 這裡是「餐廳老闆」的工作：
// 買冰箱（DB）、買食材（Repository）、雇廚師（Usecase）、雇服務生（Handler）
// 把一切組裝起來，讓餐廳開始運作
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================")
	fmt.Println(" 第二十四課：Clean Architecture 進階")
	fmt.Println("==========================================")

	// ──── 1. 初始化日誌（最先初始化，所有元件都需要它）────
	zapLogger, err := zap.NewDevelopment() // 開發模式（人類可讀）
	if err != nil {
		log.Fatalf("初始化日誌失敗: %v", err)
	}
	defer zapLogger.Sync() // 確保日誌都被寫出

	zapLogger.Info("日誌初始化完成")

	// ──── 2. 初始化資料庫 ────
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // GORM 日誌靜默（用 zap 代替）
	})
	if err != nil {
		zapLogger.Fatal("資料庫連線失敗", zap.Error(err))
	}

	if err := db.AutoMigrate(&ArticleModel{}); err != nil {
		zapLogger.Fatal("資料庫遷移失敗", zap.Error(err))
	}
	zapLogger.Info("資料庫初始化完成")

	// ──── 3. 依賴注入鏈 ────
	//
	//   db + logger
	//      ↓
	//   ArticleRepository（GormArticleRepository）
	//      ↓
	//   ArticleUsecase（articleUsecase）
	//      ↓
	//   ArticleHandler
	//      ↓
	//   Router

	articleRepo    := NewGormArticleRepository(db, zapLogger)       // Repository
	articleUsecase := NewArticleUsecase(articleRepo, zapLogger)     // Usecase
	articleHandler := NewArticleHandler(articleUsecase, zapLogger)  // Handler

	zapLogger.Info("依賴注入完成",
		zap.String("repository", "GormArticleRepository"),
		zap.String("usecase", "articleUsecase"),
		zap.String("handler", "ArticleHandler"),
	)

	// ──── 4. 建立示範資料 ────
	seedData(articleRepo, zapLogger) // 插入測試文章

	// ──── 5. 設定 Gin Router ────
	gin.SetMode(gin.DebugMode) // 開發模式
	router := gin.New()        // 建立 router（不使用預設 middleware）

	// 套用中介層
	router.Use(LoggingMiddleware(zapLogger)) // 請求日誌
	router.Use(gin.Recovery())              // Panic 恢復

	// ──── Health Check 路由（不加版本號，K8s/Docker 直接探測）────
	// 這些路由不應該有 /api/v1 前綴，因為 K8s 會直接呼叫 /healthz
	router.GET("/healthz", healthzHandler)       // Liveness：我還活著嗎？
	router.GET("/readyz", readyzHandler(db))     // Readiness：我準備好了嗎？

	// ──── API 版本控制 ────
	//
	// 為什麼需要 API 版本控制？
	//   - v1 的回應格式有問題，v2 修正了但改變了 JSON 欄位名稱
	//   - 有些客戶端還在用 v1（手機 App 舊版），不能直接破壞它們
	//   - 解法：v1 和 v2 並行運行，新客戶端用 v2，舊客戶端繼續用 v1
	//
	// 常見版本控制方式：
	//   URL 路徑（推薦）：/api/v1/articles、/api/v2/articles
	//   Header：Accept: application/vnd.api+json;version=2
	//   查詢參數：/api/articles?version=2

	jwtMw := JWTMiddleware(zapLogger)

	// API v1（現有版本）
	apiV1 := router.Group("/api/v1")
	router.POST("/api/v1/login", loginHandler) // 登入（取得 Token）
	articleHandler.RegisterRoutes(apiV1, jwtMw)

	// API v2（示範：加上 X-API-Version header）
	// 實際上 v2 可以有不同的 Handler，這裡只示範版本控制的結構
	apiV2 := router.Group("/api/v2")
	apiV2.Use(func(c *gin.Context) {
		c.Header("X-API-Version", "2")  // 在所有 v2 回應加上版本 header
		c.Next()
	})
	// v2 的文章列表（示範：回應格式略有不同，加了 api_version 欄位）
	apiV2.GET("/articles", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
		articles, total, err := articleUsecase.ListPublished(page, pageSize)
		if err != nil {
			handleError(c, err)
			return
		}
		responses := make([]articleResponse, len(articles))
		for i, a := range articles {
			responses[i] = toResponse(a)
		}
		c.JSON(http.StatusOK, gin.H{ // v2 回應格式（比 v1 多了 api_version）
			"api_version": "v2",
			"data":        responses,
			"meta":        gin.H{"total": total, "page": page, "page_size": pageSize},
		})
	})

	// ──── 6. 標記應用程式為就緒（初始化完成）────
	appReady = true // readyz 端點現在會回 200
	zapLogger.Info("應用程式就緒")

	// ──── 7. Graceful Shutdown（優雅關閉）────
	//
	// 什麼是 Graceful Shutdown？
	//   當收到 Ctrl+C 或 SIGTERM 時：
	//
	//   ❌ 硬關閉（router.Run 直接結束）：
	//      → 正在處理的請求直接中斷
	//      → 客戶端收到 Connection Reset 錯誤
	//      → 可能資料寫到一半
	//
	//   ✅ Graceful Shutdown：
	//      1. 停止接受新連線
	//      2. 等待所有進行中的請求完成
	//      3. 清理資源（關閉 DB 連線、刷新日誌）
	//      4. 程式正常退出
	//
	// K8s 的 SIGTERM 流程：
	//   K8s 要關閉 Pod → 發送 SIGTERM → 等 terminationGracePeriodSeconds（預設 30 秒）
	//   → 如果還沒結束，發送 SIGKILL（強制殺死）

	addr := ":8080"
	srv := &http.Server{ // 用 http.Server 代替 router.Run（才能做 Graceful Shutdown）
		Addr:    addr,
		Handler: router,
	}

	// 在獨立的 goroutine 中啟動伺服器（讓 main goroutine 等待信號）
	go func() {
		zapLogger.Info("伺服器啟動", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// http.ErrServerClosed 是正常關閉時的錯誤，不需要記錄
			zapLogger.Fatal("伺服器啟動失敗", zap.Error(err))
		}
	}()

	fmt.Println()
	fmt.Println("API 端點：")
	fmt.Println("  GET  /healthz                  → Liveness 探針")
	fmt.Println("  GET  /readyz                   → Readiness 探針")
	fmt.Println("  POST /api/v1/login             → 取得 Token")
	fmt.Println("  GET  /api/v1/articles          → 文章列表（v1）")
	fmt.Println("  GET  /api/v2/articles          → 文章列表（v2，有 api_version 欄位）")
	fmt.Println()
	fmt.Println("按 Ctrl+C 優雅關閉...")

	// 等待系統信號（SIGINT = Ctrl+C，SIGTERM = Docker/K8s 發的關閉信號）
	quit := make(chan os.Signal, 1)                      // 建立信號接收 channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 訂閱 SIGINT 和 SIGTERM
	sig := <-quit                                         // 阻塞等待信號
	zapLogger.Info("收到關閉信號", zap.String("signal", sig.String()))

	// 開始 Graceful Shutdown
	appReady = false // 先把 readyz 設為不就緒（Load Balancer 停止導流量）

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 最多等 30 秒
	defer cancel()

	zapLogger.Info("開始 Graceful Shutdown（最多等 30 秒）")
	if err := srv.Shutdown(ctx); err != nil { // 等待所有請求完成後關閉
		zapLogger.Error("Graceful Shutdown 失敗", zap.Error(err))
	} else {
		zapLogger.Info("伺服器已優雅關閉")
	}
}

// ==========================================================================
// Health Check Handlers（健康檢查端點）
// ==========================================================================
//
// 為什麼需要 Health Check？
//   在 K8s / Docker Swarm 等容器平台，Health Check 是「必備」功能：
//
//   /healthz（Liveness Probe）
//     → 問「你還活著嗎？」
//     → 如果回 200，K8s 認為 Pod 正常運作
//     → 如果回非 200 或超時，K8s 重啟這個 Pod
//     → 只要程式在跑就回 200，不要做複雜檢查
//
//   /readyz（Readiness Probe）
//     → 問「你準備好接受流量了嗎？」
//     → 如果回 200，Load Balancer 把流量導進來
//     → 如果回非 200，Load Balancer 暫時不導流量（但不重啟）
//     → 應該檢查所有依賴（DB 連線、Redis 連線）是否正常

// appStatus 儲存應用程式的就緒狀態（全域，由 main 設定）
var appReady = false // 預設未就緒，初始化完成後設為 true

// healthzHandler /healthz — Liveness（存活探針）
// 只要程式在跑就回 200
func healthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",                          // 存活狀態
		"time":    time.Now().Format(time.RFC3339), // 目前時間
	})
}

// readyzHandler /readyz — Readiness（就緒探針）
// 檢查所有依賴都正常後才回 200
func readyzHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果應用程式還沒初始化完成
		if !appReady {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "應用程式初始化中",
			})
			return
		}

		// 檢查資料庫連線（ping DB）
		sqlDB, err := db.DB()           // 取得底層的 *sql.DB
		if err != nil || sqlDB.Ping() != nil { // 如果取得失敗或 Ping 失敗
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "資料庫連線異常",
			})
			return
		}

		// 所有依賴都正常
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",                       // 就緒狀態
			"goroutines": runtime.NumGoroutine(),       // 目前 goroutine 數量
			"time":      time.Now().Format(time.RFC3339),
		})
	}
}

// loginHandler 示範用登入 API（直接用傳入的 user_id 產生 token）
// 正式環境：這裡應該驗證帳號密碼、查資料庫
func loginHandler(c *gin.Context) {
	var req struct {
		UserID   uint   `json:"user_id" binding:"required"`
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "請提供 user_id 和 username"})
		return
	}

	token, err := GenerateToken(req.UserID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "產生 Token 失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"user_id":  req.UserID,
		"username": req.Username,
	})
}

// seedData 插入示範資料
func seedData(repo ArticleRepository, logger *zap.Logger) {
	articles := []CreateArticleInput{
		{Title: "Go 入門：Hello World", Content: "fmt.Println(\"Hello, World!\")", AuthorID: 1},
		{Title: "Go 並發：Goroutine 基礎", Content: "go func() { ... }()", AuthorID: 1},
		{Title: "Redis 快取策略", Content: "Cache-Aside 是最常用的快取模式", AuthorID: 2},
		{Title: "Clean Architecture 的好處", Content: "依賴方向從外到內，核心不依賴框架", AuthorID: 2},
	}

	published := "published"
	for _, input := range articles {
		article, err := repo.Create(input)
		if err != nil {
			logger.Error("插入示範資料失敗", zap.Error(err))
			continue
		}
		// 把示範資料都設為已發布
		repo.Update(article.ID, UpdateArticleInput{Status: &published})
	}
	logger.Info("示範資料插入完成", zap.Int("count", len(articles)))
}
