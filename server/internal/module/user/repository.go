package user

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

// FindAll 查询全部账户（按校区过滤时传 campusID>0）
func (r *repository) FindAll(ctx context.Context, campusID uint) ([]model.User, error) {
	var users []model.User
	q := r.db.WithContext(ctx).Order("id ASC")
	if campusID > 0 {
		q = q.Where("campus_id = ?", campusID)
	}
	err := q.Find(&users).Error
	return users, err
}

// FindByID 按主键查询
func (r *repository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 按用户名查询（登录、查重用）
func (r *repository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 新增账户
func (r *repository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新账户
func (r *repository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
