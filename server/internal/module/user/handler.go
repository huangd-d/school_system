package user

import (
	"strconv"

	"school-system/internal/model"
	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- 请求结构体 ----

type CreateUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=hq_admin campus_operator activity_contact"`
	CampusID uint   `json:"campus_id" binding:"required,min=1"`
}

type UpdateUserReq struct {
	Role     string `json:"role" binding:"required,oneof=hq_admin campus_operator activity_contact"`
	CampusID uint   `json:"campus_id" binding:"required,min=1"`
}

type ResetPasswordReq struct {
	Password string `json:"password" binding:"required"`
}

// ---- 响应结构体 ----

type UserResp struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CampusID  uint   `json:"campus_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ---- 辅助函数 ----

// getOperator 从 context 获取当前操作人信息（由 auth 中间件注入）
func getOperator(c *gin.Context) (userID, campusID uint, role string) {
	if v, ok := c.Get("user_id"); ok {
		userID = v.(uint)
	}
	if v, ok := c.Get("campus_id"); ok {
		campusID = v.(uint)
	}
	if v, ok := c.Get("role"); ok {
		role = v.(string)
	}
	return
}

// toResp 模型转响应
func toResp(u model.User) UserResp {
	return UserResp{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CampusID:  u.CampusID,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ---- Handler ----

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List 获取账户列表
func (h *Handler) List(c *gin.Context) {
	_, campusID, role := getOperator(c)
	users, err := h.svc.List(c.Request.Context(), role, campusID)
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]UserResp, 0, len(users))
	for _, u := range users {
		resp = append(resp, toResp(u))
	}
	response.OK(c, resp)
}

// Create 新建账户
func (h *Handler) Create(c *gin.Context) {
	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	user, err := h.svc.Create(c.Request.Context(), req.Username, req.Password, req.Role, req.CampusID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toResp(*user))
}

// Update 编辑账户
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	user, err := h.svc.Update(c.Request.Context(), uint(id), req.Role, req.CampusID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toResp(*user))
}

// Disable 禁用账户
func (h *Handler) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	operatorID, _, _ := getOperator(c)
	if err := h.svc.Disable(c.Request.Context(), uint(id), operatorID); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}

// ResetPassword 重置密码
func (h *Handler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	if err := h.svc.ResetPassword(c.Request.Context(), uint(id), req.Password); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}
