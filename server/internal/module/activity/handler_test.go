package activity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/activity"
	"school-system/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---- 辅助函数 ----

func setupGin(method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func setAuth(c *gin.Context, userID, campusID uint, role string) {
	c.Set("user_id", userID)
	c.Set("campus_id", campusID)
	c.Set("role", role)
}

func setParam(c *gin.Context, key, value string) {
	c.Params = []gin.Param{{Key: key, Value: value}}
}

type respBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func parseResp(w *httptest.ResponseRecorder) (code int, message string, data json.RawMessage) {
	var r respBody
	json.Unmarshal(w.Body.Bytes(), &r)
	return r.Code, r.Message, r.Data
}

func toJSONBody(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

// ---- 测试数据 ----

func sampleActivity() model.Activity {
	return model.Activity{
		ID: 1, Name: "测试活动", CampusID: 1, PlannedExecutions: 10,
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:    model.ActivityNotStarted, CreatedBy: 1,
		CreatedAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

// ============================================================
//  List — GET /api/v1/activities
// ============================================================

func TestActivityHandler_List_Success(t *testing.T) {
	mockSvc := &MockActivityService{
		ListFn: func(ctx context.Context, userID, campusID uint, role string) ([]activity.ActivityWithSummary, error) {
			assert.Equal(t, model.RoleHQAdmin, role)
			return []activity.ActivityWithSummary{
				{
					Activity:      sampleActivity(),
					ContactIDs:    []uint{2, 3},
					TotalExecuted: 5,
				},
			}, nil
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/activities", nil)
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.List(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), "测试活动")
	assert.Contains(t, string(data), `"total_executed":5`)
	assert.Contains(t, string(data), `"contact_ids":[2,3]`)
}

func TestActivityHandler_List_Error(t *testing.T) {
	mockSvc := &MockActivityService{
		ListFn: func(ctx context.Context, userID, campusID uint, role string) ([]activity.ActivityWithSummary, error) {
			return nil, apperror.ErrInternal
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/activities", nil)
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.List(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInternal.Code, code)
}

// ============================================================
//  Create — POST /api/v1/activities
// ============================================================

func TestActivityHandler_Create_Success(t *testing.T) {
	mockSvc := &MockActivityService{
		CreateFn: func(ctx context.Context, name string, campusID uint, contactIDs []uint, plannedExec int, startDate, endDate string, createdBy uint, creatorRole string, creatorCampusID uint) (*model.Activity, error) {
			assert.Equal(t, "新活动", name)
			assert.Equal(t, uint(1), campusID)
			assert.Equal(t, 10, plannedExec)
			a := sampleActivity()
			a.Name = name
			return &a, nil
		},
	}
	h := activity.NewHandler(mockSvc)

	reqBody := activity.CreateActivityReq{
		Name:              "新活动",
		CampusID:          1,
		ContactIDs:        []uint{2},
		PlannedExecutions: 10,
		StartDate:         "2025-01-01",
		EndDate:           "2025-12-31",
	}

	c, w := setupGin("POST", "/api/v1/activities", toJSONBody(reqBody))
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Create(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), "新活动")
}

func TestActivityHandler_Create_MissingRequired(t *testing.T) {
	h := activity.NewHandler(&MockActivityService{})

	c, w := setupGin("POST", "/api/v1/activities", toJSONBody(map[string]string{
		"campus_id": "1",
	}))
	// 缺少 name、planned_executions、start_date、end_date

	h.Create(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestActivityHandler_Create_ServiceError(t *testing.T) {
	mockSvc := &MockActivityService{
		CreateFn: func(ctx context.Context, name string, campusID uint, contactIDs []uint, plannedExec int, startDate, endDate string, createdBy uint, creatorRole string, creatorCampusID uint) (*model.Activity, error) {
			return nil, apperror.ErrActivityCampusNotFound
		},
	}
	h := activity.NewHandler(mockSvc)

	reqBody := activity.CreateActivityReq{
		Name:              "活动",
		CampusID:          99,
		PlannedExecutions: 5,
		StartDate:         "2025-01-01",
		EndDate:           "2025-12-31",
	}
	c, w := setupGin("POST", "/api/v1/activities", toJSONBody(reqBody))

	h.Create(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, apperror.ErrActivityCampusNotFound.Code, code)
	assert.Contains(t, msg, "校区不存在")
}

// ============================================================
//  Update — PUT /api/v1/activities/:id
// ============================================================

func TestActivityHandler_Update_Success(t *testing.T) {
	mockSvc := &MockActivityService{
		UpdateFn: func(ctx context.Context, id uint, name string, contactIDs []uint, plannedExec int, operatorRole string, operatorCampusID uint) (*model.Activity, error) {
			assert.Equal(t, uint(5), id)
			a := sampleActivity()
			a.ID = 5
			a.Name = name
			return &a, nil
		},
	}
	h := activity.NewHandler(mockSvc)

	reqBody := activity.UpdateActivityReq{Name: "已更名"}
	c, w := setupGin("PUT", "/api/v1/activities/5", toJSONBody(reqBody))
	setParam(c, "id", "5")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Update(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), "已更名")
}

func TestActivityHandler_Update_InvalidID(t *testing.T) {
	h := activity.NewHandler(&MockActivityService{})

	c, w := setupGin("PUT", "/api/v1/activities/abc", toJSONBody(activity.UpdateActivityReq{Name: "x"}))
	setParam(c, "id", "abc")

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestActivityHandler_Update_NotFound(t *testing.T) {
	mockSvc := &MockActivityService{
		UpdateFn: func(ctx context.Context, id uint, name string, contactIDs []uint, plannedExec int, operatorRole string, operatorCampusID uint) (*model.Activity, error) {
			return nil, apperror.ErrActivityNotFound
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/activities/999", toJSONBody(activity.UpdateActivityReq{Name: "x"}))
	setParam(c, "id", "999")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrActivityNotFound.Code, code)
}

// ============================================================
//  Detail — GET /api/v1/activities/:id
// ============================================================

func TestActivityHandler_Detail_Success(t *testing.T) {
	mockSvc := &MockActivityService{
		DetailFn: func(ctx context.Context, id uint) (*activity.ActivityDetail, error) {
			assert.Equal(t, uint(1), id)
			return &activity.ActivityDetail{
				Activity:      sampleActivity(),
				Contacts:      []model.User{{ID: 2, Username: "联系人A", Phone: "138", Role: model.RoleActivityContact}},
				Executions:    []model.ExecutionRecord{{ID: 1, ActivityID: 1, Count: 3, RecordedBy: 2}},
				TotalExecuted: 3,
			}, nil
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/activities/1", nil)
	setParam(c, "id", "1")

	h.Detail(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), "测试活动")
	assert.Contains(t, string(data), "联系人A")
	assert.Contains(t, string(data), `"total_executed":3`)
}

func TestActivityHandler_Detail_InvalidID(t *testing.T) {
	h := activity.NewHandler(&MockActivityService{})

	c, w := setupGin("GET", "/api/v1/activities/xyz", nil)
	setParam(c, "id", "xyz")

	h.Detail(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestActivityHandler_Detail_NotFound(t *testing.T) {
	mockSvc := &MockActivityService{
		DetailFn: func(ctx context.Context, id uint) (*activity.ActivityDetail, error) {
			return nil, apperror.ErrActivityNotFound
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/activities/999", nil)
	setParam(c, "id", "999")

	h.Detail(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrActivityNotFound.Code, code)
}

// ============================================================
//  AddExecution — POST /api/v1/activities/:id/executions
// ============================================================

func TestActivityHandler_AddExecution_Success(t *testing.T) {
	mockSvc := &MockActivityService{
		AddExecutionFn: func(ctx context.Context, activityID uint, count int, recordedBy uint, operatorRole string) error {
			assert.Equal(t, uint(3), activityID)
			assert.Equal(t, 2, count)
			assert.Equal(t, uint(1), recordedBy)
			return nil
		},
	}
	h := activity.NewHandler(mockSvc)

	reqBody := activity.AddExecutionReq{Count: 2}
	c, w := setupGin("POST", "/api/v1/activities/3/executions", toJSONBody(reqBody))
	setParam(c, "id", "3")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.AddExecution(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
}

func TestActivityHandler_AddExecution_InvalidID(t *testing.T) {
	h := activity.NewHandler(&MockActivityService{})

	c, w := setupGin("POST", "/api/v1/activities/abc/executions", toJSONBody(activity.AddExecutionReq{Count: 1}))
	setParam(c, "id", "abc")

	h.AddExecution(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestActivityHandler_AddExecution_MissingCount(t *testing.T) {
	h := activity.NewHandler(&MockActivityService{})

	c, w := setupGin("POST", "/api/v1/activities/1/executions", toJSONBody(map[string]string{}))
	setParam(c, "id", "1")

	h.AddExecution(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestActivityHandler_AddExecution_ServiceError(t *testing.T) {
	mockSvc := &MockActivityService{
		AddExecutionFn: func(ctx context.Context, activityID uint, count int, recordedBy uint, operatorRole string) error {
			return apperror.ErrActivityExecExceedPlanned
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/activities/1/executions", toJSONBody(activity.AddExecutionReq{Count: 10}))
	setParam(c, "id", "1")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.AddExecution(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrActivityExecExceedPlanned.Code, code)
}

// ============================================================
//  Archive — PUT /api/v1/activities/:id/archive
// ============================================================

func TestActivityHandler_Archive_Success(t *testing.T) {
	mockSvc := &MockActivityService{
		ArchiveFn: func(ctx context.Context, id uint, operatorRole string, operatorCampusID uint) error {
			assert.Equal(t, uint(3), id)
			assert.Equal(t, model.RoleHQAdmin, operatorRole)
			return nil
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/activities/3/archive", nil)
	setParam(c, "id", "3")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Archive(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
}

func TestActivityHandler_Archive_InvalidID(t *testing.T) {
	h := activity.NewHandler(&MockActivityService{})

	c, w := setupGin("PUT", "/api/v1/activities/xyz/archive", nil)
	setParam(c, "id", "xyz")

	h.Archive(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestActivityHandler_Archive_NotSettled(t *testing.T) {
	mockSvc := &MockActivityService{
		ArchiveFn: func(ctx context.Context, id uint, operatorRole string, operatorCampusID uint) error {
			return apperror.ErrActivityNotSettled
		},
	}
	h := activity.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/activities/1/archive", nil)
	setParam(c, "id", "1")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Archive(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrActivityNotSettled.Code, code)
}
