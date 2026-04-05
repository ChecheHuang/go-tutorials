# 第十五課：PostgreSQL + Schema Design + Index Optimization

> **一句話總結**：生產環境不用 SQLite——學會 PostgreSQL、設計好的 Schema、和正確的索引，是後端工程師的基本功。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：理解正規化、學會設計資料表 |
| 🔴 資深工程師 | **必備**：索引最佳化、EXPLAIN 分析、效能調校 |

## 執行方式

```bash
go run ./tutorials/15-postgresql/
```

## 下一課預告

**第十六課：Error Wrapping** — 用 `fmt.Errorf %w` 和 `errors.Is/As` 建立完整的錯誤鏈。
