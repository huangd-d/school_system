package campus

import (
	"context"
	"fmt"
	"school-system/internal/model"
	"school-system/pkg/apperror"
)

// Repository 校区数据访问接口
type Repository interface {
	FindAll(ctx context.Context) ([]model.Campus, error)
	FindByID(ctx context.Context, id uint) (*model.Campus, error)
	FindByName(ctx context.Context, name string) (*model.Campus, error)
	FindByType(ctx context.Context, campusType string) ([]model.Campus, error)
	Create(ctx context.Context, campus *model.Campus) error
	Update(ctx context.Context, campus *model.Campus) error
	Delete(ctx context.Context, id uint) error
	CountUsers(ctx context.Context, campusID uint) (int64, error)
	CountActivities(ctx context.Context, campusID uint) (int64, error)
}

// Service 校区业务逻辑
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List 查询全部校区
func (s *Service) List(ctx context.Context) ([]model.Campus, error) {
	campuses, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询校区列表失败: %v", err)
	}
	return campuses, nil
}

// Create 新建校区
func (s *Service) Create(ctx context.Context, name string, campusType string) (*model.Campus, error) {
	// 名称非空
	if name == "" {
		return nil, apperror.ErrCampusNameEmpty
	}
	if len(name) > 100 {
		return nil, apperror.ErrCampusNameTooLong
	}

	// 类型校验
	if campusType != model.CampusTypeNormal && campusType != model.CampusTypeHQ {
		return nil, apperror.ErrCampusTypeInvalid
	}

	// 总部唯一
	if campusType == model.CampusTypeHQ {
		existing, _ := s.repo.FindByType(ctx, model.CampusTypeHQ)
		if len(existing) > 0 {
			return nil, apperror.ErrCampusHQExists
		}
	}

	// 名称唯一
	if existing, _ := s.repo.FindByName(ctx, name); existing != nil {
		return nil, apperror.New(apperror.ErrCampusNameDup.Code,
			fmt.Sprintf("校区名称「%s」已存在", name))
	}

	campus := &model.Campus{
		Name: name,
		Type: campusType,
	}

	if err := s.repo.Create(ctx, campus); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "创建校区失败: %v", err)
	}

	return campus, nil
}

// Update 编辑校区（仅允许修改名称）
func (s *Service) Update(ctx context.Context, id uint, name string) (*model.Campus, error) {
	if name == "" {
		return nil, apperror.ErrCampusNameEmpty
	}
	if len(name) > 100 {
		return nil, apperror.ErrCampusNameTooLong
	}

	campus, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrCampusNotFound
	}

	// 名称唯一（排除自身）
	if existing, _ := s.repo.FindByName(ctx, name); existing != nil && existing.ID != id {
		return nil, apperror.New(apperror.ErrCampusNameDup.Code,
			fmt.Sprintf("校区名称「%s」已存在", name))
	}

	campus.Name = name

	if err := s.repo.Update(ctx, campus); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "更新校区失败: %v", err)
	}

	return campus, nil
}

// Delete 删除校区
func (s *Service) Delete(ctx context.Context, id uint) error {
	campus, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.ErrCampusNotFound
	}

	// 总部不可删除
	if campus.Type == model.CampusTypeHQ {
		return apperror.ErrCampusHQDelete
	}

	// 存在关联账户时不可删除
	userCount, err := s.repo.CountUsers(ctx, id)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "查询校区关联账户失败: %v", err)
	}
	if userCount > 0 {
		return apperror.New(apperror.ErrCampusHasUsers.Code,
			fmt.Sprintf("该校区下还有 %d 个账户，请先转移或删除账户后再删除校区", userCount))
	}

	// 存在关联活动时不可删除
	activityCount, err := s.repo.CountActivities(ctx, id)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "查询校区关联活动失败: %v", err)
	}
	if activityCount > 0 {
		return apperror.New(apperror.ErrCampusHasActivities.Code,
			fmt.Sprintf("该校区下还有 %d 个活动，请先转移或删除活动后再删除校区", activityCount))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "删除校区失败: %v", err)
	}

	return nil
}
