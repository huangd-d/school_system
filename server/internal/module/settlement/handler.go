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
	Overview(ctx context.Context) ([]SettlementOverviewItem, error)
}

// 编译期校验 *Service 实现了 ServiceInterface
var _ ServiceInterface = (*Service)(nil)

// SettlementResp 结算记录响应（时间字段统一 yyyy-MM-dd，金额单位：分）
type SettlementResp struct {
	ID                  uint   `json:"id"`
	ActivityID          uint   `json:"activity_id"`
	Status              string `json:"status"`
	OperatorID          uint   `json:"operator_id"`
	TotalReturnedAmount int64  `json:"total_returned_amount"`
	CreatedAt           string `json:"created_at"`
}

func toSettlementResp(s model.Settlement) SettlementResp {
	return SettlementResp{
		ID:                  s.ID,
		ActivityID:          s.ActivityID,
		Status:              s.Status,
		OperatorID:          s.OperatorID,
		TotalReturnedAmount: s.TotalReturnedAmount,
		CreatedAt:           s.CreatedAt.Format("2006-01-02"),
	}
}

// Handler 结算 HTTP 处理
type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler { return &Handler{svc: svc} }

// Preview 结算预览 GET /api/v1/settlements/preview/:activity_id
func (h *Handler) Preview(c *gin.Context) {
	if _, _, role := getOperator(c); role != model.RoleHQAdmin {
		response.Err(c, apperror.ErrSettlementPermDenied)
		return
	}

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
	operatorID, _, role := getOperator(c)
	if role != model.RoleHQAdmin {
		response.Err(c, apperror.ErrSettlementPermDenied)
		return
	}

	activityID, err := strconv.ParseUint(c.Param("activity_id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	settlement, err := h.svc.Execute(c.Request.Context(), uint(activityID), operatorID)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toSettlementResp(*settlement))
}

// Reverse 回撤结算 POST /api/v1/settlements/reverse/:settlement_id
func (h *Handler) Reverse(c *gin.Context) {
	operatorID, _, role := getOperator(c)
	if role != model.RoleHQAdmin {
		response.Err(c, apperror.ErrSettlementPermDenied)
		return
	}

	settlementID, err := strconv.ParseUint(c.Param("settlement_id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	if err := h.svc.Reverse(c.Request.Context(), uint(settlementID), operatorID); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}

// ListByActivity 查询活动的结算记录 GET /api/v1/settlements/by-activity/:activity_id
func (h *Handler) ListByActivity(c *gin.Context) {
	if _, _, role := getOperator(c); role != model.RoleHQAdmin {
		response.Err(c, apperror.ErrSettlementPermDenied)
		return
	}

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

	resp := make([]SettlementResp, 0, len(settlements))
	for _, s := range settlements {
		resp = append(resp, toSettlementResp(s))
	}

	response.OK(c, resp)
}

// Overview 结算管理概览 GET /api/v1/settlements/overview
func (h *Handler) Overview(c *gin.Context) {
	if _, _, role := getOperator(c); role != model.RoleHQAdmin {
		response.Err(c, apperror.ErrSettlementPermDenied)
		return
	}

	items, err := h.svc.Overview(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, items)
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
