# 第二十二課：Benchmark + Load Testing

> **一句話總結**：「我覺得很快」不算數——用數據證明你的程式碼夠快。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | **重點**：學會用 `go test -bench` 測量效能 |
| 🔴 資深工程師 | **必備**：效能瓶頸分析、壓力測試、容量規劃 |

## 執行方式

```bash
# 執行 Demo
go run ./tutorials/22-benchmark/

# 執行正式 Benchmark
go test -bench=. -benchmem ./tutorials/22-benchmark/
```

## 下一課預告

**第二十三課：Docker 容器化** — 把應用程式打包成 Docker 映像，用 docker-compose 一鍵啟動。
