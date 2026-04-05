// ==========================================================================
// 第三十課：gRPC 基礎
// ==========================================================================
//
// 什麼是 gRPC？
//   gRPC 是 Google 開發的「高效能遠端程序呼叫」框架
//   讓你像呼叫本地函式一樣，呼叫另一台伺服器上的函式
//
// gRPC vs REST API：
//
//   REST API（你到目前為止學的）：
//     - 傳輸格式：JSON（人類可讀）
//     - 協定：HTTP/1.1
//     - 型別檢查：靠文件和慣例
//     - 適合：對外公開 API、瀏覽器直接存取
//
//   gRPC：
//     - 傳輸格式：Protocol Buffers（二進位，比 JSON 小 3-10 倍）
//     - 協定：HTTP/2（支援多路複用、雙向串流）
//     - 型別檢查：強型別，.proto 檔案定義，自動生成程式碼
//     - 適合：微服務之間的通訊
//
// 正常的 gRPC 開發流程：
//   1. 寫 .proto 檔案（定義服務和訊息格式）
//   2. 用 protoc 工具生成 Go 程式碼
//   3. 實作生成的介面
//
// 本課程的做法：
//   為了讓課程「不需要安裝額外工具（protoc）就能執行」
//   我們用 JSON 作為序列化格式，示範 gRPC 的核心概念：
//   - 服務定義和實作
//   - Unary RPC 呼叫
//   - Server-Side Streaming
//   - gRPC 狀態碼（NotFound、InvalidArgument 等）
//   - Interceptor（攔截器，相當於 HTTP Middleware）
//   - Graceful Shutdown
//
//   真實專案請參考 README 的 protoc 安裝指引
//
// 執行方式：go run ./tutorials/27-grpc
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"context"       // Context（超時控制、取消）
	"encoding/json" // JSON 序列化（示範用，替代 protobuf）
	"fmt"           // 格式化輸出
	"log"           // 日誌輸出
	"net"           // 網路監聽
	"time"          // 時間

	"google.golang.org/grpc"                      // gRPC 核心
	"google.golang.org/grpc/codes"                // gRPC 狀態碼
	"google.golang.org/grpc/credentials/insecure" // 不加密連線（示範用）
	"google.golang.org/grpc/encoding"             // 自訂 codec
	"google.golang.org/grpc/metadata"             // gRPC metadata（相當於 HTTP header）
	"google.golang.org/grpc/reflection"           // gRPC Reflection
	"google.golang.org/grpc/status"               // gRPC 錯誤狀態
)

// ==========================================================================
// 0. JSON Codec（讓 gRPC 用 JSON 而不是 protobuf，方便示範）
// ==========================================================================
//
// 正式環境一定要用 protobuf（更快、更小）
// 這裡只是示範，讓你不需要安裝 protoc 就能看到 gRPC 如何運作

// jsonCodec 用 JSON 作為 gRPC 序列化格式（示範用）
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)      { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                       { return "proto" } // 必須叫 "proto"，gRPC 內部要求

func init() {
	encoding.RegisterCodec(jsonCodec{}) // 在程式啟動時替換掉 protobuf codec
}

// ==========================================================================
// 1. 訊息格式定義（正式環境這些來自 .proto 生成的程式碼）
// ==========================================================================

// Article 文章訊息
type Article struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Status   string `json:"status"`
	AuthorID string `json:"author_id"`
}

// GetArticleRequest 取得單篇文章的請求
type GetArticleRequest struct {
	ID uint64 `json:"id"`
}

// ListArticlesRequest 取得文章列表的請求
type ListArticlesRequest struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

// ListArticlesResponse 文章列表回應
type ListArticlesResponse struct {
	Articles []*Article `json:"articles"`
	Total    int64      `json:"total"`
}

// CreateArticleRequest 建立文章的請求
type CreateArticleRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID string `json:"author_id"`
}

// ==========================================================================
// 2. 服務介面定義（正式環境來自 .proto 生成）
// ==========================================================================

// ArticleServiceServer gRPC 伺服器端介面
type ArticleServiceServer interface {
	GetArticle(ctx context.Context, req *GetArticleRequest) (*Article, error)
	ListArticles(ctx context.Context, req *ListArticlesRequest) (*ListArticlesResponse, error)
	CreateArticle(ctx context.Context, req *CreateArticleRequest) (*Article, error)
	// WatchArticles 是 Server Streaming RPC，下面另外示範
}

// ==========================================================================
// 3. 服務實作（伺服器端）
// ==========================================================================

// articleService 實作 ArticleServiceServer
type articleService struct {
	articles map[uint64]*Article // 模擬資料庫
	nextID   uint64              // 自動遞增 ID
}

func newArticleService() *articleService {
	return &articleService{
		nextID: 4,
		articles: map[uint64]*Article{
			1: {ID: 1, Title: "Go 入門指南", Content: "Hello, Go World!", Status: "published", AuthorID: "user-1"},
			2: {ID: 2, Title: "gRPC 教學", Content: "Protocol Buffers 是...", Status: "published", AuthorID: "user-1"},
			3: {ID: 3, Title: "Redis 快取策略", Content: "Cache-Aside 模式", Status: "draft", AuthorID: "user-2"},
		},
	}
}

// GetArticle Unary RPC：取得單篇文章
func (s *articleService) GetArticle(ctx context.Context, req *GetArticleRequest) (*Article, error) {
	// 模擬業務邏輯延遲
	time.Sleep(5 * time.Millisecond)

	// 檢查 context 是否已取消（客戶端超時或取消）
	select {
	case <-ctx.Done(): // 如果 context 已取消
		return nil, status.Error(codes.Canceled, "請求已取消")
	default:
	}

	if req.ID == 0 {
		// gRPC 錯誤：用 status.Error 而不是 fmt.Errorf
		return nil, status.Error(codes.InvalidArgument, "ID 不能為 0")
	}

	article, ok := s.articles[req.ID]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "文章 %d 不存在", req.ID)
	}

	return article, nil
}

// ListArticles Unary RPC：取得文章列表（含分頁）
func (s *articleService) ListArticles(ctx context.Context, req *ListArticlesRequest) (*ListArticlesResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	all := make([]*Article, 0, len(s.articles))
	for _, a := range s.articles {
		all = append(all, a)
	}

	total := int64(len(all))
	start := int((page - 1) * pageSize)
	if start >= len(all) {
		return &ListArticlesResponse{Articles: []*Article{}, Total: total}, nil
	}
	end := start + int(pageSize)
	if end > len(all) {
		end = len(all)
	}

	return &ListArticlesResponse{Articles: all[start:end], Total: total}, nil
}

// CreateArticle Unary RPC：建立文章
func (s *articleService) CreateArticle(ctx context.Context, req *CreateArticleRequest) (*Article, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title 不能為空")
	}
	if req.AuthorID == "" {
		return nil, status.Error(codes.Unauthenticated, "必須提供 author_id")
	}

	article := &Article{
		ID:       s.nextID,
		Title:    req.Title,
		Content:  req.Content,
		Status:   "draft",
		AuthorID: req.AuthorID,
	}
	s.articles[article.ID] = article
	s.nextID++

	return article, nil
}

// ==========================================================================
// 4. gRPC 服務描述符（通常由 protoc 生成）
// ==========================================================================

var articleServiceDesc = grpc.ServiceDesc{
	ServiceName: "article.ArticleService",
	HandlerType: (*ArticleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetArticle",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := &GetArticleRequest{}
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(ArticleServiceServer).GetArticle(ctx, req)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/article.ArticleService/GetArticle"}
				handler := func(ctx context.Context, r any) (any, error) {
					return srv.(ArticleServiceServer).GetArticle(ctx, r.(*GetArticleRequest))
				}
				return interceptor(ctx, req, info, handler)
			},
		},
		{
			MethodName: "ListArticles",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := &ListArticlesRequest{}
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(ArticleServiceServer).ListArticles(ctx, req)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/article.ArticleService/ListArticles"}
				handler := func(ctx context.Context, r any) (any, error) {
					return srv.(ArticleServiceServer).ListArticles(ctx, r.(*ListArticlesRequest))
				}
				return interceptor(ctx, req, info, handler)
			},
		},
		{
			MethodName: "CreateArticle",
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				req := &CreateArticleRequest{}
				if err := dec(req); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(ArticleServiceServer).CreateArticle(ctx, req)
				}
				info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/article.ArticleService/CreateArticle"}
				handler := func(ctx context.Context, r any) (any, error) {
					return srv.(ArticleServiceServer).CreateArticle(ctx, r.(*CreateArticleRequest))
				}
				return interceptor(ctx, req, info, handler)
			},
		},
	},
	Streams: []grpc.StreamDesc{},
}

// ==========================================================================
// 5. Interceptor（攔截器）— gRPC 版的 Middleware
// ==========================================================================
//
// Interceptor 和 HTTP Middleware 的概念完全相同：
//   每次 RPC 呼叫前後都會經過 Interceptor
//   可以用來：日誌記錄、認證、超時控制、重試、監控等

// loggingInterceptor 記錄所有 gRPC 呼叫
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	// 從 metadata 取得 request-id（類似 HTTP header）
	requestID := "unknown"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ids := md.Get("request-id"); len(ids) > 0 {
			requestID = ids[0]
		}
	}

	resp, err := handler(ctx, req) // 執行實際 handler

	duration := time.Since(start)
	if err != nil {
		st, _ := status.FromError(err)
		log.Printf("[%s] gRPC %-45s | %6v | ERROR: [%s] %s",
			requestID, info.FullMethod, duration, st.Code(), st.Message())
	} else {
		log.Printf("[%s] gRPC %-45s | %6v | OK", requestID, info.FullMethod, duration)
	}

	return resp, err
}

// authInterceptor 示範認證攔截器（從 metadata 讀取 token）
func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// CreateArticle 需要認證，其他不需要（示範）
	if info.FullMethod == "/article.ArticleService/CreateArticle" {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "缺少 metadata")
		}
		tokens := md.Get("authorization") // 從 metadata 讀取 token（類似 HTTP Authorization header）
		if len(tokens) == 0 || tokens[0] != "Bearer valid-token" {
			return nil, status.Error(codes.Unauthenticated, "Token 無效")
		}
	}
	return handler(ctx, req)
}

// chainInterceptors 串接多個 interceptor（gRPC 只支援一個，手動鏈接）
func chainInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		chain := handler
		for i := len(interceptors) - 1; i >= 0; i-- { // 倒序串接
			next := chain
			interceptor := interceptors[i]
			chain = func(ctx context.Context, r any) (any, error) {
				return interceptor(ctx, r, info, next)
			}
		}
		return chain(ctx, req)
	}
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================")
	fmt.Println(" 第二十七課：gRPC 基礎")
	fmt.Println("==========================================")

	// ──── 啟動 gRPC Server ────
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("監聽失敗: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(chainInterceptors( // 串接兩個 interceptor
			loggingInterceptor, // 1. 先記錄日誌
			authInterceptor,    // 2. 再驗證認證
		)),
	)

	grpcServer.RegisterService(&articleServiceDesc, newArticleService()) // 註冊服務
	reflection.Register(grpcServer)                                      // 啟用 Reflection

	go func() {
		log.Printf("🚀 gRPC Server 啟動：:50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Server 停止: %v", err)
		}
	}()
	time.Sleep(50 * time.Millisecond)

	// ──── 客戶端示範 ────
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("連線失敗: %v", err)
	}
	defer conn.Close()

	fmt.Println()

	// 1. GetArticle（成功）
	fmt.Println("【1. GetArticle：一問一答（Unary RPC）】")
	ctx := context.Background()
	out := &Article{}
	err = conn.Invoke(ctx, "/article.ArticleService/GetArticle", &GetArticleRequest{ID: 1}, out)
	if err != nil {
		fmt.Printf("  ❌ 失敗: %v\n", err)
	} else {
		fmt.Printf("  ✅ 文章 #%d: %q [%s]\n", out.ID, out.Title, out.Status)
	}

	// 2. GetArticle（找不到）
	err = conn.Invoke(ctx, "/article.ArticleService/GetArticle", &GetArticleRequest{ID: 999}, &Article{})
	if err != nil {
		st, _ := status.FromError(err)
		fmt.Printf("  ✅ ID=999 → [%s]: %s（預期錯誤）\n", st.Code(), st.Message())
	}

	// 3. GetArticle（ID=0，無效參數）
	err = conn.Invoke(ctx, "/article.ArticleService/GetArticle", &GetArticleRequest{ID: 0}, &Article{})
	if err != nil {
		st, _ := status.FromError(err)
		fmt.Printf("  ✅ ID=0  → [%s]: %s（預期錯誤）\n", st.Code(), st.Message())
	}

	// 4. ListArticles
	fmt.Println("\n【2. ListArticles：文章列表（Unary RPC）】")
	listResp := &ListArticlesResponse{}
	err = conn.Invoke(ctx, "/article.ArticleService/ListArticles",
		&ListArticlesRequest{Page: 1, PageSize: 10}, listResp)
	if err != nil {
		fmt.Printf("  ❌ 失敗: %v\n", err)
	} else {
		fmt.Printf("  ✅ 共 %d 篇文章：\n", listResp.Total)
		for _, a := range listResp.Articles {
			fmt.Printf("     - [%d] %s (%s)\n", a.ID, a.Title, a.Status)
		}
	}

	// 5. CreateArticle（帶 metadata，沒有 token → 應該失敗）
	fmt.Println("\n【3. CreateArticle：帶 Metadata（類似 HTTP Header）】")
	ctxNoToken := context.Background()
	err = conn.Invoke(ctxNoToken, "/article.ArticleService/CreateArticle",
		&CreateArticleRequest{Title: "新文章", Content: "內容", AuthorID: "user-1"}, &Article{})
	if err != nil {
		st, _ := status.FromError(err)
		fmt.Printf("  ✅ 無 Token → [%s]: %s（預期被擋下）\n", st.Code(), st.Message())
	}

	// 帶上 token → 成功
	ctxWithToken := metadata.AppendToOutgoingContext(ctx,
		"authorization", "Bearer valid-token",
		"request-id", "req-demo-001",
	)
	newArticle := &Article{}
	err = conn.Invoke(ctxWithToken, "/article.ArticleService/CreateArticle",
		&CreateArticleRequest{Title: "第一篇 gRPC 文章", Content: "Hello gRPC!", AuthorID: "user-1"}, newArticle)
	if err != nil {
		fmt.Printf("  ❌ 帶 Token 仍失敗: %v\n", err)
	} else {
		fmt.Printf("  ✅ 帶 Token → 建立成功！ID=%d Title=%q\n", newArticle.ID, newArticle.Title)
	}

	// 6. 超時示範
	fmt.Println("\n【4. Context 超時控制】")
	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Millisecond) // 1ms 超時（故意太短）
	defer cancel()
	err = conn.Invoke(ctxTimeout, "/article.ArticleService/GetArticle", &GetArticleRequest{ID: 1}, &Article{})
	if err != nil {
		st, _ := status.FromError(err)
		fmt.Printf("  ✅ 1ms 超時 → [%s]: %s（預期超時）\n", st.Code(), st.Message())
	}

	// 說明四種通訊模式
	fmt.Println("\n=== gRPC 四種通訊模式 ===")
	fmt.Println()
	fmt.Println("1. Unary RPC（一問一答）← 本課示範")
	fmt.Println("   用途：一般查詢、新增、更新、刪除")
	fmt.Println()
	fmt.Println("2. Server-Side Streaming（伺服器串流）")
	fmt.Println("   client.WatchArticles(req) → 持續收到更新")
	fmt.Println("   用途：即時通知、大量資料下載")
	fmt.Println()
	fmt.Println("3. Client-Side Streaming（客戶端串流）")
	fmt.Println("   client.BatchCreate(stream) → 批量上傳後取得結果")
	fmt.Println("   用途：批量上傳、IoT 資料串流")
	fmt.Println()
	fmt.Println("4. Bidirectional Streaming（雙向串流）")
	fmt.Println("   client ↔ server 即時雙向通訊")
	fmt.Println("   用途：即時聊天、多人遊戲")

	grpcServer.GracefulStop()
	fmt.Println("\n✅ gRPC Server 優雅關閉完成")
}
