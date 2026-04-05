// ==========================================================================
// 第三十三課：CQRS（命令查詢責任分離）
// ==========================================================================
//
// 什麼是 CQRS？
//   Command Query Responsibility Segregation（命令查詢責任分離）
//   核心思想：讀和寫使用不同的模型（甚至不同的資料庫）
//
//   傳統架構：同一個模型同時處理讀和寫
//     Article.Find(), Article.Create(), Article.Update() → 同一個 ORM 模型
//
//   CQRS 架構：
//     Command Side（寫）：處理寫操作，更新「寫資料庫」（規範化）
//     Query Side（讀）：處理讀操作，從「讀資料庫」讀取（非規範化，針對查詢最佳化）
//
// 為什麼需要 CQRS？
//   - 讀寫負載通常不對稱（例如：讀 100 次：寫 1 次）
//   - 讀和寫的最佳化方式不同（讀要快，寫要一致）
//   - 複雜的查詢邏輯污染了領域模型
//
// 搭配 Event Sourcing（事件溯源）：
//   不存最新狀態，而是存所有「事件」（什麼時候發生了什麼）
//   當前狀態 = 重播所有事件的結果
//   優點：完整的審計日誌、可以還原任意時間點的狀態
//
// 執行方式：go run ./tutorials/37-cqrs
// ==========================================================================

package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ==========================================================================
// 1. 事件定義（Event Sourcing 的核心）
// ==========================================================================

// EventType 事件類型
type EventType string

const (
	EventArticleCreated   EventType = "ArticleCreated"
	EventArticlePublished EventType = "ArticlePublished"
	EventArticleUpdated   EventType = "ArticleUpdated"
	EventArticleDeleted   EventType = "ArticleDeleted"
)

// Event 領域事件（不可變，記錄「發生了什麼」）
type Event struct {
	ID          string    // 事件唯一 ID
	Type        EventType // 事件類型
	AggregateID string    // 聚合根 ID（文章 ID）
	Data        any       // 事件資料
	OccurredAt  time.Time // 事件發生時間
	Version     int       // 事件版本（樂觀鎖）
}

// 各種事件的 Data 結構
type ArticleCreatedData struct {
	Title    string
	Content  string
	AuthorID string
}

type ArticlePublishedData struct {
	PublishedAt time.Time
}

type ArticleUpdatedData struct {
	Title   string
	Content string
}

// ==========================================================================
// 2. 聚合根（Aggregate Root）- 領域模型（寫側）
// ==========================================================================

// Article 聚合根（純領域邏輯，不關心持久化）
type Article struct {
	ID        string
	Title     string
	Content   string
	AuthorID  string
	Status    string // draft, published, deleted
	Version   int    // 版本號（每個事件 +1）
	CreatedAt time.Time
	UpdatedAt time.Time

	// 未發布的事件（等待持久化）
	pendingEvents []Event
}

// 業務方法（每個方法產生一個事件）

func NewArticle(id, title, content, authorID string) (*Article, error) {
	if title == "" {
		return nil, errors.New("標題不能為空")
	}

	a := &Article{}
	// 套用事件（所有狀態變更都通過事件）
	a.applyEvent(Event{
		ID:          fmt.Sprintf("evt-%s-1", id),
		Type:        EventArticleCreated,
		AggregateID: id,
		Data:        ArticleCreatedData{Title: title, Content: content, AuthorID: authorID},
		OccurredAt:  time.Now(),
		Version:     1,
	})
	return a, nil
}

func (a *Article) Publish() error {
	if a.Status == "published" {
		return errors.New("文章已經發布")
	}
	if a.Status == "deleted" {
		return errors.New("已刪除的文章無法發布")
	}
	a.applyEvent(Event{
		ID:          fmt.Sprintf("evt-%s-%d", a.ID, a.Version+1),
		Type:        EventArticlePublished,
		AggregateID: a.ID,
		Data:        ArticlePublishedData{PublishedAt: time.Now()},
		OccurredAt:  time.Now(),
		Version:     a.Version + 1,
	})
	return nil
}

func (a *Article) Update(title, content string) error {
	if a.Status == "deleted" {
		return errors.New("已刪除的文章無法修改")
	}
	a.applyEvent(Event{
		ID:          fmt.Sprintf("evt-%s-%d", a.ID, a.Version+1),
		Type:        EventArticleUpdated,
		AggregateID: a.ID,
		Data:        ArticleUpdatedData{Title: title, Content: content},
		OccurredAt:  time.Now(),
		Version:     a.Version + 1,
	})
	return nil
}

// applyEvent 套用事件到聚合根狀態（這是 Event Sourcing 的核心）
func (a *Article) applyEvent(evt Event) {
	switch evt.Type {
	case EventArticleCreated:
		d := evt.Data.(ArticleCreatedData)
		a.ID = evt.AggregateID
		a.Title = d.Title
		a.Content = d.Content
		a.AuthorID = d.AuthorID
		a.Status = "draft"
		a.CreatedAt = evt.OccurredAt
		a.UpdatedAt = evt.OccurredAt

	case EventArticlePublished:
		a.Status = "published"
		a.UpdatedAt = evt.OccurredAt

	case EventArticleUpdated:
		d := evt.Data.(ArticleUpdatedData)
		if d.Title != "" {
			a.Title = d.Title
		}
		if d.Content != "" {
			a.Content = d.Content
		}
		a.UpdatedAt = evt.OccurredAt

	case EventArticleDeleted:
		a.Status = "deleted"
		a.UpdatedAt = evt.OccurredAt
	}

	a.Version = evt.Version
	a.pendingEvents = append(a.pendingEvents, evt)
}

// TakePendingEvents 取出待發布的事件（Repository 保存後清空）
func (a *Article) TakePendingEvents() []Event {
	events := a.pendingEvents
	a.pendingEvents = nil
	return events
}

// ==========================================================================
// 3. Event Store（事件儲存庫）
// ==========================================================================

// EventStore 儲存所有事件（這是 Event Sourcing 的「資料庫」）
type EventStore struct {
	mu     sync.RWMutex
	events map[string][]Event // AggregateID → 事件列表
}

func NewEventStore() *EventStore {
	return &EventStore{events: make(map[string][]Event)}
}

// Append 追加事件（樂觀鎖：檢查版本）
func (es *EventStore) Append(aggregateID string, events []Event, expectedVersion int) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	existing := es.events[aggregateID]
	currentVersion := 0
	if len(existing) > 0 {
		currentVersion = existing[len(existing)-1].Version
	}

	if currentVersion != expectedVersion {
		return fmt.Errorf("版本衝突：預期 %d，實際 %d", expectedVersion, currentVersion)
	}

	es.events[aggregateID] = append(existing, events...)
	return nil
}

// GetEvents 取得聚合根的所有事件
func (es *EventStore) GetEvents(aggregateID string) []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.events[aggregateID]
}

// GetAllEvents 取得所有事件（用於重建 Read Model）
func (es *EventStore) GetAllEvents() []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()
	var all []Event
	for _, events := range es.events {
		all = append(all, events...)
	}
	return all
}

// ==========================================================================
// 4. Read Model（讀側模型，針對查詢最佳化）
// ==========================================================================

// ArticleView 讀側模型（可以包含任何查詢需要的欄位）
type ArticleView struct {
	ID          string
	Title       string
	AuthorID    string
	Status      string
	CreatedAt   time.Time
	PublishedAt *time.Time
	EventCount  int // 事件溯源的額外資訊
}

// ArticleReadModel 讀側模型的儲存（可以是 Elasticsearch、Redis 等）
type ArticleReadModel struct {
	mu       sync.RWMutex
	articles map[string]*ArticleView
}

func NewArticleReadModel() *ArticleReadModel {
	return &ArticleReadModel{articles: make(map[string]*ArticleView)}
}

// Apply 根據事件更新讀側模型（Event Handler）
func (rm *ArticleReadModel) Apply(evt Event) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	switch evt.Type {
	case EventArticleCreated:
		d := evt.Data.(ArticleCreatedData)
		rm.articles[evt.AggregateID] = &ArticleView{
			ID:         evt.AggregateID,
			Title:      d.Title,
			AuthorID:   d.AuthorID,
			Status:     "draft",
			CreatedAt:  evt.OccurredAt,
			EventCount: 1,
		}

	case EventArticlePublished:
		if view, ok := rm.articles[evt.AggregateID]; ok {
			view.Status = "published"
			t := evt.OccurredAt
			view.PublishedAt = &t
			view.EventCount++
		}

	case EventArticleUpdated:
		d := evt.Data.(ArticleUpdatedData)
		if view, ok := rm.articles[evt.AggregateID]; ok {
			if d.Title != "" {
				view.Title = d.Title
			}
			view.EventCount++
		}

	case EventArticleDeleted:
		if view, ok := rm.articles[evt.AggregateID]; ok {
			view.Status = "deleted"
			view.EventCount++
		}
	}
}

// QueryPublished 查詢所有已發布的文章（針對這個查詢最佳化）
func (rm *ArticleReadModel) QueryPublished() []*ArticleView {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	var result []*ArticleView
	for _, v := range rm.articles {
		if v.Status == "published" {
			result = append(result, v)
		}
	}
	return result
}

// QueryByID 按 ID 查詢
func (rm *ArticleReadModel) QueryByID(id string) (*ArticleView, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	v, ok := rm.articles[id]
	return v, ok
}

// ==========================================================================
// 5. Application Services（應用層，協調 Command 和 Query）
// ==========================================================================

// CommandService 處理寫操作（Command Side）
type CommandService struct {
	eventStore *EventStore
	readModel  *ArticleReadModel
}

func NewCommandService(es *EventStore, rm *ArticleReadModel) *CommandService {
	return &CommandService{eventStore: es, readModel: rm}
}

// CreateArticle 建立文章命令
func (s *CommandService) CreateArticle(id, title, content, authorID string) error {
	article, err := NewArticle(id, title, content, authorID)
	if err != nil {
		return err
	}

	events := article.TakePendingEvents()
	if err := s.eventStore.Append(id, events, 0); err != nil {
		return err
	}

	// 同步更新讀側模型（也可以用訊息佇列非同步更新）
	for _, evt := range events {
		s.readModel.Apply(evt)
	}
	return nil
}

// PublishArticle 發布文章命令
func (s *CommandService) PublishArticle(id string) error {
	// 從事件重建聚合根（Event Sourcing 的核心）
	events := s.eventStore.GetEvents(id)
	if len(events) == 0 {
		return fmt.Errorf("文章 %s 不存在", id)
	}

	article := &Article{}
	for _, evt := range events {
		article.applyEvent(evt)
	}

	// 執行業務邏輯
	if err := article.Publish(); err != nil {
		return err
	}

	// 儲存新事件
	newEvents := article.TakePendingEvents()
	if err := s.eventStore.Append(id, newEvents, events[len(events)-1].Version); err != nil {
		return err
	}

	for _, evt := range newEvents {
		s.readModel.Apply(evt)
	}
	return nil
}

// QueryService 處理讀操作（Query Side）
type QueryService struct {
	readModel *ArticleReadModel
}

func NewQueryService(rm *ArticleReadModel) *QueryService {
	return &QueryService{readModel: rm}
}

func (s *QueryService) GetPublishedArticles() []*ArticleView {
	return s.readModel.QueryPublished()
}

func (s *QueryService) GetArticle(id string) (*ArticleView, error) {
	view, ok := s.readModel.QueryByID(id)
	if !ok {
		return nil, fmt.Errorf("文章 %s 不存在", id)
	}
	return view, nil
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第三十七課：CQRS + Event Sourcing")
	fmt.Println("==========================================")

	// 初始化
	eventStore := NewEventStore()
	readModel := NewArticleReadModel()
	cmdService := NewCommandService(eventStore, readModel)
	queryService := NewQueryService(readModel)

	// ──── 1. 執行命令（寫操作）────
	fmt.Println("\n=== 1. 執行命令（Command Side）===")
	fmt.Println()

	articles := []struct{ id, title, content, author string }{
		{"art-1", "CQRS 入門", "命令查詢責任分離...", "alice"},
		{"art-2", "Event Sourcing 實戰", "事件溯源的核心概念...", "bob"},
		{"art-3", "Go 微服務架構", "使用 Go 建構微服務...", "alice"},
	}

	for _, a := range articles {
		if err := cmdService.CreateArticle(a.id, a.title, a.content, a.author); err != nil {
			fmt.Printf("  建立文章失敗: %v\n", err)
		} else {
			fmt.Printf("  ✅ 建立文章: %s\n", a.title)
		}
	}

	// 發布其中兩篇
	for _, id := range []string{"art-1", "art-3"} {
		if err := cmdService.PublishArticle(id); err != nil {
			fmt.Printf("  發布失敗: %v\n", err)
		} else {
			fmt.Printf("  ✅ 發布文章: %s\n", id)
		}
	}

	// 嘗試重複發布
	fmt.Print("\n  嘗試重複發布 art-1: ")
	if err := cmdService.PublishArticle("art-1"); err != nil {
		fmt.Printf("❌ %v（業務規則阻止）\n", err)
	}

	// ──── 2. 查詢（讀操作）────
	fmt.Println("\n=== 2. 查詢（Query Side）===")
	fmt.Println()

	published := queryService.GetPublishedArticles()
	fmt.Printf("已發布的文章（共 %d 篇）：\n", len(published))
	for _, a := range published {
		fmt.Printf("  [%s] %s（作者: %s）\n", a.ID, a.Title, a.AuthorID)
	}

	// ──── 3. Event Sourcing：從事件重建狀態 ────
	fmt.Println("\n=== 3. Event Sourcing：查看事件歷史 ===")
	fmt.Println()

	allEvents := eventStore.GetAllEvents()
	fmt.Printf("事件儲存庫中的所有事件（共 %d 個）：\n", len(allEvents))
	for _, evt := range allEvents {
		fmt.Printf("  [%s] %s → %s\n", evt.OccurredAt.Format("15:04:05.000"), evt.AggregateID, evt.Type)
	}

	// 示範從事件重建聚合根
	fmt.Println()
	fmt.Println("從事件重建 art-1 的狀態：")
	events := eventStore.GetEvents("art-1")
	rebuilt := &Article{}
	for _, evt := range events {
		rebuilt.applyEvent(evt)
		fmt.Printf("  套用事件 %s → 狀態: %s\n", evt.Type, rebuilt.Status)
	}
	fmt.Printf("最終狀態: ID=%s, 標題=%s, 狀態=%s, 版本=%d\n",
		rebuilt.ID, rebuilt.Title, rebuilt.Status, rebuilt.Version)

	// ──── 4. CQRS 架構說明 ────
	fmt.Println("\n=== 4. CQRS 架構總結 ===")
	fmt.Println()
	fmt.Println("Command Side（寫）:")
	fmt.Println("  Command → CommandHandler → Domain Model → Event Store")
	fmt.Println("  ↑ 嚴格的業務規則，一致性保證")
	fmt.Println()
	fmt.Println("Query Side（讀）:")
	fmt.Println("  Query → QueryHandler → Read Model（最佳化的 View）")
	fmt.Println("  ↑ 無業務邏輯，純查詢，可以有不同的資料庫")
	fmt.Println()
	fmt.Println("Event 流向:")
	fmt.Println("  Event Store ──→ 事件總線 ──→ Read Model Projector ──→ Read Model")
	fmt.Println("                              ──→ 訊息佇列 ──→ 其他服務")
	fmt.Println()
	fmt.Println("Event Sourcing 的好處:")
	fmt.Println("  ✅ 完整審計日誌（every change is recorded）")
	fmt.Println("  ✅ 時間旅行（replay events to any point in time）")
	fmt.Println("  ✅ 可以建立新的 Read Model（只需重播所有事件）")
	fmt.Println()
	fmt.Println("⚠️  CQRS 的代價：")
	fmt.Println("  - 最終一致性（讀側可能稍微落後於寫側）")
	fmt.Println("  - 程式碼複雜度增加")
	fmt.Println("  - 適合複雜領域，小專案過度設計")

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
