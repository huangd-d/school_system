package campus_test

import (
	"context"

	"school-system/internal/model"
	"school-system/internal/module/campus"
)

// ---- MockCampusRepo ----

// MockCampusRepo 实现 campus.Repository 接口，每个方法对应一个可替换的函数字段。
type MockCampusRepo struct {
	FindAllFn        func(ctx context.Context) ([]model.Campus, error)
	FindByIDFn       func(ctx context.Context, id uint) (*model.Campus, error)
	FindByNameFn     func(ctx context.Context, name string) (*model.Campus, error)
	FindByTypeFn     func(ctx context.Context, campusType string) ([]model.Campus, error)
	CreateFn         func(ctx context.Context, c *model.Campus) error
	UpdateFn         func(ctx context.Context, c *model.Campus) error
	DeleteFn         func(ctx context.Context, id uint) error
	CountUsersFn     func(ctx context.Context, campusID uint) (int64, error)
	CountActivitiesFn func(ctx context.Context, campusID uint) (int64, error)
}

func (m *MockCampusRepo) FindAll(ctx context.Context) ([]model.Campus, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return nil, nil
}

func (m *MockCampusRepo) FindByID(ctx context.Context, id uint) (*model.Campus, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockCampusRepo) FindByName(ctx context.Context, name string) (*model.Campus, error) {
	if m.FindByNameFn != nil {
		return m.FindByNameFn(ctx, name)
	}
	return nil, nil
}

func (m *MockCampusRepo) FindByType(ctx context.Context, campusType string) ([]model.Campus, error) {
	if m.FindByTypeFn != nil {
		return m.FindByTypeFn(ctx, campusType)
	}
	return nil, nil
}

func (m *MockCampusRepo) Create(ctx context.Context, c *model.Campus) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, c)
	}
	return nil
}

func (m *MockCampusRepo) Update(ctx context.Context, c *model.Campus) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, c)
	}
	return nil
}

func (m *MockCampusRepo) Delete(ctx context.Context, id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *MockCampusRepo) CountUsers(ctx context.Context, campusID uint) (int64, error) {
	if m.CountUsersFn != nil {
		return m.CountUsersFn(ctx, campusID)
	}
	return 0, nil
}

func (m *MockCampusRepo) CountActivities(ctx context.Context, campusID uint) (int64, error) {
	if m.CountActivitiesFn != nil {
		return m.CountActivitiesFn(ctx, campusID)
	}
	return 0, nil
}

// 编译期检查：确保 MockCampusRepo 实现了 campus.Repository
var _ campus.Repository = (*MockCampusRepo)(nil)

// ---- MockCampusService ----

// MockCampusService 实现 campus.ServiceInterface 接口，用于 handler 层测试。
type MockCampusService struct {
	ListFn   func(ctx context.Context) ([]model.Campus, error)
	CreateFn func(ctx context.Context, name string, campusType string) (*model.Campus, error)
	UpdateFn func(ctx context.Context, id uint, name string) (*model.Campus, error)
	DeleteFn func(ctx context.Context, id uint) error
}

func (m *MockCampusService) List(ctx context.Context) ([]model.Campus, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return nil, nil
}

func (m *MockCampusService) Create(ctx context.Context, name string, campusType string) (*model.Campus, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, name, campusType)
	}
	return nil, nil
}

func (m *MockCampusService) Update(ctx context.Context, id uint, name string) (*model.Campus, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, name)
	}
	return nil, nil
}

func (m *MockCampusService) Delete(ctx context.Context, id uint) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

// 编译期检查：确保 MockCampusService 实现了 campus.ServiceInterface
var _ campus.ServiceInterface = (*MockCampusService)(nil)
