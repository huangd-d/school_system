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

func (r *repository) FindSnapshots(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	var snapshots []model.AmortizationSnapshot
	err := r.db.WithContext(ctx).
		Where("activity_id = ? AND date BETWEEN ? AND ?", activityID, start, end).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *repository) FindSnapshotsByCampus(ctx context.Context, campusID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	var snapshots []model.AmortizationSnapshot
	err := r.db.WithContext(ctx).
		Joins("JOIN activities ON activities.id = amortization_snapshots.activity_id").
		Where("activities.campus_id = ? AND amortization_snapshots.date BETWEEN ? AND ?", campusID, start, end).
		Find(&snapshots).Error
	return snapshots, err
}
