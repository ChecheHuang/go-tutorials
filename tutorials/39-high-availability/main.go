// ==========================================================================
// 第三十九課：高可用架構（High Availability）
// ==========================================================================
//
// 這一課回答：
//   「100 萬人排隊中 Redis crash，怎麼讓系統不停擺？」
//
// 核心觀念：
//   1. 主從複製（Replication）— 資料不只一份
//   2. Sentinel / Cluster — 自動故障轉移
//   3. Multi-Queue Failover — 佇列不依賴單點
//   4. Health Check + 自動切換 — 偵測故障並切換
//   5. 6 層流量防線 — 讓 100 萬 QPS 變成 100 QPS
//
// 執行方式：
//   go run ./tutorials/39-high-availability/
// ==========================================================================

package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
)

// ==========================================================================
// 1. 主從複製模擬（Replication）
// ==========================================================================

// ReplicatedStore 模擬主從複製的儲存
type ReplicatedStore struct {
	mu       sync.RWMutex
	primary  map[string]string
	replicas []map[string]string // 多個副本
}

func NewReplicatedStore(replicaCount int) *ReplicatedStore {
	replicas := make([]map[string]string, replicaCount)
	for i := range replicas {
		replicas[i] = make(map[string]string)
	}
	return &ReplicatedStore{
		primary:  make(map[string]string),
		replicas: replicas,
	}
}

// Write 寫入 primary 並同步到所有 replica
func (s *ReplicatedStore) Write(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.primary[key] = value
	// 同步到所有副本（真實環境是異步複製）
	for i := range s.replicas {
		s.replicas[i][key] = value
	}
}

// ReadFromPrimary 從 primary 讀取
func (s *ReplicatedStore) ReadFromPrimary(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.primary[key]
	return v, ok
}

// ReadFromReplica 從副本讀取（primary 掛了還能讀）
func (s *ReplicatedStore) ReadFromReplica(key string, replicaIdx int) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if replicaIdx >= len(s.replicas) {
		return "", false
	}
	v, ok := s.replicas[replicaIdx][key]
	return v, ok
}

// SimulatePrimaryCrash 模擬 primary crash
func (s *ReplicatedStore) SimulatePrimaryCrash() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.primary = nil // primary 掛了
	slog.Warn("[複製] Primary crash！")
}

// PromoteReplica 提升副本為新的 primary（Failover）
func (s *ReplicatedStore) PromoteReplica(replicaIdx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if replicaIdx < len(s.replicas) {
		s.primary = s.replicas[replicaIdx]
		slog.Info("[複製] Replica 提升為新 Primary", "replica", replicaIdx)
	}
}

// ==========================================================================
// 2. Sentinel 模擬（自動故障偵測 + 故障轉移）
// ==========================================================================

// ServiceNode 代表一個服務節點
type ServiceNode struct {
	Name    string
	Healthy bool
	Primary bool
}

// Sentinel 哨兵：監控節點健康，自動執行故障轉移
type Sentinel struct {
	mu    sync.Mutex
	nodes []*ServiceNode
}

func NewSentinel(nodes []*ServiceNode) *Sentinel {
	return &Sentinel{nodes: nodes}
}

// HealthCheck 檢查所有節點，回報不健康的
func (s *Sentinel) HealthCheck() []*ServiceNode {
	s.mu.Lock()
	defer s.mu.Unlock()

	var unhealthy []*ServiceNode
	for _, n := range s.nodes {
		if !n.Healthy {
			unhealthy = append(unhealthy, n)
		}
	}
	return unhealthy
}

// Failover 當 primary 掛掉時，自動選出新的 primary
func (s *Sentinel) Failover() *ServiceNode {
	s.mu.Lock()
	defer s.mu.Unlock()

	var currentPrimary *ServiceNode
	var candidate *ServiceNode

	for _, n := range s.nodes {
		if n.Primary && !n.Healthy {
			currentPrimary = n
		}
		if !n.Primary && n.Healthy && candidate == nil {
			candidate = n
		}
	}

	if currentPrimary == nil {
		slog.Info("[Sentinel] Primary 正常，不需要 failover")
		return nil
	}

	if candidate == nil {
		slog.Error("[Sentinel] 沒有可用的 replica 進行 failover！")
		return nil
	}

	currentPrimary.Primary = false
	candidate.Primary = true
	slog.Info("[Sentinel] Failover 完成",
		"old_primary", currentPrimary.Name,
		"new_primary", candidate.Name,
	)
	return candidate
}

// ==========================================================================
// 3. 多層佇列 Failover
// ==========================================================================

// QueueLayer 佇列層（primary + fallback）
type QueueLayer struct {
	mu             sync.Mutex
	primaryQueue   []string
	fallbackQueue  []string
	primaryHealthy bool
}

func NewQueueLayer() *QueueLayer {
	return &QueueLayer{primaryHealthy: true}
}

func (q *QueueLayer) Enqueue(item string) string {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.primaryHealthy {
		q.primaryQueue = append(q.primaryQueue, item)
		return "primary(Redis)"
	}
	// Primary 掛了，自動切換到 fallback
	q.fallbackQueue = append(q.fallbackQueue, item)
	return "fallback(DB)"
}

func (q *QueueLayer) SimulatePrimaryCrash() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.primaryHealthy = false
}

func (q *QueueLayer) RecoverPrimary() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.primaryHealthy = true
	// 把 fallback 的資料搬回 primary
	q.primaryQueue = append(q.primaryQueue, q.fallbackQueue...)
	q.fallbackQueue = nil
}

// ==========================================================================
// 4. 六層流量防線模擬
// ==========================================================================

// TrafficLayer 流量防線的一層
type TrafficLayer struct {
	Name     string
	PassRate float64 // 通過率（0.0 ~ 1.0）
}

// SimulateTrafficLayers 模擬 100 萬請求經過 6 層防線
func SimulateTrafficLayers(initialRequests int, layers []TrafficLayer) {
	requests := initialRequests
	fmt.Printf("\n  初始請求：%d\n\n", requests)

	for _, layer := range layers {
		passed := int(float64(requests) * layer.PassRate)
		blocked := requests - passed
		fmt.Printf("  %-20s │ 進入: %8d │ 通過: %8d │ 擋下: %8d\n",
			layer.Name, requests, passed, blocked)
		requests = passed
	}

	fmt.Printf("\n  最終到達後端的請求：%d（從 %d 降到 %d）\n",
		requests, initialRequests, requests)
}

// ==========================================================================
// 演示
// ==========================================================================

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║         第 39 課：高可用架構（High Availability）        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// === Demo 1: 主從複製 ===
	fmt.Println("\n📦 Demo 1: 主從複製（Primary crash 後從 Replica 讀取）")
	fmt.Println("─────────────────────────────────────────────")

	store := NewReplicatedStore(2) // 1 primary + 2 replicas
	store.Write("queue:event:1", "user-1,user-2,user-3,...100萬人")

	v, _ := store.ReadFromPrimary("queue:event:1")
	fmt.Printf("  Primary 讀取：%s\n", v)

	store.SimulatePrimaryCrash()
	fmt.Println("  💥 Primary crash！")

	v, ok := store.ReadFromReplica("queue:event:1", 0)
	if ok {
		fmt.Printf("  ✅ 從 Replica 0 讀取：%s\n", v)
	}

	store.PromoteReplica(0)
	fmt.Println("  ✅ Replica 0 已提升為新 Primary")

	// === Demo 2: Sentinel 自動 Failover ===
	fmt.Println("\n🛡️ Demo 2: Sentinel 自動故障轉移")
	fmt.Println("─────────────────────────────────────────────")

	nodes := []*ServiceNode{
		{Name: "redis-1", Healthy: true, Primary: true},
		{Name: "redis-2", Healthy: true, Primary: false},
		{Name: "redis-3", Healthy: true, Primary: false},
	}

	sentinel := NewSentinel(nodes)

	fmt.Println("  初始狀態：redis-1(Primary), redis-2(Replica), redis-3(Replica)")

	// 模擬 redis-1 crash
	nodes[0].Healthy = false
	fmt.Println("  💥 redis-1 crash！")

	unhealthy := sentinel.HealthCheck()
	fmt.Printf("  Sentinel 偵測到 %d 個不健康節點\n", len(unhealthy))

	newPrimary := sentinel.Failover()
	if newPrimary != nil {
		fmt.Printf("  ✅ 自動 failover 到 %s\n", newPrimary.Name)
	}

	// === Demo 3: 多層佇列 Failover ===
	fmt.Println("\n📬 Demo 3: Multi-Queue Failover（Redis 掛了自動切 DB）")
	fmt.Println("─────────────────────────────────────────────")

	queue := NewQueueLayer()

	for i := 1; i <= 3; i++ {
		target := queue.Enqueue(fmt.Sprintf("user-%d", i))
		fmt.Printf("  user-%d → %s\n", i, target)
	}

	queue.SimulatePrimaryCrash()
	fmt.Println("  💥 Redis crash！自動切換到 DB fallback")

	for i := 4; i <= 6; i++ {
		target := queue.Enqueue(fmt.Sprintf("user-%d", i))
		fmt.Printf("  user-%d → %s\n", i, target)
	}

	queue.RecoverPrimary()
	fmt.Println("  ✅ Redis 恢復，fallback 資料已搬回 primary")

	for i := 7; i <= 8; i++ {
		target := queue.Enqueue(fmt.Sprintf("user-%d", i))
		fmt.Printf("  user-%d → %s\n", i, target)
	}

	// === Demo 4: 六層流量防線 ===
	fmt.Println("\n🛡️ Demo 4: 六層流量防線（100 萬 → 100 QPS）")
	fmt.Println("─────────────────────────────────────────────")

	layers := []TrafficLayer{
		{Name: "1. CDN 靜態快取", PassRate: 0.3},
		{Name: "2. WAF/Bot 過濾", PassRate: 0.5},
		{Name: "3. API Gateway", PassRate: 0.8},
		{Name: "4. Rate Limiter", PassRate: 0.1},
		{Name: "5. Waiting Room", PassRate: 0.01},
		{Name: "6. Seat Lock", PassRate: 1.0},
	}

	SimulateTrafficLayers(1_000_000, layers)

	// === Demo 5: 腦裂模擬 ===
	fmt.Println("\n🧠 Demo 5: 腦裂（Split Brain）問題")
	fmt.Println("─────────────────────────────────────────────")

	fmt.Println("  情境：網路分區導致兩個 Redis 都認為自己是 Primary")
	fmt.Println()
	fmt.Println("  ┌─────────────┐     網路斷開     ┌─────────────┐")
	fmt.Println("  │ Redis-1     │  ══════╳══════  │ Redis-2     │")
	fmt.Println("  │ (Primary)   │                  │ (也升為      │")
	fmt.Println("  │ 接受寫入 A  │                  │  Primary)   │")
	fmt.Println("  │             │                  │ 接受寫入 B  │")
	fmt.Println("  └─────────────┘                  └─────────────┘")
	fmt.Println()
	fmt.Println("  問題：網路恢復後，A 和 B 的資料衝突了！")
	fmt.Println()
	fmt.Println("  解法：")
	fmt.Println("  1. Quorum（多數決）— 至少 N/2+1 個節點同意才能寫入")
	fmt.Println("  2. Fencing Token — 舊 Primary 的寫入帶過期 token，會被拒絕")
	fmt.Println("  3. Redis Cluster — 用 slot + gossip 協議避免腦裂")

	// 模擬 Quorum 投票
	fmt.Println("\n  Quorum 投票模擬（3 個 Sentinel，需要 2 票同意）：")
	votes := 0
	for i := 0; i < 3; i++ {
		if rand.Float64() > 0.3 { // 70% 機率投同意
			votes++
			fmt.Printf("    Sentinel %d：✅ 同意 failover\n", i+1)
		} else {
			fmt.Printf("    Sentinel %d：❌ 不同意\n", i+1)
		}
	}
	quorum := 3/2 + 1
	if votes >= quorum {
		fmt.Printf("  結果：%d/%d 票，達到 quorum（%d），執行 failover\n", votes, 3, quorum)
	} else {
		fmt.Printf("  結果：%d/%d 票，未達 quorum（%d），不執行 failover（防止腦裂）\n", votes, 3, quorum)
	}

	// === 總結 ===
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════")
	fmt.Println("📌 總結：高可用的五個關鍵")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  1. 複製（Replication）   — 資料永遠不只一份")
	fmt.Println("  2. 故障偵測（Sentinel）  — 自動發現故障節點")
	fmt.Println("  3. 故障轉移（Failover）  — 自動切換到健康節點")
	fmt.Println("  4. 多層降級（Fallback）  — Redis 掛了用 DB 頂")
	fmt.Println("  5. 流量防線（Defense）   — 100萬 QPS 層層過濾到 100")
	fmt.Println()
}
