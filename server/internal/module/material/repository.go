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

func (r *repository) FindCategoryByID(ctx context.Context, id uint) (*model.MaterialCategory, error) {
	var cat model.MaterialCategory
	err := r.db.WithContext(ctx).First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *repository) FindCategoryByName(ctx context.Context, name string) (*model.MaterialCategory, error) {
	var cat model.MaterialCategory
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&cat).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *repository) CreateCategory(ctx context.Context, cat *model.MaterialCategory) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *repository) UpdateCategory(ctx context.Context, cat *model.MaterialCategory) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

func (r *repository) DeleteCategory(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.MaterialCategory{}, id).Error
}

func (r *repository) CountPurchasesByCategory(ctx context.Context, categoryID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.PurchaseOrder{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

// ---- 采购 ----
func (r *repository) CreatePurchaseOrder(ctx context.Context, po *model.PurchaseOrder) error {
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *repository) FindPurchaseOrders(ctx context.Context, offset, limit int) ([]model.PurchaseOrder, int64, error) {
	var pos []model.PurchaseOrder
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.PurchaseOrder{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&pos).Error
	return pos, total, err
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

func (r *repository) FindStockWithFilter(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Stock{})
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		query = query.Where("material_name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var stocks []model.Stock
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&stocks).Error
	return stocks, total, err
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

func (r *repository) FindDistributionByID(ctx context.Context, id uint) (*model.Distribution, error) {
	var dist model.Distribution
	err := r.db.WithContext(ctx).First(&dist, id).Error
	if err != nil {
		return nil, err
	}
	return &dist, nil
}

func (r *repository) FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error) {
	var dists []model.Distribution
	err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).Find(&dists).Error
	return dists, err
}

func (r *repository) FindDistributionsWithFilter(ctx context.Context, activityID uint, keyword string, startDate, endDate string, offset, limit int) ([]DistributionWithMaterial, int64, error) {
	query := r.db.WithContext(ctx).
		Table("distributions").
		Select("distributions.*, stocks.material_name").
		Joins("LEFT JOIN stocks ON stocks.id = distributions.stock_id")

	if activityID > 0 {
		query = query.Where("distributions.activity_id = ?", activityID)
	}
	if keyword != "" {
		query = query.Where("stocks.material_name LIKE ?", "%"+keyword+"%")
	}
	if startDate != "" {
		query = query.Where("distributions.created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("distributions.created_at <= ?", endDate+" 23:59:59")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []DistributionWithMaterial
	err := query.Order("distributions.created_at DESC").Offset(offset).Limit(limit).Find(&results).Error
	return results, total, err
}

func (r *repository) FindDistributionsByStock(ctx context.Context, stockID uint) ([]model.Distribution, error) {
	var dists []model.Distribution
	err := r.db.WithContext(ctx).Where("stock_id = ?", stockID).Find(&dists).Error
	return dists, err
}

func (r *repository) SumDistributionsByStock(ctx context.Context, stockID uint) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.Distribution{}).
		Where("stock_id = ?", stockID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error
	return int(total), err
}

func (r *repository) UpdateDistribution(ctx context.Context, dist *model.Distribution) error {
	return r.db.WithContext(ctx).Save(dist).Error
}

// ---- 审计日志 ----
func (r *repository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ---- 事务支持 ----
func (r *repository) DB() *gorm.DB {
	return r.db
}
