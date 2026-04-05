# 第十七課：單元測試（Unit Testing）

> **一句話總結**：單元測試就是「自動化的品質檢查員」，寫一次，每次改程式碼後都能自動確認沒有壞掉。

## 你會學到什麼？

- 什麼是單元測試以及為什麼需要它
- Go 內建的 `testing` 套件（不需要安裝！）
- 表格驅動測試（Table-Driven Tests）——Go 社群最推薦的寫法
- `t.Run` 子測試、`t.Errorf`、`t.Fatalf` 的差異
- Mock 測試：如何用「假的資料庫」測試業務邏輯
- `net/http/httptest`：如何測試 HTTP Handler（不需要啟動伺服器！）

## 執行方式

```bash
# 跑所有測試（詳細輸出）
go test -v ./tutorials/17-testing/

# 顯示測試覆蓋率
go test -cover ./tutorials/17-testing/

# 只跑特定測試（名稱包含 TestAdd）
go test -run TestAdd ./tutorials/17-testing/

# 只跑特定子測試（斜線後是子測試名稱）
go test -run TestAdd/正數相加 ./tutorials/17-testing/
```

## 生活比喻：工廠品質檢測

```
沒有單元測試的工廠：
  工人生產零件 → 出貨 → 客戶發現壞掉 → 退貨 → 修理 → 重新出貨
  （發現問題太晚！修復成本很高）

有單元測試的工廠：
  工人生產零件 → 品質檢測機器自動測試 → 發現問題立即修
  （發現問題很早！修復成本很低）

測試程式碼 = 品質檢測機器
每次你修改程式碼，跑一次 go test，立刻知道有沒有壞掉
```

## 測試的命名規則

| 規則 | 範例 | 說明 |
|------|------|------|
| 檔名以 `_test.go` 結尾 | `main_test.go` | Go 只在測試時編譯 `_test.go` 檔案 |
| 函式以 `Test` 開頭（大寫）| `TestAdd` | 小寫 `test` 不會被執行 |
| 參數是 `*testing.T` | `func TestAdd(t *testing.T)` | `t` 是測試控制器 |
| 與被測試檔案同 package | `package main` | 才能存取同 package 的函式 |

## 最基本的測試

```go
// 被測試的函式（main.go）
func Add(a, b int) int {
    return a + b
}

// 測試函式（main_test.go）
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d，期望 5", result)
    }
}
```

執行結果：
```
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
```

## 表格驅動測試（最重要的技巧）

**問題**：如果要測試 Add 的 5 個情況，不用表格就要寫 5 個測試函式。

**解決方案**：把所有測試案例放在一個「表格」裡，用迴圈執行：

```go
func TestAddTableDriven(t *testing.T) {
    tests := []struct {        // 定義表格結構
        name     string        // 案例名稱
        a, b     int           // 輸入
        expected int           // 期望輸出
    }{
        {"正數相加", 2, 3, 5},
        {"負數相加", -1, -2, -3},
        {"零相加", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {  // t.Run 建立子測試
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d，期望 %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

執行結果（`-v` 模式）：
```
=== RUN   TestAddTableDriven
=== RUN   TestAddTableDriven/正數相加
=== RUN   TestAddTableDriven/負數相加
=== RUN   TestAddTableDriven/零相加
--- PASS: TestAddTableDriven (0.00s)
```

## t.Errorf vs t.Fatalf

| 函式 | 行為 | 使用時機 |
|------|------|---------|
| `t.Errorf(...)` | 記錄錯誤，**繼續執行** | 後面的測試是獨立的 |
| `t.Fatalf(...)` | 記錄錯誤，**立即停止** | 後面的測試依賴前面的結果 |
| `t.Fatal(...)` | 同 Fatalf，不格式化 | 簡單的致命錯誤 |
| `t.Log(...)` | 記錄訊息（-v 才顯示）| 除錯用的額外資訊 |

```go
// 典型用法：先確認沒有錯誤（用 Fatalf），再驗證結果（用 Errorf）
user, err := service.GetUser(1)
if err != nil {
    t.Fatalf("不應該有錯誤: %v", err)  // 如果有 error，user 是 nil，下面會 panic
}
if user.Name != "Alice" {
    t.Errorf("名稱不符: %s", user.Name)  // 這裡不會 panic，可以用 Errorf
}
```

## Mock 測試

**問題**：UserService 需要資料庫，但測試時不想真的連接資料庫（慢、不穩定、難控制）。

**解決方案**：用 Mock（假的實作）替換真實的資料庫。

```
真實環境：UserService → 真實資料庫（慢、需要連線）
測試環境：UserService → mockUserRepository（快、記憶體中的 map）
```

**前提**：UserService 必須依賴「介面」而不是具體的資料庫類別：

```go
// ✅ 可測試：依賴介面
type UserService struct {
    repo UserRepository  // 介面！可以換成 Mock
}

// ❌ 不可測試：直接依賴具體實作
type UserService struct {
    db *gorm.DB  // 具體的 GORM DB，無法換成 Mock
}
```

**實作 Mock**：

```go
type mockUserRepository struct {
    users map[int]*User  // 用 map 模擬資料庫
}

func (m *mockUserRepository) FindByID(id int) (*User, error) {
    user, exists := m.users[id]
    if !exists {
        return nil, errors.New("使用者不存在")
    }
    return user, nil
}
```

**使用 Mock**：

```go
func TestUserServiceGetUser(t *testing.T) {
    repo := newMockRepo()           // 建立假資料庫
    service := NewUserService(repo)  // 注入假資料庫

    user, err := service.GetUser(1)
    // ...
}
```

## HTTP Handler 測試（httptest）

**問題**：如何測試 HTTP Handler？難道要真的啟動伺服器？

**解決方案**：`net/http/httptest` 套件提供「假的請求和回應」，不需要啟動伺服器！

```go
func TestGetUserHandler(t *testing.T) {
    handler := NewUserHandler(service)

    // 1. 建立假的請求
    req := httptest.NewRequest(http.MethodGet, "/users?id=1", nil)

    // 2. 建立假的回應記錄器
    w := httptest.NewRecorder()

    // 3. 直接呼叫 Handler（不需要伺服器！）
    handler.GetUserHandler(w, req)

    // 4. 驗證結果
    if w.Code != http.StatusOK {
        t.Errorf("狀態碼 = %d，期望 200", w.Code)
    }

    body, _ := io.ReadAll(w.Result().Body)
    if !strings.Contains(string(body), "Alice") {
        t.Errorf("回應應該包含 Alice")
    }
}
```

**httptest 的兩個核心工具**：

| 工具 | 用途 | 類比 |
|------|------|------|
| `httptest.NewRequest(...)` | 建立假的 HTTP 請求 | 模擬「客戶端發出請求」 |
| `httptest.NewRecorder()` | 建立假的回應記錄器 | 模擬「伺服器端接收回應」|

## 測試覆蓋率（Coverage）

**什麼是覆蓋率？** 你的測試覆蓋了多少比例的程式碼行數。

```bash
go test -cover ./tutorials/17-testing/
# 輸出：coverage: 85.7% of statements
```

**100% 覆蓋率不是目標**，但低覆蓋率代表有程式碼完全沒被測試：
- 60% 以下：測試很少，風險較高
- 80% 以上：比較安全
- 100%：有時候強求反而造成「測試品質」下降（為測試而測試）

## 在部落格專案中的對應

```
internal/usecase/user_usecase_test.go  → 測試業務邏輯（用 Mock Repository）
internal/handler/user_handler_test.go  → 測試 HTTP Handler（用 httptest）
```

測試結構：
```go
// 用 Mock Repository 測試 Usecase
func TestUserUsecase_Login(t *testing.T) {
    mockRepo := &mockUserRepo{
        user: &domain.User{Email: "alice@test.com", Password: hashedPwd},
    }
    usecase := NewUserUsecase(mockRepo)
    // ...
}
```

## 常見問題 FAQ

### Q: 為什麼測試函式的名稱要以 `Test` 大寫開頭？

Go 用命名慣例來區分「測試函式」和「普通函式」。`testing.T` 的 `go test` 工具只會執行大寫 `Test` 開頭的函式。如果你寫 `testAdd`（小寫 t），`go test` 不會執行它。

### Q: 測試跑多快？

非常快！不依賴資料庫或網路的純函式測試，通常在毫秒內完成。Go 的測試是原生支援，不需要外部框架。

### Q: 需要第三方測試框架嗎（如 testify）？

Go 內建的 `testing` 套件已經夠用了。`testify` 提供更方便的斷言函式（如 `assert.Equal`），但不是必要的。Go 社群偏向用標準庫，保持依賴最小化。

### Q: 測試檔案會被編譯到正式版本嗎？

不會！`_test.go` 結尾的檔案只在執行 `go test` 時才會被編譯。`go build` 會忽略它們。

## 練習

1. **新增測試案例**：為 `Divide` 函式新增更多表格測試（如 `Divide(0, 5)`, `Divide(-10, 2)`)
2. **讓 Mock 失敗**：修改 `mockUserRepository.Create`，讓它回傳 `errors.New("資料庫寫入失敗")`，測試 `UserService.CreateUser` 的錯誤處理
3. **HTTP 測試**：為 `CreateUserHandler` 寫一個測試，驗證回應 Body 包含 `"建立成功"`
4. **覆蓋率**：執行 `go test -cover ./tutorials/17-testing/`，查看當前覆蓋率，嘗試新增測試讓覆蓋率提高

## 下一課預告

**第十八課：Docker 部署（Docker Deployment）** —— 學習如何把 Go 應用程式「打包」成 Docker 容器，做到「在我電腦上可以跑，在任何電腦上都可以跑」。
