# === 階段 1：建置階段 ===
# 使用 Go 官方映像進行編譯
FROM golang:1.23-alpine AS builder

# 安裝 CGO 所需的 C 編譯器（SQLite 需要）
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# 先複製 go.mod 和 go.sum，利用 Docker 快取層加速重複建置
COPY go.mod go.sum ./
RUN go mod download

# 複製所有原始碼
COPY . .

# 編譯為靜態二進位檔
RUN CGO_ENABLED=1 go build -o /app/server ./cmd/server/

# === 階段 2：執行階段 ===
# 使用最小化的 Alpine 映像，減小最終映像大小
FROM alpine:3.19

# 安裝 SQLite 執行時所需的共享函式庫與 CA 憑證
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# 從建置階段複製編譯好的二進位檔
COPY --from=builder /app/server .

# 建立資料目錄
RUN mkdir -p /app/data

# 設定環境變數
ENV GIN_MODE=release
ENV DB_DSN=/app/data/blog.db
ENV SERVER_PORT=8080

# 開放埠號
EXPOSE 8080

# 啟動伺服器
CMD ["./server"]
