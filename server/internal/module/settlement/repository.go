package settlement

import (
	"context"
	"school-system/internal/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

// ---- 已有方法 ----

func (r *repository) FindByActivityID(ctx context.Context, activityID uint) ([]model.Settlement, error) {
	var settlements []model.Settlement
	err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).Find(&settlements).Error
	return settlements, err
}

func (r *repository) CreateSettlement(ctx context.Context, s *model.Settlement) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *repository) UpdateSettlement(ctx context.Context, s *model.Settlement) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *repository) CreateRecoveryItems(ctx context.Context, items []model.RecoveryItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

// ---- 活动 ----

func (r *repository) FindActivityByID(ctx context.Context, id uint) (*model.Activity, error) {
	var activity model.Activity
	err := r.db.WithContext(ctx).First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *repository) UpdateActivityStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&model.Activity{}).Where("id = ?", id).Update("status", status).Error
}

// ---- 配发 ----

func (r *repository) FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error) {
	var distributions []model.Distribution
	err := r.db.WithContext(ctx).Where("activity_id = ?", activityID).Find(&distributions).Error
	return distributions, err
}

func (r *repository) FindStockByID(ctx context.Context, id uint) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.WithContext(ctx).First(&stock, id).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// ---- 执行 ----

func (r *repository) SumExecutions(ctx context.Context, activityID uint) (int, error) {
	var sum int
	err := r.db.WithContext(ctx).Model(&model.ExecutionRecord{}).
		Where("activity_id = ?", activityID).
		Select("COALESCE(SUM(count), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *repository) FindExecutionsByDateRange(ctx context.Context, activityID uint, start, end time.Time) ([]model.ExecutionRecord, error) {
	var records []model.ExecutionRecord
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND created_at >= ? AND created_at <= ?", activityID, start, end.Add(24*time.Hour)).
		Find(&records).Error
	return records, err
}

// ---- 库存 ----

func (r *repository) CreateStock(ctx context.Context, stock *model.Stock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

func (r *repository) DeleteStock(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Stock{}, id).Error
}

// ---- 摊销快照 ----

func (r *repository) DeleteSnapshotsByActivity(ctx context.Context, activityID uint) error {
	return r.db.WithContext(ctx).Where("activity_id = ?", activityID).Delete(&model.AmortizationSnapshot{}).Error
}

func (r *repository) UpsertSnapshot(ctx context.Context, snapshot *model.AmortizationSnapshot) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "activity_id"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"execution_count", "amortization_base", "daily_amount"}),
	}).Create(snapshot).Error
}

// ---- 审计日志 ----

func (r *repository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ---- 结算查询 ----

func (r *repository) FindSettlementByID(ctx context.Context, id uint) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.WithContext(ctx).First(&settlement, id).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

func (r *repository) FindRecoveryItemsBySettlement(ctx context.Context, settlementID uint) ([]model.RecoveryItem, error) {
	var items []model.RecoveryItem
	err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&items).Error
	return items, err
}
