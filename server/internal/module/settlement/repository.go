package settlement

import (
	"context"
	"school-system/internal/model"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

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
