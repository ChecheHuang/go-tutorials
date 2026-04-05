# Go REST API 從零開始教學系列

## 學習路線圖

完成以下 18 課後，你將能完全理解部落格 API 專案的每一行程式碼。

```
你的位置
  │
  ▼
┌──────────────────────────────────────────────────────┐
│   第一階段：Go 語言基礎（建議 1-2 天）                  │
│                                                        │
│   01 變數與型別 → 02 控制流程 → 03 函式                  │
│        → 04 結構體與方法 → 05 指標                       │
│                                                        │
│   完成後：能用 Go 寫基本程式                              │
└────────────────────────┬───────────────────────────────┘
                         ▼
┌──────────────────────────────────────────────────────┐
│   第二階段：Go 進階概念（建議 1-2 天）                    │
│                                                        │
│   06 介面 → 07 錯誤處理 → 08 套件與模組                   │
│        → 09 切片與映射                                   │
│                                                        │
│   完成後：理解 Go 的核心設計哲學                           │
└────────────────────────┬───────────────────────────────┘
                         ▼
┌──────────────────────────────────────────────────────┐
│   ★ 第三階段：架構設計（建議 1 天，最重要！）              │
│                                                        │
│   10 Clean Architecture                                │
│      Domain → Repository → Usecase → Handler            │
│      依賴反轉 → 依賴注入 → 可測試性                       │
│                                                        │
│   完成後：理解部落格專案「為什麼這樣設計」                  │
└────────────────────────┬───────────────────────────────┘
                         ▼
┌──────────────────────────────────────────────────────┐
│   第四階段：Web 開發技術（建議 2-3 天）                    │
│                                                        │
│   11 HTTP 基礎 → 12 Gin 框架 → 13 JSON 與標籤            │
│        → 14 GORM 資料庫                                  │
│                                                        │
│   完成後：能用 Gin + GORM 建立簡單的 API                   │
└────────────────────────┬───────────────────────────────┘
                         ▼
┌──────────────────────────────────────────────────────┐
│   第五階段：進階功能（建議 2-3 天）                        │
│                                                        │
│   15 中介層 → 16 JWT 認證 → 17 單元測試                   │
│        → 18 Docker 部署                                  │
│                                                        │
│   完成後：完全理解部落格 API 專案的每一行程式碼              │
└────────────────────────┬───────────────────────────────┘
                         ▼
                   🎉 閱讀 TUTORIAL.md
                 完整理解部落格 API 專案！
```

---

## 課程總覽

### 第一階段：Go 語言基礎

| # | 中文 | English | 重點內容 | 執行方式 |
|---|------|---------|---------|---------|
| [01](./01-variables-types/) | 變數與型別 | Variables & Types | `var`、`:=`、基本型別、零值、`fmt` | `go run main.go` |
| [02](./02-control-flow/) | 控制流程 | Control Flow | `if`、`for`、`switch`、`for range` | `go run main.go` |
| [03](./03-functions/) | 函式 | Functions | 多重回傳、閉包、`defer`、匿名函式 | `go run main.go` |
| [04](./04-structs-methods/) | 結構體與方法 | Structs & Methods | struct、方法、嵌套、建構函式模式 | `go run main.go` |
| [05](./05-pointers/) | 指標 | Pointers | `&`、`*`、傳值 vs 傳參考、`nil` | `go run main.go` |

### 第二階段：Go 進階概念

| # | 中文 | English | 重點內容 | 執行方式 |
|---|------|---------|---------|---------|
| [06](./06-interfaces/) | 介面 | Interfaces | 隱式實作、多型、依賴注入入門 | `go run main.go` |
| [07](./07-error-handling/) | 錯誤處理 | Error Handling | `error`、自訂錯誤、`errors.Is/As`、`panic/recover` | `go run main.go` |
| [08](./08-packages-modules/) | 套件與模組 | Packages & Modules | `go mod`、可見性規則（大小寫）、`internal/` | 閱讀 README |
| [09](./09-slices-maps/) | 切片與映射 | Slices & Maps | 切片操作、Map CRUD、`filter/map` 模式 | `go run main.go` |

### ★ 第三階段：架構設計（最重要）

| # | 中文 | English | 重點內容 | 執行方式 |
|---|------|---------|---------|---------|
| [10](./10-clean-architecture/) | **架構設計** | **Clean Architecture** | **四層結構、依賴反轉、依賴注入、可替換性** | `go run main.go` |

### 第四階段：Web 開發技術

| # | 中文 | English | 重點內容 | 執行方式 |
|---|------|---------|---------|---------|
| [11](./11-http-basics/) | HTTP 基礎 | HTTP Basics | `net/http`、Handler、狀態碼、請求/回應 | `go run main.go` |
| [12](./12-gin-framework/) | Gin 框架 | Gin Framework | 路由、路由群組、`gin.Context` API | `go run main.go` |
| [13](./13-json-binding/) | JSON 處理 | JSON & Struct Tags | `json`/`binding`/`gorm` 三種標籤 | `go run main.go` |
| [14](./14-gorm-database/) | GORM 資料庫 | GORM Database | CRUD、Preload、分頁、搜尋 | `go run main.go` |

### 第五階段：進階功能

| # | 中文 | English | 重點內容 | 執行方式 |
|---|------|---------|---------|---------|
| [15](./15-middleware/) | 中介層 | Middleware | 洋蔥模型、Logger、Auth、CORS、Recovery | `go run main.go` |
| [16](./16-jwt-auth/) | JWT 認證 | JWT Authentication | bcrypt、JWT 產生/驗證、完整登入流程 | `go run main.go` |
| [17](./17-testing/) | 單元測試 | Unit Testing | 表格驅動測試、Mock、`testing` 套件 | `go test -v` |
| [18](./18-docker/) | Docker 部署 | Docker Deployment | Dockerfile 多階段建置、docker-compose | 閱讀 README |

---

## 每一課與部落格專案的對應關係

```
教學課程                         對應的專案程式碼
──────────────────────────────  ──────────────────────────────────
01 變數與型別                    → 所有檔案中的變數宣告與型別定義
02 控制流程                      → repository/ 中的條件查詢與篩選
03 函式                          → usecase/ 中的業務邏輯函式
04 結構體與方法                   → domain/ 的 Entity、所有層的 struct + method
05 指標                          → Repository 回傳 *User、GORM 的 &user
06 介面                          → domain/ 的 Repository 介面、依賴注入
07 錯誤處理                      → 所有層的 if err != nil 模式
08 套件與模組                    → 整個專案的目錄結構、internal/ vs pkg/
09 切片與映射                    → FindAll 回傳 []Article、測試中的 Mock map

★ 10 架構設計                    → 整個專案的四層結構、main.go 的組裝方式

11 HTTP 基礎                     → 理解 Gin 底下的 net/http 原理
12 Gin 框架                      → handler/router.go 路由設定
13 JSON 處理                     → domain/ 中所有結構體的 struct tags
14 GORM 資料庫                   → repository/ 的所有檔案
15 中介層                        → middleware/ 的所有檔案
16 JWT 認證                      → middleware/jwt.go、usecase/user_usecase.go
17 單元測試                      → *_test.go 測試檔案
18 Docker 部署                   → Dockerfile、docker-compose.yml
```

---

## 學習建議

1. **按順序學習**：每一課都建立在前一課的基礎上
2. **動手執行**：每一課都有可執行的程式碼，一定要自己跑過
3. **做練習**：每課結尾都有練習題，試著自己實作
4. **第 10 課要多讀幾遍**：架構設計是理解整個專案的關鍵
5. **對照專案**：學完每一課後，回去看部落格專案中對應的程式碼
6. **不要急**：如果某一課沒完全理解，重看一次比趕進度更有用
