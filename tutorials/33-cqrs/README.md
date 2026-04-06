# 第三十三課：CQRS + Event Sourcing

> **一句話總結**：CQRS 就像餐廳的點菜單和出餐檢視板 — 廚房看的和客人看的是不同格式，各自最佳化。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🔴 資深工程師 | **重點**：了解 CQRS 的取捨，能判斷何時適合使用 |
| 🏢 架構師 | 設計複雜領域的讀寫分離架構，與 DDD 整合 |

> ⚠️ CQRS 增加系統複雜度，**只有在讀寫負載嚴重不對稱或領域複雜時才使用**。

## 你會學到什麼？

- CQRS 的核心概念：為什麼要分開讀和寫
- Write Model（Command）vs Read Model（Query）的設計差異
- Event Sourcing 基礎：用事件記錄狀態變化
- Projection（投影）：如何從事件建立讀模型
- 什麼時候不該用 CQRS（避免過度設計）

## 執行方式

```bash
go run ./tutorials/33-cqrs
```

---

## 餐廳比喻：為什麼要分開讀和寫？

```
【傳統 CRUD（同一個模型做所有事）】

  客人點餐 ──▶ ┌─────────────────┐ ──▶ 客人看菜單
  客人改單 ──▶ │  同一張表         │ ──▶ 廚房看訂單
  結帳     ──▶ │  orders 表       │ ──▶ 老闆看報表
                └─────────────────┘
  問題：菜單要圖片和描述（讀最佳化），廚房只要菜名和數量（寫最佳化），老闆要統計（聚合最佳化）
  一個表很難同時滿足三種需求


【CQRS（讀寫分離）】

  點餐（寫）:
    客人 ──「我要牛排 1 份」──▶ 點菜單（Command）──▶ 廚房系統（寫資料庫）
                                                        │
                                                   產生事件
                                                        │
                                                        ▼
  查看（讀）:                                     出餐檢視板
    客人 ──「我的餐好了嗎？」──▶ 出餐檢視板（Query）──▶（讀資料庫）
    老闆 ──「今日營業額？」──▶ 報表系統（Query）──▶（讀資料庫）

  寫入用嚴格的業務規則（不能點沒有的菜、不能超過庫存）
  讀取用最佳化的視圖（客人看到的和廚房看到的不同）
```

---

## 核心概念：為什麼要分開讀和寫？

### 傳統 CRUD 的問題

```go
// 傳統做法：同一個 Article struct 處理所有操作
type Article struct {
    ID        uint
    Title     string
    Content   string      // 寫的時候需要
    AuthorID  string
    Status    string
    Tags      []Tag       // 讀的時候需要 JOIN
    Author    User        // 讀的時候需要 JOIN
    ViewCount int         // 只有讀的時候需要
    CreatedAt time.Time
}

// 問題 1：讀需要 JOIN 3 張表（Article + Tag + User），寫只動 1 張表
// 問題 2：讀 100 次 vs 寫 1 次，但用同一個查詢邏輯
// 問題 3：想加「熱門文章排行」→ 又要改 Article struct？
```

### CQRS 的解法

```go
// Write Model（命令側）：只關心業務規則
type Article struct {
    ID       string
    Title    string
    Content  string
    AuthorID string
    Status   string    // draft → published → deleted
    Version  int
}

func (a *Article) Publish() error {
    if a.Status != "draft" {
        return errors.New("只有草稿可以發布")    // 嚴格的業務規則
    }
    a.Status = "published"
    return nil
}

// Read Model（查詢側）：針對特定查詢最佳化
type ArticleView struct {
    ID          string
    Title       string
    AuthorName  string      // 已經 JOIN 好了，不用每次查
    Status      string
    PublishedAt *time.Time
    EventCount  int         // 額外的統計資訊
}

type ArticleListView struct {
    ID          string
    Title       string
    AuthorName  string
    ViewCount   int         // 已經計算好了
    CreatedAt   time.Time
}
```

---

## Write Model（Command）vs Read Model（Query）

```
                    ┌─────────────────────────────────────────────┐
                    │              CQRS 架構                       │
                    │                                             │
  ┌─────────┐      │  ┌───────────┐      ┌──────────────┐       │
  │ 使用者   │──寫──▶│ │ Command   │──▶│ Write Model   │       │
  │         │      │  │ Handler   │   │ (Domain Model) │       │
  │         │      │  └───────────┘   └──────┬─────────┘       │
  │         │      │                          │ 產生事件          │
  │         │      │                          ▼                  │
  │         │      │                   ┌─────────────┐           │
  │         │      │                   │ Event Store │           │
  │         │      │                   └──────┬──────┘           │
  │         │      │                          │ 事件投影          │
  │         │      │                          ▼                  │
  │         │      │  ┌───────────┐    ┌──────────────┐         │
  │         │──讀──▶│ │ Query     │──▶│ Read Model    │         │
  │         │      │  │ Handler   │   │ (Optimized    │         │
  └─────────┘      │  └───────────┘   │  View)        │         │
                    │                  └──────────────┘          │
                    └─────────────────────────────────────────────┘
```

| | Write Model（命令側） | Read Model（查詢側） |
|--|----------------------|---------------------|
| **目的** | 保證業務規則正確 | 快速回應查詢 |
| **資料結構** | 正規化（3NF） | 反正規化（冗餘、預計算） |
| **資料庫** | PostgreSQL（ACID 保證） | Elasticsearch / Redis / 專用 View 表 |
| **操作** | Create / Update / Delete | Read only |
| **一致性** | 強一致性 | 最終一致性（可能延遲幾毫秒） |
| **擴展** | 垂直擴展（寫入較難水平擴展） | 水平擴展（加 Read Replica） |

---

## Event Sourcing 基礎

Event Sourcing 不存「當前狀態」，改存「所有發生過的事件」。

### 傳統 CRUD vs Event Sourcing

```
【傳統 CRUD】
  資料庫只存最新狀態：
  ┌─────┬──────────────┬────────────┐
  │ ID  │ Title        │ Status     │
  ├─────┼──────────────┼────────────┤
  │ 1   │ 新標題       │ published  │  ← 只有最新狀態，看不到歷史
  └─────┴──────────────┴────────────┘

  問題：「這篇文章原本的標題是什麼？」→ 不知道
  問題：「誰在什麼時候改的？」→ 不知道（除非額外做 audit log）


【Event Sourcing】
  資料庫存所有事件（不可變）：
  ┌──────┬────────────────────┬───────────────────────────┬─────────────┐
  │ 版本 │ 事件類型            │ 資料                       │ 時間         │
  ├──────┼────────────────────┼───────────────────────────┼─────────────┤
  │ 1    │ ArticleCreated     │ {title: "原始標題"}        │ 2024-01-01  │
  │ 2    │ ArticlePublished   │ {}                        │ 2024-01-02  │
  │ 3    │ ArticleUpdated     │ {title: "新標題"}          │ 2024-01-03  │
  └──────┴────────────────────┴───────────────────────────┴─────────────┘

  當前狀態 = 重播所有事件
  版本 1 後的狀態: {title: "原始標題", status: "draft"}
  版本 2 後的狀態: {title: "原始標題", status: "published"}
  版本 3 後的狀態: {title: "新標題", status: "published"}
```

### Event Sourcing 的核心程式碼

```go
// 1. 事件是不可變的（只能追加，不能修改或刪除）
type Event struct {
    ID          string
    Type        EventType       // ArticleCreated, ArticlePublished, ...
    AggregateID string          // 哪個聚合根的事件
    Data        any             // 事件資料
    Version     int             // 版本號（樂觀鎖）
    OccurredAt  time.Time
}

// 2. 聚合根透過「套用事件」來改變狀態
func (a *Article) applyEvent(evt Event) {
    switch evt.Type {
    case EventArticleCreated:
        d := evt.Data.(ArticleCreatedData)
        a.Title = d.Title
        a.Status = "draft"
    case EventArticlePublished:
        a.Status = "published"
    }
    a.Version = evt.Version
}

// 3. 從事件重建聚合根（每次都從頭重播）
func LoadArticle(events []Event) *Article {
    article := &Article{}
    for _, evt := range events {
        article.applyEvent(evt)
    }
    return article
}
```

### 什麼時候用 Event Sourcing？

```
✅ 適合：
  - 需要完整審計日誌（金融、醫療、法規要求）
  - 需要時間旅行（回到任意時間點看狀態）
  - 複雜的領域邏輯（DDD + Event Sourcing）
  - 多個讀模型需要從同一事件源建立

❌ 不適合：
  - 簡單 CRUD（部落格、Todo App）
  - 事件量極大且不需要歷史（IoT 資料）
  - 團隊不熟悉 Event Sourcing
```

---

## Projection（投影）：如何建立讀模型

Projection 是把事件「投影」到讀模型的過程。

```
Event Store:
  [ArticleCreated] → [ArticlePublished] → [ArticleUpdated]
         │                   │                    │
         ▼                   ▼                    ▼
  ┌────────────────────────────────────────────────────┐
  │                  Projector                          │
  │  switch evt.Type {                                 │
  │  case Created:  INSERT INTO article_views ...      │
  │  case Published: UPDATE article_views SET status.. │
  │  case Updated:  UPDATE article_views SET title...  │
  │  }                                                 │
  └────────────────────────────────────────────────────┘
         │
         ▼
  Read Model (article_views 表):
  ┌─────┬──────────┬───────────┬────────────┐
  │ ID  │ Title    │ Author    │ Status     │
  ├─────┼──────────┼───────────┼────────────┤
  │ 1   │ 新標題   │ alice     │ published  │
  └─────┴──────────┴───────────┴────────────┘
```

### 投影策略

| 策略 | 做法 | 延遲 | 一致性 |
|------|------|------|--------|
| **同步投影** | Command 處理後立即更新 Read Model | 無延遲 | 強一致 |
| **非同步投影** | 事件發到 MQ，Projector 非同步更新 | 毫秒~秒 | 最終一致 |
| **追趕式投影** | 定期從 Event Store 讀取新事件 | 秒~分鐘 | 最終一致 |

```go
// 同步投影（本課的做法）
func (s *CommandService) CreateArticle(...) error {
    // 1. 儲存事件到 Event Store
    s.eventStore.Append(id, events, 0)

    // 2. 立即更新 Read Model（同步）
    for _, evt := range events {
        s.readModel.Apply(evt)    // ← 同步投影
    }
    return nil
}

// 非同步投影（搶票系統的做法）
func (s *CommandService) CreateArticle(...) error {
    s.eventStore.Append(id, events, 0)

    // 發到 MQ，Projector 非同步處理
    for _, evt := range events {
        s.broker.Publish("article.events", evt)    // ← 非同步
    }
    return nil
}
```

---

## 什麼時候不要用 CQRS？

| 情境 | 為什麼不用 |
|------|-----------|
| **簡單 CRUD 應用** | 讀寫邏輯相同，分開只會增加程式碼量 |
| **小團隊 / 小專案** | 維護兩個模型的成本 > 收益 |
| **強一致性要求** | CQRS 天生是最終一致，如果不能容忍延遲就不適合 |
| **讀寫比例接近** | CQRS 的價值在於讀寫不對稱（讀 >> 寫） |
| **沒有複雜查詢** | 如果 SELECT * FROM articles WHERE id = ? 就夠了，不需要 CQRS |

### CQRS 的成本

```
傳統 CRUD：
  1 個 Model + 1 個 Repository + 1 個 Handler
  → 總共 3 個檔案

CQRS：
  Write Model + Read Model
  + Command Handler + Query Handler
  + Event Store + Projector
  + Event 定義
  → 總共 7+ 個檔案

如果你的 CRUD 夠用，不要過度設計！
```

---

## 與搶票系統的連結

搶票系統用了**簡化版的 CQRS**：`Order`（Write Model）和 `OrderView`（Read Model）。

### Write Model：`domain/order.go`

```go
// Order 是寫入模型 — 嚴格的業務規則
type Order struct {
    ID        uint        `gorm:"primaryKey"`
    EventID   uint        `gorm:"index;not null"`
    UserID    string      `gorm:"index;not null"`
    Quantity  int         `gorm:"not null;default:1"`
    Amount    float64     `gorm:"not null"`
    Status    OrderStatus `gorm:"size:20;not null;default:pending"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Write Repository：只有寫入操作
type OrderWriteRepository interface {
    Create(order *Order) error
    UpdateStatus(id uint, status OrderStatus) error
}
```

### Read Model：`domain/order.go` 的 `OrderView`

```go
// OrderView 是讀取模型 — 已經 JOIN 好的視圖
type OrderView struct {
    ID        uint        `json:"id"`
    EventID   uint        `json:"event_id"`
    EventName string      `json:"event_name"`    // ← 已經 JOIN 好，不用每次查
    UserID    string      `json:"user_id"`
    Quantity  int         `json:"quantity"`
    Amount    float64     `json:"amount"`
    Status    OrderStatus `json:"status"`
    CreatedAt time.Time   `json:"created_at"`
}

// Read Repository：只有查詢操作（JOIN orders + events）
type OrderReadRepository interface {
    FindByID(id uint) (*OrderView, error)
    FindByUserID(userID string) ([]OrderView, error)
    FindByEventID(eventID uint) ([]OrderView, error)
}
```

### 為什麼搶票系統要這樣設計？

```
寫入時：
  搶票 → 只寫 orders 表（快！不需要 JOIN）
  → OrderWriteRepository.Create(order)

讀取時：
  查訂單 → JOIN orders + events（需要活動名稱）
  → OrderReadRepository.FindByUserID(userID)
  → 回傳 OrderView（已經 JOIN 好的資料）

好處：
  - 寫入時不需要額外查詢 events 表，減少鎖競爭
  - 讀取時用專門的查詢，可以加索引最佳化
  - 可以獨立擴展讀和寫（讀 Replica + 寫 Primary）
```

注意：搶票系統沒有用 Event Sourcing，而是用**傳統的 CQRS**（同一個資料庫，不同的 Repository 介面）。這是最務實的做法。

---

## CQRS 架構全貌

```
  ┌─────────┐         ┌──────────────────┐         ┌────────────────┐
  │ 前端     │──POST──▶│ Command Handler  │────────▶│ Write DB       │
  │         │         │ (驗證、業務規則)  │         │ (PostgreSQL)   │
  │         │         └──────────────────┘         └───────┬────────┘
  │         │                                              │ 事件
  │         │                                              ▼
  │         │                                     ┌─────────────────┐
  │         │                                     │ Message Queue   │
  │         │                                     │ (Kafka/Redis)   │
  │         │                                     └────────┬────────┘
  │         │                                              │
  │         │                                    ┌─────────┼──────────┐
  │         │                                    ▼         ▼          ▼
  │         │         ┌──────────────────┐  Projector  Projector  Projector
  │         │──GET───▶│ Query Handler    │      │         │          │
  │         │         │ (純查詢，無邏輯) │      ▼         ▼          ▼
  └─────────┘         └────────┬─────────┘  Read DB   Search    Analytics
                               │            (Redis)   (ES)      (ClickHouse)
                               │               │
                               └───────────────┘
```

---

## 常見問題 FAQ

### Q: CQRS 一定要搭配 Event Sourcing 嗎？

不一定。CQRS 只是「讀寫分離」，你可以用普通的 DB 表做寫模型，另一張 DB 表做讀模型。Event Sourcing 是進階選項，增加了不少複雜度，只在需要完整審計軌跡或時間旅行查詢時才值得。

### Q: 讀模型多久更新一次？

取決於一致性需求。同步投影是即時的（但慢），非同步投影有延遲（通常毫秒到秒級）。搶票系統用非同步投影，使用者看到的庫存可能有 1-2 秒的延遲，但這在搶票場景是可接受的。

### Q: 部落格系統需要 CQRS 嗎？

**不需要。** 部落格的讀寫比例雖然是讀多寫少，但用 Redis 快取就能解決效能問題。CQRS 適合搶票系統這種讀寫模型差異大、需要獨立擴展的場景。

### Q: Event Store 會不會越來越大？

會。解決方案：1) **Snapshot**——每 N 個事件存一個快照，重建時從快照開始（本課練習 4），2) **歸檔**——把舊事件移到冷儲存，3) **壓縮**——合併相關事件。

### Q: 搶票系統的 Order vs OrderView 為什麼要分開？

`Order`（寫模型）包含完整的業務邏輯（狀態機、驗證），欄位設計為方便更新。`OrderView`（讀模型）包含展示需要的所有資訊（含使用者名稱、活動名稱），欄位設計為方便查詢，避免 JOIN。

---

## 練習題

### 練習 1：新增 DeleteArticle 命令
在本課的 `CommandService` 中新增刪除功能：
- 新增 `EventArticleDeleted` 事件類型
- 在 `Article` 聚合根中新增 `Delete()` 方法（業務規則：已刪除的不能再刪）
- 在 `ArticleReadModel.Apply` 中處理刪除事件
- 在 `QueryService` 中確認刪除的文章不會出現在查詢結果中

### 練習 2：實作第二個 Read Model
新增一個 `AuthorStatsReadModel`：
- 統計每個作者有幾篇文章、幾篇已發布
- 每個事件都更新這個統計
- 新增 `QueryService.GetAuthorStats(authorID)` 方法
- 體會：同一個 Event Store，可以建立多個不同用途的 Read Model

### 練習 3：實作非同步投影
將目前的同步投影改為非同步：
- 建立一個 `EventBus`（用 Go channel）
- `CommandService` 儲存事件後，發到 `EventBus`
- 另一個 goroutine 監聽 `EventBus`，更新 Read Model
- 觀察：Command 回傳成功後，Query 可能還沒更新（最終一致性）

### 練習 4：實作 Snapshot 最佳化
Event Sourcing 的問題：事件越多，重建聚合根越慢。
- 每 10 個事件，儲存一個 Snapshot（當前狀態的快照）
- 重建時：從最近的 Snapshot 開始，只重播之後的事件
- 比較有 Snapshot 和沒有 Snapshot 的效能差異

### 練習 5：版本衝突處理
本課的 `EventStore.Append` 已經有樂觀鎖（檢查 expectedVersion）。
- 模擬兩個 goroutine 同時修改同一篇文章
- 觀察版本衝突的錯誤
- 實作重試機制：衝突時重新讀取事件、重建聚合根、再次嘗試
- 思考：為什麼用樂觀鎖而不是悲觀鎖？
