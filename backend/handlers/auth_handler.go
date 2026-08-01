package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yeung/system-hardening/backend/configs"
	"github.com/yeung/system-hardening/backend/services"
	"github.com/yeung/system-hardening/backend/utils"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token       string             `json:"token"`
	ExpiresIn   int                `json:"expires_in"`
	UserInfo    *UserInfoResponse  `json:"user_info"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Details  map[string]string `json:"details,omitempty"`
}

// AuthHandler 认证处理器结构
type AuthHandler struct {
	ldapService *services.LDAPService
	jwtConfig   configs.JWTConfig
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(ldapService *services.LDAPService, jwtConfig configs.JWTConfig) *AuthHandler {
	return &AuthHandler{
		ldapService: ldapService,
		jwtConfig:   jwtConfig,
	}
}

// LoginHandler 处理登录请求
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误：" + err.Error(),
		})
		return
	}

	// 验证用户名和密码
	authenticated, err := h.ldapService.AuthenticateUser(req.Username, req.Password)
	if !authenticated || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户名或密码错误，或无权限登录",
		})
		return
	}

	// 构建邮箱地址（使用域后缀）
	email := req.Username + "@" + h.jwtConfig.SecretKey // TODO: 从配置中获取域名部分

	// 生成 JWT token
	token, err := utils.GenerateToken(req.Username, email, "", h.jwtConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成认证 token 失败",
		})
		return
	}

	// 获取用户详细信息
	userDetails, _ := h.ldapService.GetUserDetails(req.Username)

	// 返回成功响应
	response := LoginResponse{
		Token:     token,
		ExpiresIn: h.jwtConfig.ExpiryHour * 3600, // 转换为秒
		UserInfo: &UserInfoResponse{
			Username: req.Username,
			Email:    email,
			Details:  userDetails,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetProfileHandler 获取当前用户信息
func (h *AuthHandler) GetProfileHandler(c *gin.Context) {
	username := c.GetString("username")
	email := c.GetString("email")

	userDetails, err := h.ldapService.GetUserDetails(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_info": gin.H{
			"username": username,
			"email":    email,
			"details":  userDetails,
		},
	})
}
