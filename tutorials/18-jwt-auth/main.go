// 第十八課：JWT 認證（JSON Web Token Authentication）
// 認證（Authentication）就是「證明你是誰」
// 就像進電影院要出示門票一樣，JWT 就是你的「數位門票」
//
// 本課學習：
//   1. bcrypt 密碼雜湊：安全地儲存密碼
//   2. JWT Token 產生：登入成功後發一張「門票」
//   3. JWT Token 驗證：每次請求出示「門票」
//   4. 完整認證流程：註冊 → 登入 → 取得 Token → 存取受保護資源
//
// 需要安裝：
//   go get github.com/golang-jwt/jwt/v5
//   go get golang.org/x/crypto/bcrypt
//
// 執行方式：go run ./tutorials/16-jwt-auth
package main

import (
	"fmt"      // 格式化輸出
	"strings"  // 字串處理
	"time"     // 時間相關

	"github.com/golang-jwt/jwt/v5" // JWT 套件：產生和驗證 Token
	"golang.org/x/crypto/bcrypt"   // bcrypt 套件：密碼雜湊
)

// ========================================
// 什麼是 JWT？
// ========================================
//
// JWT 就像電影票：
//   - 上面有你的資訊（座位號、場次）     → Payload（負載）
//   - 有電影院的蓋章（防偽標記）          → Signature（簽名）
//   - 有票的類型說明                      → Header（標頭）
//   - 有使用期限（場次時間）              → Expiration（過期時間）
//
// JWT 的結構（三段用 . 分隔）：
//   eyJhbGciOi...  .  eyJ1c2VyX2lk...  .  SflKxwRJSM...
//   └── Header ──┘    └── Payload ───┘    └─ Signature ─┘
//
// Header:    {"alg": "HS256", "typ": "JWT"}      → 演算法和類型
// Payload:   {"user_id": 1, "exp": 1234567890}   → 攜帶的資料
// Signature: HMACSHA256(header + "." + payload, 密鑰) → 防止竄改
//
// 認證流程：
//   1. 使用者用帳號密碼登入
//   2. 伺服器驗證密碼正確 → 發一張 JWT（門票）
//   3. 使用者每次請求帶上 JWT：Authorization: Bearer <token>
//   4. 伺服器驗證 JWT 的簽名和有效期 → 允許存取

// ========================================
// 什麼是 bcrypt？
// ========================================
//
// bcrypt 是一種「單向雜湊演算法」
// 就像把雞蛋打成蛋液：可以打散，但不能還原成完整的蛋
//
//   明文密碼 "mypassword"
//       ↓ bcrypt 雜湊（加入隨機鹽值）
//   "$2a$10$N9qo8uLOickgx2ZMRZoMye..."
//
// 為什麼不能直接存明文密碼？
//   - 資料庫被駭客入侵 → 所有密碼直接外洩
//   - 用 bcrypt 雜湊 → 駭客拿到雜湊值也無法反推原始密碼

// jwtSecret 是簽名用的密鑰（生產環境必須用環境變數，不能寫在程式碼裡）
const jwtSecret = "my-super-secret-key-for-tutorial" // JWT 簽名密鑰

// ========================================
// 第一節：bcrypt 密碼雜湊
// ========================================

// hashPassword 將明文密碼轉為 bcrypt 雜湊值
// 參數：password - 使用者輸入的明文密碼
// 回傳：雜湊後的字串、錯誤
func hashPassword(password string) (string, error) {
	// bcrypt.DefaultCost = 10，數字越高越安全但越慢
	// GenerateFromPassword 會自動加入隨機鹽值（salt）
	// 所以同一個密碼每次雜湊的結果都不一樣！
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),    // 將字串轉為位元組切片
		bcrypt.DefaultCost,  // 使用預設的計算成本（10）
	)
	if err != nil { // 如果雜湊過程出錯
		return "", fmt.Errorf("密碼雜湊失敗: %w", err) // 回傳錯誤
	}
	return string(bytes), nil // 將位元組切片轉回字串並回傳
}

// checkPassword 比對明文密碼與雜湊值是否匹配
// 參數：password - 使用者輸入的密碼，hash - 資料庫中儲存的雜湊值
// 回傳：true 表示密碼正確，false 表示密碼錯誤
func checkPassword(password, hash string) bool {
	// CompareHashAndPassword 會從 hash 中提取鹽值
	// 然後用同樣的鹽值雜湊 password，比對結果
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),     // 資料庫中的雜湊值
		[]byte(password), // 使用者輸入的密碼
	)
	return err == nil // err == nil 表示匹配成功
}

// ========================================
// 第二節：JWT Token 產生
// ========================================

// generateToken 為指定的使用者 ID 產生一個 JWT Token
// 參數：userID - 使用者的唯一識別碼
// 回傳：Token 字串、錯誤
func generateToken(userID uint) (string, error) {
	// Claims 就是 Token 中攜帶的資料（像門票上印的資訊）
	claims := jwt.MapClaims{                                  // 建立一個 Claims 映射
		"user_id": userID,                                    // 使用者 ID（自訂欄位）
		"exp":     time.Now().Add(24 * time.Hour).Unix(),     // 過期時間：24 小時後
		"iat":     time.Now().Unix(),                         // 簽發時間：現在
	}

	// 用 HS256 演算法建立一個新的 Token 物件
	// HS256 = HMAC-SHA256，是最常用的 JWT 簽名演算法
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // 建立 Token

	// 用密鑰簽名，產生最終的 Token 字串
	// 這一步會把 Header.Payload 用密鑰簽名，產生 Header.Payload.Signature
	tokenString, err := token.SignedString([]byte(jwtSecret)) // 簽名並取得字串
	if err != nil {                                           // 如果簽名失敗
		return "", fmt.Errorf("Token 簽名失敗: %w", err)        // 回傳錯誤
	}
	return tokenString, nil // 回傳 Token 字串
}

// ========================================
// 第三節：JWT Token 驗證
// ========================================

// validateToken 驗證 JWT Token 並回傳其中的 Claims
// 參數：tokenString - 客戶端傳來的 Token 字串
// 回傳：Claims（Token 中的資料）、錯誤
func validateToken(tokenString string) (jwt.MapClaims, error) {
	// jwt.Parse 會解析 Token 字串，並用回呼函式提供密鑰
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 安全檢查：確認簽名演算法是 HMAC 系列
		// 這是為了防止「演算法替換攻擊」
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // 如果不是 HMAC 演算法
			return nil, fmt.Errorf("不支援的簽名演算法: %v", token.Header["alg"]) // 拒絕
		}
		return []byte(jwtSecret), nil // 回傳密鑰讓 jwt.Parse 驗證簽名
	})

	if err != nil { // 如果解析或驗證失敗（過期、簽名錯誤等）
		return nil, fmt.Errorf("Token 驗證失敗: %w", err) // 回傳錯誤
	}

	// 從 Token 中提取 Claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid { // 確認 Token 有效
		return claims, nil // 回傳 Claims
	}

	return nil, fmt.Errorf("無效的 Token") // Token 無效
}

// ========================================
// 第四節：模擬使用者資料
// ========================================

// SimulatedUser 模擬的使用者結構
type SimulatedUser struct {
	ID             uint   // 使用者 ID
	Username       string // 使用者名稱
	Email          string // 電子信箱
	HashedPassword string // bcrypt 雜湊後的密碼（絕對不存明文！）
}

// ========================================
// 主程式：完整認證流程展示
// ========================================

func main() {
	// ===== 第一部分：bcrypt 密碼雜湊示範 =====
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第一部分：bcrypt 密碼雜湊")                   // 標題
	fmt.Println("==========================================") // 分隔線

	originalPassword := "mySecurePassword123" // 原始密碼
	fmt.Printf("原始密碼：%s\n", originalPassword)   // 顯示原始密碼

	// 雜湊密碼（就像把雞蛋打成蛋液，無法還原）
	hash1, err := hashPassword(originalPassword) // 第一次雜湊
	if err != nil {                              // 檢查錯誤
		fmt.Printf("雜湊失敗：%v\n", err) // 顯示錯誤
		return                             // 提早返回
	}
	fmt.Printf("雜湊結果 1：%s\n", hash1) // 顯示第一次雜湊結果

	// 同一個密碼，第二次雜湊的結果會不同（因為隨機鹽值）
	hash2, err := hashPassword(originalPassword) // 第二次雜湊
	if err != nil {                              // 檢查錯誤
		fmt.Printf("雜湊失敗：%v\n", err) // 顯示錯誤
		return                             // 提早返回
	}
	fmt.Printf("雜湊結果 2：%s\n", hash2) // 顯示第二次雜湊結果
	fmt.Printf("兩次結果相同嗎？%v\n", hash1 == hash2) // false！因為鹽值不同

	// 驗證密碼
	fmt.Println()                                                     // 空行
	fmt.Printf("驗證正確密碼：%v\n", checkPassword(originalPassword, hash1)) // true
	fmt.Printf("驗證錯誤密碼：%v\n", checkPassword("wrongPassword", hash1))   // false
	fmt.Printf("用第二個雜湊驗證：%v\n", checkPassword(originalPassword, hash2)) // true（雖然雜湊不同，但密碼相同就能通過）

	// ===== 第二部分：JWT Token 產生與驗證 =====
	fmt.Println()                                  // 空行
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第二部分：JWT Token 產生與驗證")           // 標題
	fmt.Println("==========================================") // 分隔線

	// 產生 Token（假設使用者 ID 為 42）
	userID := uint(42)                         // 模擬的使用者 ID
	token, err := generateToken(userID)        // 產生 JWT Token
	if err != nil {                            // 檢查錯誤
		fmt.Printf("產生 Token 失敗：%v\n", err) // 顯示錯誤
		return                                   // 提早返回
	}

	fmt.Printf("使用者 ID：%d\n", userID) // 顯示使用者 ID
	fmt.Printf("JWT Token：%s\n", token)  // 顯示完整的 Token

	// 解析 Token 的三個部分
	parts := strings.Split(token, ".") // 用 . 分割 Token
	fmt.Println()                       // 空行
	fmt.Println("--- JWT 結構解析 ---")  // 小標題
	fmt.Printf("Header（標頭）：  %s\n", parts[0])    // 第一段：演算法資訊
	fmt.Printf("Payload（負載）： %s\n", parts[1])    // 第二段：攜帶的資料
	fmt.Printf("Signature（簽名）：%s\n", parts[2])   // 第三段：防偽簽名

	// 驗證 Token
	fmt.Println()                         // 空行
	fmt.Println("--- 驗證 Token ---")     // 小標題
	claims, err := validateToken(token)   // 驗證並解析 Token
	if err != nil {                       // 如果驗證失敗
		fmt.Printf("驗證失敗：%v\n", err)   // 顯示錯誤
		return                              // 提早返回
	}

	// 從 Claims 中提取資料
	fmt.Printf("Token 有效！\n")                                      // Token 有效
	fmt.Printf("user_id = %.0f\n", claims["user_id"].(float64))     // 提取 user_id（JSON 數字預設是 float64）
	fmt.Printf("簽發時間（iat）= %.0f\n", claims["iat"].(float64))    // 提取簽發時間
	fmt.Printf("過期時間（exp）= %.0f\n", claims["exp"].(float64))    // 提取過期時間

	// ===== 第三部分：Token 被竄改的情況 =====
	fmt.Println()                                  // 空行
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第三部分：竄改 Token 的檢測")             // 標題
	fmt.Println("==========================================") // 分隔線

	// 模擬攻擊者竄改 Token（改變最後一個字元）
	tamperedToken := token[:len(token)-1] + "X"         // 竄改簽名的最後一個字元
	fmt.Printf("原始 Token 結尾：...%s\n", token[len(token)-10:])       // 顯示原始結尾
	fmt.Printf("竄改 Token 結尾：...%s\n", tamperedToken[len(tamperedToken)-10:]) // 顯示竄改結尾

	_, err = validateToken(tamperedToken)   // 嘗試驗證被竄改的 Token
	if err != nil {                         // 驗證失敗（預期行為）
		fmt.Printf("竄改被檢測到！錯誤：%v\n", err) // 顯示錯誤訊息
	}

	// ===== 第四部分：完整認證流程模擬 =====
	fmt.Println()                                  // 空行
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第四部分：完整認證流程模擬")               // 標題
	fmt.Println("==========================================") // 分隔線

	// 步驟 1：使用者註冊（密碼用 bcrypt 雜湊後儲存）
	fmt.Println()                                // 空行
	fmt.Println("[步驟 1] 使用者註冊")            // 步驟標題
	registerPassword := "alice-password-123"     // 使用者設定的密碼
	hashedPwd, err := hashPassword(registerPassword) // 雜湊密碼
	if err != nil {                                   // 檢查錯誤
		fmt.Printf("註冊失敗：%v\n", err) // 顯示錯誤
		return                             // 提早返回
	}

	// 模擬將使用者存入資料庫
	alice := SimulatedUser{                       // 建立使用者
		ID:             1,                        // 使用者 ID
		Username:       "alice",                  // 使用者名稱
		Email:          "alice@example.com",      // 電子信箱
		HashedPassword: hashedPwd,                // 儲存雜湊後的密碼（不是明文！）
	}
	fmt.Printf("註冊成功！使用者：%s（ID: %d）\n", alice.Username, alice.ID)  // 顯示結果
	fmt.Printf("資料庫中儲存的密碼（雜湊）：%s\n", alice.HashedPassword)       // 顯示雜湊密碼

	// 步驟 2：使用者登入（驗證密碼 → 發 Token）
	fmt.Println()                                     // 空行
	fmt.Println("[步驟 2] 使用者登入")                   // 步驟標題
	loginPassword := "alice-password-123"              // 使用者輸入的密碼
	fmt.Printf("輸入的密碼：%s\n", loginPassword)       // 顯示輸入的密碼

	// 比對密碼
	if !checkPassword(loginPassword, alice.HashedPassword) { // 驗證密碼
		fmt.Println("登入失敗：密碼錯誤")                     // 密碼不符
		return                                                // 提早返回
	}
	fmt.Println("密碼驗證通過！")                              // 密碼正確

	// 密碼正確，發一張 JWT Token
	loginToken, err := generateToken(alice.ID) // 產生 Token
	if err != nil {                            // 檢查錯誤
		fmt.Printf("Token 產生失敗：%v\n", err) // 顯示錯誤
		return                                   // 提早返回
	}
	fmt.Printf("發給使用者的 Token：%s\n", loginToken) // 顯示 Token

	// 步驟 3：使用者帶著 Token 存取受保護的資源
	fmt.Println()                                        // 空行
	fmt.Println("[步驟 3] 存取受保護的資源")                // 步驟標題

	// 模擬 HTTP 請求的 Authorization header
	authHeader := "Bearer " + loginToken                 // 模擬 Authorization header
	fmt.Printf("Authorization: %s...\n", authHeader[:50]) // 只顯示前 50 個字元

	// 從 header 中提取 Token
	headerParts := strings.SplitN(authHeader, " ", 2)  // 用空格分割成兩部分
	if len(headerParts) != 2 || headerParts[0] != "Bearer" { // 檢查格式
		fmt.Println("格式錯誤！應該是：Bearer <token>")   // 格式錯誤
		return                                             // 提早返回
	}

	extractedToken := headerParts[1]                    // 提取 Token 部分
	loginClaims, err := validateToken(extractedToken)   // 驗證 Token
	if err != nil {                                     // 如果驗證失敗
		fmt.Printf("Token 無效：%v\n", err)              // 顯示錯誤
		return                                            // 提早返回
	}

	// Token 驗證成功，取得使用者資訊
	authenticatedUserID := uint(loginClaims["user_id"].(float64)) // 從 Token 提取 user_id
	fmt.Printf("Token 驗證成功！使用者 ID：%d\n", authenticatedUserID) // 顯示使用者 ID

	// 確認是同一個使用者
	if authenticatedUserID == alice.ID { // 比對使用者 ID
		fmt.Printf("歡迎回來，%s！你已通過認證。\n", alice.Username) // 歡迎訊息
	}

	// 步驟 4：錯誤密碼登入失敗
	fmt.Println()                                          // 空行
	fmt.Println("[步驟 4] 錯誤密碼登入（應該失敗）")          // 步驟標題
	wrongPassword := "wrong-password"                       // 錯誤的密碼
	fmt.Printf("輸入的密碼：%s\n", wrongPassword)            // 顯示輸入
	if checkPassword(wrongPassword, alice.HashedPassword) { // 驗證密碼
		fmt.Println("登入成功（不應該發生！）")                 // 不應該到這裡
	} else {
		fmt.Println("登入失敗：密碼錯誤（正確行為！）")          // 密碼不符，符合預期
	}

	// ===== 總結 =====
	fmt.Println()                                  // 空行
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 總結")                                     // 標題
	fmt.Println("==========================================") // 分隔線
	fmt.Println("1. 永遠不要儲存明文密碼，使用 bcrypt 雜湊")        // 重點 1
	fmt.Println("2. JWT 由 Header.Payload.Signature 三部分組成")  // 重點 2
	fmt.Println("3. JWT 的簽名可以防止 Token 被竄改")               // 重點 3
	fmt.Println("4. 認證流程：註冊→雜湊密碼→登入→發Token→驗Token") // 重點 4
	fmt.Println("5. 密鑰（Secret）要用環境變數，不能寫在程式碼裡")   // 重點 5
}
