package report

import (
	"context"
	"school-system/internal/model"
	"time"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

// ---- 已有方法 ----

func (r *repository) FindSnapshots(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	var snapshots []model.AmortizationSnapshot
	q := r.db.WithContext(ctx).Where("activity_id = ?", activityID)
	if !start.IsZero() {
		q = q.Where("date >= ?", start)
	}
	if !end.IsZero() {
		q = q.Where("date <= ?", end)
	}
	err := q.Order("date").Find(&snapshots).Error
	return snapshots, err
}

func (r *repository) FindSnapshotsByCampus(ctx context.Context, campusID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	var snapshots []model.AmortizationSnapshot
	err := r.db.WithContext(ctx).
		Table("amortization_snapshots").
		Joins("JOIN activities ON activities.id = amortization_snapshots.activity_id").
		Where("activities.campus_id = ?", campusID).
		Where("amortization_snapshots.date >= ? AND amortization_snapshots.date <= ?", start, end).
		Order("amortization_snapshots.date").
		Find(&snapshots).Error
	return snapshots, err
}

// ---- 新增方法 ----

func (r *repository) FindSnapshotsByDateRange(ctx context.Context, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	var snapshots []model.AmortizationSnapshot
	err := r.db.WithContext(ctx).
		Where("date >= ? AND date <= ?", start, end).
		Order("date").
		Find(&snapshots).Error
	return snapshots, err
}

func (r *repository) FindActivitiesByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error) {
	var activities []model.Activity
	q := r.db.WithContext(ctx)
	if campusID != 0 {
		q = q.Where("campus_id = ?", campusID)
	}
	err := q.Find(&activities).Error
	return activities, err
}

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

func (r *repository) FindAllActivities(ctx context.Context) ([]model.Activity, error) {
	var activities []model.Activity
	err := r.db.WithContext(ctx).Find(&activities).Error
	return activities, err
}

func (r *repository) FindDistributionAggByCategory(ctx context.Context, start, end time.Time) ([]CategoryAggRow, error) {
	var rows []CategoryAggRow
	err := r.db.WithContext(ctx).
		Table("distributions").
		Select("stocks.category_id, material_categories.name as category_name, "+
			"CAST(SUM(distributions.quantity) AS INTEGER) as total_quantity, "+
			"SUM(distributions.quantity * stocks.unit_price) as total_amount").
		Joins("JOIN stocks ON stocks.id = distributions.stock_id").
		Joins("JOIN material_categories ON material_categories.id = stocks.category_id").
		Where("distributions.created_at >= ? AND distributions.created_at <= ?",
			start, end.Add(24*time.Hour)).
		Group("stocks.category_id").
		Order("total_amount DESC").
		Find(&rows).Error
	return rows, err
}
