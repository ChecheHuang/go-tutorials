# 第二十課：Redis 快取

> **一句話總結**：Redis 是一個「超快速的記憶體資料庫」，把常用資料放在記憶體裡，速度比資料庫快 100 倍以上。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：Redis 快取是效能最佳化的第一步 |
| 🔴 資深工程師 | 快取策略（Cache-Aside/Write-Through）、快取雪崩/穿透/擊穿 |

## 你會學到什麼？

- 什麼是 Redis，為什麼需要它（冰箱 vs 超市比喻）
- `github.com/redis/go-redis/v9` 套件的基本用法
- **Cache-Aside 模式**：最常見的快取策略
- **Session 儲存**：用 Redis 管理使用者登入狀態
- **Rate Limiting**：限制每個用戶的 API 請求頻率
- **排行榜**：用 Sorted Set 實作即時排名
- **計數器**：用 INCR 實作原子計數（文章瀏覽數）

## 啟動 Redis（三選一）

### 選項 A：Docker（推薦，最簡單）

```bash
# 在專案根目錄執行
docker-compose -f docker-compose.dev.yml up -d

# 確認 Redis 正在運行
docker-compose -f docker-compose.dev.yml ps

# 用 redis-cli 測試
docker exec -it go-tutorials-redis redis-cli ping
# 應該回傳：PONG
```

### 選項 B：Upstash 雲端免費

1. 前往 [https://upstash.com](https://upstash.com) 免費註冊（不需要信用卡）
2. 建立 Redis Database → 選最近的地區
3. 複製 `REDIS_URL`（格式：`rediss://default:xxx@xxx.upstash.io:6379`）
4. 設定環境變數：
```bash
export REDIS_URL="rediss://default:你的密碼@你的端點.upstash.io:6379"
```

### 選項 C：本地安裝

```bash
# macOS
brew install redis
brew services start redis

# Ubuntu/Debian
sudo apt install redis-server
sudo service redis-server start

# 驗證安裝
redis-cli ping  # 應該回傳 PONG
```

## 執行方式

```bash
# 確認 Redis 已啟動後
go run ./tutorials/20-redis
```

## 生活比喻：冰箱 vs 超市

```
你需要一樣東西（使用者資料、商品資訊、排行榜...）

沒有快取：
  每次都去超市（資料庫） → 耗時 50-500ms
  10 個人同時查 → 超市大排長龍

有 Redis 快取：
  先看冰箱（Redis） → 耗時 0.1ms
  冰箱有（Cache Hit） → 直接拿，超快！
  冰箱沒有（Cache Miss） → 去超市買，買回來順便放冰箱
  下次 10 個人查 → 全從冰箱拿，超市不用動

Redis 就是你的超大冰箱：
  - 比超市快 100-1000 倍
  - 但東西有保存期限（TTL，過期自動刪除）
  - 空間有限（比硬碟小得多）
```

## 什麼是 go-redis/v9？

`github.com/redis/go-redis/v9` 是 Redis 官方推薦的 Go 客戶端：

```bash
go get github.com/redis/go-redis/v9
```

```go
import "github.com/redis/go-redis/v9"

// 連接 Redis
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 或者用連線字串
opt, _ := redis.ParseURL("redis://localhost:6379")
client := redis.NewClient(opt)

// 所有操作都需要 Context
ctx := context.Background()
```

## Redis 指令速查

| 操作 | Go 程式碼 | 說明 |
|------|----------|------|
| 儲存（有期限）| `client.Set(ctx, "key", "val", 5*time.Minute)` | 5 分鐘後自動刪除 |
| 儲存（永久）| `client.Set(ctx, "key", "val", 0)` | 不過期 |
| 取得 | `val, err := client.Get(ctx, "key").Result()` | `redis.Nil` = 不存在 |
| 刪除 | `client.Del(ctx, "key1", "key2")` | 可一次刪多個 |
| 是否存在 | `client.Exists(ctx, "key").Result()` | 回傳 1 或 0 |
| 剩餘時間 | `client.TTL(ctx, "key").Result()` | 剩幾秒過期 |
| 計數 +1 | `count, _ := client.Incr(ctx, "key").Result()` | 原子操作 |
| 計數 +N | `client.IncrBy(ctx, "key", 100)` | 一次加 N |
| 批量取得 | `client.MGet(ctx, "k1", "k2").Result()` | 一次取多個 |

## Cache-Aside 模式

最常見的快取策略，讀取時先查快取，寫入時讓快取失效：

```go
func getUser(client *redis.Client, userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)

    // 1. 先查 Redis
    data, err := client.Get(ctx, cacheKey).Result()
    if err == nil {
        // Cache Hit！直接用快取
        var user User
        json.Unmarshal([]byte(data), &user)
        return &user, nil
    }

    // 2. Cache Miss → 查資料庫
    user := db.FindUser(userID)

    // 3. 存入快取（下次就快了）
    jsonData, _ := json.Marshal(user)
    client.Set(ctx, cacheKey, jsonData, 5*time.Minute)

    return user, nil
}

// 更新時讓快取失效
func updateUser(client *redis.Client, user *User) {
    db.Save(user)
    client.Del(ctx, fmt.Sprintf("user:%d", user.ID))  // 刪除快取
    // 下次讀取時會重新從 DB 載入最新資料
}
```

## Rate Limiting

```go
func checkRateLimit(client *redis.Client, userID int, maxReqs int) bool {
    key := fmt.Sprintf("rate:%d:%d", userID, time.Now().Minute())

    count, _ := client.Incr(ctx, key).Result()
    if count == 1 {
        client.Expire(ctx, key, time.Minute)  // 只有第一次才設定過期
    }

    return count <= int64(maxReqs)
}
```

## 排行榜（Sorted Set）

```go
// 新增/更新分數
client.ZAdd(ctx, "leaderboard", redis.Z{Score: 5000, Member: "Bob"})

// 增加分數
client.ZIncrBy(ctx, "leaderboard", 100, "Alice")

// 取前 3 名（從高到低）
top3, _ := client.ZRevRangeWithScores(ctx, "leaderboard", 0, 2).Result()

// 查詢某人的排名（0 = 第 1 名）
rank, _ := client.ZRevRank(ctx, "leaderboard", "Alice").Result()
fmt.Printf("第 %d 名", rank+1)
```

## 在部落格專案中的應用

| 功能 | Redis 用法 | key 格式 |
|------|-----------|----------|
| 文章列表快取 | Cache-Aside | `articles:page:{page}` |
| 用戶資料快取 | Cache-Aside | `user:{id}` |
| 登入 Session | String + TTL | `session:{uuid}` |
| API 限流 | INCR + EXPIRE | `rate:{user_id}:{minute}` |
| 文章瀏覽數 | INCR | `article:{id}:views` |
| 熱門文章排行 | Sorted Set | `articles:trending` |

## Redis Key 命名慣例

```
✅ 好的命名（用冒號分層）：
  user:42
  user:42:profile
  article:100:views
  session:abc123

❌ 不好的命名：
  user42
  User_42_Profile
  articleviews100
```

## 常見問題 FAQ

### Q: 快取和資料庫的資料不一致怎麼辦？

這叫做「快取穿透/擊穿/雪崩」問題，有幾種解決方式：
- **Cache-Aside**（我們這課學的）：更新 DB 時刪除快取，讀取時重新載入
- **Write-Through**：寫入 DB 的同時更新快取（強一致性）
- **TTL 兜底**：就算沒主動刪除，快取也會在 TTL 後自動過期

通常用 Cache-Aside + 合理的 TTL 就夠了。

### Q: Redis 的資料斷電後會消失嗎？

預設情況是的（記憶體）。但我們的 docker-compose.dev.yml 啟用了 `--appendonly yes`（AOF 持久化），Redis 會把每個寫入操作記錄到磁碟，重啟後可以恢復。

### Q: redis.Nil 是什麼？

`redis.Nil` 是一個特殊的 error 值，表示「key 不存在」。需要特別判斷：

```go
val, err := client.Get(ctx, "key").Result()
if err == redis.Nil {
    // key 不存在（這是正常情況）
} else if err != nil {
    // 真正的錯誤（連線問題等）
} else {
    // 成功取得 val
}
```

### Q: 要不要每個功能都加快取？

**不要**！快取讓系統更複雜，只在真正需要時才加：
- 讀多寫少的資料（用戶個人資料、文章內容）
- 計算複雜的結果（排行榜、統計）
- 需要計數但不想每次查 DB（瀏覽數）

## 練習

1. **快取失效實驗**：修改 `getUser`，在第一次 Cache Miss 後，立刻刪除 Redis 的 key，再呼叫第二次，觀察每次都是 Cache Miss 的情況
2. **多層快取 key**：修改 Rate Limiting，改為「每小時最多 100 次」（提示：用 `time.Now().Hour()`）
3. **文章排行榜**：用 Sorted Set 實作「最近 24 小時熱門文章」排行榜，每次有人瀏覽就加分
4. **Session 續期**：在 Session 儲存示範中，加入「每次驗證成功時自動續期」的邏輯

## 下一課預告

**第二十一課：結構化日誌（Structured Logging）** —— 學習用 `zap` 取代 `fmt.Println`，輸出帶有 trace ID、level、caller 的 JSON 格式日誌，讓線上問題排查更容易。
