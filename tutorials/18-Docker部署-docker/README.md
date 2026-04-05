# 第十七課：Docker 容器化部署

## 學習目標

- 理解 Docker 的基本概念（映像、容器、Dockerfile）
- 學會撰寫多階段建置的 Dockerfile
- 了解 docker-compose 的用途
- 能夠將 Go 應用程式容器化

## 前置需求

- 安裝 [Docker Desktop](https://www.docker.com/products/docker-desktop/)

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

## 練習

1. 修改 `docker-compose.yml`，把 `SERVER_PORT` 改為 `3000`，並更新 ports 映射
2. 用 `docker exec -it blog-api sh` 進入容器，查看 `/app/data/blog.db` 是否存在
3. 執行 `docker images` 比較 `golang:1.23-alpine` 和最終映像的大小差異
