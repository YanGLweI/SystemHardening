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

// UploadWindowsData 处理 Windows 加固数据上传
func (h *ClientHandler) UploadWindowsData(c *gin.Context) {
	h.controller.UploadWindowsData(c)
}

// Heartbeat 处理心跳请求
func (h *ClientHandler) Heartbeat(c *gin.Context) {
	h.controller.Heartbeat(c)
}

// CheckUpdate 处理检查更新请求
func (h *ClientHandler) CheckUpdate(c *gin.Context) {
	h.controller.CheckUpdate(c)
}

// GetCheckSchedule 处理客户端获取加固检查计划请求
func (h *ClientHandler) GetCheckSchedule(c *gin.Context) {
	h.controller.GetCheckScheduleForClient(c)
}
