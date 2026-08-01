package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/controllers"
)

type ClientHandler struct {
	controller *controllers.ClientController
}

// NewClientHandler 创建客户端处理器
func NewClientHandler(controller *controllers.ClientController) *ClientHandler {
	return &ClientHandler{controller: controller}
}

// RequestTempToken 处理请求临时 Token
func (h *ClientHandler) RequestTempToken(c *gin.Context) {
	h.controller.RequestTempToken(c)
}

// Register 处理客户端注册
func (h *ClientHandler) Register(c *gin.Context) {
	h.controller.Register(c)
}

// RefreshToken 处理 Token 刷新
func (h *ClientHandler) RefreshToken(c *gin.Context) {
	h.controller.RefreshToken(c)
}

// UploadData 处理加固数据上传
func (h *ClientHandler) UploadData(c *gin.Context) {
	h.controller.UploadData(c)
}
