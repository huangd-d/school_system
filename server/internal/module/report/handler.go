package report

import (
	"context"
	"school-system/pkg/apperror"
	"school-system/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceInterface 报表服务接口（便于测试 mock）
type ServiceInterface interface {
	ByActivity(ctx context.Context, activityID uint) (*ActivityReport, error)
	ByDateRange(ctx context.Context, start, end time.Time) ([]DateRangeItem, error)
	ByCampus(ctx context.Context, campusID uint, start, end time.Time) (*CampusReport, error)
	ByCategory(ctx context.Context, start, end time.Time) ([]CategoryReportItem, error)
}

// 编译期校验 *Service 实现了 ServiceInterface
var _ ServiceInterface = (*Service)(nil)

// Handler 报表 HTTP 处理
type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler { return &Handler{svc: svc} }

// ByActivity 按活动维度 GET /api/v1/reports/by-activity?activity_id=1
func (h *Handler) ByActivity(c *gin.Context) {
	activityIDStr := c.Query("activity_id")
	activityID, err := strconv.ParseUint(activityIDStr, 10, 64)
	if err != nil || activityID == 0 {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	result, err := h.svc.ByActivity(c.Request.Context(), uint(activityID))
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, result)
}

// ByDateRange 按日期范围维度 GET /api/v1/reports/by-date-range?start_date=2024-01-01&end_date=2024-12-31
func (h *Handler) ByDateRange(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	result, err := h.svc.ByDateRange(c.Request.Context(), start, end)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, result)
}

// ByCampus 按校区维度 GET /api/v1/reports/by-campus?campus_id=1&start_date=...&end_date=...
func (h *Handler) ByCampus(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	var campusID uint
	if cidStr := c.Query("campus_id"); cidStr != "" {
		cid, parseErr := strconv.ParseUint(cidStr, 10, 64)
		if parseErr != nil {
			response.Err(c, apperror.ErrInvalidParam)
			return
		}
		campusID = uint(cid)
	}

	result, err := h.svc.ByCampus(c.Request.Context(), campusID, start, end)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, result)
}

// ByCategory 按品类维度 GET /api/v1/reports/by-category?start_date=...&end_date=...
func (h *Handler) ByCategory(c *gin.Context) {
	start, end, err := parseDateRange(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	result, err := h.svc.ByCategory(c.Request.Context(), start, end)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, result)
}

// ---- 辅助函数 ----

// parseDateRange 从查询参数解析日期范围
func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, apperror.New(apperror.ErrInvalidParam.Code, "缺少日期参数")
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return time.Time{}, time.Time{}, apperror.New(apperror.ErrInvalidParam.Code, "日期格式无效，应为 YYYY-MM-DD")
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return time.Time{}, time.Time{}, apperror.New(apperror.ErrInvalidParam.Code, "日期格式无效，应为 YYYY-MM-DD")
	}

	// end 设为当天末尾
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())

	return start, end, nil
}
