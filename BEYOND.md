# 42 課之外：延伸學習指南

> 42 課涵蓋了從 Junior 到 Staff 的核心技能。
> 這份文件列出**課程沒有教但實際工作會用到**的技術，幫助你知道「還有什麼要學」。

## 為什麼這些沒有放進課程？

| 類型 | 原因 |
|------|------|
| 特定工具（Grafana、ELK） | 是「設定」不是「程式碼」，看文件就能上手 |
| 雲平台（AWS、GCP） | 太依賴平台，每家 API 不同 |
| 其他資料庫（MongoDB、ES） | 概念相通，學完 PostgreSQL 後看文件就能遷移 |
| 方法論（TDD、DDD） | 偏理論，需要在實際專案中體會 |

---

## 一、可觀測性完整棧（Observability Stack）

課程教了 Prometheus（指標）、OpenTelemetry（追蹤）、slog（日誌），但生產環境通常需要完整的可觀測性平台。

### Grafana（儀表板）

| 項目 | 說明 |
|------|------|
| 是什麼 | Prometheus 指標的視覺化儀表板 |
| 什麼時候學 | 學完第 35 課 Prometheus 之後 |
| 怎麼學 | `docker run -p 3000:3000 grafana/grafana`，匯入 Prometheus data source |
| 重點 | 建立 Dashboard、設定 Alert Rule、Panel Query 語法（PromQL） |

```
你的 Go 服務 → Prometheus 抓指標 → Grafana 顯示圖表
                                    → Grafana Alert → Slack/Email
```

### ELK Stack（日誌聚合）

| 項目 | 說明 |
|------|------|
| 是什麼 | Elasticsearch + Logstash + Kibana，集中管理所有服務的日誌 |
| 替代方案 | Loki + Grafana（更輕量，推薦 Go 專案用） |
| 什麼時候學 | 學完第 21 課 slog 之後，當你有多個服務需要集中看日誌時 |
| 怎麼學 | 先用 Loki：`docker-compose` 啟動 Loki + Grafana |

```
Go 服務 (slog JSON) → Loki/Logstash → Kibana/Grafana 搜尋日誌
```

### Jaeger / Zipkin（追蹤後端）

| 項目 | 說明 |
|------|------|
| 是什麼 | OpenTelemetry 的追蹤資料儲存和視覺化後端 |
| 什麼時候學 | 學完第 37 課 OpenTelemetry 之後 |
| 怎麼學 | `docker run -p 16686:16686 jaegertracing/all-in-one`，把 OTel exporter 從 stdout 改成 Jaeger |

---

## 二、雲平台（Cloud Provider）

課程用 Docker + Kubernetes 教部署，但實際工作會用雲平台的 managed service。

### AWS（最多職缺要求）

| 服務 | 對應課程概念 | 說明 |
|------|------------|------|
| ECS / EKS | 38 課 Kubernetes | 託管容器服務 |
| RDS | 15 課 PostgreSQL | 託管資料庫 |
| ElastiCache | 26 課 Redis | 託管 Redis |
| SQS / SNS | 32 課 Message Queue | 託管訊息佇列 |
| CloudWatch | 35 課 Prometheus | 監控和日誌 |
| API Gateway | 17 課 Middleware | 託管 API 閘道 |
| ECR | 34 課 CI/CD | Docker Registry |
| IAM | 18 課 JWT | 權限管理（但概念完全不同） |

### GCP

| 服務 | 對應 AWS | 說明 |
|------|---------|------|
| GKE | EKS | 託管 Kubernetes |
| Cloud SQL | RDS | 託管 PostgreSQL |
| Memorystore | ElastiCache | 託管 Redis |
| Cloud Pub/Sub | SQS/SNS | 訊息佇列 |
| Cloud Run | ECS Fargate | 無伺服器容器 |

### 怎麼學

1. 先把 42 課學完，理解底層原理
2. 選一個雲平台（建議 AWS，職缺最多）
3. 用 Free Tier 把部落格部署上去
4. 考 AWS SAA（Solutions Architect Associate）系統化學習

---

## 三、其他資料庫

課程教了 SQLite（開發）+ PostgreSQL（概念），但不同場景用不同 DB。

### MongoDB（文件型資料庫）

| 項目 | 說明 |
|------|------|
| 什麼時候用 | Schema 經常變動、巢狀 JSON 資料、快速原型開發 |
| Go Driver | `go.mongodb.org/mongo-driver` |
| 跟 PostgreSQL 差在哪 | 不需要事先定義 Schema、不支援 JOIN（要用 Aggregation）、不支援 ACID Transaction（新版支援但效能差） |
| GORM 支援嗎 | 不支援，需要用原生 driver |

### Elasticsearch（搜尋引擎）

| 項目 | 說明 |
|------|------|
| 什麼時候用 | 全文搜尋、日誌搜尋、商品搜尋 |
| Go Client | `github.com/elastic/go-elasticsearch` |
| 跟 PostgreSQL 差在哪 | 專門做搜尋最佳化（倒排索引）、不適合當主要資料庫 |
| 常見架構 | PostgreSQL（主 DB）→ 同步到 Elasticsearch（搜尋用） |

### 什麼時候該學

```
PostgreSQL 夠用嗎？
  ├─ 是 → 不需要學其他 DB
  └─ 不夠 → 什麼場景不夠？
       ├─ 全文搜尋效能差 → Elasticsearch
       ├─ Schema 太常變動 → MongoDB
       ├─ 需要時序資料 → InfluxDB / TimescaleDB
       └─ 需要圖關係查詢 → Neo4j
```

---

## 四、方法論

### TDD（Test-Driven Development）

| 項目 | 說明 |
|------|------|
| 是什麼 | 先寫測試，再寫實作，反覆循環（Red → Green → Refactor） |
| 課程教的 | 第 19 課教了怎麼「寫測試」，但不是 TDD 流程 |
| 差別 | TDD 是先寫一個會失敗的測試 → 寫最少的程式碼讓它通過 → 重構 |
| 怎麼練 | 用 Go 做 Kata 練習（如 [exercism.org/tracks/go](https://exercism.org/tracks/go)） |

```
TDD 流程：
  1. 寫測試（Red）   → go test → FAIL
  2. 寫實作（Green） → go test → PASS
  3. 重構（Refactor）→ go test → PASS
  重複
```

### DDD（Domain-Driven Design）

| 項目 | 說明 |
|------|------|
| 是什麼 | 根據業務領域（Domain）來設計軟體架構 |
| 課程教的 | 第 10 課 Clean Architecture 是 DDD 的簡化版 |
| 差別 | DDD 多了 Bounded Context、Aggregate Root、Domain Event、Ubiquitous Language |
| 什麼時候學 | 當你的系統有超過 5 個微服務，且業務邏輯很複雜時 |

```
Clean Architecture（課程教的）：
  domain → usecase → repository → handler

DDD（進階）：
  Bounded Context
    ├─ Aggregate Root（實體 + 業務規則）
    ├─ Domain Service（跨 Aggregate 的業務邏輯）
    ├─ Domain Event（aggregate 之間的溝通）
    └─ Repository（只對 Aggregate Root 操作）
```

### 12-Factor App

| 項目 | 說明 |
|------|------|
| 是什麼 | 12 條雲原生應用程式的設計原則 |
| 課程教的 | 其實已經覆蓋了大部分（Config=環境變數、Logs=stdout、Concurrency=goroutine） |
| 怎麼學 | 讀 [12factor.net](https://12factor.net/)，對照部落格專案檢查每一條 |

---

## 五、進階 Go 生態

### 常用但課程沒教的套件

| 套件 | 用途 | 什麼時候需要 |
|------|------|------------|
| `sqlc` | SQL → Go 程式碼生成（取代 GORM） | 不想用 ORM、追求效能 |
| `ent` | Facebook 的 Go ORM | 需要圖查詢、複雜關聯 |
| `cobra` | CLI 框架 | 要做 CLI 工具 |
| `fx` | Uber 的 DI 框架 | 不想用 Wire、偏好 runtime DI |
| `zap` | Uber 的高效能日誌 | slog 效能不夠時 |
| `validator` | 結構體驗證 | 複雜的輸入驗證規則 |
| `casbin` | 權限管理（RBAC/ABAC） | 複雜的權限模型 |

### 進階主題

| 主題 | 什麼時候需要 |
|------|------------|
| Go Runtime 原理（GMP 調度模型） | 要做效能極致優化時 |
| CGO | 需要呼叫 C 函式庫時 |
| Plugin 系統 | 需要動態載入模組時 |
| WASM | Go 編譯到 WebAssembly |

---

## 學習優先順序建議

學完 42 課後，按照你的職涯方向選擇：

### 如果你要面試後端職位

1. **AWS/GCP 基礎**（最多職缺要求）
2. **系統設計面試練習**（用 42 課學到的模式回答）
3. **Grafana + ELK**（面試常問「怎麼監控」）

### 如果你要做微服務架構

1. **DDD**（Bounded Context 設計）
2. **Kafka**（取代 Channel MQ）
3. **Service Mesh**（Istio / Linkerd）

### 如果你要做高效能系統

1. **sqlc**（取代 GORM，減少反射開銷）
2. **Go Runtime 原理**（GMP、GC 調校）
3. **eBPF / 系統程式設計**

---

## 推薦學習資源

| 資源 | 類型 | 適合 |
|------|------|------|
| [Go by Example](https://gobyexample.com/) | 線上教學 | 快速查語法 |
| [Exercism Go Track](https://exercism.org/tracks/go) | 練習題 | TDD 練習 |
| [System Design Primer](https://github.com/donnemartin/system-design-primer) | GitHub | 面試準備 |
| [Designing Data-Intensive Applications](https://dataintensive.net/) | 書 | 分散式系統聖經 |
| [Go 語言設計與實現](https://draveness.me/golang/) | 部落格 | Go Runtime 深入 |
| [Alex Xu - System Design Interview](https://www.amazon.com/dp/B08CMF2CQF) | 書 | 系統設計面試 |
