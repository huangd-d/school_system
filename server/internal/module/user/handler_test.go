package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/user"
	"school-system/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- 辅助函数 ----

// setupGin 返回 Gin 测试上下文，注入 auth 信息。
func setupGin(method, path string, body io.Reader) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, body)
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

// setAuth 注入操作人信息到 context。
func setAuth(c *gin.Context, userID, campusID uint, role string) {
	c.Set("user_id", userID)
	c.Set("campus_id", campusID)
	c.Set("role", role)
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
//  List — GET /api/v1/users
// ============================================================

func TestUserHandler_List_Success(t *testing.T) {
	mockSvc := &MockUserService{
		ListFn: func(ctx context.Context, operatorRole string, operatorCampusID uint) ([]model.User, error) {
			require.Equal(t, model.RoleHQAdmin, operatorRole) // 验证传入了正确的 role
			return []model.User{
				{
					ID: 1, Username: "admin", Phone: "13800138000",
					Role: model.RoleHQAdmin, CampusID: 1, Status: model.UserStatusActive,
					CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/users", nil)
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.List(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), "admin")
}

func TestUserHandler_List_Error(t *testing.T) {
	mockSvc := &MockUserService{
		ListFn: func(ctx context.Context, operatorRole string, operatorCampusID uint) ([]model.User, error) {
			return nil, apperror.ErrInternal
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/users", nil)
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.List(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInternal.Code, code)
}

// ============================================================
//  Create — POST /api/v1/users
// ============================================================

func TestUserHandler_Create_Success(t *testing.T) {
	mockSvc := &MockUserService{
		CreateFn: func(ctx context.Context, username, password, phone, role string, campusID uint) (*model.User, error) {
			assert.Equal(t, "newuser", username)
			return &model.User{
				ID: 10, Username: username, Phone: phone, Role: role, CampusID: campusID,
				Status: model.UserStatusActive,
				CreatedAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	h := user.NewHandler(mockSvc)

	reqBody := user.CreateUserReq{
		Username: "newuser",
		Password: "pass123",
		Phone:    "13800138000",
		Role:     model.RoleHQAdmin,
		CampusID: 1,
	}

	c, w := setupGin("POST", "/api/v1/users", toJSONBody(reqBody))

	h.Create(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), "newuser")
	assert.Contains(t, string(data), `"id":10`)
}

func TestUserHandler_Create_MissingRequired(t *testing.T) {
	h := user.NewHandler(&MockUserService{}) // service 不会被调用

	// 缺少 username（required）
	c, w := setupGin("POST", "/api/v1/users", toJSONBody(map[string]string{
		"password": "pass",
	}))

	h.Create(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestUserHandler_Create_InvalidRoleEnum(t *testing.T) {
	h := user.NewHandler(&MockUserService{})

	// role 不在枚举值内
	c, w := setupGin("POST", "/api/v1/users", toJSONBody(map[string]interface{}{
		"username":  "user",
		"password":  "pass",
		"role":      "bad_role",
		"campus_id": 1,
	}))

	h.Create(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestUserHandler_Create_ServiceError(t *testing.T) {
	mockSvc := &MockUserService{
		CreateFn: func(ctx context.Context, username, password, phone, role string, campusID uint) (*model.User, error) {
			return nil, apperror.New(apperror.ErrUserUsernameDup.Code, "用户名「dup」已存在")
		},
	}
	h := user.NewHandler(mockSvc)

	reqBody := user.CreateUserReq{
		Username: "dup",
		Password: "pass",
		Phone:    "123",
		Role:     model.RoleHQAdmin,
		CampusID: 1,
	}
	c, w := setupGin("POST", "/api/v1/users", toJSONBody(reqBody))

	h.Create(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, apperror.ErrUserUsernameDup.Code, code)
	assert.Contains(t, msg, "dup")
}

// ============================================================
//  Update — PUT /api/v1/users/:id
// ============================================================

func TestUserHandler_Update_Success(t *testing.T) {
	mockSvc := &MockUserService{
		UpdateFn: func(ctx context.Context, id uint, username, phone, role string, campusID uint) (*model.User, error) {
			assert.Equal(t, uint(5), id)
			assert.Equal(t, model.RoleCampusOperator, role)
			return &model.User{
				ID: 5, Username: "user", Phone: phone, Role: role, CampusID: campusID,
				Status: model.UserStatusActive,
				CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	h := user.NewHandler(mockSvc)

	reqBody := user.UpdateUserReq{Role: model.RoleCampusOperator, CampusID: 2}
	c, w := setupGin("PUT", "/api/v1/users/5", toJSONBody(reqBody))
	setParam(c, "id", "5")

	h.Update(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), `"role":"campus_operator"`)
	assert.Contains(t, string(data), `"campus_id":2`)
}

func TestUserHandler_Update_InvalidID(t *testing.T) {
	h := user.NewHandler(&MockUserService{})

	c, w := setupGin("PUT", "/api/v1/users/abc", toJSONBody(user.UpdateUserReq{
		Role: model.RoleHQAdmin, CampusID: 1,
	}))
	setParam(c, "id", "abc")

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestUserHandler_Update_UserNotFound(t *testing.T) {
	mockSvc := &MockUserService{
		UpdateFn: func(ctx context.Context, id uint, username, phone, role string, campusID uint) (*model.User, error) {
			return nil, apperror.ErrUserNotFound
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/users/999", toJSONBody(user.UpdateUserReq{
		Role: model.RoleHQAdmin, CampusID: 1,
	}))
	setParam(c, "id", "999")

	h.Update(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrUserNotFound.Code, code)
}

// ============================================================
//  Disable — PUT /api/v1/users/:id/disable
// ============================================================

func TestUserHandler_Disable_Success(t *testing.T) {
	mockSvc := &MockUserService{
		DisableFn: func(ctx context.Context, id uint, operatorID uint) error {
			assert.Equal(t, uint(3), id)
			assert.Equal(t, uint(1), operatorID)
			return nil
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/users/3/disable", nil)
	setAuth(c, 1, 1, model.RoleHQAdmin)
	setParam(c, "id", "3")

	h.Disable(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
}

func TestUserHandler_Disable_Self(t *testing.T) {
	mockSvc := &MockUserService{
		DisableFn: func(ctx context.Context, id uint, operatorID uint) error {
			return apperror.ErrUserDisableSelf
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/users/1/disable", nil)
	setAuth(c, 1, 1, model.RoleHQAdmin) // operatorID = 1
	setParam(c, "id", "1")               // target = 1 → 自己禁用自己

	h.Disable(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrUserDisableSelf.Code, code)
}

func TestUserHandler_Disable_InvalidID(t *testing.T) {
	h := user.NewHandler(&MockUserService{})

	c, w := setupGin("PUT", "/api/v1/users/xyz/disable", nil)
	setParam(c, "id", "xyz")

	h.Disable(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

// ============================================================
//  Enable — PUT /api/v1/users/:id/enable
// ============================================================

func TestUserHandler_Enable_Success(t *testing.T) {
	mockSvc := &MockUserService{
		EnableFn: func(ctx context.Context, id uint) error {
			assert.Equal(t, uint(3), id)
			return nil
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/users/3/enable", nil)
	setParam(c, "id", "3")

	h.Enable(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
}

func TestUserHandler_Enable_InvalidID(t *testing.T) {
	h := user.NewHandler(&MockUserService{})

	c, w := setupGin("PUT", "/api/v1/users/xyz/enable", nil)
	setParam(c, "id", "xyz")

	h.Enable(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestUserHandler_Enable_UserNotFound(t *testing.T) {
	mockSvc := &MockUserService{
		EnableFn: func(ctx context.Context, id uint) error {
			return apperror.ErrUserNotFound
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/users/999/enable", nil)
	setParam(c, "id", "999")

	h.Enable(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrUserNotFound.Code, code)
}

// ============================================================
//  ResetPassword — PUT /api/v1/users/:id/reset-pwd
// ============================================================

func TestUserHandler_ResetPassword_Success(t *testing.T) {
	mockSvc := &MockUserService{
		ResetPasswordFn: func(ctx context.Context, id uint, newPassword string) error {
			assert.Equal(t, uint(2), id)
			assert.Equal(t, "new_secret", newPassword)
			return nil
		},
	}
	h := user.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/users/2/reset-pwd", toJSONBody(user.ResetPasswordReq{
		Password: "new_secret",
	}))
	setParam(c, "id", "2")

	h.ResetPassword(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
}

func TestUserHandler_ResetPassword_Empty(t *testing.T) {
	h := user.NewHandler(&MockUserService{})

	// 缺少 required 字段 → binding error
	c, w := setupGin("PUT", "/api/v1/users/2/reset-pwd", toJSONBody(map[string]string{}))
	setParam(c, "id", "2")

	h.ResetPassword(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestUserHandler_ResetPassword_InvalidID(t *testing.T) {
	h := user.NewHandler(&MockUserService{})

	c, w := setupGin("PUT", "/api/v1/users/abc/reset-pwd", toJSONBody(user.ResetPasswordReq{
		Password: "pass",
	}))
	setParam(c, "id", "abc")

	h.ResetPassword(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}
