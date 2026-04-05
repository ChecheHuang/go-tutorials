# 第三十課：CI/CD（持續整合/持續部署）

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 CI/CD 概念，能讀懂 GitHub Actions 設定 |
| 🔴 資深工程師 | **必備**：能設計完整 Pipeline，包含多環境部署、回滾策略 |
| 🏢 DevOps/SRE | 核心技能，負責維護和最佳化 CI/CD 系統 |

## 本課內容

- **`.github/workflows/ci.yml`**：完整的 GitHub Actions Pipeline
- **`Dockerfile`**：多階段建構，最小化映像大小

## Pipeline 流程

```
push/PR → Lint → Test → Build → Deploy Staging → Deploy Production
                                                  ↑
                                          （需要手動批准）
```

## 核心概念

### GitHub Actions 基礎

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go test ./...
```

### 多階段 Docker 建構

```dockerfile
# 建構階段（~1GB）
FROM golang:1.23-alpine AS builder
COPY . .
RUN go build -o /app ./cmd/server

# 執行階段（~20MB）
FROM scratch
COPY --from=builder /app /app
ENTRYPOINT ["/app"]
```

### 環境保護規則

- **Staging**：自動部署，無需批准
- **Production**：需要指定人員批准
- 使用 `environment: production` 設定環境

## 常用 GitHub Actions 指令

```bash
# 查看 workflow 執行狀態
gh run list
gh run view <run-id>

# 重新觸發失敗的 workflow
gh run rerun <run-id>

# 查看 secrets（只能查看名稱，不能看值）
gh secret list
```

## 最佳實踐

- 使用 `cache: true` 快取 Go modules
- `CGO_ENABLED=0` 生成靜態二進位
- `-ldflags="-w -s"` 減小二進位大小
- 使用 `GITHUB_TOKEN` 推送到 GitHub Container Registry
- 部署後執行 smoke test 確認服務正常
