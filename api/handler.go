package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"trading/business"
)

// StockHandler 股票数据 HTTP 处理器
type StockHandler struct {
	svc       business.StockService
	scheduler business.Scheduler
}

// NewStockHandler 创建 StockHandler
func NewStockHandler(svc business.StockService, scheduler business.Scheduler) *StockHandler {
	return &StockHandler{svc: svc, scheduler: scheduler}
}

// response 统一 JSON 响应结构
type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func respondSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, response{Code: 0, Message: "success", Data: data})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, response{Code: status, Message: message, Data: nil})
}
