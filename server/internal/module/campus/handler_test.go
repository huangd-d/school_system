package campus_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/campus"
	"school-system/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---- 辅助函数 ----

// setupGin 返回 Gin 测试上下文。
func setupGin(method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

// setParam 设置 URL 路径参数（如 :id）。
func setParam(c *gin.Context, key, value string) {
	c.Params = []gin.Param{{Key: key, Value: value}}
}

// respBody 统一响应 JSON 结构（用于解析）。
type respBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// parseResp 解析统一响应 JSON。
func parseResp(w *httptest.ResponseRecorder) (code int, message string, data json.RawMessage) {
	var r respBody
	json.Unmarshal(w.Body.Bytes(), &r)
	return r.Code, r.Message, r.Data
}

// toJSONBody 将结构体转为 io.Reader。
func toJSONBody(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

// ============================================================
//  List — GET /api/v1/campuses
// ============================================================

func TestCampusHandler_List_Success(t *testing.T) {
	mockSvc := &MockCampusService{
		ListFn: func(ctx context.Context) ([]model.Campus, error) {
			return []model.Campus{
				{ID: 1, Name: "总部", Type: model.CampusTypeHQ},
				{ID: 2, Name: "校区A", Type: model.CampusTypeNormal},
			}, nil
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/campuses", nil)
	h.List(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), "总部")
	assert.Contains(t, string(data), "校区A")
}

func TestCampusHandler_List_Error(t *testing.T) {
	mockSvc := &MockCampusService{
		ListFn: func(ctx context.Context) ([]model.Campus, error) {
			return nil, apperror.ErrInternal
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/campuses", nil)
	h.List(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInternal.Code, code)
}

// ============================================================
//  Create — POST /api/v1/campuses
// ============================================================

func TestCampusHandler_Create_Success(t *testing.T) {
	mockSvc := &MockCampusService{
		CreateFn: func(ctx context.Context, name string, campusType string) (*model.Campus, error) {
			assert.Equal(t, "新校区", name)
			assert.Equal(t, model.CampusTypeNormal, campusType)
			return &model.Campus{ID: 5, Name: name, Type: campusType}, nil
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/campuses", toJSONBody(map[string]string{
		"name": "新校区",
		"type": "normal",
	}))

	h.Create(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), `"name":"新校区"`)
	assert.Contains(t, string(data), `"type":"normal"`)
	assert.Contains(t, string(data), `"id":5`)
}

func TestCampusHandler_Create_MissingRequired(t *testing.T) {
	h := campus.NewHandler(&MockCampusService{}) // service 不会被调用

	// 缺少 name 字段（required）
	c, w := setupGin("POST", "/api/v1/campuses", toJSONBody(map[string]string{
		"type": "normal",
	}))

	h.Create(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestCampusHandler_Create_InvalidType(t *testing.T) {
	h := campus.NewHandler(&MockCampusService{})

	// type 不在枚举值内（oneof=hq normal）
	c, w := setupGin("POST", "/api/v1/campuses", toJSONBody(map[string]string{
		"name": "新校区",
		"type": "bad_value",
	}))

	h.Create(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestCampusHandler_Create_ServiceError(t *testing.T) {
	mockSvc := &MockCampusService{
		CreateFn: func(ctx context.Context, name string, campusType string) (*model.Campus, error) {
			return nil, apperror.New(apperror.ErrCampusNameDup.Code, "校区名称「已存在」已存在")
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/campuses", toJSONBody(map[string]string{
		"name": "已存在",
		"type": "normal",
	}))

	h.Create(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, apperror.ErrCampusNameDup.Code, code)
	assert.Contains(t, msg, "已存在")
}

// ============================================================
//  Update — PUT /api/v1/campuses/:id
// ============================================================

func TestCampusHandler_Update_Success(t *testing.T) {
	mockSvc := &MockCampusService{
		UpdateFn: func(ctx context.Context, id uint, name string) (*model.Campus, error) {
			assert.Equal(t, uint(3), id)
			assert.Equal(t, "新名称", name)
			return &model.Campus{ID: 3, Name: name, Type: model.CampusTypeNormal}, nil
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/campuses/3", toJSONBody(map[string]string{
		"name": "新名称",
	}))
	setParam(c, "id", "3")

	h.Update(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), `"name":"新名称"`)
	assert.Contains(t, string(data), `"id":3`)
}

func TestCampusHandler_Update_InvalidID(t *testing.T) {
	h := campus.NewHandler(&MockCampusService{})

	c, w := setupGin("PUT", "/api/v1/campuses/abc", toJSONBody(map[string]string{
		"name": "新名称",
	}))
	setParam(c, "id", "abc")

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestCampusHandler_Update_MissingName(t *testing.T) {
	h := campus.NewHandler(&MockCampusService{})

	// 缺少 name 字段（required）
	c, w := setupGin("PUT", "/api/v1/campuses/3", toJSONBody(map[string]string{}))
	setParam(c, "id", "3")

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestCampusHandler_Update_ServiceError(t *testing.T) {
	mockSvc := &MockCampusService{
		UpdateFn: func(ctx context.Context, id uint, name string) (*model.Campus, error) {
			return nil, apperror.ErrCampusNotFound
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/campuses/999", toJSONBody(map[string]string{
		"name": "新名称",
	}))
	setParam(c, "id", "999")

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrCampusNotFound.Code, code)
}

// ============================================================
//  Delete — DELETE /api/v1/campuses/:id
// ============================================================

func TestCampusHandler_Delete_Success(t *testing.T) {
	mockSvc := &MockCampusService{
		DeleteFn: func(ctx context.Context, id uint) error {
			assert.Equal(t, uint(3), id)
			return nil
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("DELETE", "/api/v1/campuses/3", nil)
	setParam(c, "id", "3")

	h.Delete(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	// data 应为 null
	assert.Equal(t, "null", string(data))
}

func TestCampusHandler_Delete_InvalidID(t *testing.T) {
	h := campus.NewHandler(&MockCampusService{})

	c, w := setupGin("DELETE", "/api/v1/campuses/abc", nil)
	setParam(c, "id", "abc")

	h.Delete(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestCampusHandler_Delete_ServiceError(t *testing.T) {
	mockSvc := &MockCampusService{
		DeleteFn: func(ctx context.Context, id uint) error {
			return apperror.ErrCampusHQDelete
		},
	}
	h := campus.NewHandler(mockSvc)

	c, w := setupGin("DELETE", "/api/v1/campuses/1", nil)
	setParam(c, "id", "1")

	h.Delete(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrCampusHQDelete.Code, code)
}
