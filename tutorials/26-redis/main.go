// ==========================================================================
// 第二十六課：Redis 快取
// ==========================================================================
//
// 什麼是 Redis？
//   Redis（Remote Dictionary Server）是一個「超快速的記憶體資料庫」
//   它把資料存在記憶體（RAM）裡，所以速度非常快：
//
//   一般資料庫（SQLite/MySQL）：查詢耗時約 1-100 毫秒
//   Redis：查詢耗時約 0.1 毫秒（100 倍以上的差距！）
//
// 生活比喻：冰箱 vs 超市
//
//   資料庫 = 超市（Supermarket）
//     ✅ 東西很多、永久保存
//     ❌ 要開車去、花時間找
//
//   Redis = 家裡的冰箱（Fridge）
//     ✅ 拿東西超快、開門就有
//     ❌ 空間有限、食物會過期
//
//   快取策略：
//     需要東西 → 先看冰箱有沒有
//     冰箱有  → 直接拿（超快！）
//     冰箱沒有 → 去超市買，買回來放冰箱，下次就快了
//
// 什麼是 github.com/redis/go-redis/v9？
//   這是 Go 語言連接 Redis 最主流的客戶端套件
//   v9 是第 9 版，支援 Redis 7.x 的所有功能
//   提供類型安全的 API，每個操作都有對應的 Go 方法
//
// 執行前置需求（三選一）：
//   A. Docker：docker-compose -f docker-compose.dev.yml up -d
//   B. Upstash：https://upstash.com 免費申請，設定 REDIS_URL 環境變數
//   C. 本地安裝：brew install redis（Mac）/ apt install redis-server（Linux）
//
// 執行方式：go run ./tutorials/20-redis
// ==========================================================================

package main // 宣告這是 main 套件

import ( // 匯入所有需要的套件
	"context"          // 標準庫：Context，Redis 操作都需要它
	"encoding/json"    // 標準庫：JSON 序列化/反序列化
	"fmt"              // 標準庫：格式化輸出
	"os"               // 標準庫：讀取環境變數
	"time"             // 標準庫：時間相關功能

	"github.com/redis/go-redis/v9" // 第三方套件：Redis 客戶端，提供所有 Redis 操作的 Go API
)

// ==========================================================================
// 全域變數
// ==========================================================================

// ctx 是所有 Redis 操作都需要的 Context
// 在這個教學中我們用一個簡單的 Background Context
// 在真實專案中，應該用 HTTP 請求的 Context（帶有超時和取消功能）
var ctx = context.Background() // 建立一個永不取消的 root Context

// ==========================================================================
// 資料結構定義
// ==========================================================================

// CachedUser 模擬從資料庫取出並放入 Redis 快取的使用者資料
type CachedUser struct { // 定義快取的使用者結構
	ID       int    `json:"id"`        // 使用者 ID（json tag 讓 JSON 欄位名稱是小寫）
	Username string `json:"username"`  // 使用者名稱
	Email    string `json:"email"`     // 電子信箱
}

// ==========================================================================
// 連接 Redis
// ==========================================================================

// connectRedis 建立並回傳 Redis 客戶端連線
func connectRedis() *redis.Client { // 回傳 redis.Client 指標
	// 從環境變數讀取 Redis 連線位址
	// 如果沒有設定，預設用本地的 Redis（Docker 或本地安裝）
	redisURL := os.Getenv("REDIS_URL") // 讀取環境變數
	if redisURL == "" {                // 如果沒有設定
		redisURL = "redis://localhost:6379" // 使用預設的本地 Redis
	}

	// redis.ParseURL 把連線字串解析成 Options 結構
	// 支援格式：redis://localhost:6379 或 rediss://user:pass@host:port（SSL）
	opt, err := redis.ParseURL(redisURL) // 解析連線字串
	if err != nil {                       // 如果格式錯誤
		panic(fmt.Sprintf("Redis URL 格式錯誤: %v", err)) // 直接 panic（設定錯誤是不可恢復的）
	}

	client := redis.NewClient(opt) // 建立 Redis 客戶端（還沒有真正連線）

	// Ping 測試連線是否成功
	// Ping 就像打電話確認對方有沒有接
	if err := client.Ping(ctx).Err(); err != nil { // 發送 PING 指令
		panic(fmt.Sprintf("無法連接 Redis，請確認 Redis 已啟動：%v\n"+
			"提示：執行 docker-compose -f docker-compose.dev.yml up -d", err)) // 提示如何啟動
	}

	fmt.Printf("成功連接 Redis：%s\n", redisURL) // 連線成功的提示
	return client                               // 回傳客戶端
}

// ==========================================================================
// 1. Redis 基本操作
// ==========================================================================
//
// Redis 最基本的操作：
//   SET key value [EX seconds]  → 儲存一個值（可設定過期時間）
//   GET key                     → 取得一個值
//   DEL key                     → 刪除一個值
//   EXISTS key                  → 確認 key 是否存在
//   TTL key                     → 查看還有幾秒過期
//   EXPIRE key seconds          → 設定過期時間

// demonstrateBasicOps 示範 Redis 基本操作
func demonstrateBasicOps(client *redis.Client) { // 接受 Redis 客戶端
	fmt.Println("\n=== 1. 基本操作（SET / GET / DEL）===") // 印出標題

	// ---- SET：儲存資料 ----
	// client.Set(ctx, key, value, expiration)
	// expiration = 0 表示永不過期
	// expiration > 0 表示幾秒後自動刪除
	err := client.Set(ctx, "greeting", "你好，Redis！", 5*time.Minute).Err() // 儲存，5 分鐘後過期
	if err != nil { // 如果操作失敗
		fmt.Printf("SET 失敗: %v\n", err) // 印出錯誤
		return                             // 提前返回
	}
	fmt.Println("SET greeting = 你好，Redis！（5 分鐘後過期）") // 印出成功訊息

	// ---- GET：取得資料 ----
	// .Result() 回傳兩個值：資料值 和 錯誤
	// 特別注意：如果 key 不存在，err 是 redis.Nil（不是一般的錯誤）
	val, err := client.Get(ctx, "greeting").Result() // 取得 greeting 的值
	if err == redis.Nil {                             // 如果 key 不存在
		fmt.Println("greeting 不存在") // 印出提示
	} else if err != nil { // 如果是其他錯誤（連線問題等）
		fmt.Printf("GET 失敗: %v\n", err) // 印出錯誤
	} else { // 成功取得
		fmt.Printf("GET greeting = %s\n", val) // 印出取得的值
	}

	// ---- TTL：查看剩餘時間 ----
	// TTL = Time To Live（還能活多久）
	ttl, _ := client.TTL(ctx, "greeting").Result() // 查看 greeting 的剩餘時間
	fmt.Printf("TTL greeting = %v（還有這麼久才過期）\n", ttl) // 印出剩餘時間

	// ---- EXISTS：確認 key 是否存在 ----
	exists, _ := client.Exists(ctx, "greeting").Result()            // 回傳存在的 key 數量
	fmt.Printf("EXISTS greeting = %v（1 表示存在，0 表示不存在）\n", exists) // 印出結果

	// ---- DEL：刪除資料 ----
	client.Del(ctx, "greeting")                     // 刪除 greeting
	val, err = client.Get(ctx, "greeting").Result() // 再次取得
	if err == redis.Nil {                            // 應該是 Nil（已刪除）
		fmt.Println("DEL 之後 GET greeting = (不存在，已被刪除)") // 確認刪除成功
	}

	// ---- 永不過期 vs 有期限 ----
	client.Set(ctx, "permanent", "永久保存", 0)             // 0 = 永不過期
	client.Set(ctx, "temporary", "30 秒後消失", 30*time.Second) // 30 秒後自動刪除
	fmt.Println("\n儲存了兩個 key：permanent（永久）和 temporary（30秒後過期）")

	// ---- 清理：刪除本節建立的 key ----
	client.Del(ctx, "permanent", "temporary") // 一次刪除多個 key
}

// ==========================================================================
// 2. Cache-Aside 模式（最常見的快取策略）
// ==========================================================================
//
// Cache-Aside（旁路快取）流程：
//
//   讀取時：
//     1. 先查 Redis（快）
//     2. 有 → 直接回傳（Cache Hit，快取命中）
//     3. 沒有 → 查資料庫 → 存入 Redis → 回傳（Cache Miss，快取未命中）
//
//   寫入時：
//     1. 更新資料庫
//     2. 讓 Redis 的快取失效（DEL key）
//     3. 下次讀取時會重新從 DB 載入最新資料

// simulateDBQuery 模擬「慢速的資料庫查詢」
// 在真實專案中，這裡是真正的 GORM 查詢
func simulateDBQuery(userID int) *CachedUser { // 接受使用者 ID，回傳使用者資料
	fmt.Printf("  [資料庫] 查詢使用者 ID=%d（模擬 100ms 延遲）\n", userID) // 印出查詢日誌
	time.Sleep(100 * time.Millisecond) // 模擬資料庫查詢耗時（100ms）
	// 模擬資料庫回傳的資料
	return &CachedUser{ // 建立並回傳使用者資料
		ID:       userID,                                          // 使用者 ID
		Username: fmt.Sprintf("user_%d", userID),                  // 使用者名稱
		Email:    fmt.Sprintf("user%d@example.com", userID),       // 電子信箱
	}
}

// getUser 實作 Cache-Aside 模式取得使用者資料
func getUser(client *redis.Client, userID int) (*CachedUser, error) { // 回傳使用者和錯誤
	cacheKey := fmt.Sprintf("user:%d", userID) // 快取的 key 格式：user:1, user:2...

	// 步驟 1：先查 Redis 快取
	cachedData, err := client.Get(ctx, cacheKey).Result() // 嘗試從快取取得

	if err == nil { // Cache Hit：快取有資料！
		fmt.Printf("  [Redis] Cache HIT：找到 user:%d 的快取\n", userID) // 印出命中訊息
		var user CachedUser                                              // 宣告 User 變數
		json.Unmarshal([]byte(cachedData), &user)                        // JSON 反序列化（把字串轉回 struct）
		return &user, nil                                                // 直接回傳快取的資料
	}

	if err != redis.Nil { // 不是「key 不存在」的錯誤，而是連線等其他問題
		return nil, fmt.Errorf("Redis 錯誤: %w", err) // 回傳錯誤
	}

	// 步驟 2：Cache Miss：快取沒有，查資料庫
	fmt.Printf("  [Redis] Cache MISS：user:%d 不在快取，查詢資料庫\n", userID) // 印出未命中訊息
	user := simulateDBQuery(userID)                                          // 查詢資料庫（慢！）

	// 步驟 3：把資料庫結果存入 Redis 快取（下次就快了）
	jsonData, _ := json.Marshal(user)                                    // 把 struct 轉成 JSON 字串
	client.Set(ctx, cacheKey, jsonData, 5*time.Minute)                   // 存入 Redis，5 分鐘後自動過期
	fmt.Printf("  [Redis] 已將 user:%d 存入快取（5 分鐘後過期）\n", userID) // 印出存入訊息

	return user, nil // 回傳從資料庫取得的資料
}

// demonstrateCacheAside 示範 Cache-Aside 模式
func demonstrateCacheAside(client *redis.Client) { // 接受 Redis 客戶端
	fmt.Println("\n=== 2. Cache-Aside 模式（冰箱 vs 超市）===") // 印出標題

	// 清理可能殘留的快取，確保示範效果
	client.Del(ctx, "user:1") // 刪除 user:1 的快取

	// 第一次取得：Cache Miss（快取沒有，去查 DB）
	fmt.Println("\n--- 第一次取得 user:1（冰箱是空的，要去超市）---") // 小標題
	start := time.Now()                                           // 記錄開始時間
	user, _ := getUser(client, 1)                                 // 取得使用者
	fmt.Printf("  取得：%s（耗時 %v）\n", user.Username, time.Since(start)) // 印出結果和耗時

	// 第二次取得：Cache Hit（快取有了，直接拿）
	fmt.Println("\n--- 第二次取得 user:1（冰箱有了，直接拿）---") // 小標題
	start = time.Now()                                        // 重新記錄時間
	user, _ = getUser(client, 1)                              // 再次取得（這次會命中快取）
	fmt.Printf("  取得：%s（耗時 %v，快很多！）\n", user.Username, time.Since(start)) // 印出結果和耗時

	// 更新時讓快取失效
	fmt.Println("\n--- 更新資料後讓快取失效 ---") // 小標題
	client.Del(ctx, "user:1")              // 刪除快取（讓它失效）
	fmt.Println("  已更新資料庫，快取已清除（下次讀取會重新從 DB 載入）") // 說明
}

// ==========================================================================
// 3. Session 儲存
// ==========================================================================
//
// 為什麼要把 Session 存在 Redis？
//   JWT Token：無狀態，伺服器不儲存，難以即時撤銷
//   Session：有狀態，儲存在 Redis，可以即時讓使用者登出
//
//   想像：JWT 是印好的票券（無法立刻作廢），Session 是會員卡系統（可以當場停用）
//
//   Session 的 key 通常是隨機產生的 UUID，存在瀏覽器的 Cookie 裡
//   Redis 儲存：session:{uuid} → { user_id, username, ... }

// SessionData 儲存在 Redis 中的 Session 資料
type SessionData struct { // 定義 Session 資料結構
	UserID   int    `json:"user_id"`  // 使用者 ID
	Username string `json:"username"` // 使用者名稱
	LoginAt  string `json:"login_at"` // 登入時間
}

// demonstrateSession 示範用 Redis 儲存 Session
func demonstrateSession(client *redis.Client) { // 接受 Redis 客戶端
	fmt.Println("\n=== 3. Session 儲存 ===") // 印出標題

	// ---- 登入：建立 Session ----
	fmt.Println("\n--- 使用者登入，建立 Session ---") // 小標題

	// 在真實專案中，sessionID 要用 uuid.New().String() 產生隨機的
	sessionID := "abc123xyz789"            // 模擬的 Session ID（真實應用要用 uuid）
	sessionKey := "session:" + sessionID   // Redis key 格式：session:abc123xyz789

	session := SessionData{ // 建立 Session 資料
		UserID:   42,                      // 使用者 ID
		Username: "alice",                 // 使用者名稱
		LoginAt:  time.Now().Format(time.RFC3339), // 登入時間（標準格式）
	}

	sessionJSON, _ := json.Marshal(session) // 把 struct 轉成 JSON
	// 儲存到 Redis，並設定 30 分鐘後自動過期（Session 超時）
	client.Set(ctx, sessionKey, sessionJSON, 30*time.Minute) // 儲存 Session
	fmt.Printf("  Session 已建立：%s\n", sessionID)          // 印出 Session ID
	fmt.Printf("  使用者：%s（ID: %d）\n", session.Username, session.UserID) // 印出使用者資訊
	fmt.Println("  過期時間：30 分鐘後自動失效") // 說明過期機制

	// ---- 驗證：每次請求時檢查 Session ----
	fmt.Println("\n--- 驗證 Session（每次 API 請求都要做）---") // 小標題

	storedData, err := client.Get(ctx, sessionKey).Result() // 從 Redis 取得 Session
	if err == redis.Nil {                                    // Session 不存在或已過期
		fmt.Println("  Session 不存在或已過期，請重新登入") // 需要重新登入
	} else if err != nil { // 其他錯誤
		fmt.Printf("  Session 驗證錯誤: %v\n", err) // 印出錯誤
	} else { // Session 有效
		var loadedSession SessionData                    // 宣告變數
		json.Unmarshal([]byte(storedData), &loadedSession) // 反序列化
		fmt.Printf("  Session 有效！歡迎回來，%s\n", loadedSession.Username) // 歡迎訊息
	}

	// ---- 每次請求都「續期」：重置過期時間 ----
	// 這樣只要使用者一直在操作，Session 就不會過期
	client.Expire(ctx, sessionKey, 30*time.Minute) // 重置過期時間為 30 分鐘
	fmt.Println("  Session 已續期（30 分鐘重新計時）") // 印出續期訊息

	// ---- 登出：刪除 Session ----
	fmt.Println("\n--- 使用者登出，刪除 Session ---") // 小標題
	client.Del(ctx, sessionKey)               // 刪除 Session（立即生效）
	fmt.Println("  Session 已刪除，使用者已登出") // 印出登出訊息

	// 驗證 Session 已刪除
	_, err = client.Get(ctx, sessionKey).Result() // 嘗試取得已刪除的 Session
	if err == redis.Nil {                          // 確認已不存在
		fmt.Println("  驗證：Session 確實已不存在") // 確認登出成功
	}
}

// ==========================================================================
// 4. Rate Limiting（請求限流）
// ==========================================================================
//
// 什麼是 Rate Limiting？
//   限制每個使用者在一段時間內最多能做幾次操作
//   例如：每分鐘最多 10 次 API 呼叫
//
// 生活比喻：
//   就像自助餐的補菜規定：「每人每 10 分鐘最多取 3 次」
//   服務員在旁邊計數，超過就說「請等一下」
//
// 實作原理（INCR + EXPIRE 組合）：
//   1. 每次請求：INCR rate_limit:{user_id}:{time_window}（計數加 1）
//   2. 第一次請求：設定過期時間（EXPIRE，計時開始）
//   3. 每次請求：檢查計數是否超過上限
//   4. 過期後：計數器自動歸零，新的時間窗口開始

// checkRateLimit 檢查使用者是否超過請求限制
// 回傳：是否允許請求、目前計數、錯誤
func checkRateLimit(client *redis.Client, userID int, maxRequests int) (bool, int64, error) {
	// key 格式：rate:{user_id}:{當前分鐘}（每分鐘一個新 key）
	// time.Now().Minute() 取得當前的分鐘數（0-59）
	// 這樣每分鐘都會有一個新的計數器
	key := fmt.Sprintf("rate:%d:%d", userID, time.Now().Minute()) // 建立限流 key

	// INCR：計數器加 1，並取得加 1 後的值
	// 如果 key 不存在，INCR 會自動建立並設為 0，然後加 1 = 1
	count, err := client.Incr(ctx, key).Result() // 計數 +1
	if err != nil {                               // 如果操作失敗
		return false, 0, err // 回傳錯誤
	}

	if count == 1 { // 如果是這個時間窗口的第一次請求
		// 設定 1 分鐘後過期（計時開始）
		// 注意：只有 count==1 時才設定，否則每次請求都重置計時器
		client.Expire(ctx, key, time.Minute) // 設定 60 秒後過期
	}

	allowed := count <= int64(maxRequests) // 如果計數沒超過上限，就允許
	return allowed, count, nil             // 回傳結果
}

// demonstrateRateLimit 示範 Rate Limiting
func demonstrateRateLimit(client *redis.Client) { // 接受 Redis 客戶端
	fmt.Println("\n=== 4. Rate Limiting（請求限流）===") // 印出標題

	const userID = 101    // 模擬的使用者 ID
	const maxRequests = 3 // 每分鐘最多 3 次（示範用，真實通常是 60 或 100）

	// 清理可能殘留的限流計數（確保示範效果）
	client.Del(ctx, fmt.Sprintf("rate:%d:%d", userID, time.Now().Minute())) // 清除計數

	fmt.Printf("\n使用者 %d 的 API 請求限制：每分鐘最多 %d 次\n", userID, maxRequests) // 說明限制

	// 模擬 5 次 API 呼叫
	for i := 1; i <= 5; i++ { // 發出 5 次請求
		allowed, count, _ := checkRateLimit(client, userID, maxRequests) // 檢查限流

		if allowed { // 如果允許
			fmt.Printf("  第 %d 次請求：✅ 允許（目前計數：%d/%d）\n", i, count, maxRequests) // 印出允許
		} else { // 如果被限流
			fmt.Printf("  第 %d 次請求：❌ 被限流（計數 %d 已超過 %d）\n", i, count, maxRequests) // 印出拒絕
		}
	}

	// 清理本節建立的 key
	client.Del(ctx, fmt.Sprintf("rate:%d:%d", userID, time.Now().Minute()))
}

// ==========================================================================
// 5. 排行榜（Sorted Set）
// ==========================================================================
//
// Redis Sorted Set（有序集合）是 Redis 最強大的資料結構之一
// 每個成員都有一個「分數」（Score），Redis 會自動排序
//
// 就像電玩的排行榜：
//   玩家名稱  → Member（成員）
//   玩家分數  → Score（分數）
//   Redis 自動維護從高到低的排序
//
// 主要指令：
//   ZADD key score member  → 新增或更新成員的分數
//   ZINCRBY key n member   → 增加成員的分數
//   ZREVRANK key member    → 取得成員的排名（從高分開始，第 1 名 = 0）
//   ZREVRANGE key 0 N      → 取得前 N+1 名（從高分到低分）
//   ZSCORE key member      → 取得成員的分數

// demonstrateLeaderboard 示範排行榜功能
func demonstrateLeaderboard(client *redis.Client) { // 接受 Redis 客戶端
	fmt.Println("\n=== 5. 排行榜（Sorted Set）===") // 印出標題

	leaderboardKey := "tutorial:leaderboard" // 排行榜的 Redis key
	client.Del(ctx, leaderboardKey)           // 清理舊資料，確保示範效果

	// ---- ZADD：新增玩家分數 ----
	fmt.Println("\n--- 新增玩家分數 ---") // 小標題

	// redis.Z 是代表一個有序集合成員的結構：{ Score: 分數, Member: 成員 }
	players := []redis.Z{ // 定義玩家分數列表
		{Score: 3500, Member: "Alice"},  // Alice：3500 分
		{Score: 5000, Member: "Bob"},    // Bob：5000 分
		{Score: 2800, Member: "Carol"},  // Carol：2800 分
		{Score: 4200, Member: "Dave"},   // Dave：4200 分
		{Score: 1500, Member: "Eve"},    // Eve：1500 分
	}

	client.ZAdd(ctx, leaderboardKey, players...) // 批量新增（... 把切片展開成多個參數）
	fmt.Println("  已新增 5 位玩家的分數") // 印出新增訊息

	// ---- ZREVRANGE：取得前 3 名（從高分到低分）----
	fmt.Println("\n--- 排行榜前 3 名 ---") // 小標題

	// ZRevRangeWithScores：從高分到低分排列，取第 0 到第 2 名（共 3 名）
	top3, _ := client.ZRevRangeWithScores(ctx, leaderboardKey, 0, 2).Result() // 取前 3 名
	for i, z := range top3 { // 走訪前 3 名
		fmt.Printf("  第 %d 名：%s（%.0f 分）\n", i+1, z.Member, z.Score) // 印出排名、名稱、分數
	}

	// ---- ZINCRBY：增加玩家分數 ----
	fmt.Println("\n--- Alice 完成任務，增加 800 分 ---") // 小標題
	newScore, _ := client.ZIncrBy(ctx, leaderboardKey, 800, "Alice").Result() // Alice +800 分
	fmt.Printf("  Alice 新分數：%.0f 分\n", newScore)                          // 印出新分數

	// ---- ZREVRANK：查詢排名 ----
	fmt.Println("\n--- 查詢各玩家排名（加分後）---") // 小標題

	players2 := []string{"Alice", "Bob", "Carol"} // 要查詢的玩家列表
	for _, name := range players2 {                // 走訪每個玩家
		// ZRevRank：從高分開始計算排名（0 = 第 1 名）
		rank, _ := client.ZRevRank(ctx, leaderboardKey, name).Result() // 取得排名
		score, _ := client.ZScore(ctx, leaderboardKey, name).Result()  // 取得分數
		fmt.Printf("  %s：第 %d 名，%.0f 分\n", name, rank+1, score)  // +1 因為從 0 開始
	}

	// ---- 更新後的完整排行榜 ----
	fmt.Println("\n--- 完整排行榜（更新後）---") // 小標題
	allPlayers, _ := client.ZRevRangeWithScores(ctx, leaderboardKey, 0, -1).Result() // -1 表示取全部
	for i, z := range allPlayers { // 走訪全部玩家
		fmt.Printf("  #%d %s：%.0f 分\n", i+1, z.Member, z.Score) // 印出排名、名稱、分數
	}

	// 清理本節建立的 key
	client.Del(ctx, leaderboardKey)
}

// ==========================================================================
// 6. 計數器（文章瀏覽數）
// ==========================================================================
//
// INCR（Increment）是 Redis 最簡單也最有用的指令之一
// 對一個 key 的值加 1，如果 key 不存在就從 0 開始加
// 速度極快，而且是原子操作（不會有競態條件問題）
//
// 應用場景：
//   - 文章瀏覽數
//   - 商品庫存扣減
//   - 用戶登入失敗次數
//   - API 呼叫統計

// demonstrateCounter 示範計數器功能
func demonstrateCounter(client *redis.Client) { // 接受 Redis 客戶端
	fmt.Println("\n=== 6. 計數器（文章瀏覽數）===") // 印出標題

	// 清理舊資料
	client.Del(ctx, "article:1:views", "article:2:views") // 清除舊的瀏覽數

	// 模擬多個使用者瀏覽文章
	fmt.Println("\n--- 模擬多人瀏覽文章 1 ---") // 小標題

	for i := 1; i <= 5; i++ { // 模擬 5 次瀏覽
		// INCR：原子地加 1，回傳加 1 後的值
		views, _ := client.Incr(ctx, "article:1:views").Result() // 文章 1 的瀏覽數 +1
		fmt.Printf("  第 %d 個人瀏覽了文章 1，目前瀏覽數：%d\n", i, views)  // 印出瀏覽數
	}

	// 模擬批量增加（例如從舊系統匯入）
	fmt.Println("\n--- 文章 2 匯入舊資料（一次加 100）---") // 小標題
	client.IncrBy(ctx, "article:2:views", 100)          // INCRBY：一次加指定的數量
	views2, _ := client.Get(ctx, "article:2:views").Result() // 取得目前瀏覽數
	fmt.Printf("  文章 2 瀏覽數：%s\n", views2)              // 印出瀏覽數

	// 查詢多篇文章的瀏覽數（MGet：一次取多個 key）
	fmt.Println("\n--- 一次查詢多篇文章的瀏覽數（MGet）---") // 小標題
	// MGet 效率比逐一 Get 更高，只需要一次網路往返
	values, _ := client.MGet(ctx, "article:1:views", "article:2:views").Result() // 批量取得
	fmt.Printf("  文章 1 瀏覽數：%v\n", values[0])                                 // 印出文章 1
	fmt.Printf("  文章 2 瀏覽數：%v\n", values[1])                                 // 印出文章 2

	// 清理本節建立的 key
	client.Del(ctx, "article:1:views", "article:2:views")
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第二十課：Redis 快取")                     // 標題
	fmt.Println("==========================================") // 分隔線
	fmt.Println()                                            // 空行

	// 提示使用者如何啟動 Redis
	fmt.Println("確認 Redis 已啟動（三選一）：") // 提示
	fmt.Println("  A. Docker: docker-compose -f docker-compose.dev.yml up -d") // 選項 A
	fmt.Println("  B. Upstash: 設定 REDIS_URL 環境變數")                         // 選項 B
	fmt.Println("  C. 本地安裝: brew install redis && brew services start redis") // 選項 C
	fmt.Println() // 空行

	client := connectRedis() // 連接 Redis（如果失敗會 panic 並顯示提示）
	defer client.Close()     // 程式結束時關閉連線（釋放資源）

	demonstrateBasicOps(client)    // 示範 1：基本操作（SET/GET/DEL/TTL）
	demonstrateCacheAside(client)  // 示範 2：Cache-Aside 快取模式
	demonstrateSession(client)     // 示範 3：Session 儲存
	demonstrateRateLimit(client)   // 示範 4：Rate Limiting
	demonstrateLeaderboard(client) // 示範 5：排行榜（Sorted Set）
	demonstrateCounter(client)     // 示範 6：計數器（INCR）

	fmt.Println("\n==========================================") // 分隔線
	fmt.Println(" 教學完成！")                                // 結尾
	fmt.Println("==========================================") // 分隔線
}
