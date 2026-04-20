# Go Tutorials 章節列表與核心考點

## 基礎語法

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 01 | variables-types | 變數與型別 | var vs :=、基本型別、zero value、型別轉換 |
| 02 | control-flow | 流程控制 | if/else、for（Go 唯一的迴圈）、switch、defer |
| 03 | functions | 函式 | 多回傳值、具名回傳值、variadic、first-class function |
| 04 | structs-methods | 結構體與方法 | struct 定義、value vs pointer receiver、嵌入 |
| 05 | pointers | 指標 | &/\*、何時用指標、nil pointer |
| 06 | interfaces | 介面 | 隱式實作、空介面、型別斷言、型別 switch |
| 07 | error-handling | 錯誤處理 | error interface、errors.New、多回傳值慣例 |
| 08 | packages-modules | 套件與模組 | go.mod、import 路徑、exported vs unexported |
| 09 | slices-maps | 切片與映射 | slice append/copy、map CRUD、nil slice vs empty slice |

## Web 開發

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 10 | clean-architecture | 整潔架構 | 四層結構、依賴方向、介面隔離 |
| 11 | http-basics | HTTP 基礎 | net/http、Handler interface、ServeMux |
| 12 | gin-framework | Gin 框架 | Router group、Context、gin.H |
| 13 | json-binding | JSON 綁定 | ShouldBindJSON、struct tag、驗證 |
| 14 | gorm-database | GORM 資料庫 | Model、AutoMigrate、CRUD 方法 |
| 15 | postgresql | PostgreSQL | dsn 連線、GORM with PG、migrations |
| 16 | error-wrapping | 錯誤包裝 | fmt.Errorf %w、errors.Is/As、sentinel error |
| 17 | middleware | 中介層 | gin.HandlerFunc、c.Next()、abort |
| 18 | jwt-auth | JWT 認證 | Claims、簽發/驗證、Bearer token |

## 工程品質

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 19 | testing | 測試 | testing.T、table-driven tests、mock interface |
| 20 | config | 設定管理 | viper、環境變數、設定結構體 |
| 21 | structured-logging | 結構化日誌 | zerolog/zap、JSON 輸出、log level |
| 22 | benchmark | 效能測試 | testing.B、b.N、基準比較 |
| 23 | docker | Docker | Dockerfile、multi-stage build、compose |

## 進階特性

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 24 | generics | 泛型 | type constraint、comparable、any |
| 25 | goroutines | Goroutines | go 關鍵字、WaitGroup、channel、select |
| 26 | redis | Redis | SET/GET/TTL、Session、Cache pattern |
| 27 | websocket | WebSocket | gorilla/ws、讀寫 goroutine、hub 模式 |
| 28 | database-advanced | 進階資料庫 | 事務、鎖、N+1 問題、Preload |
| 29 | clean-arch-advanced | 進階整潔架構 | Repository pattern、DI、use case 組合 |

## 基礎設施

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 30 | grpc | gRPC | proto 定義、Unary vs Stream、生成程式碼 |
| 31 | wire | Wire 依賴注入 | Provider、Injector、wire.go |
| 32 | message-queue | 訊息佇列 | RabbitMQ/Kafka、Producer/Consumer、ACK |
| 33 | cqrs | CQRS | Command vs Query 分離、Event sourcing |
| 34 | cicd | CI/CD | GitHub Actions、pipeline 階段、artifact |

## 可觀測性

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 35 | prometheus | Prometheus | Counter/Gauge/Histogram、/metrics、PromQL |
| 36 | pprof | pprof 效能分析 | CPU profile、heap profile、火焰圖解讀 |
| 37 | opentelemetry | OpenTelemetry | Span、Trace、Context 傳播 |

## 高可用與分散式

| 編號 | 目錄名稱 | 主題 | 核心考點 |
|------|---------|------|---------|
| 38 | kubernetes | Kubernetes | Deployment/Service/ConfigMap、滾動更新 |
| 39 | circuit-breaker | 熔斷器 | Closed/Open/Half-Open 狀態、gobreaker |
| 40 | failure-modeling | 故障建模 | 故障類型、注入測試、降級策略 |
| 41 | high-availability | 高可用 | 多副本、健康檢查、graceful shutdown |
| 42 | distributed-consistency | 分散式一致性 | CAP 定理、最終一致性、Saga pattern |
