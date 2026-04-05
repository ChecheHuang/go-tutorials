package handler

import (
	"net/http/pprof"

	"blog-api/internal/middleware"
	"blog-api/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	r.Use(middleware.Metrics())  // Prometheus 指標（放在最前面才能測量所有請求）
	r.Use(middleware.Logger())   // 結構化日誌
	r.Use(middleware.Recovery()) // Panic 恢復
	r.Use(middleware.CORS())     // 跨域設定

	// 系統端點
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// pprof 效能分析端點（僅 debug 模式啟用）
	if cfg.Server.Mode == "debug" {
		pprofGroup := r.Group("/debug/pprof")
		{
			pprofGroup.GET("/", gin.WrapF(pprof.Index))
			pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
			pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
			pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
			pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
			pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
			pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
			pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
			pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
			pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
		}
	}

	// Swagger API 文件
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 路由群組
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

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

		v1.GET("/articles", articleHandler.GetAll)
		v1.GET("/articles/:id", articleHandler.GetByID)
		v1.GET("/articles/:id/comments", commentHandler.GetByArticleID)
	}

	return r
}
