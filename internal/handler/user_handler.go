// Package handler 實作 HTTP 請求處理器
// 負責解析請求、呼叫 usecase、回傳回應
// 教學對應：第 12 課（Gin 框架）、第 18 課（JWT 認證）
package handler

import (
	"blog-api/internal/domain"
	"blog-api/internal/usecase"
	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 處理使用者相關的 HTTP 請求
type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// NewUserHandler 建立使用者 Handler 實例
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// Register 處理使用者註冊請求
// @Summary     使用者註冊
// @Description 建立新的使用者帳號
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.RegisterRequest true "註冊資訊"
// @Success     201 {object} response.Response{data=domain.User}
// @Failure     400 {object} response.Response
// @Router      /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "請求參數驗證失敗："+err.Error())
		return
	}

	user, err := h.userUsecase.Register(req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Created(c, user)
}

// Login 處理使用者登入請求
// @Summary     使用者登入
// @Description 驗證使用者身份並回傳 JWT Token
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body domain.LoginRequest true "登入資訊"
// @Success     200 {object} response.Response{data=domain.LoginResponse}
// @Failure     400 {object} response.Response
// @Router      /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "請求參數驗證失敗："+err.Error())
		return
	}

	loginResp, err := h.userUsecase.Login(req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, loginResp)
}

// GetProfile 取得目前登入使用者的資料
// @Summary     取得個人資料
// @Description 取得目前登入使用者的詳細資料
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} response.Response{data=domain.User}
// @Failure     401 {object} response.Response
// @Router      /auth/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := h.userUsecase.GetByID(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, user)
}
