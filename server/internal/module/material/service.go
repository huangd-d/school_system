package material

import (
	"context"
	"school-system/internal/model"
)

// Repository 物资数据访问接口
type Repository interface {
	// 分类
	FindAllCategories(ctx context.Context) ([]model.MaterialCategory, error)
	CreateCategory(ctx context.Context, cat *model.MaterialCategory) error
	// 采购
	CreatePurchaseOrder(ctx context.Context, po *model.PurchaseOrder) error
	// 库存
	FindAllStock(ctx context.Context) ([]model.Stock, error)
	FindStockByID(ctx context.Context, id uint) (*model.Stock, error)
	CreateStock(ctx context.Context, stock *model.Stock) error
	UpdateStock(ctx context.Context, stock *model.Stock) error
	// 派发
	CreateDistribution(ctx context.Context, dist *model.Distribution) error
	FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindDistributionsByStock(ctx context.Context, stockID uint) ([]model.Distribution, error)
}

// Service 物资业务逻辑
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListCategories(ctx context.Context) ([]model.MaterialCategory, error) {
	return nil, nil
}

func (s *Service) CreateCategory(ctx context.Context, name, note string) (*model.MaterialCategory, error) {
	return nil, nil
}

func (s *Service) Purchase(ctx context.Context, materialName string, categoryID uint, quantity int, totalAmount float64, notes string, purchaserID uint) (*model.Stock, error) {
	return nil, nil
}

func (s *Service) ListStock(ctx context.Context) ([]model.Stock, error) { return nil, nil }

func (s *Service) Distribute(ctx context.Context, stockID uint, activityID uint, quantity int, operatorID uint, reason string) (*model.Distribution, error) {
	return nil, nil
}

func (s *Service) AdjustDistribution(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string) error {
	return nil
}
