# 第十八課：Docker 容器化部署

> **一句話總結**：Docker 把你的程式和所有依賴打包成一個「容器」，做到「在我電腦上可以跑，在任何電腦上都可以跑」。

## 你會學到什麼？

- 什麼是 Docker，為什麼需要它（解決什麼問題）
- Docker 的三個核心概念：Dockerfile、Image（映像）、Container（容器）
- 撰寫多階段建置的 Dockerfile（把映像從 800MB 縮到 15MB！）
- 用 docker-compose 一鍵啟動應用程式
- 常用的 Docker 指令速查

## 前置需求

- 安裝 [Docker Desktop](https://www.docker.com/products/docker-desktop/)（Windows/Mac 都有）

## Docker 基本概念

### 三個核心概念

```
Dockerfile     →  映像（Image）  →  容器（Container）
（食譜）           （蛋糕模具）        （實際的蛋糕）

Dockerfile: 描述如何建置映像的「說明書」
Image:      唯讀的模板，包含應用程式和所有依賴
Container:  Image 的執行實例（可以有多個容器使用同一個 Image）
```

### 為什麼需要 Docker？

沒有 Docker：
```
「在我電腦上可以跑」→ 但在別人電腦/伺服器上不行
原因：Go 版本不同、缺少 SQLite、環境變數不同...
```

有 Docker：
```
把應用程式 + 所有依賴打包成一個容器
→ 在任何安裝了 Docker 的機器上都能跑
```

## Dockerfile 逐行解析

```dockerfile
# === 階段 1：建置（Builder Stage）===
FROM golang:1.23-alpine AS builder
# 使用 Go 1.23 的 Alpine 版本作為基底映像
# Alpine 是超輕量的 Linux 發行版（~5MB）
# AS builder 是給這個階段取名字

RUN apk add --no-cache gcc musl-dev
# apk 是 Alpine 的套件管理器
# gcc 和 musl-dev 是 C 編譯器（SQLite 需要 CGO）
# --no-cache 不保留下載快取，減小映像大小

WORKDIR /app
# 設定工作目錄，後續命令都在這個目錄下執行

COPY go.mod go.sum ./
RUN go mod download
# 先只複製 go.mod 和 go.sum，然後下載依賴
# 這是 Docker 的快取優化技巧：
# 只有 go.mod 變化時才重新下載依賴
# 如果只改了程式碼，這一層會使用快取

COPY . .
# 複製所有原始碼

RUN CGO_ENABLED=1 go build -o /app/server ./cmd/server/
# 編譯 Go 程式
# CGO_ENABLED=1：啟用 CGO（SQLite 需要）

# === 階段 2：執行（Runtime Stage）===
FROM alpine:3.19
# 使用乾淨的 Alpine 作為最終映像
# 不需要 Go 編譯器了，只需要執行環境

RUN apk add --no-cache ca-certificates sqlite-libs
# ca-certificates：HTTPS 需要的 CA 憑證
# sqlite-libs：SQLite 的共享函式庫

WORKDIR /app

COPY --from=builder /app/server .
# 從 builder 階段複製編譯好的二進位檔
# 只複製這一個檔案！不需要原始碼和 Go 工具鏈

RUN mkdir -p /app/data
# 建立資料目錄

ENV GIN_MODE=release
ENV DB_DSN=/app/data/blog.db
ENV SERVER_PORT=8080
# 設定環境變數

EXPOSE 8080
# 聲明容器使用的埠號（文件性質，不會自動開放）

CMD ["./server"]
# 容器啟動時執行的命令
```

### 多階段建置的好處

| | 建置階段 | 最終映像 |
|---|---|---|
| 包含 | Go 工具鏈 + 原始碼 + 依賴 | 只有二進位檔 |
| 大小 | ~800MB | ~15MB |

## docker-compose.yml 逐行解析

```yaml
version: "3.8"

services:
  blog-api:                          # 服務名稱
    build: .                          # 使用當前目錄的 Dockerfile 建置
    container_name: blog-api          # 容器名稱
    ports:
      - "8080:8080"                   # 主機埠:容器埠
    environment:                      # 環境變數
      - GIN_MODE=release
      - JWT_SECRET=change-me
    volumes:
      - blog-data:/app/data           # 掛載資料卷（持久化）
    restart: unless-stopped           # 自動重啟策略

volumes:
  blog-data:                          # 定義命名資料卷
```

### Volume（資料卷）

```
沒有 Volume：容器刪除後，SQLite 資料庫也跟著消失
有 Volume：  資料庫儲存在主機上，容器刪除後資料仍在
```

## 常用 Docker 指令

### 映像相關

```bash
# 建置映像
docker build -t blog-api .

# 查看所有映像
docker images

# 刪除映像
docker rmi blog-api
```

### 容器相關

```bash
# 啟動容器
docker run -d -p 8080:8080 --name blog blog-api

# 查看執行中的容器
docker ps

# 查看容器日誌
docker logs blog
docker logs -f blog    # 持續追蹤

# 進入容器（除錯用）
docker exec -it blog sh

# 停止容器
docker stop blog

# 刪除容器
docker rm blog
```

### docker-compose 相關

```bash
# 建置並啟動（背景執行）
docker-compose up -d --build

# 查看日誌
docker-compose logs -f

# 停止
docker-compose down

# 停止並刪除資料卷
docker-compose down -v
```

## 動手做

### 步驟 1：建置映像

```bash
cd /path/to/blog-api
docker build -t blog-api .
```

### 步驟 2：啟動容器

```bash
docker run -d -p 8080:8080 -e JWT_SECRET=my-secret blog-api
```

### 步驟 3：測試

```bash
curl http://localhost:8080/api/v1/articles
```

### 或者用 docker-compose（更簡單）

```bash
docker-compose up -d --build
curl http://localhost:8080/api/v1/articles
docker-compose logs -f
docker-compose down
```

## 在專案中的對應

部落格專案根目錄的 `Dockerfile` 和 `docker-compose.yml` 就是用這一課學到的技術寫成的。

## 常見問題 FAQ

### Q: Docker 容器和虛擬機（VM）有什麼差別？

```
虛擬機（VM）：
  模擬整個電腦（含 CPU、記憶體、作業系統）
  → 很重（幾 GB），啟動慢（幾分鐘）

Docker 容器：
  共用主機的作業系統核心，只隔離應用程式
  → 很輕（幾 MB），啟動快（幾秒）
```

### Q: Dockerfile 中的 `RUN`、`COPY`、`CMD` 差在哪？

| 指令 | 執行時機 | 用途 |
|------|---------|------|
| `RUN` | **建置映像時**（`docker build`） | 安裝套件、執行設定命令 |
| `COPY` | **建置映像時**（`docker build`） | 複製檔案到映像裡 |
| `CMD` | **啟動容器時**（`docker run`） | 容器啟動時執行的命令 |

### Q: 什麼是 Layer（層）？

Dockerfile 的每一行指令都建立一個「層」，Docker 會快取這些層。如果某一層沒有變化，就不需要重新執行，大幅加速建置速度。這就是為什麼先 `COPY go.mod go.sum` 再 `RUN go mod download`，而不是直接 `COPY . .`。

## 練習

1. **查看映像大小**：執行 `docker images`，比較 `golang:1.23-alpine`（建置用）和你建置的映像的大小差異
2. **修改埠號**：把 `docker-compose.yml` 的 `SERVER_PORT` 改為 `3000`，並更新 `ports` 映射為 `"3000:3000"`
3. **進入容器**：用 `docker exec -it blog-api sh` 進入容器，用 `ls /app` 查看裡面有什麼
4. **測試持久化**：用 docker-compose 啟動後，新增一篇文章，然後 `docker-compose down` 再 `docker-compose up -d`，確認文章還在（因為有 Volume）

## 整個教學系列完成！

恭喜你完成了 18 課的 Go 完整教學！

```
你現在掌握了：
  ✅ Go 語言基礎（變數、型別、控制流程、函式、結構體、指標）
  ✅ Go 進階概念（介面、錯誤處理、套件、切片、映射）
  ✅ 架構設計（Clean Architecture、依賴注入）
  ✅ Web 開發（HTTP、Gin、JSON、GORM）
  ✅ 進階功能（中介層、JWT 認證、單元測試、Docker）

下一步：閱讀 TUTORIAL.md，理解完整的部落格 API 專案！
```
