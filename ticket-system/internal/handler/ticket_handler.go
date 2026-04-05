package handler

import (
	"net/http"
	"strconv"

	"ticket-system/internal/domain"
	"ticket-system/internal/usecase"
	"ticket-system/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	uc *usecase.TicketUsecase
}

func NewTicketHandler(uc *usecase.TicketUsecase) *TicketHandler {
	return &TicketHandler{uc: uc}
}

// GrabTicket 搶票 API
// POST /api/v1/tickets/grab
func (h *TicketHandler) GrabTicket(c *gin.Context) {
	var cmd domain.OrderCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "參數驗證失敗：" + err.Error()})
		return
	}

	order, err := h.uc.GrabTicket(c.Request.Context(), cmd)
	if err != nil {
		status := apperror.HTTPStatus(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "搶票成功，訂單處理中",
		"order":   order,
	})
}

// GetStock 取得即時庫存
// GET /api/v1/events/:id/stock
func (h *TicketHandler) GetStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的活動 ID"})
		return
	}

	stock, err := h.uc.GetStock(c.Request.Context(), uint(id))
	if err != nil {
		status := apperror.HTTPStatus(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stock)
}

// GetOrder 查詢訂單
// GET /api/v1/orders/:id
func (h *TicketHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的訂單 ID"})
		return
	}

	view, err := h.uc.GetOrder(c.Request.Context(), uint(id))
	if err != nil {
		status := apperror.HTTPStatus(err)
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, view)
}

// GetUserOrders 查詢使用者的所有訂單
// GET /api/v1/users/:user_id/orders
func (h *TicketHandler) GetUserOrders(c *gin.Context) {
	userID := c.Param("user_id")

	views, err := h.uc.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, views)
}

// GetEvents 取得所有活動
// GET /api/v1/events
func (h *TicketHandler) GetEvents(c *gin.Context) {
	events, err := h.uc.GetAllEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}
