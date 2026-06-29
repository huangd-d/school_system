package campus

import (
	"context"
	"school-system/internal/model"

	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindAll 查询全部校区
func (r *repository) FindAll(ctx context.Context) ([]model.Campus, error) {
	var campuses []model.Campus
	err := r.db.WithContext(ctx).Order("id ASC").Find(&campuses).Error
	return campuses, err
}

// FindByID 按主键查询
func (r *repository) FindByID(ctx context.Context, id uint) (*model.Campus, error) {
	var campus model.Campus
	err := r.db.WithContext(ctx).First(&campus, id).Error
	if err != nil {
		return nil, err
	}
	return &campus, nil
}

// FindByName 按名称查重
func (r *repository) FindByName(ctx context.Context, name string) (*model.Campus, error) {
	var campus model.Campus
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&campus).Error
	if err != nil {
		return nil, err
	}
	return &campus, nil
}

// FindByType 按类型查询（用于检查总部是否已存在）
func (r *repository) FindByType(ctx context.Context, campusType string) ([]model.Campus, error) {
	var campuses []model.Campus
	err := r.db.WithContext(ctx).Where("type = ?", campusType).Find(&campuses).Error
	return campuses, err
}

// Create 新增校区
func (r *repository) Create(ctx context.Context, campus *model.Campus) error {
	return r.db.WithContext(ctx).Create(campus).Error
}

// Update 更新校区
func (r *repository) Update(ctx context.Context, campus *model.Campus) error {
	return r.db.WithContext(ctx).Save(campus).Error
}

// Delete 删除校区
func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Campus{}, id).Error
}

// CountUsers 统计该校区下的账户数
func (r *repository) CountUsers(ctx context.Context, campusID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("campus_id = ?", campusID).Count(&count).Error
	return count, err
}

// CountActivities 统计该校区下的活动数
func (r *repository) CountActivities(ctx context.Context, campusID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Activity{}).Where("campus_id = ?", campusID).Count(&count).Error
	return count, err
}
