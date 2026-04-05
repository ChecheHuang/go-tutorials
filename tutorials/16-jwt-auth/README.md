# 第十五課：JWT 認證

## 學習目標

- 理解 JWT 的結構和運作原理
- 學會 bcrypt 密碼雜湊
- 實作完整的註冊/登入/認證流程
- 理解為什麼 JWT 適合 REST API

## 執行方式

```bash
cd tutorials/15-jwt-auth
go mod init jwt-demo && go mod tidy
go run main.go
```

## 重點筆記

### JWT 的三部分

```
eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoxfQ.xxxxx
└──── Header ────────┘└──── Payload ───────┘└Sig┘

Header:    演算法（HS256）
Payload:   攜帶的資料（user_id、過期時間）
Signature: 用密鑰簽名，確保未被竄改
```

### JWT vs Session

| | JWT | Session |
|---|---|---|
| 儲存位置 | 客戶端 | 伺服器端 |
| 擴展性 | 好（無狀態） | 需要共享 Session Store |
| 撤銷 | 較難（等過期） | 容易（刪除 Session） |
| 適合 | REST API、微服務 | 傳統 Web 應用 |

### bcrypt 密碼安全

```go
// 永遠不要儲存明文密碼！
hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
// hash = "$2a$10$N9qo8uLOickgx2ZMRZoMye..." （每次不同！因為有隨機鹽值）

bcrypt.CompareHashAndPassword(hash, []byte("password"))  // nil = 匹配
bcrypt.CompareHashAndPassword(hash, []byte("wrong"))     // error = 不匹配
```

### 完整認證流程

```
註冊：password → bcrypt 雜湊 → 存入資料庫
登入：password → bcrypt 比對 → 產生 JWT → 回傳 Token
請求：Bearer Token → 驗證簽名 → 提取 user_id → 處理請求
```

## 練習

1. 把 JWT 的有效期改為 1 分鐘，觀察過期後的錯誤訊息
2. 在 Claims 中加入 `role` 欄位（admin/user），實作角色檢查中介層
3. 嘗試篡改 Token 的 Payload，觀察簽名驗證失敗的錯誤
