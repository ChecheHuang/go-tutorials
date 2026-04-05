# ==========================================================================
# 多階段建構（Multi-stage Build）
# ==========================================================================
# 第一階段：建構（包含完整 Go 工具鏈，映像大）
# 第二階段：執行（只包含二進位檔，映像小）
# 結果：最終映像從 ~1GB → ~20MB

# ──── 階段 1：建構 ────
FROM golang:1.23-alpine AS builder

# 安裝建構依賴
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 先複製 go.mod/go.sum，利用 Docker 快取（依賴沒變就不重新下載）
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼
COPY . .

# 建構（CGO_ENABLED=0 生成靜態二進位，不依賴 C 庫）
ARG VERSION=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -o /app/server \
    ./cmd/server

# ──── 階段 2：執行 ────
FROM scratch  # 完全空白的映像（最小化）
# 如果需要 shell 除錯，改用：FROM alpine:3.19

# 從建構階段複製必要檔案
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/server /server

# 設定時區
ENV TZ=Asia/Taipei

# 使用非 root 用戶（安全最佳實踐）
USER 65534:65534  # nobody:nobody

# 暴露埠號
EXPOSE 8080

# Health check（Kubernetes readiness probe 也會用到）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/server", "--health-check"]

ENTRYPOINT ["/server"]
