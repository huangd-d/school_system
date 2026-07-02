package user

import (
	"context"
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
	Phone    string `json:"phone"`
	Role     string `json:"role" binding:"required,oneof=hq_admin campus_operator activity_contact"`
	CampusID uint   `json:"campus_id" binding:"required,min=1"`
}

type UpdateUserReq struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
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
	Phone     string `json:"phone"`
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

// checkHQAdmin 校验是否为总部管理员，非管理员返回 40003
func checkHQAdmin(c *gin.Context) bool {
	_, _, role := getOperator(c)
	if role != model.RoleHQAdmin {
		response.Err(c, apperror.ErrForbidden)
		return false
	}
	return true
}

// toResp 模型转响应
func toResp(u model.User) UserResp {
	return UserResp{
		ID:        u.ID,
		Username:  u.Username,
		Phone:     u.Phone,
		Role:      u.Role,
		CampusID:  u.CampusID,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ---- ServiceInterface（handler 依赖的服务接口，便于测试时注入 mock）----

type ServiceInterface interface {
	List(ctx context.Context, operatorRole string, operatorCampusID uint) ([]model.User, error)
	Create(ctx context.Context, username, password, phone, role string, campusID uint) (*model.User, error)
	Update(ctx context.Context, id uint, username, phone, role string, campusID uint) (*model.User, error)
	Disable(ctx context.Context, id uint, operatorID uint) error
	Enable(ctx context.Context, id uint) error
	ResetPassword(ctx context.Context, id uint, newPassword string) error
}

// ---- Handler ----

type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler {
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
	if !checkHQAdmin(c) {
		return
	}

	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	user, err := h.svc.Create(c.Request.Context(), req.Username, req.Password, req.Phone, req.Role, req.CampusID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toResp(*user))
}

// Update 编辑账户
func (h *Handler) Update(c *gin.Context) {
	if !checkHQAdmin(c) {
		return
	}

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

	user, err := h.svc.Update(c.Request.Context(), uint(id), req.Username, req.Phone, req.Role, req.CampusID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toResp(*user))
}

// Disable 禁用账户
func (h *Handler) Disable(c *gin.Context) {
	if !checkHQAdmin(c) {
		return
	}

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

// Enable 启用账户
func (h *Handler) Enable(c *gin.Context) {
	if !checkHQAdmin(c) {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	if err := h.svc.Enable(c.Request.Context(), uint(id)); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}

// ResetPassword 重置密码
func (h *Handler) ResetPassword(c *gin.Context) {
	if !checkHQAdmin(c) {
		return
	}

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
