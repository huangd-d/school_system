package material

import (
	"context"
	"fmt"
	"school-system/internal/model"
	"school-system/pkg/apperror"
	"time"

	"gorm.io/gorm"
)

// Repository 物资数据访问接口
type Repository interface {
	// 分类
	FindAllCategories(ctx context.Context) ([]model.MaterialCategory, error)
	FindCategoryByID(ctx context.Context, id uint) (*model.MaterialCategory, error)
	FindCategoryByName(ctx context.Context, name string) (*model.MaterialCategory, error)
	CreateCategory(ctx context.Context, cat *model.MaterialCategory) error
	UpdateCategory(ctx context.Context, cat *model.MaterialCategory) error
	DeleteCategory(ctx context.Context, id uint) error
	CountPurchasesByCategory(ctx context.Context, categoryID uint) (int64, error)

	// 采购
	CreatePurchaseOrder(ctx context.Context, po *model.PurchaseOrder) error
	FindPurchaseOrders(ctx context.Context, offset, limit int) ([]model.PurchaseOrder, int64, error)

	// 库存
	FindAllStock(ctx context.Context) ([]model.Stock, error)
	FindStockByID(ctx context.Context, id uint) (*model.Stock, error)
	FindStockWithFilter(ctx context.Context, categoryID uint, keyword string, offset, limit int) ([]model.Stock, int64, error)
	CreateStock(ctx context.Context, stock *model.Stock) error
	UpdateStock(ctx context.Context, stock *model.Stock) error

	// 派发
	CreateDistribution(ctx context.Context, dist *model.Distribution) error
	FindDistributionByID(ctx context.Context, id uint) (*model.Distribution, error)
	FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindDistributionsByStock(ctx context.Context, stockID uint) ([]model.Distribution, error)
	FindDistributionsWithFilter(ctx context.Context, activityID uint, keyword string, startDate, endDate string, offset, limit int) ([]DistributionWithMaterial, int64, error)
	SumDistributionsByStock(ctx context.Context, stockID uint) (int, error)
	UpdateDistribution(ctx context.Context, dist *model.Distribution) error

	// 审计日志
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error

	// 事务支持
	DB() *gorm.DB
}

// ActivityLookup 活动查询接口（由 activity 模块满足，避免跨模块直接依赖）
type ActivityLookup interface {
	FindByID(ctx context.Context, id uint) (*model.Activity, error)
	FindByIDs(ctx context.Context, ids []uint) ([]model.Activity, error)
}

// StockListResult 库存列表查询结果
type StockListResult struct {
	Stocks []model.Stock
	Total  int64
}

// PurchaseOrderListResult 采购单列表查询结果
type PurchaseOrderListResult struct {
	Orders []model.PurchaseOrder
	Total  int64
}

// DistributionWithMaterial 派发记录 + 物资名称 + 活动名称
type DistributionWithMaterial struct {
	ID           uint      `json:"id"`
	StockID      uint      `json:"stock_id"`
	MaterialName string    `json:"material_name"`
	ActivityID   uint      `json:"activity_id"`
	ActivityName string    `json:"activity_name"`
	Quantity     int       `json:"quantity"`
	OperatorID   uint      `json:"operator_id"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

// DistributionListResult 派发记录分页结果
type DistributionListResult struct {
	Distributions []DistributionWithMaterial
	Total         int64
}

// Service 物资业务逻辑
type Service struct {
	repo       Repository
	activities ActivityLookup
}

// NewService 创建物资 service
func NewService(repo Repository, activities ActivityLookup) *Service {
	return &Service{repo: repo, activities: activities}
}

// ---- 分类管理 ----

// ListCategories 查询所有分类
func (s *Service) ListCategories(ctx context.Context) ([]model.MaterialCategory, error) {
	cats, err := s.repo.FindAllCategories(ctx)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询分类列表失败: %v", err)
	}
	return cats, nil
}

// CreateCategory 创建分类（仅总部管理员）
func (s *Service) CreateCategory(ctx context.Context, name, note string, operatorRole string) (*model.MaterialCategory, error) {
	if operatorRole != model.RoleHQAdmin {
		return nil, apperror.ErrMaterialPermDenied
	}
	if name == "" {
		return nil, apperror.ErrMaterialCategoryNameEmpty
	}
	if len(name) > 50 {
		return nil, apperror.ErrMaterialCategoryNameTooLong
	}

	// 重名检查
	existing, err := s.repo.FindCategoryByName(ctx, name)
	if err == nil && existing != nil {
		return nil, apperror.Newf(apperror.ErrMaterialCategoryNameDup.Code, "分类名称「%s」已存在", name)
	}

	if len(note) > 200 {
		note = note[:200]
	}

	cat := &model.MaterialCategory{
		Name: name,
		Note: note,
	}
	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "创建分类失败: %v", err)
	}
	return cat, nil
}

// UpdateCategory 更新分类（仅总部管理员）
func (s *Service) UpdateCategory(ctx context.Context, id uint, name, note string, operatorRole string) (*model.MaterialCategory, error) {
	if operatorRole != model.RoleHQAdmin {
		return nil, apperror.ErrMaterialPermDenied
	}

	cat, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrMaterialCategoryNotFound
	}

	if name == "" {
		return nil, apperror.ErrMaterialCategoryNameEmpty
	}
	if len(name) > 50 {
		return nil, apperror.ErrMaterialCategoryNameTooLong
	}

	// 重名检查（排除自身）
	existing, err := s.repo.FindCategoryByName(ctx, name)
	if err == nil && existing != nil && existing.ID != id {
		return nil, apperror.Newf(apperror.ErrMaterialCategoryNameDup.Code, "分类名称「%s」已存在", name)
	}

	cat.Name = name
	if len(note) > 200 {
		note = note[:200]
	}
	cat.Note = note

	if err := s.repo.UpdateCategory(ctx, cat); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "更新分类失败: %v", err)
	}
	return cat, nil
}

// DeleteCategory 删除分类（仅总部管理员，且分类下无采购记录）
func (s *Service) DeleteCategory(ctx context.Context, id uint, operatorRole string) error {
	if operatorRole != model.RoleHQAdmin {
		return apperror.ErrMaterialPermDenied
	}

	if _, err := s.repo.FindCategoryByID(ctx, id); err != nil {
		return apperror.ErrMaterialCategoryNotFound
	}

	count, err := s.repo.CountPurchasesByCategory(ctx, id)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "查询分类采购记录失败: %v", err)
	}
	if count > 0 {
		return apperror.ErrMaterialCategoryHasPurchase
	}

	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "删除分类失败: %v", err)
	}
	return nil
}

// ---- 采购管理 ----

// Purchase 采购入库（仅总部管理员，创建采购单+库存，事务保护）
func (s *Service) Purchase(ctx context.Context, materialName string, categoryID uint, quantity int, totalAmount int64, notes string, purchaserID uint, operatorRole string) (*model.Stock, error) {
	if operatorRole != model.RoleHQAdmin {
		return nil, apperror.ErrMaterialPermDenied
	}
	if materialName == "" {
		return nil, apperror.ErrMaterialNameEmpty
	}
	if quantity <= 0 {
		return nil, apperror.ErrMaterialQuantityInvalid
	}
	if totalAmount <= 0 {
		return nil, apperror.ErrMaterialAmountInvalid
	}

	// 分类存在性校验
	if _, err := s.repo.FindCategoryByID(ctx, categoryID); err != nil {
		return nil, apperror.ErrMaterialCategoryNotExist
	}

	// 计算单价（单位：分，四舍五入到分）
	unitPrice := (totalAmount + int64(quantity)/2) / int64(quantity)

	if len(notes) > 500 {
		notes = notes[:500]
	}

	var stock *model.Stock

	err := s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建采购单
		po := &model.PurchaseOrder{
			MaterialName: materialName,
			CategoryID:   categoryID,
			Quantity:     quantity,
			TotalAmount:  totalAmount,
			UnitPrice:    unitPrice,
			Notes:        notes,
			PurchaserID:  purchaserID,
		}
		if err := tx.Create(po).Error; err != nil {
			return fmt.Errorf("创建采购单失败: %w", err)
		}

		// 2. 创建库存记录
		stock = &model.Stock{
			PurchaseOrderID: po.ID,
			CategoryID:      categoryID,
			MaterialName:    materialName,
			TotalQuantity:   quantity,
			RemainingQty:    quantity,
			UnitPrice:       unitPrice,
			Source:          "purchase",
		}
		if err := tx.Create(stock).Error; err != nil {
			return fmt.Errorf("创建库存记录失败: %w", err)
		}

		// 3. 写入审计日志
		auditLog := &model.AuditLog{
			OperatorID:    purchaserID,
			OperationType: model.AuditOpPurchase,
			EntityType:    "stock",
			EntityID:      stock.ID,
			AfterValue:    fmt.Sprintf("name:%s,qty:%d,amount:%d(分)", materialName, quantity, totalAmount),
			ImpactAmount:  totalAmount,
		}
		if err := tx.Create(auditLog).Error; err != nil {
			return fmt.Errorf("写入审计日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "采购入库失败: %v", err)
	}

	return stock, nil
}

// ListPurchaseOrders 分页查询采购单列表
func (s *Service) ListPurchaseOrders(ctx context.Context, page, pageSize int) (*PurchaseOrderListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	orders, total, err := s.repo.FindPurchaseOrders(ctx, offset, pageSize)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询采购单列表失败: %v", err)
	}

	return &PurchaseOrderListResult{Orders: orders, Total: total}, nil
}

// ---- 库存管理 ----

// ListStock 分页查询库存列表（支持按分类、关键字筛选）
func (s *Service) ListStock(ctx context.Context, categoryID uint, keyword string, page, pageSize int) (*StockListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	stocks, total, err := s.repo.FindStockWithFilter(ctx, categoryID, keyword, offset, pageSize)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询库存列表失败: %v", err)
	}

	return &StockListResult{Stocks: stocks, Total: total}, nil
}

// ListStockDistributions 获取某库存的所有派发记录
func (s *Service) ListStockDistributions(ctx context.Context, stockID uint) ([]model.Distribution, error) {
	if _, err := s.repo.FindStockByID(ctx, stockID); err != nil {
		return nil, apperror.ErrMaterialStockNotFound
	}
	dists, err := s.repo.FindDistributionsByStock(ctx, stockID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询派发记录失败: %v", err)
	}
	return dists, nil
}

// GetStock 获取单条库存详情
func (s *Service) GetStock(ctx context.Context, id uint) (*model.Stock, error) {
	stock, err := s.repo.FindStockByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrMaterialStockNotFound
	}
	return stock, nil
}

// ---- 派发管理 ----

// Distribute 派发物资到活动（仅总部管理员，事务保护）
func (s *Service) Distribute(ctx context.Context, stockID uint, activityID uint, quantity int, operatorID uint, reason string, operatorRole string) (*model.Distribution, error) {
	if operatorRole != model.RoleHQAdmin {
		return nil, apperror.ErrMaterialPermDenied
	}
	if quantity <= 0 {
		return nil, apperror.ErrMaterialDistQuantityInvalid
	}

	// 库存存在性校验
	stock, err := s.repo.FindStockByID(ctx, stockID)
	if err != nil {
		return nil, apperror.ErrMaterialStockNotFound
	}

	// 活动存在性校验
	if _, err := s.activities.FindByID(ctx, activityID); err != nil {
		return nil, apperror.ErrMaterialActivityNotFound
	}

	// 库存余量校验
	if quantity > stock.RemainingQty {
		return nil, apperror.ErrMaterialStockInsufficient
	}

	if len(reason) > 500 {
		reason = reason[:500]
	}

	var dist *model.Distribution

	err = s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 扣减库存余量
		stock.RemainingQty -= quantity
		if err := tx.Save(stock).Error; err != nil {
			return fmt.Errorf("更新库存失败: %w", err)
		}

		// 2. 创建派发记录
		dist = &model.Distribution{
			StockID:    stockID,
			ActivityID: activityID,
			Quantity:   quantity,
			OperatorID: operatorID,
			Reason:     reason,
		}
		if err := tx.Create(dist).Error; err != nil {
			return fmt.Errorf("创建派发记录失败: %w", err)
		}

		// 3. 写入审计日志
		auditLog := &model.AuditLog{
			OperatorID:    operatorID,
			OperationType: model.AuditOpDistribute,
			EntityType:    "distribution",
			EntityID:      dist.ID,
			AfterValue:    fmt.Sprintf("stock_id:%d,activity_id:%d,qty:%d", stockID, activityID, quantity),
		}
		if err := tx.Create(auditLog).Error; err != nil {
			return fmt.Errorf("写入审计日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "派发物资失败: %v", err)
	}

	return dist, nil
}

// AdjustDistribution 调整派发数量（仅总部管理员，支持追加和退回，事务保护）
func (s *Service) AdjustDistribution(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string, operatorRole string) error {
	if operatorRole != model.RoleHQAdmin {
		return apperror.ErrMaterialPermDenied
	}
	if newQuantity <= 0 {
		return apperror.ErrMaterialAdjQuantityZero
	}

	// 查找派发记录
	dist, err := s.repo.FindDistributionByID(ctx, distributionID)
	if err != nil {
		return apperror.ErrMaterialDistNotFound
	}

	// 查找关联库存
	stock, err := s.repo.FindStockByID(ctx, dist.StockID)
	if err != nil {
		return apperror.ErrMaterialStockNotFound
	}

	oldQuantity := dist.Quantity
	diff := newQuantity - oldQuantity

	if diff == 0 {
		return nil // 幂等，无变化
	}

	if diff > 0 {
		// 追加配发：检查库存余量
		if diff > stock.RemainingQty {
			return apperror.ErrMaterialStockInsufficient
		}
	}

	if len(reason) > 500 {
		reason = reason[:500]
	}

	return s.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 更新库存（diff 为正时扣减，diff 为负时退回）
		stock.RemainingQty -= diff
		if err := tx.Save(stock).Error; err != nil {
			return fmt.Errorf("更新库存失败: %w", err)
		}

		// 2. 更新派发记录
		dist.Quantity = newQuantity
		dist.Reason = reason
		dist.OperatorID = operatorID
		if err := tx.Save(dist).Error; err != nil {
			return fmt.Errorf("更新派发记录失败: %w", err)
		}

		// 3. 写入审计日志
		auditLog := &model.AuditLog{
			OperatorID:    operatorID,
			OperationType: model.AuditOpAdjustDist,
			EntityType:    "distribution",
			EntityID:      distributionID,
			BeforeValue:   fmt.Sprintf("qty:%d", oldQuantity),
			AfterValue:    fmt.Sprintf("qty:%d", newQuantity),
		}
		if err := tx.Create(auditLog).Error; err != nil {
			return fmt.Errorf("写入审计日志失败: %w", err)
		}

		return nil
	})
}

// ---- 派发记录查询 ----

// ListDistributions 查询全部派发记录（支持按活动、物资名称、时间段筛选，分页）
func (s *Service) ListDistributions(ctx context.Context, activityID uint, keyword string, startDate, endDate string, page, pageSize int) (*DistributionListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dists, total, err := s.repo.FindDistributionsWithFilter(ctx, activityID, keyword, startDate, endDate, offset, pageSize)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询派发记录列表失败: %v", err)
	}

	// 批量获取活动名称
	if len(dists) > 0 {
		activityIDs := make([]uint, 0, len(dists))
		seen := make(map[uint]bool)
		for _, d := range dists {
			if !seen[d.ActivityID] {
				activityIDs = append(activityIDs, d.ActivityID)
				seen[d.ActivityID] = true
			}
		}

		activities, err := s.activities.FindByIDs(ctx, activityIDs)
		if err != nil {
			return nil, apperror.Newf(apperror.ErrInternal.Code, "查询活动信息失败: %v", err)
		}

		nameMap := make(map[uint]string, len(activities))
		for _, a := range activities {
			nameMap[a.ID] = a.Name
		}

		for i := range dists {
			dists[i].ActivityName = nameMap[dists[i].ActivityID]
		}
	}

	return &DistributionListResult{Distributions: dists, Total: total}, nil
}
