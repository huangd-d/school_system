package settlement

import (
	"context"
	"school-system/internal/model"
	"school-system/pkg/apperror"
	"school-system/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ServiceInterface 结算服务接口（便于测试 mock）
type ServiceInterface interface {
	Preview(ctx context.Context, activityID uint) (*PreviewResult, error)
	Execute(ctx context.Context, activityID uint, operatorID uint) (*model.Settlement, error)
	Reverse(ctx context.Context, settlementID uint, operatorID uint) error
	ListByActivity(ctx context.Context, activityID uint) ([]model.Settlement, error)
}

// 编译期校验 *Service 实现了 ServiceInterface
var _ ServiceInterface = (*Service)(nil)

// Handler 结算 HTTP 处理
type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler { return &Handler{svc: svc} }

// Preview 结算预览 POST /api/v1/settlements/preview/:activity_id
func (h *Handler) Preview(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("activity_id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	result, err := h.svc.Preview(c.Request.Context(), uint(activityID))
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, result)
}

// Execute 执行结算 POST /api/v1/settlements/execute/:activity_id
func (h *Handler) Execute(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("activity_id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	operatorID, _, _ := getOperator(c)

	settlement, err := h.svc.Execute(c.Request.Context(), uint(activityID), operatorID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, settlement)
}

// Reverse 回撤结算 POST /api/v1/settlements/reverse/:settlement_id
func (h *Handler) Reverse(c *gin.Context) {
	settlementID, err := strconv.ParseUint(c.Param("settlement_id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	operatorID, _, _ := getOperator(c)

	if err := h.svc.Reverse(c.Request.Context(), uint(settlementID), operatorID); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}

// ListByActivity 查询活动的结算记录 GET /api/v1/settlements/by-activity/:activity_id
func (h *Handler) ListByActivity(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("activity_id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	settlements, err := h.svc.ListByActivity(c.Request.Context(), uint(activityID))
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, settlements)
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
