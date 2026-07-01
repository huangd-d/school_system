package material_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/material"
	"school-system/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- 辅助：快速构造 mock 依赖 ----

func newMockRepo() *MockMaterialRepo {
	return &MockMaterialRepo{}
}

func newMockActivities() *MockActivityLookup {
	return &MockActivityLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Activity, error) {
			return &model.Activity{ID: id, Name: "测试活动", CampusID: 1, Status: model.ActivityInProgress}, nil
		},
	}
}

func newService(repo material.Repository, activities material.ActivityLookup) *material.Service {
	return material.NewService(repo, activities)
}

func assertAppError(t *testing.T, err error, expected *apperror.AppError) {
	t.Helper()
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr), "错误不是 *apperror.AppError: %v", err)
	assert.Equal(t, expected.Code, appErr.Code, "错误码不匹配: got %d, want %d", appErr.Code, expected.Code)
}

// ============================================================
//  ListCategories
// ============================================================

func TestMaterialService_ListCategories_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllCategoriesFn = func(ctx context.Context) ([]model.MaterialCategory, error) {
		return []model.MaterialCategory{
			{ID: 1, Name: "教材"},
			{ID: 2, Name: "文具"},
		}, nil
	}
	svc := newService(mockRepo, newMockActivities())

	cats, err := svc.ListCategories(context.Background())
	require.NoError(t, err)
	assert.Len(t, cats, 2)
	assert.Equal(t, "教材", cats[0].Name)
}

func TestMaterialService_ListCategories_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllCategoriesFn = func(ctx context.Context) ([]model.MaterialCategory, error) {
		return nil, errors.New("db error")
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.ListCategories(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查询分类列表失败")
}

// ============================================================
//  CreateCategory
// ============================================================

func TestMaterialService_CreateCategory_EmptyName(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.CreateCategory(context.Background(), "", "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialCategoryNameEmpty)
}

func TestMaterialService_CreateCategory_NameTooLong(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.CreateCategory(context.Background(), strings.Repeat("x", 51), "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialCategoryNameTooLong)
}

func TestMaterialService_CreateCategory_NotHQAdmin(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.CreateCategory(context.Background(), "教材", "", model.RoleCampusOperator)
	assertAppError(t, err, apperror.ErrMaterialPermDenied)
}

func TestMaterialService_CreateCategory_NameDuplicate(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByNameFn = func(ctx context.Context, name string) (*model.MaterialCategory, error) {
		return &model.MaterialCategory{ID: 1, Name: "已存在"}, nil
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.CreateCategory(context.Background(), "已存在", "", model.RoleHQAdmin)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrMaterialCategoryNameDup.Code, appErr.Code)
	assert.Contains(t, err.Error(), "已存在")
}

func TestMaterialService_CreateCategory_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByNameFn = func(ctx context.Context, name string) (*model.MaterialCategory, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.CreateCategoryFn = func(ctx context.Context, cat *model.MaterialCategory) error {
		cat.ID = 10
		return nil
	}
	svc := newService(mockRepo, newMockActivities())

	cat, err := svc.CreateCategory(context.Background(), "教材", "教学用书", model.RoleHQAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint(10), cat.ID)
	assert.Equal(t, "教材", cat.Name)
	assert.Equal(t, "教学用书", cat.Note)
}

func TestMaterialService_CreateCategory_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByNameFn = func(ctx context.Context, name string) (*model.MaterialCategory, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.CreateCategoryFn = func(ctx context.Context, cat *model.MaterialCategory) error {
		return errors.New("db error")
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.CreateCategory(context.Background(), "教材", "", model.RoleHQAdmin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "创建分类失败")
}

// ============================================================
//  UpdateCategory
// ============================================================

func TestMaterialService_UpdateCategory_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.UpdateCategory(context.Background(), 999, "新名称", "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialCategoryNotFound)
}

func TestMaterialService_UpdateCategory_NotHQAdmin(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.UpdateCategory(context.Background(), 1, "新名称", "", model.RoleActivityContact)
	assertAppError(t, err, apperror.ErrMaterialPermDenied)
}

func TestMaterialService_UpdateCategory_Success(t *testing.T) {
	existing := &model.MaterialCategory{ID: 1, Name: "旧名称", Note: "旧备注"}
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return existing, nil
	}
	mockRepo.FindCategoryByNameFn = func(ctx context.Context, name string) (*model.MaterialCategory, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.UpdateCategoryFn = func(ctx context.Context, cat *model.MaterialCategory) error {
		return nil
	}
	svc := newService(mockRepo, newMockActivities())

	cat, err := svc.UpdateCategory(context.Background(), 1, "新名称", "新备注", model.RoleHQAdmin)
	require.NoError(t, err)
	assert.Equal(t, "新名称", cat.Name)
	assert.Equal(t, "新备注", cat.Note)
}

func TestMaterialService_UpdateCategory_NameDuplicate_Other(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return &model.MaterialCategory{ID: 1, Name: "旧名称"}, nil
	}
	mockRepo.FindCategoryByNameFn = func(ctx context.Context, name string) (*model.MaterialCategory, error) {
		return &model.MaterialCategory{ID: 2, Name: "新名称"}, nil // 不同 ID，冲突
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.UpdateCategory(context.Background(), 1, "新名称", "", model.RoleHQAdmin)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrMaterialCategoryNameDup.Code, appErr.Code)
}

// ============================================================
//  DeleteCategory
// ============================================================

func TestMaterialService_DeleteCategory_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	err := svc.DeleteCategory(context.Background(), 999, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialCategoryNotFound)
}

func TestMaterialService_DeleteCategory_NotHQAdmin(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	err := svc.DeleteCategory(context.Background(), 1, model.RoleCampusOperator)
	assertAppError(t, err, apperror.ErrMaterialPermDenied)
}

func TestMaterialService_DeleteCategory_HasPurchases(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return &model.MaterialCategory{ID: 1, Name: "教材"}, nil
	}
	mockRepo.CountPurchasesByCategoryFn = func(ctx context.Context, categoryID uint) (int64, error) {
		return 3, nil
	}
	svc := newService(mockRepo, newMockActivities())

	err := svc.DeleteCategory(context.Background(), 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialCategoryHasPurchase)
}

func TestMaterialService_DeleteCategory_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return &model.MaterialCategory{ID: 1, Name: "教材"}, nil
	}
	mockRepo.CountPurchasesByCategoryFn = func(ctx context.Context, categoryID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.DeleteCategoryFn = func(ctx context.Context, id uint) error {
		return nil
	}
	svc := newService(mockRepo, newMockActivities())

	err := svc.DeleteCategory(context.Background(), 1, model.RoleHQAdmin)
	require.NoError(t, err)
}

// ============================================================
//  Purchase
// ============================================================

func TestMaterialService_Purchase_EmptyMaterialName(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Purchase(context.Background(), "", 1, 10, 100, "", 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialNameEmpty)
}

func TestMaterialService_Purchase_InvalidQuantity(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Purchase(context.Background(), "教材", 1, 0, 100, "", 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialQuantityInvalid)
}

func TestMaterialService_Purchase_NegativeQuantity(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Purchase(context.Background(), "教材", 1, -5, 100, "", 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialQuantityInvalid)
}

func TestMaterialService_Purchase_InvalidAmount(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Purchase(context.Background(), "教材", 1, 10, 0, "", 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialAmountInvalid)
}

func TestMaterialService_Purchase_CategoryNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindCategoryByIDFn = func(ctx context.Context, id uint) (*model.MaterialCategory, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.Purchase(context.Background(), "教材", 999, 10, 100, "", 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialCategoryNotExist)
}

func TestMaterialService_Purchase_NotHQAdmin(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Purchase(context.Background(), "教材", 1, 10, 100, "", 1, model.RoleCampusOperator)
	assertAppError(t, err, apperror.ErrMaterialPermDenied)
}

// ============================================================
//  ListStock
// ============================================================

func TestMaterialService_ListStock_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockWithFilterFn = func(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error) {
		return []model.Stock{
			{ID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 80},
		}, 1, nil
	}
	svc := newService(mockRepo, newMockActivities())

	result, err := svc.ListStock(context.Background(), 0, "", 1, 20)
	require.NoError(t, err)
	assert.Len(t, result.Stocks, 1)
	assert.Equal(t, int64(1), result.Total)
}

func TestMaterialService_ListStock_DefaultPagination(t *testing.T) {
	mockRepo := newMockRepo()
	callCount := 0
	mockRepo.FindStockWithFilterFn = func(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error) {
		callCount++
		assert.Equal(t, 0, offset) // page=1 → offset=0
		assert.Equal(t, 20, limit) // pageSize=0 → default 20
		return nil, 0, nil
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.ListStock(context.Background(), 0, "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestMaterialService_ListStock_FilterByCategory(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockWithFilterFn = func(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error) {
		assert.Equal(t, uint(2), categoryID)
		return nil, 0, nil
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.ListStock(context.Background(), 2, "", 1, 20)
	require.NoError(t, err)
}

// ============================================================
//  GetStock
// ============================================================

func TestMaterialService_GetStock_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.GetStock(context.Background(), 999)
	assertAppError(t, err, apperror.ErrMaterialStockNotFound)
}

func TestMaterialService_GetStock_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return &model.Stock{ID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 80}, nil
	}
	svc := newService(mockRepo, newMockActivities())

	stock, err := svc.GetStock(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "语文教材", stock.MaterialName)
	assert.Equal(t, 100, stock.TotalQuantity)
}

// ============================================================
//  Distribute
// ============================================================

func TestMaterialService_Distribute_InvalidQuantity(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Distribute(context.Background(), 1, 1, 0, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialDistQuantityInvalid)
}

func TestMaterialService_Distribute_NotHQAdmin(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	_, err := svc.Distribute(context.Background(), 1, 1, 10, 1, "", model.RoleCampusOperator)
	assertAppError(t, err, apperror.ErrMaterialPermDenied)
}

func TestMaterialService_Distribute_StockNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.Distribute(context.Background(), 999, 1, 10, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialStockNotFound)
}

func TestMaterialService_Distribute_ActivityNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return &model.Stock{ID: 1, RemainingQty: 100}, nil
	}
	mockActivities := &MockActivityLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Activity, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := newService(mockRepo, mockActivities)

	_, err := svc.Distribute(context.Background(), 1, 999, 10, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialActivityNotFound)
}

func TestMaterialService_Distribute_InsufficientStock(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return &model.Stock{ID: 1, RemainingQty: 5}, nil
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.Distribute(context.Background(), 1, 1, 10, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialStockInsufficient)
}

// ============================================================
//  AdjustDistribution
// ============================================================

func TestMaterialService_AdjustDist_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindDistributionByIDFn = func(ctx context.Context, id uint) (*model.Distribution, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	err := svc.AdjustDistribution(context.Background(), 999, 10, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialDistNotFound)
}

func TestMaterialService_AdjustDist_ZeroQuantity(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	err := svc.AdjustDistribution(context.Background(), 1, 0, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialAdjQuantityZero)
}

func TestMaterialService_AdjustDist_NegativeQuantity(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	err := svc.AdjustDistribution(context.Background(), 1, -5, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialAdjQuantityZero)
}

func TestMaterialService_AdjustDist_NotHQAdmin(t *testing.T) {
	svc := newService(newMockRepo(), newMockActivities())
	err := svc.AdjustDistribution(context.Background(), 1, 10, 1, "", model.RoleCampusOperator)
	assertAppError(t, err, apperror.ErrMaterialPermDenied)
}

func TestMaterialService_AdjustDist_StockNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindDistributionByIDFn = func(ctx context.Context, id uint) (*model.Distribution, error) {
		return &model.Distribution{ID: 1, StockID: 999, Quantity: 5}, nil
	}
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, newMockActivities())

	err := svc.AdjustDistribution(context.Background(), 1, 10, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialStockNotFound)
}

func TestMaterialService_AdjustDist_Increase_ExceedStock(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindDistributionByIDFn = func(ctx context.Context, id uint) (*model.Distribution, error) {
		return &model.Distribution{ID: 1, StockID: 1, Quantity: 5}, nil
	}
	mockRepo.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return &model.Stock{ID: 1, RemainingQty: 2}, nil // 只有2剩余，需要加5（10-5=5 > 2）
	}
	svc := newService(mockRepo, newMockActivities())

	err := svc.AdjustDistribution(context.Background(), 1, 10, 1, "", model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrMaterialStockInsufficient)
}

// ============================================================
//  ListPurchaseOrders
// ============================================================

func TestMaterialService_ListPurchaseOrders_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindPurchaseOrdersFn = func(ctx context.Context, offset, limit int) ([]model.PurchaseOrder, int64, error) {
		return []model.PurchaseOrder{
			{ID: 1, MaterialName: "语文教材", Quantity: 100, TotalAmount: 5000},
		}, 1, nil
	}
	svc := newService(mockRepo, newMockActivities())

	result, err := svc.ListPurchaseOrders(context.Background(), 1, 20)
	require.NoError(t, err)
	assert.Len(t, result.Orders, 1)
	assert.Equal(t, int64(1), result.Total)
}

func TestMaterialService_ListPurchaseOrders_DefaultPagination(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindPurchaseOrdersFn = func(ctx context.Context, offset, limit int) ([]model.PurchaseOrder, int64, error) {
		assert.Equal(t, 0, offset)
		assert.Equal(t, 20, limit)
		return nil, 0, nil
	}
	svc := newService(mockRepo, newMockActivities())

	_, err := svc.ListPurchaseOrders(context.Background(), 0, 0)
	require.NoError(t, err)
}
