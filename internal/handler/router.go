package handler

import (
	"blog-api/internal/middleware"
	"blog-api/pkg/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 設定所有路由與中介層
// 這是整個 API 的路由入口，負責將 URL 路徑對應到正確的 handler
func SetupRouter(
	cfg *config.Config,
	userHandler *UserHandler,
	articleHandler *ArticleHandler,
	commentHandler *CommentHandler,
) *gin.Engine {
	// 設定 Gin 模式並關閉預設的路由日誌（使用自訂的彩色路由表）
	gin.SetMode(gin.ReleaseMode)

	// 建立 Gin 引擎（不使用預設中介層）
	r := gin.New()

	// 註冊全域中介層（每個請求都會經過）
	r.Use(middleware.Logger())   // 日誌記錄
	r.Use(middleware.Recovery()) // Panic 恢復
	r.Use(middleware.CORS())     // 跨域設定

	// Swagger API 文件路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由群組
	v1 := r.Group("/api/v1")
	{
		// === 認證相關路由（不需要登入）===
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register) // 註冊
			auth.POST("/login", userHandler.Login)       // 登入
		}

		// === 需要 JWT 認證的路由 ===
		authenticated := v1.Group("")
		authenticated.Use(middleware.JWTAuth(cfg)) // 套用 JWT 認證中介層
		{
			// 使用者相關
			authenticated.GET("/auth/profile", userHandler.GetProfile) // 取得個人資料

			// 文章 CRUD
			authenticated.POST("/articles", articleHandler.Create)      // 建立文章
			authenticated.PUT("/articles/:id", articleHandler.Update)   // 更新文章
			authenticated.DELETE("/articles/:id", articleHandler.Delete) // 刪除文章

			// 留言 CRUD
			authenticated.POST("/articles/:id/comments", commentHandler.Create)    // 建立留言
			authenticated.PUT("/comments/:id", commentHandler.Update)              // 更新留言
			authenticated.DELETE("/comments/:id", commentHandler.Delete)           // 刪除留言
		}

		// === 公開路由（不需要登入）===
		v1.GET("/articles", articleHandler.GetAll)                        // 取得文章列表
		v1.GET("/articles/:id", articleHandler.GetByID)                  // 取得文章詳情
		v1.GET("/articles/:id/comments", commentHandler.GetByArticleID)  // 取得文章留言
	}

	return r
}
