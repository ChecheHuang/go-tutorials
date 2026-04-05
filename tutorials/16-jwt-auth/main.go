// 第十五課：JWT 認證
// 學習 JWT（JSON Web Token）的原理與實作
//
// 執行前安裝：
//   go get github.com/golang-jwt/jwt/v5
//   go get golang.org/x/crypto/bcrypt
//   go get github.com/gin-gonic/gin
//
// 執行方式：go run main.go
package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ========================================
// JWT 基本概念
// ========================================
//
// JWT 由三部分組成，用 . 分隔：
//   xxxxx.yyyyy.zzzzz
//   Header.Payload.Signature
//
// Header:    {"alg": "HS256", "typ": "JWT"}     → Base64 編碼
// Payload:   {"user_id": 1, "exp": 1234567890}  → Base64 編碼
// Signature: HMACSHA256(header + "." + payload, secret)
//
// 流程：
// 1. 使用者登入 → 伺服器產生 JWT → 回傳給客戶端
// 2. 客戶端每次請求帶上 JWT（Authorization: Bearer <token>）
// 3. 伺服器驗證 JWT 的簽名和有效期

// JWT 密鑰（生產環境要用環境變數）
const jwtSecret = "my-super-secret-key-for-tutorial"

// ========================================
// 使用者模型與模擬資料庫
// ========================================

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"` // 不會出現在 JSON 回應中
}

// 模擬資料庫
var users []User
var nextID uint = 1

// ========================================
// 1. 密碼雜湊（bcrypt）
// ========================================

// hashPassword 將明文密碼轉為 bcrypt 雜湊值
func hashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10，越高越安全但越慢
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword 比對密碼與雜湊值
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ========================================
// 2. JWT Token 產生與驗證
// ========================================

// generateToken 為使用者產生 JWT Token
func generateToken(userID uint) (string, error) {
	// 建立 Claims（Token 中攜帶的資料）
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // 24 小時後過期
		"iat":     time.Now().Unix(),                     // 簽發時間
	}

	// 用 HS256 演算法建立 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 用密鑰簽名並取得字串形式的 Token
	return token.SignedString([]byte(jwtSecret))
}

// validateToken 驗證 JWT Token 並回傳 Claims
func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 確認演算法是 HMAC（防止演算法替換攻擊）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支援的簽名演算法: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("無效的 Token")
}

// ========================================
// 3. JWT 中介層
// ========================================

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取得 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization header"})
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "格式錯誤，請使用: Bearer <token>"})
			c.Abort()
			return
		}

		// 驗證 Token
		claims, err := validateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 無效: " + err.Error()})
			c.Abort()
			return
		}

		// 從 Claims 提取 user_id 並存入 Context
		userID := uint(claims["user_id"].(float64))
		c.Set("user_id", userID)

		c.Next()
	}
}

// ========================================
// 4. Handler
// ========================================

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// register 使用者註冊
func register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 雜湊密碼
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼加密失敗"})
		return
	}

	// 建立使用者
	user := User{
		ID:       nextID,
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}
	nextID++
	users = append(users, user)

	c.JSON(http.StatusCreated, gin.H{
		"message": "註冊成功",
		"user":    user, // Password 不會出現（json:"-"）
	})
}

// login 使用者登入
func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找使用者
	var foundUser *User
	for i := range users {
		if users[i].Email == req.Email {
			foundUser = &users[i]
			break
		}
	}

	if foundUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "信箱或密碼錯誤"})
		return
	}

	// 驗證密碼
	if !checkPassword(req.Password, foundUser.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "信箱或密碼錯誤"})
		return
	}

	// 產生 JWT
	token, err := generateToken(foundUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "產生 Token 失敗"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登入成功",
		"token":   token,
		"user":    foundUser,
	})
}

// profile 取得個人資料（需要認證）
func profile(c *gin.Context) {
	userID := c.GetUint("user_id") // 從 JWT 中介層取得

	// 查找使用者
	for _, user := range users {
		if user.ID == userID {
			c.JSON(http.StatusOK, gin.H{"user": user})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "使用者不存在"})
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// ===== 示範 bcrypt =====
	fmt.Println("=== bcrypt 密碼雜湊示範 ===")
	hash, _ := hashPassword("mypassword")
	fmt.Println("明文: mypassword")
	fmt.Println("雜湊:", hash)
	fmt.Println("驗證正確密碼:", checkPassword("mypassword", hash))
	fmt.Println("驗證錯誤密碼:", checkPassword("wrongpassword", hash))

	// ===== 示範 JWT =====
	fmt.Println("\n=== JWT Token 示範 ===")
	token, _ := generateToken(1)
	fmt.Println("Token:", token[:50]+"...")
	claims, _ := validateToken(token)
	fmt.Printf("解析 Claims: user_id=%.0f\n", claims["user_id"])

	// ===== 路由設定 =====
	fmt.Println("\n=== 啟動伺服器 ===")

	// 公開路由
	r.POST("/register", register)
	r.POST("/login", login)

	// 受保護路由
	auth := r.Group("/")
	auth.Use(JWTMiddleware())
	{
		auth.GET("/profile", profile)
	}

	fmt.Println("伺服器啟動於 http://localhost:9090")
	fmt.Println()
	fmt.Println("測試步驟：")
	fmt.Println("1. 註冊:")
	fmt.Println(`   curl -X POST http://localhost:9090/register -H "Content-Type: application/json" -d '{"username":"alice","email":"alice@test.com","password":"123456"}'`)
	fmt.Println()
	fmt.Println("2. 登入（取得 Token）:")
	fmt.Println(`   curl -X POST http://localhost:9090/login -H "Content-Type: application/json" -d '{"email":"alice@test.com","password":"123456"}'`)
	fmt.Println()
	fmt.Println("3. 存取受保護路由（帶上 Token）:")
	fmt.Println(`   curl -H "Authorization: Bearer <貼上你的token>" http://localhost:9090/profile`)

	r.Run(":9090")
}
