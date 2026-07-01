package campus

import (
	"context"
	"strconv"

	"school-system/internal/model"
	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- 请求结构体 ----

// CreateCampusReq 创建校区请求
type CreateCampusReq struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=hq normal"`
}

// UpdateCampusReq 编辑校区请求
type UpdateCampusReq struct {
	Name string `json:"name" binding:"required"`
}

// ---- 响应结构体 ----

// CampusResp 校区响应
type CampusResp struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ---- Service 接口 ----

// ServiceInterface 校区业务接口（供 Handler 依赖，便于单元测试 mock）
type ServiceInterface interface {
	List(ctx context.Context) ([]model.Campus, error)
	Create(ctx context.Context, name string, campusType string) (*model.Campus, error)
	Update(ctx context.Context, id uint, name string) (*model.Campus, error)
	Delete(ctx context.Context, id uint) error
}

// ---- Handler ----

// Handler 校区 HTTP 处理
type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler {
	return &Handler{svc: svc}
}

// List 获取校区列表
func (h *Handler) List(c *gin.Context) {
	campuses, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]CampusResp, 0, len(campuses))
	for _, v := range campuses {
		resp = append(resp, CampusResp{ID: v.ID, Name: v.Name, Type: v.Type})
	}

	response.OK(c, resp)
}

// Create 新建校区
func (h *Handler) Create(c *gin.Context) {
	var req CreateCampusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	campus, err := h.svc.Create(c.Request.Context(), req.Name, req.Type)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, CampusResp{ID: campus.ID, Name: campus.Name, Type: campus.Type})
}

// Update 编辑校区
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req UpdateCampusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	campus, err := h.svc.Update(c.Request.Context(), uint(id), req.Name)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, CampusResp{ID: campus.ID, Name: campus.Name, Type: campus.Type})
}

// Delete 删除校区
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}
