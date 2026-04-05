# 第三十七課：CQRS + Event Sourcing

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🔴 資深工程師 | **重點**：了解 CQRS 的取捨，能判斷何時適合使用 |
| 🏢 架構師 | 設計複雜領域的讀寫分離架構，與 DDD 整合 |

> ⚠️ CQRS 增加系統複雜度，**只有在讀寫負載嚴重不對稱或領域複雜時才使用**。

## 核心概念

### CQRS（命令查詢責任分離）

```
Command（寫）:
User → CommandHandler → Aggregate → EventStore → ReadModel

Query（讀）:
User → QueryHandler → ReadModel（最佳化的 View）
```

### Event Sourcing（事件溯源）

```
不存「當前狀態」，改存「事件列表」：

事件 1: ArticleCreated { title: "..." }
事件 2: ArticlePublished { at: 2024-01-01 }
事件 3: ArticleUpdated { title: "新標題" }

當前狀態 = 重播所有事件的結果
```

## 程式碼結構

```go
// Aggregate（聚合根）：領域邏輯
type Article struct { ... }
func (a *Article) Publish() error { ... }  // 產生事件

// Event Store：儲存所有事件（不可變）
type EventStore struct { ... }
func (es *EventStore) Append(events []Event) error { ... }

// Read Model：針對查詢最佳化的視圖
type ArticleReadModel struct { ... }
func (rm *ArticleReadModel) Apply(evt Event) { ... }  // 更新視圖

// Command/Query Service：應用層
cmdService.CreateArticle(...)    // 寫操作
queryService.GetPublished()      // 讀操作
```

## 優缺點

**優點：**
- 讀寫可以獨立擴展
- 完整的審計日誌
- 可以時間旅行（還原任意時間點的狀態）
- 讀側可以有多個不同的 View

**缺點：**
- 最終一致性（讀側可能稍微落後）
- 程式碼量大幅增加
- 除錯相對困難
- 過度設計的風險高

## 執行方式

```bash
go run ./tutorials/37-cqrs
```
