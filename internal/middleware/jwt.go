// Package middleware 提供 Gin 中介層
// 中介層在請求到達 handler 之前執行，用於處理認證、日誌、跨域等通用邏輯
// 教學對應：第 17 課（中介層）、第 18 課（JWT 認證）
package middleware

import (
	"strings"

	"blog-api/pkg/config"
	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth 回傳 JWT 認證中介層
// 從 Authorization header 中取得 Bearer Token 並驗證
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 Header 取得 Authorization 值
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少認證 Token")
			c.Abort() // 中止後續的 handler 執行
			return
		}

		// 檢查 Bearer Token 格式：「Bearer <token>」
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Token 格式錯誤，請使用 Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析並驗證 JWT Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 確認簽名演算法是 HMAC（防止演算法替換攻擊）
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Token 無效或已過期")
			c.Abort()
			return
		}

		// 從 Token 中提取使用者 ID，並存入 Gin Context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "Token 解析失敗")
			c.Abort()
			return
		}

		// JWT 中的數字預設會被解析為 float64
		userID := uint(claims["user_id"].(float64))
		c.Set("user_id", userID) // 將 user_id 存入 context，供後續 handler 使用

		c.Next() // 繼續執行下一個 handler
	}
}
