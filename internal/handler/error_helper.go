package handler

import (
	"blog-api/pkg/apperror"
	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// handleError 根據 apperror 類型自動選擇正確的 HTTP 狀態碼回傳
func handleError(c *gin.Context, err error) {
	status := apperror.HTTPStatus(err)
	response.Error(c, status, err.Error())
}
