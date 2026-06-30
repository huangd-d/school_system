package user_test

import (
	"context"

	"school-system/internal/model"
	"school-system/internal/module/user"
)

// ---- MockUserRepo ----

// MockUserRepo 实现 user.Repository 接口，每个方法对应一个可替换的函数字段。
type MockUserRepo struct {
	FindAllFn        func(ctx context.Context, campusID uint) ([]model.User, error)
	FindByIDFn       func(ctx context.Context, id uint) (*model.User, error)
	FindByUsernameFn func(ctx context.Context, username string) (*model.User, error)
	CreateFn         func(ctx context.Context, user *model.User) error
	UpdateFn         func(ctx context.Context, user *model.User) error
}

func (m *MockUserRepo) FindAll(ctx context.Context, campusID uint) ([]model.User, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx, campusID)
	}
	return nil, nil
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.FindByUsernameFn != nil {
		return m.FindByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *MockUserRepo) Create(ctx context.Context, u *model.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, u)
	}
	return nil
}

func (m *MockUserRepo) Update(ctx context.Context, u *model.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, u)
	}
	return nil
}

// 编译期检查：确保 MockUserRepo 实现了 user.Repository
var _ user.Repository = (*MockUserRepo)(nil)

// ---- MockCampusLookup ----

// MockCampusLookup 实现 user.CampusLookup 接口。
type MockCampusLookup struct {
	FindByIDFn func(ctx context.Context, id uint) (*model.Campus, error)
}

func (m *MockCampusLookup) FindByID(ctx context.Context, id uint) (*model.Campus, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

// 编译期检查：确保 MockCampusLookup 实现了 user.CampusLookup
var _ user.CampusLookup = (*MockCampusLookup)(nil)

// ---- MockUserService ----

// MockUserService 实现 user.ServiceInterface 接口，用于 handler 层测试。
type MockUserService struct {
	ListFn          func(ctx context.Context, operatorRole string, operatorCampusID uint) ([]model.User, error)
	CreateFn        func(ctx context.Context, username, password, phone, role string, campusID uint) (*model.User, error)
	UpdateFn        func(ctx context.Context, id uint, username, phone, role string, campusID uint) (*model.User, error)
	DisableFn       func(ctx context.Context, id uint, operatorID uint) error
	EnableFn        func(ctx context.Context, id uint) error
	ResetPasswordFn func(ctx context.Context, id uint, newPassword string) error
}

func (m *MockUserService) List(ctx context.Context, operatorRole string, operatorCampusID uint) ([]model.User, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, operatorRole, operatorCampusID)
	}
	return nil, nil
}

func (m *MockUserService) Create(ctx context.Context, username, password, phone, role string, campusID uint) (*model.User, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, username, password, phone, role, campusID)
	}
	return nil, nil
}

func (m *MockUserService) Update(ctx context.Context, id uint, username, phone, role string, campusID uint) (*model.User, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, username, phone, role, campusID)
	}
	return nil, nil
}

func (m *MockUserService) Disable(ctx context.Context, id uint, operatorID uint) error {
	if m.DisableFn != nil {
		return m.DisableFn(ctx, id, operatorID)
	}
	return nil
}

func (m *MockUserService) Enable(ctx context.Context, id uint) error {
	if m.EnableFn != nil {
		return m.EnableFn(ctx, id)
	}
	return nil
}

func (m *MockUserService) ResetPassword(ctx context.Context, id uint, newPassword string) error {
	if m.ResetPasswordFn != nil {
		return m.ResetPasswordFn(ctx, id, newPassword)
	}
	return nil
}

// 编译期检查：确保 MockUserService 实现了 user.ServiceInterface
var _ user.ServiceInterface = (*MockUserService)(nil)
