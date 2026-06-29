package auth

import (
	"strings"

	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler 认证模块 HTTP 处理
type Handler struct {
	svc *Service
}

// NewHandler 创建 Handler
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResp 登录响应
type LoginResp struct {
	Token string `json:"token"`
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
		CampusID uint   `json:"campus_id"`
	} `json:"user"`
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	token, user, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := LoginResp{Token: token}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	resp.User.Role = user.Role
	resp.User.CampusID = user.CampusID

	response.OK(c, resp)
}

// Refresh 刷新 Token
func (h *Handler) Refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Err(c, apperror.ErrUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		response.Err(c, apperror.ErrUnauthorized)
		return
	}

	newToken, err := h.svc.Refresh(tokenString)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, gin.H{"token": newToken})
}
