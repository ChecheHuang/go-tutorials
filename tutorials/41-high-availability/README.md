# 第四十一課：高可用架構（High Availability）

> **一句話總結**：100 萬人排隊中 Redis crash，系統不能停——這就是高可用。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🔴 資深工程師 | **重點**：理解 Replication、Failover 概念 |
| ⚫ Staff / Principal | **核心能力**：設計多層防線、處理腦裂問題 |

## 你會學到什麼？

- 主從複製（Replication）：資料不只一份
- Sentinel 自動故障轉移：偵測 crash 並自動切換
- Multi-Queue Failover：Redis 掛了自動切 DB
- 六層流量防線：100 萬 QPS 降到 100
- 腦裂（Split Brain）：兩個 Primary 同時存在的災難

## 執行方式

```bash
go run ./tutorials/41-high-availability/
```

## 六層流量防線

這是搶票系統最重要的架構——不是「怎麼處理 100 萬請求」，而是「怎麼讓 100 萬變成 100」：

```
100 萬請求
    ↓
┌─ 1. CDN 靜態快取 ──────────────── 擋掉 70%（靜態資源不打後端）
│   30 萬
├─ 2. WAF / Bot 過濾 ────────────── 擋掉 50%（機器人、黃牛）
│   15 萬
├─ 3. API Gateway ───────────────── 擋掉 20%（認證失敗、格式錯誤）
│   12 萬
├─ 4. Rate Limiter ──────────────── 擋掉 90%（每人每秒 1 次）
│   1.2 萬
├─ 5. Waiting Room Queue ────────── 每秒只放 120 人進入
│   120
└─ 6. Seat Lock ─────────────────── 鎖座位，避免超賣
    120
```

## Redis 高可用方案對比

| 方案 | 自動 Failover | 寫入擴展 | 適合場景 |
|------|:---:|:---:|---------|
| 主從複製 | ❌ 手動 | ❌ | 小型系統 |
| Sentinel | ✅ 自動 | ❌ | 中型系統（搶票 queue）|
| Cluster | ✅ 自動 | ✅ 分片 | 大型系統（百萬 QPS）|

## 腦裂（Split Brain）

```
正常：
  [Primary] ←→ [Replica-1] ←→ [Replica-2]

網路分區：
  [Primary]     ╳     [Replica-1] ←→ [Replica-2]
   寫入 A                 ↑ 被 Sentinel 選為新 Primary
                         寫入 B

網路恢復：
  A 和 B 衝突了！哪個是對的？
```

**解法**：Quorum（多數決）— 需要 N/2+1 個 Sentinel 同意才能 failover。

## 面試題對照

| 面試問題 | 考的概念 | 本課位置 |
|---------|---------|---------|
| Redis 掛了系統怎麼不崩？ | Replication + Failover | Demo 1-2 |
| Queue 掛了排隊資料怎麼辦？ | Multi-Queue Fallback | Demo 3 |
| 100 萬人怎麼不打爆後端？ | 六層流量防線 | Demo 4 |
| 兩個 Primary 同時寫入？ | Split Brain + Quorum | Demo 5 |

## 下一課預告

**第四十課：分散式一致性** — CAP 定理實戰、Saga vs 2PC、最終一致性、Inventory Token。
