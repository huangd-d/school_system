package material_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/material"
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

// ============================================================
//  ListCategories — GET /api/v1/materials/categories
// ============================================================

func TestMaterialHandler_ListCategories_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListCategoriesFn: func(ctx context.Context) ([]model.MaterialCategory, error) {
			return []model.MaterialCategory{
				{ID: 1, Name: "教材"},
				{ID: 2, Name: "文具"},
			}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/categories", nil)
	h.ListCategories(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), "教材")
	assert.Contains(t, string(data), "文具")
}

func TestMaterialHandler_ListCategories_Error(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListCategoriesFn: func(ctx context.Context) ([]model.MaterialCategory, error) {
			return nil, apperror.ErrInternal
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/categories", nil)
	h.ListCategories(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInternal.Code, code)
}

// ============================================================
//  CreateCategory — POST /api/v1/materials/categories
// ============================================================

func TestMaterialHandler_CreateCategory_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		CreateCategoryFn: func(ctx context.Context, name, note string, operatorRole string) (*model.MaterialCategory, error) {
			assert.Equal(t, "教材", name)
			assert.Equal(t, model.RoleHQAdmin, operatorRole)
			return &model.MaterialCategory{ID: 5, Name: name, Note: note}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/materials/categories", toJSONBody(map[string]string{
		"name": "教材",
		"note": "教学用书",
	}))
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.CreateCategory(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Contains(t, string(data), `"name":"教材"`)
	assert.Contains(t, string(data), `"id":5`)
}

func TestMaterialHandler_CreateCategory_MissingName(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("POST", "/api/v1/materials/categories", toJSONBody(map[string]string{
		"note": "test",
	}))

	h.CreateCategory(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestMaterialHandler_CreateCategory_ServiceError(t *testing.T) {
	mockSvc := &MockMaterialService{
		CreateCategoryFn: func(ctx context.Context, name, note string, operatorRole string) (*model.MaterialCategory, error) {
			return nil, apperror.New(apperror.ErrMaterialCategoryNameDup.Code, "分类名称「教材」已存在")
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/materials/categories", toJSONBody(map[string]string{
		"name": "教材",
	}))
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.CreateCategory(c)

	code, msg, _ := parseResp(w)
	assert.Equal(t, apperror.ErrMaterialCategoryNameDup.Code, code)
	assert.Contains(t, msg, "教材")
}

// ============================================================
//  UpdateCategory — PUT /api/v1/materials/categories/:id
// ============================================================

func TestMaterialHandler_UpdateCategory_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		UpdateCategoryFn: func(ctx context.Context, id uint, name, note string, operatorRole string) (*model.MaterialCategory, error) {
			assert.Equal(t, uint(3), id)
			return &model.MaterialCategory{ID: 3, Name: name, Note: note}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/materials/categories/3", toJSONBody(map[string]string{
		"name": "新名称",
	}))
	setParam(c, "id", "3")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.UpdateCategory(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), `"name":"新名称"`)
	assert.Contains(t, string(data), `"id":3`)
}

func TestMaterialHandler_UpdateCategory_InvalidID(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("PUT", "/api/v1/materials/categories/abc", toJSONBody(map[string]string{
		"name": "新名称",
	}))
	setParam(c, "id", "abc")

	h.UpdateCategory(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

// ============================================================
//  DeleteCategory — DELETE /api/v1/materials/categories/:id
// ============================================================

func TestMaterialHandler_DeleteCategory_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		DeleteCategoryFn: func(ctx context.Context, id uint, operatorRole string) error {
			assert.Equal(t, uint(3), id)
			return nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("DELETE", "/api/v1/materials/categories/3", nil)
	setParam(c, "id", "3")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.DeleteCategory(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Equal(t, "null", string(data))
}

func TestMaterialHandler_DeleteCategory_InvalidID(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("DELETE", "/api/v1/materials/categories/abc", nil)
	setParam(c, "id", "abc")

	h.DeleteCategory(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

// ============================================================
//  Purchase — POST /api/v1/materials/purchase
// ============================================================

func TestMaterialHandler_Purchase_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		PurchaseFn: func(ctx context.Context, materialName string, categoryID uint, quantity int, totalAmount float64, notes string, purchaserID uint, operatorRole string) (*model.Stock, error) {
			assert.Equal(t, "语文教材", materialName)
			assert.Equal(t, uint(1), categoryID)
			assert.Equal(t, 100, quantity)
			return &model.Stock{ID: 10, MaterialName: materialName, TotalQuantity: quantity, RemainingQty: quantity, UnitPrice: 50.0, Source: "purchase"}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/materials/purchase", toJSONBody(map[string]interface{}{
		"material_name": "语文教材",
		"category_id":   1,
		"quantity":      100,
		"total_amount":  5000,
	}))
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Purchase(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), `"material_name":"语文教材"`)
	assert.Contains(t, string(data), `"id":10`)
}

func TestMaterialHandler_Purchase_MissingRequired(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("POST", "/api/v1/materials/purchase", toJSONBody(map[string]interface{}{
		"category_id": 1,
	}))

	h.Purchase(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

// ============================================================
//  ListStock — GET /api/v1/materials/stock
// ============================================================

func TestMaterialHandler_ListStock_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListStockFn: func(ctx context.Context, categoryID uint, keyword string, page, pageSize int) (*material.StockListResult, error) {
			return &material.StockListResult{
				Stocks: []model.Stock{
					{ID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 80},
				},
				Total: 1,
			}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/stock", nil)
	h.ListStock(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), "语文教材")
	assert.Contains(t, string(data), `"total":1`)
}

// ============================================================
//  GetStock — GET /api/v1/materials/stock/:id
// ============================================================

func TestMaterialHandler_GetStock_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		GetStockFn: func(ctx context.Context, id uint) (*model.Stock, error) {
			return &model.Stock{ID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 80}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/stock/1", nil)
	setParam(c, "id", "1")

	h.GetStock(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), "语文教材")
}

func TestMaterialHandler_GetStock_InvalidID(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("GET", "/api/v1/materials/stock/abc", nil)
	setParam(c, "id", "abc")

	h.GetStock(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

// ============================================================
//  Distribute — POST /api/v1/materials/distribute
// ============================================================

func TestMaterialHandler_Distribute_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		DistributeFn: func(ctx context.Context, stockID uint, activityID uint, quantity int, operatorID uint, reason string, operatorRole string) (*model.Distribution, error) {
			return &model.Distribution{ID: 5, StockID: stockID, ActivityID: activityID, Quantity: quantity}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("POST", "/api/v1/materials/distribute", toJSONBody(map[string]interface{}{
		"stock_id":    1,
		"activity_id": 2,
		"quantity":    10,
		"reason":      "活动需要",
	}))
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.Distribute(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), `"id":5`)
	assert.Contains(t, string(data), `"quantity":10`)
}

func TestMaterialHandler_Distribute_MissingRequired(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("POST", "/api/v1/materials/distribute", toJSONBody(map[string]interface{}{
		"stock_id": 1,
	}))

	h.Distribute(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

// ============================================================
//  AdjustDistribution — PUT /api/v1/materials/distribute/:id
// ============================================================

func TestMaterialHandler_AdjustDist_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		AdjustDistributionFn: func(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string, operatorRole string) error {
			assert.Equal(t, uint(5), distributionID)
			assert.Equal(t, 15, newQuantity)
			return nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/materials/distribute/5", toJSONBody(map[string]interface{}{
		"quantity": 15,
		"reason":   "追加配发",
	}))
	setParam(c, "id", "5")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.AdjustDistribution(c)

	code, msg, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Equal(t, "success", msg)
	assert.Equal(t, "null", string(data))
}

func TestMaterialHandler_AdjustDist_InvalidID(t *testing.T) {
	h := material.NewHandler(&MockMaterialService{})

	c, w := setupGin("PUT", "/api/v1/materials/distribute/abc", toJSONBody(map[string]interface{}{
		"quantity": 15,
	}))
	setParam(c, "id", "abc")

	h.AdjustDistribution(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInvalidParam.Code, code)
}

func TestMaterialHandler_AdjustDist_ServiceError(t *testing.T) {
	mockSvc := &MockMaterialService{
		AdjustDistributionFn: func(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string, operatorRole string) error {
			return apperror.ErrMaterialDistNotFound
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("PUT", "/api/v1/materials/distribute/999", toJSONBody(map[string]interface{}{
		"quantity": 15,
	}))
	setParam(c, "id", "999")
	setAuth(c, 1, 1, model.RoleHQAdmin)

	h.AdjustDistribution(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrMaterialDistNotFound.Code, code)
}

// ============================================================
//  ListPurchaseOrders — GET /api/v1/materials/purchases
// ============================================================

func TestMaterialHandler_ListPurchaseOrders_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListPurchaseOrdersFn: func(ctx context.Context, page, pageSize int) (*material.PurchaseOrderListResult, error) {
			return &material.PurchaseOrderListResult{
				Orders: []model.PurchaseOrder{
					{ID: 1, MaterialName: "语文教材", Quantity: 100, TotalAmount: 5000},
				},
				Total: 1,
			}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/purchases", nil)
	h.ListPurchaseOrders(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), "语文教材")
	assert.Contains(t, string(data), `"total":1`)
}

// ============================================================
//  ListDistributions — GET /api/v1/materials/distributions
// ============================================================

func TestMaterialHandler_ListDistributions_Success(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListDistributionsFn: func(ctx context.Context, activityID uint, keyword string, startDate, endDate string, page, pageSize int) (*material.DistributionListResult, error) {
			return &material.DistributionListResult{
				Distributions: []material.DistributionWithMaterial{
					{ID: 1, StockID: 1, MaterialName: "语文教材", ActivityID: 1, ActivityName: "开学典礼", Quantity: 10, Reason: "活动需要"},
				},
				Total: 1,
			}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/distributions?activity_id=1&keyword=教材", nil)
	h.ListDistributions(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), "语文教材")
	assert.Contains(t, string(data), "开学典礼")
	assert.Contains(t, string(data), `"total":1`)
}

func TestMaterialHandler_ListDistributions_DefaultParams(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListDistributionsFn: func(ctx context.Context, activityID uint, keyword string, startDate, endDate string, page, pageSize int) (*material.DistributionListResult, error) {
			// 默认参数：activityID=0, keyword="", startDate="", endDate="", page=1, pageSize=20
			assert.Equal(t, uint(0), activityID)
			assert.Equal(t, "", keyword)
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, pageSize)
			return &material.DistributionListResult{Distributions: nil, Total: 0}, nil
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/distributions", nil)
	h.ListDistributions(c)

	code, _, data := parseResp(w)
	assert.Equal(t, 0, code)
	assert.Contains(t, string(data), `"total":0`)
}

func TestMaterialHandler_ListDistributions_ServiceError(t *testing.T) {
	mockSvc := &MockMaterialService{
		ListDistributionsFn: func(ctx context.Context, activityID uint, keyword string, startDate, endDate string, page, pageSize int) (*material.DistributionListResult, error) {
			return nil, apperror.ErrInternal
		},
	}
	h := material.NewHandler(mockSvc)

	c, w := setupGin("GET", "/api/v1/materials/distributions", nil)
	h.ListDistributions(c)

	code, _, _ := parseResp(w)
	assert.Equal(t, apperror.ErrInternal.Code, code)
}
