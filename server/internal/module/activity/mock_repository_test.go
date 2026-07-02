package activity_test

import (
	"context"

	"school-system/internal/model"
	"school-system/internal/module/activity"
)

// ---- MockActivityRepo ----

// MockActivityRepo 实现 activity.Repository 接口，每个方法对应一个可替换的函数字段。
type MockActivityRepo struct {
	FindAllFn            func(ctx context.Context) ([]model.Activity, error)
	FindByCampusIDFn     func(ctx context.Context, campusID uint) ([]model.Activity, error)
	FindByContactUserIDFn func(ctx context.Context, userID uint) ([]model.Activity, error)
	FindByIDFn           func(ctx context.Context, id uint) (*model.Activity, error)
	FindByIDsFn          func(ctx context.Context, ids []uint) ([]model.Activity, error)
	CreateFn             func(ctx context.Context, a *model.Activity) error
	UpdateFn             func(ctx context.Context, a *model.Activity) error
	SetContactsFn        func(ctx context.Context, activityID uint, userIDs []uint) error
	FindContactIDsFn     func(ctx context.Context, activityID uint) ([]uint, error)
	CreateExecutionFn    func(ctx context.Context, exec *model.ExecutionRecord) error
	FindExecutionsFn     func(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error)
	SumExecutionsFn      func(ctx context.Context, activityID uint) (int, error)
}

func (m *MockActivityRepo) FindAll(ctx context.Context) ([]model.Activity, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx)
	}
	return nil, nil
}

func (m *MockActivityRepo) FindByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error) {
	if m.FindByCampusIDFn != nil {
		return m.FindByCampusIDFn(ctx, campusID)
	}
	return nil, nil
}

func (m *MockActivityRepo) FindByContactUserID(ctx context.Context, userID uint) ([]model.Activity, error) {
	if m.FindByContactUserIDFn != nil {
		return m.FindByContactUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockActivityRepo) FindByID(ctx context.Context, id uint) (*model.Activity, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockActivityRepo) FindByIDs(ctx context.Context, ids []uint) ([]model.Activity, error) {
	if m.FindByIDsFn != nil {
		return m.FindByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *MockActivityRepo) Create(ctx context.Context, a *model.Activity) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, a)
	}
	return nil
}

func (m *MockActivityRepo) Update(ctx context.Context, a *model.Activity) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, a)
	}
	return nil
}

func (m *MockActivityRepo) SetContacts(ctx context.Context, activityID uint, userIDs []uint) error {
	if m.SetContactsFn != nil {
		return m.SetContactsFn(ctx, activityID, userIDs)
	}
	return nil
}

func (m *MockActivityRepo) FindContactIDs(ctx context.Context, activityID uint) ([]uint, error) {
	if m.FindContactIDsFn != nil {
		return m.FindContactIDsFn(ctx, activityID)
	}
	return nil, nil
}

func (m *MockActivityRepo) CreateExecution(ctx context.Context, exec *model.ExecutionRecord) error {
	if m.CreateExecutionFn != nil {
		return m.CreateExecutionFn(ctx, exec)
	}
	return nil
}

func (m *MockActivityRepo) FindExecutions(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error) {
	if m.FindExecutionsFn != nil {
		return m.FindExecutionsFn(ctx, activityID)
	}
	return nil, nil
}

func (m *MockActivityRepo) SumExecutions(ctx context.Context, activityID uint) (int, error) {
	if m.SumExecutionsFn != nil {
		return m.SumExecutionsFn(ctx, activityID)
	}
	return 0, nil
}

// 编译期检查：确保 MockActivityRepo 实现了 activity.Repository
var _ activity.Repository = (*MockActivityRepo)(nil)

// ---- MockCampusLookup ----

// MockCampusLookup 实现 activity.CampusLookup 接口。
type MockCampusLookup struct {
	FindByIDFn func(ctx context.Context, id uint) (*model.Campus, error)
}

func (m *MockCampusLookup) FindByID(ctx context.Context, id uint) (*model.Campus, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

var _ activity.CampusLookup = (*MockCampusLookup)(nil)

// ---- MockUserLookup ----

// MockUserLookup 实现 activity.UserLookup 接口。
type MockUserLookup struct {
	FindByIDFn func(ctx context.Context, id uint) (*model.User, error)
}

func (m *MockUserLookup) FindByID(ctx context.Context, id uint) (*model.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

var _ activity.UserLookup = (*MockUserLookup)(nil)

// ---- MockActivityService ----

// MockActivityService 实现 activity.ServiceInterface 接口，用于 handler 层测试。
type MockActivityService struct {
	ListFn         func(ctx context.Context, userID, campusID uint, role string) ([]activity.ActivityWithSummary, error)
	CreateFn       func(ctx context.Context, name string, campusID uint, contactIDs []uint, plannedExec int, startDate, endDate string, createdBy uint, creatorRole string, creatorCampusID uint) (*model.Activity, error)
	UpdateFn       func(ctx context.Context, id uint, name string, contactIDs []uint, plannedExec int, operatorRole string, operatorCampusID uint) (*model.Activity, error)
	DetailFn       func(ctx context.Context, id uint) (*activity.ActivityDetail, error)
	AddExecutionFn func(ctx context.Context, activityID uint, count int, recordedBy uint, operatorRole string) error
	ArchiveFn      func(ctx context.Context, id uint, operatorRole string, operatorCampusID uint) error
}

func (m *MockActivityService) List(ctx context.Context, userID, campusID uint, role string) ([]activity.ActivityWithSummary, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, userID, campusID, role)
	}
	return nil, nil
}

func (m *MockActivityService) Create(ctx context.Context, name string, campusID uint, contactIDs []uint, plannedExec int, startDate, endDate string, createdBy uint, creatorRole string, creatorCampusID uint) (*model.Activity, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, name, campusID, contactIDs, plannedExec, startDate, endDate, createdBy, creatorRole, creatorCampusID)
	}
	return nil, nil
}

func (m *MockActivityService) Update(ctx context.Context, id uint, name string, contactIDs []uint, plannedExec int, operatorRole string, operatorCampusID uint) (*model.Activity, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, name, contactIDs, plannedExec, operatorRole, operatorCampusID)
	}
	return nil, nil
}

func (m *MockActivityService) Detail(ctx context.Context, id uint) (*activity.ActivityDetail, error) {
	if m.DetailFn != nil {
		return m.DetailFn(ctx, id)
	}
	return nil, nil
}

func (m *MockActivityService) AddExecution(ctx context.Context, activityID uint, count int, recordedBy uint, operatorRole string) error {
	if m.AddExecutionFn != nil {
		return m.AddExecutionFn(ctx, activityID, count, recordedBy, operatorRole)
	}
	return nil
}

func (m *MockActivityService) Archive(ctx context.Context, id uint, operatorRole string, operatorCampusID uint) error {
	if m.ArchiveFn != nil {
		return m.ArchiveFn(ctx, id, operatorRole, operatorCampusID)
	}
	return nil
}

var _ activity.ServiceInterface = (*MockActivityService)(nil)
