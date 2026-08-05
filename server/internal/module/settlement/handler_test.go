package settlement_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/settlement"
	"school-system/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---- 辅助函数 ----

func setupGin(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
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

func parseResp(w *httptest.ResponseRecorder) int {
	var r struct {
		Code int `json:"code"`
	}
	json.Unmarshal(w.Body.Bytes(), &r)
	return r.Code
}

// ---- MockService ----

// MockSettlementService 实现 settlement.ServiceInterface（权限测试中不会被调用）
type MockSettlementService struct{}

func (m *MockSettlementService) Preview(ctx context.Context, activityID uint) (*settlement.PreviewResult, error) {
	return nil, apperror.ErrInternal
}
func (m *MockSettlementService) Execute(ctx context.Context, activityID uint, operatorID uint) (*model.Settlement, error) {
	return nil, apperror.ErrInternal
}
func (m *MockSettlementService) Reverse(ctx context.Context, settlementID uint, operatorID uint) error {
	return apperror.ErrInternal
}
func (m *MockSettlementService) ListByActivity(ctx context.Context, activityID uint) ([]model.Settlement, error) {
	return nil, apperror.ErrInternal
}
func (m *MockSettlementService) Overview(ctx context.Context) ([]settlement.SettlementOverviewItem, error) {
	return nil, apperror.ErrInternal
}

// ---- 权限校验测试 ----

// TestSettlementHandler_NonHQAdmin 验证四个结算接口均拒绝非总部管理员（45008）
func TestSettlementHandler_NonHQAdmin(t *testing.T) {
	h := settlement.NewHandler(&MockSettlementService{})

	cases := []struct {
		name   string
		invoke func(c *gin.Context)
	}{
		{
			name: "Preview",
			invoke: func(c *gin.Context) {
				setParam(c, "activity_id", "1")
				h.Preview(c)
			},
		},
		{
			name: "Execute",
			invoke: func(c *gin.Context) {
				setParam(c, "activity_id", "1")
				h.Execute(c)
			},
		},
		{
			name: "Reverse",
			invoke: func(c *gin.Context) {
				setParam(c, "settlement_id", "1")
				h.Reverse(c)
			},
		},
		{
			name: "ListByActivity",
			invoke: func(c *gin.Context) {
				setParam(c, "activity_id", "1")
				h.ListByActivity(c)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := setupGin("GET", "/api/v1/settlements/test")
			setAuth(c, 2, 1, model.RoleCampusOperator) // 非 hq_admin
			tc.invoke(c)
			assert.Equal(t, apperror.ErrSettlementPermDenied.Code, parseResp(w), "非管理员应被拒绝")
		})
	}
}

// TestSettlementHandler_HQAdminAllowed 验证 hq_admin 可正常进入（mock 服务返回内部错误而非权限错误）
func TestSettlementHandler_HQAdminAllowed(t *testing.T) {
	h := settlement.NewHandler(&MockSettlementService{})

	c, w := setupGin("GET", "/api/v1/settlements/test")
	setAuth(c, 1, 1, model.RoleHQAdmin)
	setParam(c, "activity_id", "1")
	h.Preview(c)

	code := parseResp(w)
	assert.NotEqual(t, apperror.ErrSettlementPermDenied.Code, code, "hq_admin 不应被权限拦截")
	assert.Equal(t, apperror.ErrInternal.Code, code, "应走到服务层（mock 返回内部错误）")
}
