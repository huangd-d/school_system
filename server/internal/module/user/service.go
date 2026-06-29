package user

import (
	"context"
	"fmt"
	"school-system/internal/model"
	"school-system/pkg/apperror"

	"golang.org/x/crypto/bcrypt"
)

// Repository 账户数据访问接口
type Repository interface {
	FindAll(ctx context.Context, campusID uint) ([]model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
}

// CampusLookup 校区查询接口（由 campus 模块满足，避免跨模块直接依赖）
type CampusLookup interface {
	FindByID(ctx context.Context, id uint) (*model.Campus, error)
}

// Service 账户业务逻辑
type Service struct {
	repo     Repository
	campuses CampusLookup
}

func NewService(repo Repository, campuses CampusLookup) *Service {
	return &Service{repo: repo, campuses: campuses}
}

// List 查询账户列表（总部管理员看全部，其他角色仅看本校区）
func (s *Service) List(ctx context.Context, operatorRole string, operatorCampusID uint) ([]model.User, error) {
	var filterCampusID uint
	if operatorRole != model.RoleHQAdmin {
		filterCampusID = operatorCampusID
	}
	users, err := s.repo.FindAll(ctx, filterCampusID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询账户列表失败: %v", err)
	}
	return users, nil
}

// Create 新建账户
func (s *Service) Create(ctx context.Context, username, password, role string, campusID uint) (*model.User, error) {
	// 参数校验
	if username == "" {
		return nil, apperror.ErrUserUsernameEmpty
	}
	if len(username) > 50 {
		return nil, apperror.ErrUserUsernameTooLong
	}
	if password == "" {
		return nil, apperror.ErrUserPasswordEmpty
	}
	if role != model.RoleHQAdmin && role != model.RoleCampusOperator && role != model.RoleActivityContact {
		return nil, apperror.ErrUserRoleInvalid
	}
	if campusID == 0 {
		return nil, apperror.ErrUserCampusRequired
	}

	// 验证校区存在且角色与校区类型匹配
	campus, err := s.campuses.FindByID(ctx, campusID)
	if err != nil {
		return nil, apperror.ErrUserCampusNotFound
	}
	if role == model.RoleHQAdmin && campus.Type != model.CampusTypeHQ {
		return nil, apperror.ErrUserHQAdminCampus
	}
	if role != model.RoleHQAdmin && campus.Type == model.CampusTypeHQ {
		return nil, apperror.ErrUserNormalCampus
	}

	// 用户名唯一
	if existing, _ := s.repo.FindByUsername(ctx, username); existing != nil {
		return nil, apperror.New(apperror.ErrUserUsernameDup.Code,
			fmt.Sprintf("用户名「%s」已存在", username))
	}

	// 密码 bcrypt 哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "密码加密失败: %v", err)
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		CampusID:     campusID,
		Status:       model.UserStatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "创建账户失败: %v", err)
	}

	return user, nil
}

// Update 编辑账户（修改角色和校区）
func (s *Service) Update(ctx context.Context, id uint, role string, campusID uint) (*model.User, error) {
	if role != model.RoleHQAdmin && role != model.RoleCampusOperator && role != model.RoleActivityContact {
		return nil, apperror.ErrUserRoleInvalid
	}
	if campusID == 0 {
		return nil, apperror.ErrUserCampusRequired
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}

	// 验证校区存在且角色与校区类型匹配
	campus, err := s.campuses.FindByID(ctx, campusID)
	if err != nil {
		return nil, apperror.ErrUserCampusNotFound
	}
	if role == model.RoleHQAdmin && campus.Type != model.CampusTypeHQ {
		return nil, apperror.ErrUserHQAdminCampus
	}
	if role != model.RoleHQAdmin && campus.Type == model.CampusTypeHQ {
		return nil, apperror.ErrUserNormalCampus
	}

	user.Role = role
	user.CampusID = campusID

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "更新账户失败: %v", err)
	}

	return user, nil
}

// Disable 禁用账户（不能禁用自己）
func (s *Service) Disable(ctx context.Context, id uint, operatorID uint) error {
	if id == operatorID {
		return apperror.ErrUserDisableSelf
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.ErrUserNotFound
	}

	user.Status = model.UserStatusDisabled
	if err := s.repo.Update(ctx, user); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "禁用账户失败: %v", err)
	}

	return nil
}

// ResetPassword 重置密码
func (s *Service) ResetPassword(ctx context.Context, id uint, newPassword string) error {
	if newPassword == "" {
		return apperror.ErrUserPasswordEmpty
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return apperror.ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "密码加密失败: %v", err)
	}

	user.PasswordHash = string(hash)
	if err := s.repo.Update(ctx, user); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "重置密码失败: %v", err)
	}

	return nil
}
