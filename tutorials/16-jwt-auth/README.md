# 第十六課：JWT 認證（JWT Authentication）

> **一句話總結**：JWT 就是一張「數位門票」，登入成功後伺服器發給你，之後每次請求出示這張票，伺服器就知道你是誰。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：API 認證是後端必備技能 |
| 🔴 資深工程師 | Token 刷新策略、黑名單、多設備登入設計 |

## 你會學到什麼？

- 什麼是 JWT（JSON Web Token）以及它的三段結構
- 為什麼不能直接儲存明文密碼，bcrypt 雜湊是什麼
- 如何用 `github.com/golang-jwt/jwt/v5` 產生和驗證 Token
- 如何用 `golang.org/x/crypto/bcrypt` 安全地處理密碼
- 完整的認證流程：註冊 → 登入 → 發 Token → 驗 Token

## 需要安裝

```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
```

## 執行方式

```bash
go run ./tutorials/16-jwt-auth
```

## 生活比喻：電影院門票

```
沒有 JWT 的世界（每次都要重新驗證）：
  你：「我是 Alice，我要進去看電影」
  工作人員：「請出示身份證和護照...」← 很麻煩！
  （每次都要這樣做）

有 JWT 的世界（一次驗證，之後拿票）：
  你：「我是 Alice，這是我的帳號密碼」
  工作人員：「驗證成功！給你一張票」← 門票 = JWT Token
  （之後進進出出）
  你：「這是我的票」← 出示 Token
  工作人員：「票沒有被偽造，請進！」← 驗證簽名
```

## JWT 的三段結構

JWT 由三部分組成，用 `.` 分隔：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0MiwiZXhwIjoxNjk5OTk5OTk5fQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
└──────────── Header ──────────────────┘└────────────── Payload ──────────────────────┘└─────── Signature ───────────────────────────────┘
```

| 部分 | 內容 | 說明 |
|------|------|------|
| **Header（標頭）** | `{"alg": "HS256", "typ": "JWT"}` | 使用的演算法 |
| **Payload（負載）** | `{"user_id": 42, "exp": 1699999999}` | 攜帶的資料 |
| **Signature（簽名）** | `HMAC-SHA256(header.payload, 密鑰)` | 防偽防竄改 |

**重要**：Header 和 Payload 只是 Base64 編碼（不是加密），任何人都能解碼看到內容。只有 Signature 是用密鑰生成的，所以不能竄改 Payload（簽名會對不上）。

## 什麼是 bcrypt？為什麼不存明文密碼？

```
危險做法（儲存明文）：
  資料庫：{ username: "alice", password: "mypassword123" }
  如果資料庫被駭 → 所有密碼直接外洩！

安全做法（bcrypt 雜湊）：
  資料庫：{ username: "alice", password: "$2a$10$N9qo8uLOickgx2ZMRZoMye..." }
  如果資料庫被駭 → 駭客拿到的是雜湊值，無法反推原始密碼
```

**bcrypt 的特點**：

```
1. 單向不可逆：打散了就回不去（像把雞蛋打成蛋液）
2. 加入隨機鹽值：同一個密碼，每次雜湊結果都不同
3. 比對時不需要知道原始密碼：只需要用同樣方式計算，看結果是否吻合
```

## 使用的第三方套件說明

### `github.com/golang-jwt/jwt/v5`

這是目前 Go 生態最主流的 JWT 套件（官方推薦）：
- 支援多種簽名演算法（HS256、RS256 等）
- 自動驗證過期時間（`exp`）
- 類型安全的 Claims 設計

```go
// 產生 Token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte("my-secret"))

// 驗證 Token
token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte("my-secret"), nil
})
```

### `golang.org/x/crypto/bcrypt`

這是 Go 官方維護的密碼學套件（`x/` 表示官方擴充套件）：

```go
// 雜湊密碼（Cost 越高越安全但越慢，10 是預設值）
hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

// 驗證密碼（自動處理鹽值，不需要你手動提取）
err := bcrypt.CompareHashAndPassword(hash, []byte("password"))
// err == nil → 密碼正確
// err != nil → 密碼錯誤
```

## 完整認證流程圖

```
┌─────────────────────────────────────────────────────────┐
│ 1. 註冊（Register）                                      │
│                                                          │
│  用戶輸入密碼 "mypassword"                               │
│      ↓ bcrypt.GenerateFromPassword                       │
│  "$2a$10$N9qo8uLO..."（雜湊值）→ 存入資料庫              │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 2. 登入（Login）                                         │
│                                                          │
│  用戶輸入 "mypassword"                                   │
│      ↓ bcrypt.CompareHashAndPassword                     │
│  對比資料庫的雜湊值 → 匹配！                              │
│      ↓ jwt.NewWithClaims                                 │
│  產生 JWT Token → 回傳給用戶                             │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 3. 存取受保護資源                                        │
│                                                          │
│  用戶請求帶上 Authorization: Bearer <token>              │
│      ↓ jwt.Parse                                         │
│  驗證簽名 + 檢查是否過期                                  │
│      ↓ 成功                                              │
│  從 Claims 取出 user_id → 處理請求                       │
└─────────────────────────────────────────────────────────┘
```

## JWT vs Session：什麼時候用哪個？

| 比較點 | JWT（Token） | Session（Cookie） |
|--------|-------------|-----------------|
| 儲存位置 | **客戶端**（存在手機/瀏覽器） | **伺服器端**（存在記憶體/Redis） |
| 伺服器狀態 | **無狀態**（Stateless） | **有狀態**（Stateful） |
| 擴展性 | 好（多台伺服器不需共享） | 需要共享 Session Store |
| 撤銷 Token | 較難（要等到過期） | 容易（直接刪除 Session） |
| 適合場景 | REST API、微服務、手機 App | 傳統網頁、需要即時登出 |

**部落格專案選 JWT 的原因**：REST API 天生無狀態，JWT 符合這個設計。

## 程式碼速查

```go
// 密碼雜湊
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 密碼比對
err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
// err == nil → 密碼正確

// 產生 Token
claims := jwt.MapClaims{
    "user_id": userID,
    "exp":     time.Now().Add(24 * time.Hour).Unix(),  // 24 小時後過期
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(jwtSecret))

// 驗證 Token
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("不支援的演算法")
    }
    return []byte(jwtSecret), nil
})
claims := token.Claims.(jwt.MapClaims)
userID := uint(claims["user_id"].(float64))
```

## 安全注意事項

1. **密鑰必須用環境變數**，不能寫死在程式碼裡：
   ```go
   // ❌ 危險：密鑰暴露在程式碼中
   const jwtSecret = "my-secret-key"

   // ✅ 安全：從環境變數讀取
   jwtSecret := os.Getenv("JWT_SECRET")
   ```

2. **設定合理的過期時間**：Token 越久越危險（被盜後難撤銷）
   - 短期 Token：15 分鐘（存取用）
   - 長期 Refresh Token：30 天（用來換新 Token）

3. **Payload 不要存敏感資訊**：因為 Payload 可以被解碼看到，不要存密碼、信用卡號等

4. **驗證演算法**：使用 `token.Method.(*jwt.SigningMethodHMAC)` 防止演算法替換攻擊

## 在部落格專案中的對應

`internal/middleware/jwt.go`：
```go
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 從 Authorization Header 取得 Token
        // 2. 驗證 Token
        // 3. 把 user_id 存入 Context
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

`internal/usecase/user_usecase.go`：
```go
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
    user, err := u.userRepo.FindByEmail(req.Email)
    // 統一錯誤訊息（不告訴攻擊者是「信箱不存在」還是「密碼錯誤」）
    if err != nil {
        return nil, errors.New("信箱或密碼錯誤")
    }
    if err := bcrypt.CompareHashAndPassword(...); err != nil {
        return nil, errors.New("信箱或密碼錯誤")
    }
    // 產生 JWT Token 並回傳
}
```

## 常見問題 FAQ

### Q: JWT 的 Payload 是加密的嗎？可以存使用者資料嗎？

不是！Payload 只是 Base64 編碼，任何人都能解碼。不要存密碼、信用卡號等敏感資訊。只存最基本的識別資訊（user_id、role）。

### Q: 為什麼 `claims["user_id"]` 取出來是 `float64`？

因為 JWT Payload 是 JSON，JSON 的數字預設解析成 `float64`。所以取出後要做型別轉換：
```go
userID := uint(claims["user_id"].(float64))
```

### Q: Token 過期後怎麼辦？

用戶需要重新登入（取得新 Token）。進階做法是使用 Refresh Token 機制：短期存取 Token（15 分鐘）+ 長期 Refresh Token（30 天），用 Refresh Token 換新的存取 Token，不需要重新輸入密碼。

## 練習

1. **修改過期時間**：把 Token 有效期改為 1 分鐘，等 1 分鐘後再驗證，觀察過期錯誤訊息
2. **加入 role 欄位**：在 Claims 中加入 `"role": "admin"`，取出並印出
3. **竄改測試**：把程式輸出的 Token 中間某個字元改掉，再用 `validateToken` 驗證，觀察錯誤
4. **環境變數**：把 `jwtSecret` 改成從 `os.Getenv("JWT_SECRET")` 讀取，用 `JWT_SECRET=mysecret go run ./tutorials/16-jwt-auth` 執行

## 下一課預告

**第十七課：單元測試（Unit Testing）** —— 學習如何用 Go 內建的 `testing` 套件撰寫自動化測試，包括表格驅動測試、Mock 測試，以及如何測試 HTTP Handler。
