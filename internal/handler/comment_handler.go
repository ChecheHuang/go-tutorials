package handler

import (
	"strconv"

	"blog-api/internal/domain"
	"blog-api/internal/usecase"
	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommentHandler 處理留言相關的 HTTP 請求
type CommentHandler struct {
	commentUsecase usecase.CommentUsecase
}

// NewCommentHandler 建立留言 Handler 實例
func NewCommentHandler(commentUsecase usecase.CommentUsecase) *CommentHandler {
	return &CommentHandler{commentUsecase: commentUsecase}
}

// Create 建立新留言
// @Summary     建立留言
// @Description 在指定文章下建立新留言（需要登入）
// @Tags        comments
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                        true "文章 ID" example(1)
// @Param       request body domain.CreateCommentRequest true "留言內容"
// @Success     201 {object} response.Response{data=domain.Comment}
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /articles/{id}/comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的文章 ID")
		return
	}

	var req domain.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "請求參數驗證失敗："+err.Error())
		return
	}

	userID := c.GetUint("user_id")

	comment, err := h.commentUsecase.Create(uint(articleID), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, comment)
}

// GetByArticleID 取得指定文章的所有留言
// @Summary     取得文章留言
// @Description 取得指定文章下的所有留言
// @Tags        comments
// @Produce     json
// @Param       id path int true "文章 ID" example(1)
// @Success     200 {object} response.Response{data=[]domain.Comment}
// @Failure     400 {object} response.Response
// @Router      /articles/{id}/comments [get]
func (h *CommentHandler) GetByArticleID(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的文章 ID")
		return
	}

	comments, err := h.commentUsecase.GetByArticleID(uint(articleID))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, comments)
}

// Update 更新留言
// @Summary     更新留言
// @Description 更新指定留言（只有留言者本人可以更新）
// @Tags        comments
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id      path int                        true "留言 ID" example(1)
// @Param       request body domain.UpdateCommentRequest true "更新內容"
// @Success     200 {object} response.Response{data=domain.Comment}
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /comments/{id} [put]
func (h *CommentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的留言 ID")
		return
	}

	var req domain.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "請求參數驗證失敗："+err.Error())
		return
	}

	userID := c.GetUint("user_id")

	comment, err := h.commentUsecase.Update(uint(id), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, comment)
}

// Delete 刪除留言
// @Summary     刪除留言
// @Description 刪除指定留言（只有留言者本人可以刪除）
// @Tags        comments
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "留言 ID" example(1)
// @Success     200 {object} response.Response
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "無效的留言 ID")
		return
	}

	userID := c.GetUint("user_id")

	if err := h.commentUsecase.Delete(uint(id), userID); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "留言已刪除"})
}
