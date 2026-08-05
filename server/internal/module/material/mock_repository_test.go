package material_test

import (
	"context"

	"school-system/internal/model"
	"school-system/internal/module/material"

	"gorm.io/gorm"
)

// ---- MockMaterialRepo ----

// MockMaterialRepo 实现 material.Repository 接口，每个方法对应一个可替换的函数字段。
type MockMaterialRepo struct {
	FindAllCategoriesFn         func(ctx context.Context) ([]model.MaterialCategory, error)
	FindCategoryByIDFn          func(ctx context.Context, id uint) (*model.MaterialCategory, error)
	FindCategoryByNameFn        func(ctx context.Context, name string) (*model.MaterialCategory, error)
	CreateCategoryFn            func(ctx context.Context, cat *model.MaterialCategory) error
	UpdateCategoryFn            func(ctx context.Context, cat *model.MaterialCategory) error
	DeleteCategoryFn            func(ctx context.Context, id uint) error
	CountPurchasesByCategoryFn  func(ctx context.Context, categoryID uint) (int64, error)
	CreatePurchaseOrderFn       func(ctx context.Context, po *model.PurchaseOrder) error
	FindPurchaseOrdersFn        func(ctx context.Context, offset, limit int) ([]model.PurchaseOrder, int64, error)
	FindAllStockFn              func(ctx context.Context) ([]model.Stock, error)
	FindStockByIDFn             func(ctx context.Context, id uint) (*model.Stock, error)
	FindStockWithFilterFn       func(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error)
	CreateStockFn               func(ctx context.Context, stock *model.Stock) error
	UpdateStockFn               func(ctx context.Context, stock *model.Stock) error
	CreateDistributionFn        func(ctx context.Context, dist *model.Distribution) error
	FindDistributionByIDFn      func(ctx context.Context, id uint) (*model.Distribution, error)
	FindDistributionsByActivityFn func(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindDistributionsByStockFn  func(ctx context.Context, stockID uint) ([]model.Distribution, error)
	FindDistributionsWithFilterFn func(ctx context.Context, activityID uint, keyword string, startDate, endDate string, offset, limit int) ([]material.DistributionWithMaterial, int64, error)
	SumDistributionsByStockFn   func(ctx context.Context, stockID uint) (int, error)
	UpdateDistributionFn        func(ctx context.Context, dist *model.Distribution) error
	CreateAuditLogFn            func(ctx context.Context, log *model.AuditLog) error
	DBFn                        func() *gorm.DB
}

func (m *MockMaterialRepo) FindAllCategories(ctx context.Context) ([]model.MaterialCategory, error) {
	if m.FindAllCategoriesFn != nil {
		return m.FindAllCategoriesFn(ctx)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindCategoryByID(ctx context.Context, id uint) (*model.MaterialCategory, error) {
	if m.FindCategoryByIDFn != nil {
		return m.FindCategoryByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindCategoryByName(ctx context.Context, name string) (*model.MaterialCategory, error) {
	if m.FindCategoryByNameFn != nil {
		return m.FindCategoryByNameFn(ctx, name)
	}
	return nil, nil
}

func (m *MockMaterialRepo) CreateCategory(ctx context.Context, cat *model.MaterialCategory) error {
	if m.CreateCategoryFn != nil {
		return m.CreateCategoryFn(ctx, cat)
	}
	return nil
}

func (m *MockMaterialRepo) UpdateCategory(ctx context.Context, cat *model.MaterialCategory) error {
	if m.UpdateCategoryFn != nil {
		return m.UpdateCategoryFn(ctx, cat)
	}
	return nil
}

func (m *MockMaterialRepo) DeleteCategory(ctx context.Context, id uint) error {
	if m.DeleteCategoryFn != nil {
		return m.DeleteCategoryFn(ctx, id)
	}
	return nil
}

func (m *MockMaterialRepo) CountPurchasesByCategory(ctx context.Context, categoryID uint) (int64, error) {
	if m.CountPurchasesByCategoryFn != nil {
		return m.CountPurchasesByCategoryFn(ctx, categoryID)
	}
	return 0, nil
}

func (m *MockMaterialRepo) CreatePurchaseOrder(ctx context.Context, po *model.PurchaseOrder) error {
	if m.CreatePurchaseOrderFn != nil {
		return m.CreatePurchaseOrderFn(ctx, po)
	}
	return nil
}

func (m *MockMaterialRepo) FindPurchaseOrders(ctx context.Context, offset, limit int) ([]model.PurchaseOrder, int64, error) {
	if m.FindPurchaseOrdersFn != nil {
		return m.FindPurchaseOrdersFn(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockMaterialRepo) FindAllStock(ctx context.Context) ([]model.Stock, error) {
	if m.FindAllStockFn != nil {
		return m.FindAllStockFn(ctx)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindStockByID(ctx context.Context, id uint) (*model.Stock, error) {
	if m.FindStockByIDFn != nil {
		return m.FindStockByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindStockWithFilter(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error) {
	if m.FindStockWithFilterFn != nil {
		return m.FindStockWithFilterFn(ctx, categoryID, keyword, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockMaterialRepo) CreateStock(ctx context.Context, stock *model.Stock) error {
	if m.CreateStockFn != nil {
		return m.CreateStockFn(ctx, stock)
	}
	return nil
}

func (m *MockMaterialRepo) UpdateStock(ctx context.Context, stock *model.Stock) error {
	if m.UpdateStockFn != nil {
		return m.UpdateStockFn(ctx, stock)
	}
	return nil
}

func (m *MockMaterialRepo) CreateDistribution(ctx context.Context, dist *model.Distribution) error {
	if m.CreateDistributionFn != nil {
		return m.CreateDistributionFn(ctx, dist)
	}
	return nil
}

func (m *MockMaterialRepo) FindDistributionByID(ctx context.Context, id uint) (*model.Distribution, error) {
	if m.FindDistributionByIDFn != nil {
		return m.FindDistributionByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error) {
	if m.FindDistributionsByActivityFn != nil {
		return m.FindDistributionsByActivityFn(ctx, activityID)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindDistributionsByStock(ctx context.Context, stockID uint) ([]model.Distribution, error) {
	if m.FindDistributionsByStockFn != nil {
		return m.FindDistributionsByStockFn(ctx, stockID)
	}
	return nil, nil
}

func (m *MockMaterialRepo) FindDistributionsWithFilter(ctx context.Context, activityID uint, keyword string, startDate, endDate string, offset, limit int) ([]material.DistributionWithMaterial, int64, error) {
	if m.FindDistributionsWithFilterFn != nil {
		return m.FindDistributionsWithFilterFn(ctx, activityID, keyword, startDate, endDate, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockMaterialRepo) SumDistributionsByStock(ctx context.Context, stockID uint) (int, error) {
	if m.SumDistributionsByStockFn != nil {
		return m.SumDistributionsByStockFn(ctx, stockID)
	}
	return 0, nil
}

func (m *MockMaterialRepo) UpdateDistribution(ctx context.Context, dist *model.Distribution) error {
	if m.UpdateDistributionFn != nil {
		return m.UpdateDistributionFn(ctx, dist)
	}
	return nil
}

func (m *MockMaterialRepo) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	if m.CreateAuditLogFn != nil {
		return m.CreateAuditLogFn(ctx, log)
	}
	return nil
}

func (m *MockMaterialRepo) DB() *gorm.DB {
	if m.DBFn != nil {
		return m.DBFn()
	}
	return nil
}

// 编译期检查：确保 MockMaterialRepo 实现了 material.Repository
var _ material.Repository = (*MockMaterialRepo)(nil)

// ---- MockActivityLookup ----

// MockActivityLookup 实现 material.ActivityLookup 接口。
type MockActivityLookup struct {
	FindByIDFn  func(ctx context.Context, id uint) (*model.Activity, error)
	FindByIDsFn func(ctx context.Context, ids []uint) ([]model.Activity, error)
}

func (m *MockActivityLookup) FindByID(ctx context.Context, id uint) (*model.Activity, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockActivityLookup) FindByIDs(ctx context.Context, ids []uint) ([]model.Activity, error) {
	if m.FindByIDsFn != nil {
		return m.FindByIDsFn(ctx, ids)
	}
	return nil, nil
}

var _ material.ActivityLookup = (*MockActivityLookup)(nil)

// ---- MockMaterialService ----

// MockMaterialService 实现 material.ServiceInterface 接口，用于 handler 层测试。
type MockMaterialService struct {
	ListCategoriesFn      func(ctx context.Context) ([]model.MaterialCategory, error)
	CreateCategoryFn      func(ctx context.Context, name, note string, operatorRole string) (*model.MaterialCategory, error)
	UpdateCategoryFn      func(ctx context.Context, id uint, name, note string, operatorRole string) (*model.MaterialCategory, error)
	DeleteCategoryFn      func(ctx context.Context, id uint, operatorRole string) error
	PurchaseFn            func(ctx context.Context, materialName string, categoryID uint, quantity int, totalAmount int64, notes string, purchaserID uint, operatorRole string) (*model.Stock, error)
	ListPurchaseOrdersFn  func(ctx context.Context, page, pageSize int) (*material.PurchaseOrderListResult, error)
	ListStockFn           func(ctx context.Context, categoryID uint, keyword string, page, pageSize int) (*material.StockListResult, error)
	GetStockFn               func(ctx context.Context, id uint) (*model.Stock, error)
	ListStockDistributionsFn  func(ctx context.Context, stockID uint) ([]model.Distribution, error)
	ListDistributionsFn       func(ctx context.Context, activityID uint, keyword string, startDate, endDate string, page, pageSize int) (*material.DistributionListResult, error)
	DistributeFn              func(ctx context.Context, stockID uint, activityID uint, quantity int, operatorID uint, reason string, operatorRole string) (*model.Distribution, error)
	AdjustDistributionFn      func(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string, operatorRole string) error
}

func (m *MockMaterialService) ListCategories(ctx context.Context) ([]model.MaterialCategory, error) {
	if m.ListCategoriesFn != nil {
		return m.ListCategoriesFn(ctx)
	}
	return nil, nil
}

func (m *MockMaterialService) CreateCategory(ctx context.Context, name, note string, operatorRole string) (*model.MaterialCategory, error) {
	if m.CreateCategoryFn != nil {
		return m.CreateCategoryFn(ctx, name, note, operatorRole)
	}
	return nil, nil
}

func (m *MockMaterialService) UpdateCategory(ctx context.Context, id uint, name, note string, operatorRole string) (*model.MaterialCategory, error) {
	if m.UpdateCategoryFn != nil {
		return m.UpdateCategoryFn(ctx, id, name, note, operatorRole)
	}
	return nil, nil
}

func (m *MockMaterialService) DeleteCategory(ctx context.Context, id uint, operatorRole string) error {
	if m.DeleteCategoryFn != nil {
		return m.DeleteCategoryFn(ctx, id, operatorRole)
	}
	return nil
}

func (m *MockMaterialService) Purchase(ctx context.Context, materialName string, categoryID uint, quantity int, totalAmount int64, notes string, purchaserID uint, operatorRole string) (*model.Stock, error) {
	if m.PurchaseFn != nil {
		return m.PurchaseFn(ctx, materialName, categoryID, quantity, totalAmount, notes, purchaserID, operatorRole)
	}
	return nil, nil
}

func (m *MockMaterialService) ListPurchaseOrders(ctx context.Context, page, pageSize int) (*material.PurchaseOrderListResult, error) {
	if m.ListPurchaseOrdersFn != nil {
		return m.ListPurchaseOrdersFn(ctx, page, pageSize)
	}
	return nil, nil
}

func (m *MockMaterialService) ListStock(ctx context.Context, categoryID uint, keyword string, page, pageSize int) (*material.StockListResult, error) {
	if m.ListStockFn != nil {
		return m.ListStockFn(ctx, categoryID, keyword, page, pageSize)
	}
	return nil, nil
}

func (m *MockMaterialService) GetStock(ctx context.Context, id uint) (*model.Stock, error) {
	if m.GetStockFn != nil {
		return m.GetStockFn(ctx, id)
	}
	return nil, nil
}

func (m *MockMaterialService) ListStockDistributions(ctx context.Context, stockID uint) ([]model.Distribution, error) {
	if m.ListStockDistributionsFn != nil {
		return m.ListStockDistributionsFn(ctx, stockID)
	}
	return nil, nil
}

func (m *MockMaterialService) ListDistributions(ctx context.Context, activityID uint, keyword string, startDate, endDate string, page, pageSize int) (*material.DistributionListResult, error) {
	if m.ListDistributionsFn != nil {
		return m.ListDistributionsFn(ctx, activityID, keyword, startDate, endDate, page, pageSize)
	}
	return nil, nil
}

func (m *MockMaterialService) Distribute(ctx context.Context, stockID uint, activityID uint, quantity int, operatorID uint, reason string, operatorRole string) (*model.Distribution, error) {
	if m.DistributeFn != nil {
		return m.DistributeFn(ctx, stockID, activityID, quantity, operatorID, reason, operatorRole)
	}
	return nil, nil
}

func (m *MockMaterialService) AdjustDistribution(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string, operatorRole string) error {
	if m.AdjustDistributionFn != nil {
		return m.AdjustDistributionFn(ctx, distributionID, newQuantity, operatorID, reason, operatorRole)
	}
	return nil
}

var _ material.ServiceInterface = (*MockMaterialService)(nil)
