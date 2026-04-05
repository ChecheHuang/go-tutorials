package handler

import (
	"strconv"

	"blog-api/internal/domain"
	"blog-api/internal/usecase"
	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ArticleHandler 處理文章相關的 HTTP 請求
type ArticleHandler struct {
	articleUsecase usecase.ArticleUsecase
}

// NewArticleHandler 建立文章 Handler 實例
func NewArticleHandler(articleUsecase usecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{articleUsecase: articleUsecase}
}

// Create 建立新文章
// @Summary     建立文章
// @Description 建立一篇新文章（需要登入）
// @Tags        articles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       request body domain.CreateArticleRequest true "文章資訊"
// @Success     201 {object} response.Response{data=domain.Article}
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /api/v1/articles [post]
func (h *ArticleHandler) Create(c *gin.Context) {
	var req domain.CreateArticleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "請求參數驗證失敗："+err.Error())
		return
	}

	// 從 JWT context 取得目前登入的使用者 ID
	userID := c.GetUint("user_id")

	article, err := h.articleUsecase.Create(userID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, article)
}

// GetByID 取得單篇文章詳情
// @Summary     取得文章詳情
// @Description 根據 ID 取得文章詳情，包含作者與留言
// @Tags        articles
// @Produce     json
// @Param       id path int true "文章 ID" example(1)
// @Success     200 {object} response.Response{data=domain.Article}
// @Failure     404 {object} response.Response
// @Router      /api/v1/articles/{id} [get]
func (h *ArticleHandler) GetByID(c *gin.Context) {
	// 從 URL 路徑參數取得文章 ID
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的文章 ID")
		return
	}

	article, err := h.articleUsecase.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, article)
}

// GetAll 取得文章列表（支援分頁與搜尋）
// @Summary     取得文章列表
// @Description 取得文章列表，支援分頁、關鍵字搜尋與依作者篩選
// @Tags        articles
// @Produce     json
// @Param       page      query int    false "頁碼（預設 1）"
// @Param       page_size query int    false "每頁筆數（預設 10，最大 100）"
// @Param       search    query string false "搜尋關鍵字"
// @Param       user_id   query int    false "依作者 ID 篩選"
// @Success     200 {object} response.Response{data=response.PaginatedData}
// @Router      /api/v1/articles [get]
func (h *ArticleHandler) GetAll(c *gin.Context) {
	var query domain.ArticleQuery

	// 使用 ShouldBindQuery 綁定 URL 查詢參數
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "查詢參數驗證失敗："+err.Error())
		return
	}

	// 設定分頁預設值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	articles, total, err := h.articleUsecase.GetAll(query)
	if err != nil {
		response.InternalServerError(c, "取得文章列表失敗")
		return
	}

	response.Paginated(c, articles, total, query.Page, query.PageSize)
}

// Update 更新文章
// @Summary     更新文章
// @Description 更新指定文章（只有作者本人可以更新）
// @Tags        articles
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                       true "文章 ID"
// @Param       request body domain.UpdateArticleRequest true "更新資訊"
// @Success     200 {object} response.Response{data=domain.Article}
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Failure     403 {object} response.Response
// @Router      /api/v1/articles/{id} [put]
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的文章 ID")
		return
	}

	var req domain.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "請求參數驗證失敗："+err.Error())
		return
	}

	userID := c.GetUint("user_id")

	article, err := h.articleUsecase.Update(uint(id), userID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, article)
}

// Delete 刪除文章
// @Summary     刪除文章
// @Description 刪除指定文章（只有作者本人可以刪除）
// @Tags        articles
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "文章 ID" example(1)
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Failure     403 {object} response.Response
// @Router      /api/v1/articles/{id} [delete]
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的文章 ID")
		return
	}

	userID := c.GetUint("user_id")

	if err := h.articleUsecase.Delete(uint(id), userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "文章已刪除"})
}
