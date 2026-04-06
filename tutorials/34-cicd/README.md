# 第三十四課：CI/CD（持續整合/持續部署）

> **一句話總結**：CI/CD 就是把「手動測試、手動部署」全部交給機器自動完成，讓你每次 push 程式碼都能安心上線。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟢 初級工程師 | 了解 CI/CD 概念，能讀懂 `.yml` 設定檔 |
| 🟡 中級工程師 | 能寫 GitHub Actions Workflow，設定自動測試和建構 |
| 🔴 資深工程師 | **必備**：能設計完整 Pipeline，包含多環境部署、回滾策略 |
| 🏢 DevOps/SRE | 核心技能，負責維護和最佳化 CI/CD 系統 |

## 你會學到什麼？

- CI/CD 是什麼？為什麼每個團隊都需要它
- GitHub Actions 的四大元件：Workflow、Job、Step、Runner
- 如何寫一份完整的 `.github/workflows/ci.yml`
- 多階段 Docker 建構（結合第 23 課）
- CI 中的測試策略：`go test`、`go vet`、`golangci-lint`
- 三種部署策略的差異：Rolling、Blue-Green、Canary
- 如何安全管理 Secrets
- 為部落格專案設計完整的 CI/CD Pipeline

## 執行方式

```bash
# 查看本課範例 Workflow 檔案
cat tutorials/34-cicd/.github/workflows/ci.yml

# 在你的 GitHub 專案中使用
# 將 .github/workflows/ci.yml 複製到你的專案根目錄
cp -r tutorials/34-cicd/.github .github

# 查看 GitHub Actions 執行狀態（需要 gh CLI）
gh run list
gh run view <run-id>
gh run watch <run-id>  # 即時觀看執行進度
```

## 生活比喻：自動化工廠的生產線

```
手動部署（沒有 CI/CD）：

  工程師寫完程式碼
    → 手動跑測試（但常常忘記）
      → 手動打包（步驟很多容易搞錯）
        → 手動上傳到伺服器（半夜三點上線）
          → 手動確認有沒有壞掉
            → 壞了！手動回滾（已經凌晨四點了...）

自動化工廠（有 CI/CD）：

  工程師 push 程式碼
    → 🤖 自動跑 lint（程式碼風格檢查）
      → 🤖 自動跑測試（所有測試都 PASS 才繼續）
        → 🤖 自動建構 Docker Image
          → 🤖 自動部署到 Staging 環境
            → 🤖 自動部署到 Production（或等待審批）
              → 🤖 自動執行 Smoke Test
                → 出問題？🤖 自動回滾！

CI = 持續整合 = 每次 push 都自動測試、建構
CD = 持續部署 = 測試通過後自動部署到伺服器
```

## 為什麼需要 CI/CD？手動部署的恐怖故事

| 情境 | 沒有 CI/CD | 有 CI/CD |
|------|-----------|---------|
| 忘記跑測試就部署 | 線上服務壞了，使用者投訴 | 測試沒通過，部署被阻擋 |
| 部署步驟搞錯 | 少了一步，服務起不來 | 每次都是同樣的自動化步驟 |
| 半夜緊急部署 | 要叫人起來手動操作 | 合併 PR 就自動部署 |
| 多人同時開發 | 「我的改動沒問題啊？」 | 每個 PR 都獨立測試 |
| 需要回滾 | 「上一版的 binary 在哪？」 | 一鍵回滾到上一個版本 |
| 環境不一致 | 「在我電腦上能跑啊」 | Docker 確保環境一致 |

> **真實案例**：Knight Capital 在 2012 年因為手動部署出錯，45 分鐘內虧損 4.6 億美元。自動化部署不只是效率問題，更是安全問題。

## GitHub Actions 四大元件

```
┌──────────────────────────────────────────────────────┐
│  Workflow（工作流程）                                   │
│  定義在 .github/workflows/*.yml                        │
│                                                        │
│  ┌────────────────────────────────────────────┐       │
│  │  Job: test（工作）                           │       │
│  │  runs-on: ubuntu-latest  ← Runner          │       │
│  │                                             │       │
│  │  ┌─────────────────────────────────────┐   │       │
│  │  │  Step 1: actions/checkout@v4        │   │       │
│  │  ├─────────────────────────────────────┤   │       │
│  │  │  Step 2: actions/setup-go@v5        │   │       │
│  │  ├─────────────────────────────────────┤   │       │
│  │  │  Step 3: run: go test ./...         │   │       │
│  │  └─────────────────────────────────────┘   │       │
│  └────────────────────────────────────────────┘       │
│                                                        │
│  ┌────────────────────────────────────────────┐       │
│  │  Job: build（依賴 test 完成後才執行）        │       │
│  │  needs: [test]                              │       │
│  └────────────────────────────────────────────┘       │
└──────────────────────────────────────────────────────┘
```

| 元件 | 說明 | 比喻 |
|------|------|------|
| **Workflow** | 一個完整的自動化流程，定義在 YAML 檔案中 | 工廠的一條完整生產線 |
| **Job** | Workflow 中的一個獨立工作，可以平行或串列執行 | 生產線上的一個工作站 |
| **Step** | Job 中的一個步驟，按順序執行 | 工作站上的一個動作 |
| **Runner** | 執行 Job 的機器（GitHub 提供或自建） | 執行工作的機器人 |

## 完整 Workflow YAML 逐行解說

```yaml
# ===== 第一段：名稱和觸發條件 =====
name: CI                          # Workflow 名稱（顯示在 GitHub Actions 頁面）

on:                               # 什麼時候觸發這個 Workflow
  push:
    branches: [main]              # push 到 main 分支時觸發
  pull_request:
    branches: [main]              # 對 main 發 PR 時觸發

# ===== 第二段：環境變數 =====
env:
  GO_VERSION: "1.23"              # 統一管理 Go 版本
  REGISTRY: ghcr.io               # GitHub Container Registry
  IMAGE_NAME: ${{ github.repository }}  # 例如 user/blog-api

# ===== 第三段：Jobs =====
jobs:
  # ----- Job 1：程式碼品質檢查 -----
  lint:
    name: Lint & Vet
    runs-on: ubuntu-latest        # 使用 GitHub 提供的 Ubuntu Runner
    steps:
      - uses: actions/checkout@v4             # Step 1：拉取程式碼
      - uses: actions/setup-go@v5             # Step 2：安裝 Go
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true                         # 快取 Go modules（加速）
      - run: go vet ./...                     # Step 3：靜態分析
      - uses: golangci/golangci-lint-action@v4  # Step 4：golangci-lint
        with:
          version: latest

  # ----- Job 2：測試 -----
  test:
    name: Test
    runs-on: ubuntu-latest
    needs: [lint]                  # 等 lint 完成才執行
    services:                     # 啟動測試用的資料庫容器
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: blog_test
        ports:
          - 5432:5432
        options: >-               # 等待 PostgreSQL 就緒
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      - run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/blog_test?sslmode=disable
      - name: Upload coverage              # 上傳覆蓋率報告
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  # ----- Job 3：建構 Docker Image -----
  build:
    name: Build & Push Image
    runs-on: ubuntu-latest
    needs: [test]                 # 等測試通過才建構
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    permissions:
      contents: read
      packages: write             # 需要推送到 GHCR 的權限
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3      # 登入 GHCR
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5  # 建構並推送
        with:
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}

  # ----- Job 4：部署到 Staging -----
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [build]
    environment: staging          # 綁定 staging 環境
    steps:
      - name: Deploy to staging server
        run: |
          echo "Deploying ${{ github.sha }} to staging..."
          # 實際部署指令（SSH、kubectl、等等）

  # ----- Job 5：部署到 Production -----
  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [deploy-staging]
    environment: production       # 綁定 production 環境（需要審批）
    steps:
      - name: Deploy to production server
        run: |
          echo "Deploying ${{ github.sha }} to production..."
```

## Pipeline 視覺化流程

```
push/PR
  │
  ▼
┌──────┐    ┌──────┐    ┌───────┐    ┌─────────────┐    ┌──────────────────┐
│ Lint │───→│ Test │───→│ Build │───→│ Deploy      │───→│ Deploy           │
│      │    │      │    │       │    │ Staging     │    │ Production       │
│go vet│    │go    │    │Docker │    │（自動）      │    │（需要審批 ✋）    │
│lint  │    │test  │    │push   │    │             │    │                  │
└──────┘    └──────┘    └───────┘    └─────────────┘    └──────────────────┘
```

## 多階段 Docker 建構（參考第 23 課）

CI/CD 中建構 Docker Image 時，多階段建構能大幅縮小映像大小：

```dockerfile
# ===== 階段一：建構 =====
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 先複製 go.mod 和 go.sum（利用 Docker 快取層）
COPY go.mod go.sum ./
RUN go mod download

# 再複製原始碼
COPY . .

# 建構靜態二進位
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/server ./cmd/server

# ===== 階段二：執行 =====
FROM scratch

# 複製 CA 憑證（HTTPS 需要）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 只複製編譯好的二進位
COPY --from=builder /app/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
```

| 建構方式 | 映像大小 | 安全性 |
|---------|---------|--------|
| `golang:1.23` 直接執行 | ~1.2 GB | 差（包含編譯器、shell 等） |
| `alpine` 執行 | ~30 MB | 中（有基本 shell） |
| `scratch`（空映像）| ~15 MB | 佳（只有你的程式） |

> **`-ldflags="-w -s"` 是什麼？** `-w` 移除 DWARF 除錯資訊，`-s` 移除符號表，可以減少約 30% 的二進位大小。

## CI 中的測試策略

```yaml
steps:
  # 1. 靜態分析：找出語法錯誤和潛在問題
  - name: Go Vet
    run: go vet ./...

  # 2. Lint：程式碼風格和最佳實踐
  - name: golangci-lint
    uses: golangci/golangci-lint-action@v4
    with:
      version: latest
      # 可以在 .golangci.yml 中設定要啟用的 linter

  # 3. 單元測試 + Race Detector + 覆蓋率
  - name: Test
    run: go test -v -race -coverprofile=coverage.out ./...

  # 4. 整合測試（需要資料庫等外部服務）
  - name: Integration Test
    run: go test -v -tags=integration ./...
    env:
      DATABASE_URL: ${{ secrets.TEST_DATABASE_URL }}
```

| 檢查項目 | 工具 | 用途 | 速度 |
|---------|------|------|------|
| 靜態分析 | `go vet` | 找語法問題、可疑程式碼 | 很快（< 10 秒）|
| Lint | `golangci-lint` | 程式碼風格、最佳實踐 | 快（< 30 秒）|
| 單元測試 | `go test` | 驗證業務邏輯正確性 | 中等（< 2 分鐘）|
| Race Detector | `go test -race` | 找出並行競爭條件 | 中等（速度比較慢）|
| 覆蓋率 | `-coverprofile` | 確認測試覆蓋了多少程式碼 | 無額外成本 |
| 整合測試 | `-tags=integration` | 驗證與外部服務的互動 | 慢（需要資料庫）|

## 部署策略比較

```
Rolling Update（滾動更新）：
  v1 v1 v1 v1    ← 4 台伺服器
  v2 v1 v1 v1    ← 逐步替換
  v2 v2 v1 v1
  v2 v2 v2 v1
  v2 v2 v2 v2    ← 全部更新完成

Blue-Green（藍綠部署）：
  [Blue: v1] ← 流量 ───→ [Green: v2 準備中]
  [Blue: v1]              [Green: v2 就緒]
  [Blue: v1 待命]  ←───→ [Green: v2] ← 流量（一次切換）

Canary（金絲雀部署）：
  v1 v1 v1 v1    ← 100% 流量到 v1
  v2 v1 v1 v1    ← 25% 流量到 v2（觀察指標）
  v2 v2 v1 v1    ← 50% 流量到 v2（指標正常）
  v2 v2 v2 v2    ← 100% 流量到 v2（全部切換）
```

| 策略 | 優點 | 缺點 | 回滾速度 | 適用場景 |
|------|------|------|---------|---------|
| **Rolling** | 簡單、不需額外資源 | 短暫新舊版本共存 | 中等 | 大部分場景 |
| **Blue-Green** | 一次切換、回滾快 | 需要雙倍資源 | 很快（切回去） | 重要服務 |
| **Canary** | 風險最低 | 實作複雜 | 快（停止推進） | 流量大的服務 |

## Secret 管理

在 CI/CD 中，永遠不要把密碼、API Key 寫在程式碼裡。GitHub Actions 提供了 Secrets 機制：

```yaml
# 在 Workflow 中使用 Secret
env:
  DATABASE_URL: ${{ secrets.DATABASE_URL }}
  API_KEY: ${{ secrets.API_KEY }}

# Secret 的安全保障：
# 1. 加密儲存，只有執行時才解密
# 2. 不會顯示在 log 中（自動遮罩）
# 3. Fork 的 PR 無法存取 Secrets
```

```bash
# 用 gh CLI 管理 Secrets
gh secret set DATABASE_URL              # 互動式輸入
gh secret set API_KEY < api_key.txt     # 從檔案讀取
gh secret list                          # 列出所有 Secrets（只看得到名稱）
gh secret delete OLD_SECRET             # 刪除 Secret

# 設定 Environment Secret（只在特定環境可用）
gh secret set PROD_DB_URL --env production
```

| Secret 類型 | 範圍 | 用途 |
|-------------|------|------|
| Repository Secret | 整個 repo 的所有 Workflow | 通用設定（如 CODECOV_TOKEN）|
| Environment Secret | 特定環境（staging/production）| 環境特定的密碼（如 DB URL）|
| Organization Secret | 整個組織的所有 repo | 共用的 API Key |

## 部落格專案 CI/CD Pipeline 範例

把前面學到的全部串起來，為我們的部落格 API 設計完整的 Pipeline：

```yaml
name: Blog API CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  quality:
    name: Code Quality
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
          cache: true
      - run: go vet ./...
      - uses: golangci/golangci-lint-action@v4

  test:
    name: Test
    runs-on: ubuntu-latest
    needs: [quality]
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: blog_test
        ports: ["5432:5432"]
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        ports: ["6379:6379"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
          cache: true
      - name: Run migrations
        run: go run ./cmd/migrate up
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/blog_test?sslmode=disable
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/blog_test?sslmode=disable
          REDIS_URL: redis://localhost:6379
      - uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  build-and-push:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    needs: [test]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          push: true
          tags: |
            ghcr.io/${{ github.repository }}:latest
            ghcr.io/${{ github.repository }}:${{ github.sha }}

  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [build-and-push]
    environment: production
    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            docker pull ghcr.io/${{ github.repository }}:${{ github.sha }}
            docker stop blog-api || true
            docker rm blog-api || true
            docker run -d --name blog-api \
              -p 8080:8080 \
              -e DATABASE_URL="${{ secrets.PROD_DATABASE_URL }}" \
              ghcr.io/${{ github.repository }}:${{ github.sha }}
      - name: Smoke Test
        run: |
          sleep 10
          curl -f https://api.example.com/healthz || exit 1
          echo "Smoke test passed!"
```

## 常用 GitHub Actions 指令

```bash
# 查看 workflow 執行狀態
gh run list
gh run view <run-id>
gh run watch <run-id>             # 即時觀看

# 重新觸發失敗的 workflow
gh run rerun <run-id>
gh run rerun <run-id> --failed    # 只重跑失敗的 job

# 手動觸發 workflow
gh workflow run deploy.yml

# 查看 workflow 列表
gh workflow list
gh workflow view ci.yml

# 下載 artifact
gh run download <run-id>
```

## FAQ

### Q1：CI/CD 和 DevOps 有什麼關係？

CI/CD 是 DevOps 文化中最核心的實踐之一。DevOps 強調開發（Dev）和維運（Ops）的協作，而 CI/CD 就是用自動化工具來橋接這兩個角色。不過 DevOps 的範圍更廣，還包含監控（第 35 課）、可觀測性（第 37 課）、基礎設施即程式碼（IaC）等。

### Q2：GitHub Actions 是免費的嗎？

公開 Repo 完全免費。私有 Repo 每月有 2000 分鐘的免費額度（以 Ubuntu Runner 計算），超過後按分鐘計費。大部分小型專案的免費額度綽綽有餘。

### Q3：CI 跑太慢怎麼辦？

常見的加速方式：(1) 啟用 `cache: true` 快取 Go modules；(2) 將不互相依賴的 Job 設為平行執行；(3) 使用更快的 Runner（如 `ubuntu-latest-16-cores`）；(4) 只在有變動的 package 跑測試（`go test ./pkg/changed/...`）。

### Q4：如何處理部署失敗？需要手動回滾嗎？

好的 CI/CD Pipeline 應該支持自動回滾。最簡單的做法是：每次部署時用 git SHA 作為 Docker Image 的 tag（像範例中的 `${{ github.sha }}`），回滾時只要重新部署上一個版本的 image 即可。進階做法可以搭配 Kubernetes 的 rollback 機制。

### Q5：PR 上的 CI 檢查可以設為「必須通過」嗎？

可以！在 GitHub Repo 的 Settings > Branches > Branch protection rules 中，啟用「Require status checks to pass before merging」，然後選擇你的 CI Workflow。這樣就能確保所有 PR 都必須通過 CI 才能合併。

## 練習

1. 在 GitHub Actions 中加入 `golangci-lint` 靜態分析步驟
2. 加入一個 smoke test 步驟：部署後自動打 `/healthz` API 確認服務正常
3. 設定 GitHub Actions 的 dependency caching（用 `actions/cache` 快取 Go modules）
4. 寫一個手動觸發（`workflow_dispatch`）的 deploy workflow，支持選擇部署環境（staging/production）
5. 設計一個包含 lint、test、build 三個 Job 的 Workflow，其中 test 和 lint 平行執行，build 在兩者都通過後才執行

## 下一課預告

下一課我們會學習 **Prometheus 與 Grafana 監控**——部署上線之後，怎麼知道服務有沒有問題？答案就是監控和告警。CI/CD 讓你自動部署，監控讓你即時知道部署的結果是否健康。
