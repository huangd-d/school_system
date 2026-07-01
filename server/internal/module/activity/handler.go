package activity

import (
	"context"
	"strconv"

	"school-system/internal/model"
	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- 请求结构体 ----

type CreateActivityReq struct {
	Name              string `json:"name" binding:"required"`
	CampusID          uint   `json:"campus_id" binding:"required,min=1"`
	ContactIDs        []uint `json:"contact_ids"`
	PlannedExecutions int    `json:"planned_executions" binding:"required,min=1"`
	StartDate         string `json:"start_date" binding:"required"`
	EndDate           string `json:"end_date" binding:"required"`
}

type UpdateActivityReq struct {
	Name              string `json:"name"`
	ContactIDs        []uint `json:"contact_ids"`
	PlannedExecutions int    `json:"planned_executions"`
}

type AddExecutionReq struct {
	Count int `json:"count" binding:"required,min=1"`
}

// ---- 响应结构体 ----

type ActivityResp struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	CampusID          uint   `json:"campus_id"`
	PlannedExecutions int    `json:"planned_executions"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	Status            string `json:"status"`
	CreatedBy         uint   `json:"created_by"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type ActivityListResp struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	CampusID          uint   `json:"campus_id"`
	PlannedExecutions int    `json:"planned_executions"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	Status            string `json:"status"`
	CreatedBy         uint   `json:"created_by"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	ContactIDs        []uint `json:"contact_ids"`
	TotalExecuted     int    `json:"total_executed"`
}

type ActivityDetailResp struct {
	ID                uint              `json:"id"`
	Name              string            `json:"name"`
	CampusID          uint              `json:"campus_id"`
	PlannedExecutions int               `json:"planned_executions"`
	StartDate         string            `json:"start_date"`
	EndDate           string            `json:"end_date"`
	Status            string            `json:"status"`
	CreatedBy         uint              `json:"created_by"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
	Contacts          []UserBriefResp   `json:"contacts"`
	Executions        []ExecutionResp   `json:"executions"`
	TotalExecuted     int               `json:"total_executed"`
}

type UserBriefResp struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type ExecutionResp struct {
	ID         uint   `json:"id"`
	ActivityID uint   `json:"activity_id"`
	Count      int    `json:"count"`
	RecordedBy uint   `json:"recorded_by"`
	CreatedAt  string `json:"created_at"`
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

func toActivityResp(a model.Activity) ActivityResp {
	return ActivityResp{
		ID:                a.ID,
		Name:              a.Name,
		CampusID:          a.CampusID,
		PlannedExecutions: a.PlannedExecutions,
		StartDate:         a.StartDate.Format("2006-01-02"),
		EndDate:           a.EndDate.Format("2006-01-02"),
		Status:            a.Status,
		CreatedBy:         a.CreatedBy,
		CreatedAt:         a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toActivityListResp(s ActivityWithSummary) ActivityListResp {
	return ActivityListResp{
		ID:                s.ID,
		Name:              s.Name,
		CampusID:          s.CampusID,
		PlannedExecutions: s.PlannedExecutions,
		StartDate:         s.StartDate.Format("2006-01-02"),
		EndDate:           s.EndDate.Format("2006-01-02"),
		Status:            s.Status,
		CreatedBy:         s.CreatedBy,
		CreatedAt:         s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         s.UpdatedAt.Format("2006-01-02 15:04:05"),
		ContactIDs:        s.ContactIDs,
		TotalExecuted:     s.TotalExecuted,
	}
}

func toActivityDetailResp(d ActivityDetail) ActivityDetailResp {
	contacts := make([]UserBriefResp, 0, len(d.Contacts))
	for _, u := range d.Contacts {
		contacts = append(contacts, UserBriefResp{
			ID:       u.ID,
			Username: u.Username,
			Phone:    u.Phone,
			Role:     u.Role,
		})
	}

	executions := make([]ExecutionResp, 0, len(d.Executions))
	for _, e := range d.Executions {
		executions = append(executions, ExecutionResp{
			ID:         e.ID,
			ActivityID: e.ActivityID,
			Count:      e.Count,
			RecordedBy: e.RecordedBy,
			CreatedAt:  e.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return ActivityDetailResp{
		ID:                d.ID,
		Name:              d.Name,
		CampusID:          d.CampusID,
		PlannedExecutions: d.PlannedExecutions,
		StartDate:         d.StartDate.Format("2006-01-02"),
		EndDate:           d.EndDate.Format("2006-01-02"),
		Status:            d.Status,
		CreatedBy:         d.CreatedBy,
		CreatedAt:         d.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:         d.UpdatedAt.Format("2006-01-02 15:04:05"),
		Contacts:          contacts,
		Executions:        executions,
		TotalExecuted:     d.TotalExecuted,
	}
}

// ---- ServiceInterface（handler 依赖的服务接口，便于测试时注入 mock）----

type ServiceInterface interface {
	List(ctx context.Context, userID, campusID uint, role string) ([]ActivityWithSummary, error)
	Create(ctx context.Context, name string, campusID uint, contactIDs []uint,
		plannedExec int, startDate, endDate string, createdBy uint,
		creatorRole string, creatorCampusID uint) (*model.Activity, error)
	Update(ctx context.Context, id uint, name string, contactIDs []uint,
		plannedExec int, operatorRole string, operatorCampusID uint) (*model.Activity, error)
	Detail(ctx context.Context, id uint) (*ActivityDetail, error)
	AddExecution(ctx context.Context, activityID uint, count int, recordedBy uint, operatorRole string) error
	Archive(ctx context.Context, id uint, operatorRole string, operatorCampusID uint) error
}

// ---- Handler ----

type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler {
	return &Handler{svc: svc}
}

// List 获取活动列表
func (h *Handler) List(c *gin.Context) {
	userID, campusID, role := getOperator(c)

	activities, err := h.svc.List(c.Request.Context(), userID, campusID, role)
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]ActivityListResp, 0, len(activities))
	for _, a := range activities {
		resp = append(resp, toActivityListResp(a))
	}
	response.OK(c, resp)
}

// Create 新建活动
func (h *Handler) Create(c *gin.Context) {
	var req CreateActivityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	userID, campusID, role := getOperator(c)

	activity, err := h.svc.Create(
		c.Request.Context(),
		req.Name, req.CampusID, req.ContactIDs,
		req.PlannedExecutions, req.StartDate, req.EndDate,
		userID, role, campusID,
	)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toActivityResp(*activity))
}

// Update 编辑活动
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req UpdateActivityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	_, campusID, role := getOperator(c)

	activity, err := h.svc.Update(
		c.Request.Context(),
		uint(id), req.Name, req.ContactIDs,
		req.PlannedExecutions, role, campusID,
	)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toActivityResp(*activity))
}

// Detail 获取活动详情
func (h *Handler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	detail, err := h.svc.Detail(c.Request.Context(), uint(id))
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toActivityDetailResp(*detail))
}

// AddExecution 录入执行次数
func (h *Handler) AddExecution(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req AddExecutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	userID, _, role := getOperator(c)

	if err := h.svc.AddExecution(c.Request.Context(), uint(id), req.Count, userID, role); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}

// Archive 归档活动
func (h *Handler) Archive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	_, campusID, role := getOperator(c)

	if err := h.svc.Archive(c.Request.Context(), uint(id), role, campusID); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}
