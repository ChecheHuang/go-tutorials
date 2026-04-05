package handler

import (
	"blog-api/internal/middleware"
	"blog-api/pkg/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 設定所有路由與中介層
func SetupRouter(
	cfg *config.Config,
	userHandler *UserHandler,
	articleHandler *ArticleHandler,
	commentHandler *CommentHandler,
	healthHandler *HealthHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 全域中介層
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康檢查端點（Kubernetes Probe 用）
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	// Swagger API 文件
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由群組
	v1 := r.Group("/api/v1")
	{
		// 認證相關路由（不需要登入）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// 需要 JWT 認證的路由
		authenticated := v1.Group("")
		authenticated.Use(middleware.JWTAuth(cfg))
		{
			authenticated.GET("/auth/profile", userHandler.GetProfile)

			authenticated.POST("/articles", articleHandler.Create)
			authenticated.PUT("/articles/:id", articleHandler.Update)
			authenticated.DELETE("/articles/:id", articleHandler.Delete)

			authenticated.POST("/articles/:id/comments", commentHandler.Create)
			authenticated.PUT("/comments/:id", commentHandler.Update)
			authenticated.DELETE("/comments/:id", commentHandler.Delete)
		}

		// 公開路由
		v1.GET("/articles", articleHandler.GetAll)
		v1.GET("/articles/:id", articleHandler.GetByID)
		v1.GET("/articles/:id/comments", commentHandler.GetByArticleID)
	}

	return r
}
