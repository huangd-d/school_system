package material

import (
	"context"
	"school-system/internal/model"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

// ---- 分类 ----
func (r *repository) FindAllCategories(ctx context.Context) ([]model.MaterialCategory, error) {
	var cats []model.MaterialCategory
	err := r.db.WithContext(ctx).Find(&cats).Error
	return cats, err
}

func (r *repository) CreateCategory(ctx context.Context, cat *model.MaterialCategory) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

// ---- 采购 ----
func (r *repository) CreatePurchaseOrder(ctx context.Context, po *model.PurchaseOrder) error {
	return r.db.WithContext(ctx).Create(po).Error
}

// ---- 库存 ----
func (r *repository) FindAllStock(ctx context.Context) ([]model.Stock, error) {
	var stocks []model.Stock
	err := r.db.WithContext(ctx).Find(&stocks).Error
	return stocks, err
}

func (r *repository) FindStockByID(ctx context.Context, id uint) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.WithContext(ctx).First(&stock, id).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *repository) CreateStock(ctx context.Context, stock *model.Stock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

func (r *repository) UpdateStock(ctx context.Context, stock *model.Stock) error {
	return r.db.WithContext(ctx).Save(stock).Error
}

// ---- 派发 ----
func (r *repository) CreateDistribution(ctx context.Context, dist *model.Distribution) error {
	return r.db.WithContext(ctx).Create(dist).Error
}

func (r *repository) FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error) {
	var dists []model.Distribution
	err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).Find(&dists).Error
	return dists, err
}

func (r *repository) FindDistributionsByStock(ctx context.Context, stockID uint) ([]model.Distribution, error) {
	var dists []model.Distribution
	err := r.db.WithContext(ctx).Where("stock_id = ?", stockID).Find(&dists).Error
	return dists, err
}
