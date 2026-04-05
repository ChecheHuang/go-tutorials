# 實戰練習題集

> 42 課教你「怎麼做」，這份練習教你「怎麼想」。
> 分三類：系統設計、Production Incident 除錯、Code Review。

---

## 一、系統設計練習（限時 30 分鐘）

每題模擬真實面試，你需要：畫架構圖、選擇技術、討論 trade-off。

### SD-01：短網址服務（Junior）

```
需求：設計一個像 bit.ly 的短網址服務
  - 輸入長網址 → 回傳短網址
  - 訪問短網址 → 302 重導向到原始網址
  - 每天 1000 萬次讀取，1 萬次寫入

你需要回答：
  1. 短碼怎麼產生？（hash？自增 ID + base62？）
  2. 用什麼資料庫？為什麼？
  3. 怎麼處理重複的長網址？
  4. 怎麼加快讀取速度？

提示：想想第 26 課 Redis、第 15 課 Index
```

<details>
<summary>參考答案方向（點擊展開）</summary>

```
架構：
  Client → API Gateway → Go Service → Redis Cache → PostgreSQL

短碼生成：
  方案 A：自增 ID + base62 編碼（簡單，但可以被猜到）
  方案 B：nanoid / UUID 截取（隨機，不可猜測）
  方案 C：原始網址的 MD5 取前 7 碼（可能碰撞）
  → 推薦 A，加上隨機偏移量防止被猜

資料庫：
  PostgreSQL（持久化）+ Redis（快取熱門短網址）
  Index：短碼是 primary key，讀取 O(1)

重複處理：
  方案 A：對長網址建 unique index，重複就回傳已有的短碼
  方案 B：不管重複，每次都產生新的（簡單但浪費空間）

Trade-off：
  一致性 vs 效能：Cache 可能有過期的資料（短網址被刪了但 cache 還在）
  → 設定合理的 TTL（如 1 小時），或用 write-through cache
```
</details>

---

### SD-02：聊天系統（Mid）

```
需求：設計一個即時聊天系統（像 Slack 的簡化版）
  - 1 對 1 聊天
  - 群組聊天（最多 500 人）
  - 離線訊息（上線後能收到）
  - 已讀未讀狀態

你需要回答：
  1. 即時推送用什麼技術？
  2. 訊息存在哪裡？怎麼設計 Schema？
  3. 群組 500 人同時收訊息，怎麼不爆？
  4. 離線訊息怎麼處理？

提示：想想第 27 課 WebSocket、第 32 課 Message Queue、第 26 課 Redis
```

<details>
<summary>參考答案方向</summary>

```
架構：
  Client ←WebSocket→ Chat Service → Message Queue → DB
                         ↓
                    Redis (在線狀態 + 未讀計數)

即時推送：WebSocket（長連線）
  每個使用者連到一個 Chat Service 實例
  多個實例之間用 Redis Pub/Sub 轉發

訊息儲存：
  messages 表：id, conversation_id, sender_id, content, created_at
  conversations 表：id, type(1on1/group), created_at
  participants 表：conversation_id, user_id, last_read_at

  Index: (conversation_id, created_at) 複合索引

群組推送：
  不要 for 迴圈逐個 WebSocket 發送
  → 用 Redis Pub/Sub：publish 到 channel:group:{id}
  → 每個 Chat Service 訂閱相關 channel，只推給自己連線的使用者

離線訊息：
  上線時：SELECT * FROM messages WHERE conversation_id IN (...) AND created_at > last_read_at

已讀：
  Redis HASH：unread:{user_id} → {conversation_id: count}
  讀了就 HDEL
```
</details>

---

### SD-03：搶票系統（Senior）

```
需求：100 萬人同時搶 5000 張演唱會門票
  - 不能超賣
  - 不能讓系統掛掉
  - 付款失敗要回滾
  - 要公平排隊

你需要回答：
  1. 畫出完整架構（至少 5 個元件）
  2. 100 萬 QPS 怎麼降到後端可以處理的量？
  3. 怎麼防超賣？
  4. 付款成功但出票服務 crash 怎麼辦？
  5. Redis crash 排隊名單怎麼不丟？

提示：這題的完整答案就是第 39-42 課的內容
```

<details>
<summary>參考答案方向</summary>

```
六層流量防線（第 41 課）：
  CDN → WAF/Bot → API Gateway → Rate Limiter → Waiting Room → Seat Lock

防超賣（第 42 課）：
  Redis DECRBY 原子操作 + Inventory Token
  先發 5000 token，沒 token 的人不能選座位

付款失敗回滾（第 40 課）：
  Saga Pattern：扣庫存 → 建訂單 → 支付 → 出票
  任何步驟失敗 → 反向補償

Redis crash（第 41 課）：
  WAL：先寫 DB 日誌再寫 Redis
  Sentinel：自動故障轉移到 Replica

完整架構：
  Client → CDN → WAF → Gateway → Rate Limiter
    → Waiting Room (Redis Sorted Set)
    → Token Bucket (5000 tokens)
    → Booking Service (Seat Lock: SET NX EX)
    → Order Service (Saga Orchestrator)
    → Payment Service (gRPC + Circuit Breaker)
    → DB (CQRS: Write Model + Read Model)
    → WebSocket (即時通知)
```
</details>

---

### SD-04：Feed 動態牆（Senior）

```
需求：設計社群媒體的動態牆（像 Instagram/Twitter）
  - 使用者可以發文
  - 追蹤的人發文會出現在你的動態牆
  - 動態牆要依時間排序
  - 有些使用者有 100 萬追蹤者

你需要回答：
  1. Push model vs Pull model？
  2. 大 V（100 萬追蹤者）發文怎麼處理？
  3. 動態牆的排序和分頁怎麼做？
  4. 怎麼處理「你追蹤的人剛發的文」的即時性？

提示：這是 Fan-out 問題，想想第 32 課 Message Queue
```

<details>
<summary>參考答案方向</summary>

```
兩種模型：
  Push（寫擴散）：發文時推送到所有追蹤者的 feed
    ✅ 讀取快（feed 已經組好）
    ❌ 大 V 發文要推 100 萬份（寫入爆炸）

  Pull（讀擴散）：讀 feed 時即時查詢所有追蹤者的文章
    ✅ 寫入快
    ❌ 讀取慢（要查很多人的文章）

業界解法：混合模式
  普通使用者（< 1000 追蹤者）：Push
  大 V（> 10000 追蹤者）：Pull
  讀 feed 時：merge(你的 push feed, 大 V 的最新文章)

技術：
  Push feed：Redis Sorted Set（score = timestamp）
  大 V 文章：即時查 DB + Cache
  分頁：ZREVRANGEBYSCORE + cursor-based pagination

即時性：
  WebSocket 推送「有新文章」通知
  Client 收到後重新拉 feed（不直接推文章內容，避免流量）
```
</details>

---

### SD-05：分散式任務排程器（Staff）

```
需求：設計一個分散式定時任務排程器（像 cron，但跨多台機器）
  - 支援「每 5 分鐘執行一次」「每天早上 9 點」
  - 多台機器同一個任務只能執行一次（不重複）
  - 機器掛了，任務要自動轉移到其他機器
  - 要能看到任務執行歷史和狀態

你需要回答：
  1. 怎麼確保同一個任務不會被兩台機器同時執行？
  2. 機器掛了怎麼偵測？怎麼轉移任務？
  3. 任務執行失敗怎麼重試？
  4. 怎麼擴展到 1000 台機器？
```

<details>
<summary>參考答案方向</summary>

```
避免重複執行：
  分散式鎖（Redis SET NX EX）或 DB 樂觀鎖
  執行前嘗試拿鎖，拿不到就跳過

故障偵測 + 轉移：
  每台機器定期更新心跳（Redis SETEX heartbeat:{machine_id} 30）
  Watcher 偵測心跳過期 → 將該機器的任務重新分配

失敗重試：
  指數退避（第 32 課 Message Queue 概念）
  最多重試 3 次，之後進 Dead Letter Queue
  每次重試都要用 idempotency key 防重複

擴展：
  一致性 Hash（Consistent Hashing）分配任務到機器
  新增/移除機器時，只有少數任務需要重新分配

架構：
  Scheduler Service → Redis (鎖 + 心跳) → Task Queue
  Worker Pool (N 台機器) → 消費 Task Queue
  DB (任務定義 + 執行歷史)
```
</details>

---

## 二、Production Incident 除錯練習

每題模擬一個線上故障，你需要：判斷原因、找出解法、說明如何預防。

### INC-01：API 回應變慢（P95 從 50ms 飆到 3 秒）

```
情境：
  週一早上 9 點，你收到 Grafana Alert：
  「blog-api P95 latency > 2000ms」

  已知資訊：
  - 上週五沒有新部署
  - QPS 沒有明顯增加
  - CPU/Memory 正常
  - 錯誤率沒有上升（都是 200）

  你會怎麼排查？依序列出步驟。

提示：想想第 36 課 pprof、第 35 課 Prometheus、第 28 課 DB 進階
```

<details>
<summary>排查步驟</summary>

```
1. 看 Prometheus 指標
   → 哪個 endpoint 慢了？（http_request_duration_seconds by path）
   → 是所有 endpoint 都慢還是特定的？

2. 如果是特定 endpoint（例如 GET /articles）
   → 看 DB 查詢時間（如果有 DB query duration metric）
   → 用 EXPLAIN ANALYZE 看查詢計畫

3. 常見原因：
   a. DB table 長大了，缺少 index
      → 上週五前資料量小不需要 index，現在超過閾值了
      → 解法：CREATE INDEX

   b. N+1 Query
      → 文章列表每篇都查一次作者
      → 解法：Preload / JOIN

   c. 連線池耗盡
      → DB 連線數達到上限，請求在排隊
      → 解法：增加 pool size 或優化慢查詢

   d. Redis 快取過期
      → 所有快取同時到期（Cache Stampede）
      → 解法：隨機 TTL、singleflight

4. 用 pprof 確認
   → go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
   → 看火焰圖，找到哪個函式佔最多時間

5. 預防措施
   → 加 DB query duration metric
   → 設定慢查詢告警（> 100ms）
   → 定期 EXPLAIN 主要查詢
```
</details>

---

### INC-02：記憶體持續增長（Memory Leak）

```
情境：
  搶票系統上線 3 天後，Kubernetes Pod 記憶體使用量：
  - Day 1: 64MB
  - Day 2: 128MB
  - Day 3: 256MB
  - Day 4: OOMKilled，Pod 重啟

  你要怎麼找到記憶體洩漏的原因？
```

<details>
<summary>排查步驟</summary>

```
1. pprof heap profile
   → go tool pprof http://localhost:8080/debug/pprof/heap
   → top 10：看哪個函式分配最多記憶體

2. goroutine leak 檢查
   → go tool pprof http://localhost:8080/debug/pprof/goroutine
   → 如果 goroutine 數量持續增長 → goroutine leak
   → 常見原因：
     a. channel 沒人讀，goroutine 永遠阻塞
     b. context 沒有 cancel，goroutine 不會結束
     c. HTTP client 沒有設定 timeout

3. 常見記憶體洩漏：
   a. WebSocket 連線沒有正確 Unregister
      → client 斷線但 server 端的 conn 沒移除
      → Hub.clients map 無限增長

   b. slice 只 append 不清理
      → 例如 event log 只加不刪

   c. sync.Pool 用錯
      → Put 進去的物件持有大 buffer

4. 修復後驗證
   → 設定 K8s resource limits（第 38 課）
   → 加 Prometheus metric: go_memstats_alloc_bytes
   → 設定告警：記憶體 > 80% limits
```
</details>

---

### INC-03：搶票系統超賣了 5 張票

```
情境：
  活動只有 5000 張票，但最後賣出了 5005 張。
  系統用 Redis DECRBY 扣庫存，理論上不應該超賣。

  你的排查方向？
```

<details>
<summary>排查步驟</summary>

```
1. 檢查 Redis 操作是否真的原子
   → DECRBY 本身是原子的，但如果程式碼是：
     GET stock → 檢查 > 0 → DECRBY
     這就不是原子操作！GET 和 DECRBY 之間可能有其他請求

2. 正確做法 vs 錯誤做法：
   ❌ 錯誤：
     stock = redis.GET("stock")
     if stock > 0 {
         redis.DECRBY("stock", 1)  // race condition!
     }

   ✅ 正確：
     remaining = redis.DECRBY("stock", 1)
     if remaining < 0 {
         redis.INCRBY("stock", 1)  // 回滾
     }

3. 其他可能原因：
   a. Redis Cluster 腦裂
      → 兩個 master 同時接受寫入
      → 解法：設定 min-replicas-to-write

   b. 支付失敗但庫存沒回滾
      → 扣了庫存 → 支付失敗 → 忘記 INCRBY 回去
      → 解法：Saga 補償（第 40 課）

   c. 重複建訂單
      → 同一個請求被處理兩次（MQ 重送）
      → 解法：idempotency key（第 40 課）

4. 事後修復
   → 對帳：比對 Redis 庫存 vs DB 訂單數量
   → 多出的 5 張：聯絡使用者退款
   → 加上定期對帳的背景程式
```
</details>

---

### INC-04：部署後全部 502 Bad Gateway

```
情境：
  你剛 deploy 了新版本到 Kubernetes，Ingress 全部回傳 502。
  舊版本是正常的。

  你會怎麼處理？
```

<details>
<summary>排查步驟</summary>

```
1. 立即止血
   → kubectl rollout undo deployment/blog-api -n blog
   → 先回滾到上一個版本，恢復服務

2. 排查原因
   → kubectl get pods -n blog
   → Pod 狀態是什麼？Running？CrashLoopBackOff？

   如果 CrashLoopBackOff：
   → kubectl logs deployment/blog-api -n blog --previous
   → 看 crash 的原因

   如果 Running 但 502：
   → kubectl describe pod <pod-name> -n blog
   → 看 Readiness Probe 是否失敗
   → 可能原因：/readyz 路徑改了、port 改了、DB 連不上

3. 常見原因：
   a. 環境變數忘記設
      → 新版需要新的 env var，ConfigMap 沒更新

   b. DB Migration 失敗
      → 新版的 AutoMigrate 加了 NOT NULL 欄位但舊資料沒有值

   c. Health Check 路徑變了
      → 程式碼改了 /healthz → /health 但 K8s YAML 沒更新

   d. 依賴服務掛了
      → 新版依賴 Redis 但 Redis 沒啟動

4. 預防措施
   → Readiness Probe 要配合 Graceful Startup（第 29 課）
   → 部署前在 staging 環境先測
   → CI/CD 加 smoke test（部署後自動打一個 API 確認）
   → 設定 maxUnavailable: 0（零停機部署，第 38 課）
```
</details>

---

### INC-05：凌晨 3 點 Redis 記憶體爆了

```
情境：
  凌晨 3 點收到 PagerDuty：
  「Redis memory usage > 90%, eviction started」
  搶票活動是明天早上 10 點開始，現在 Redis 裡存了 50 萬人的排隊資料。

  你只有 7 小時。怎麼辦？
```

<details>
<summary>處理步驟</summary>

```
1. 立即（10 分鐘內）
   → redis-cli INFO memory：看記憶體用了多少
   → redis-cli DBSIZE：看有多少 key
   → redis-cli --bigkeys：找出佔最多空間的 key

2. 短期止血（1 小時內）
   a. 如果是快取 key 佔太多：
      → 降低 TTL 或清除非關鍵快取
      → 保留排隊資料，清掉文章快取

   b. 如果記憶體真的不夠：
      → 雲平台升級 Redis 實例（AWS ElastiCache 可以線上升級）
      → 或新增 Redis 節點做 Cluster

   c. 設定 maxmemory-policy：
      → allkeys-lru：滿了就淘汰最久沒用的 key
      → 但要確保排隊資料不會被淘汰！
      → 用不同的 Redis 實例分開：cache Redis + queue Redis

3. 根本解決（活動結束後）
   → 排隊資料不要全放 Redis，用 DB 做 WAL（第 40 課）
   → Redis 只存「當前正在處理的批次」
   → 監控加上 Redis memory 告警（< 70% 就警告）

4. 預防
   → capacity planning：50 萬人 × 每人排隊資料約 200 bytes = 100MB
   → Redis 至少要 256MB + buffer
   → 壓測：活動前用假資料跑一次
```
</details>

---

## 三、Code Review 練習

每題給你一段有問題的程式碼，找出所有問題並說明怎麼修。

### CR-01：找出並發 bug

```go
type UserCache struct {
    data map[string]*User
}

func (c *UserCache) Get(id string) *User {
    return c.data[id]
}

func (c *UserCache) Set(id string, user *User) {
    c.data[id] = user
}

// 被多個 goroutine 同時呼叫
func handleRequest(cache *UserCache, userID string) {
    user := cache.Get(userID)
    if user == nil {
        user = fetchUserFromDB(userID)
        cache.Set(userID, user)
    }
    // 使用 user...
}
```

<details>
<summary>問題與修復</summary>

```
問題 1：map 不是 goroutine safe
  → 多個 goroutine 同時讀寫 map 會 panic（fatal error: concurrent map read and map write）
  → 修復：用 sync.RWMutex 或 sync.Map

問題 2：Cache Stampede（快取雪崩）
  → 100 個 goroutine 同時發現 cache miss
  → 100 個都去查 DB
  → 修復：用 singleflight 確保同一個 key 只查一次 DB

問題 3：沒有 TTL
  → cache 只增不減，記憶體會爆
  → 修復：加上過期機制

修復後：
  type UserCache struct {
      mu   sync.RWMutex
      data map[string]*cacheEntry
      sf   singleflight.Group
  }

  type cacheEntry struct {
      user      *User
      expiresAt time.Time
  }

  func (c *UserCache) Get(id string) *User {
      c.mu.RLock()
      defer c.mu.RUnlock()
      entry, ok := c.data[id]
      if !ok || time.Now().After(entry.expiresAt) {
          return nil
      }
      return entry.user
  }

  func (c *UserCache) GetOrFetch(id string) *User {
      if user := c.Get(id); user != nil {
          return user
      }
      // singleflight: 同一個 key 只查一次 DB
      val, _, _ := c.sf.Do(id, func() (any, error) {
          user := fetchUserFromDB(id)
          c.Set(id, user, 5*time.Minute)
          return user, nil
      })
      return val.(*User)
  }
```
</details>

---

### CR-02：找出 SQL Injection

```go
func searchArticles(c *gin.Context) {
    keyword := c.Query("q")

    var articles []Article
    db.Raw("SELECT * FROM articles WHERE title LIKE '%" + keyword + "%'").Scan(&articles)

    c.JSON(200, articles)
}
```

<details>
<summary>問題與修復</summary>

```
問題：SQL Injection
  → keyword = "'; DROP TABLE articles; --"
  → 變成：SELECT * FROM articles WHERE title LIKE '%'; DROP TABLE articles; --%'
  → 整個 articles 表被刪了！

修復 1（參數化查詢）：
  db.Raw("SELECT * FROM articles WHERE title LIKE ?", "%"+keyword+"%").Scan(&articles)

修復 2（用 GORM 方法）：
  db.Where("title LIKE ?", "%"+keyword+"%").Find(&articles)

修復 3（最好）：
  // 額外限制 keyword 長度和字元
  if len(keyword) > 100 {
      c.JSON(400, gin.H{"error": "搜尋字串太長"})
      return
  }
  db.Where("title LIKE ?", "%"+keyword+"%").Find(&articles)

原則：永遠不要用字串拼接 SQL
```
</details>

---

### CR-03：找出 Goroutine Leak

```go
func processOrders(orders []Order) {
    for _, order := range orders {
        go func(o Order) {
            result := callPaymentService(o)
            if result.Success {
                updateOrderStatus(o.ID, "paid")
            }
        }(order)
    }
}

func callPaymentService(o Order) PaymentResult {
    resp, err := http.Post(paymentURL, "application/json", toJSON(o))
    if err != nil {
        return PaymentResult{Success: false}
    }
    defer resp.Body.Close()
    // 解析回應...
}
```

<details>
<summary>問題與修復</summary>

```
問題 1：Goroutine 沒有等待完成
  → processOrders 回傳時 goroutine 可能還在跑
  → 如果 main 結束了，goroutine 會被殺掉
  → 修復：用 sync.WaitGroup 或 errgroup

問題 2：HTTP Client 沒有 Timeout
  → 如果 Payment Service 掛了，http.Post 會永遠等下去
  → goroutine 永遠不會結束 = goroutine leak
  → 修復：設定 Timeout

問題 3：無限 goroutine
  → 如果 orders 有 10 萬筆，就開 10 萬個 goroutine
  → 修復：用 Worker Pool 控制並發數

問題 4：沒有 Circuit Breaker
  → Payment Service 掛了，所有請求都在等
  → 修復：gobreaker 包裝

修復後：
  func processOrders(ctx context.Context, orders []Order) error {
      g, ctx := errgroup.WithContext(ctx)
      g.SetLimit(10) // 最多 10 個並發

      client := &http.Client{Timeout: 5 * time.Second}

      for _, order := range orders {
          o := order
          g.Go(func() error {
              return processOneOrder(ctx, client, o)
          })
      }

      return g.Wait()
  }
```
</details>

---

### CR-04：找出效能問題

```go
func getArticlesWithComments(db *gorm.DB) []Article {
    var articles []Article
    db.Find(&articles)

    for i := range articles {
        var comments []Comment
        db.Where("article_id = ?", articles[i].ID).Find(&comments)
        articles[i].Comments = comments
    }

    return articles
}
```

<details>
<summary>問題與修復</summary>

```
問題：N+1 Query
  → 如果有 100 篇文章：
    1 次查 articles
    + 100 次查 comments
    = 101 次 SQL 查詢

  如果有 1000 篇文章 = 1001 次查詢！

修復 1（Preload）：
  db.Preload("Comments").Find(&articles)
  → 只有 2 次查詢：
    SELECT * FROM articles
    SELECT * FROM comments WHERE article_id IN (1,2,3,...100)

修復 2（JOIN，更高效）：
  db.Joins("Comments").Find(&articles)
  → 只有 1 次查詢（但結果集可能更大）

修復 3（加分頁，避免一次載入全部）：
  db.Preload("Comments").
     Offset(offset).
     Limit(pageSize).
     Find(&articles)
```
</details>

---

### CR-05：找出安全漏洞

```go
func deleteArticle(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetUint("user_id") // 從 JWT 取得

    var article Article
    if err := db.First(&article, id).Error; err != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
        return
    }

    db.Delete(&article)
    c.JSON(200, gin.H{"message": "已刪除"})
}
```

<details>
<summary>問題與修復</summary>

```
問題 1：IDOR（Insecure Direct Object Reference）
  → 沒有檢查 article.UserID == userID
  → 任何登入使用者都可以刪除任何人的文章！
  → 修復：加上權限檢查

問題 2：userID 取了但沒用
  → 明顯是忘了加檢查

問題 3：id 沒有做型別驗證
  → id 是字串，直接丟給 GORM
  → 雖然 GORM 會參數化，但最好還是先轉 uint

修復後：
  func deleteArticle(c *gin.Context) {
      id, err := strconv.ParseUint(c.Param("id"), 10, 32)
      if err != nil {
          c.JSON(400, gin.H{"error": "無效的文章 ID"})
          return
      }

      userID := c.GetUint("user_id")

      var article Article
      if err := db.First(&article, id).Error; err != nil {
          c.JSON(404, gin.H{"error": "文章不存在"})
          return
      }

      // 權限檢查！
      if article.UserID != userID {
          c.JSON(403, gin.H{"error": "無權限刪除此文章"})
          return
      }

      db.Delete(&article)
      c.JSON(200, gin.H{"message": "已刪除"})
  }
```
</details>

---

## 練習建議

| 類型 | 建議練法 | 時間 |
|------|---------|------|
| 系統設計 | 限時 30 分鐘，先不看答案，畫完再對照 | 每週 1 題 |
| Incident | 模擬「收到告警」的情境，列出排查步驟 | 每週 1 題 |
| Code Review | 在看到答案前，嘗試找出所有問題 | 每天 1 題 |

**最有價值的練習方式**：找一個朋友互相出題，輪流當面試官。
