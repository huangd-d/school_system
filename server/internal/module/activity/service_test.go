package activity_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/activity"
	"school-system/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- 辅助 ----

// newMockRepo 返回一个所有方法默认返回 nil 的 MockActivityRepo。
func newMockRepo() *MockActivityRepo {
	return &MockActivityRepo{}
}

// newMockCampuses 返回一个默认 FindByID 返回普通校区的 MockCampusLookup。
func newMockCampuses() *MockCampusLookup {
	return &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return &model.Campus{ID: id, Name: "测试校区", Type: model.CampusTypeNormal}, nil
		},
	}
}

// newMockUsers 返回一个默认 FindByID 返回用户（属于测试校区）的 MockUserLookup。
func newMockUsers() *MockUserLookup {
	return &MockUserLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
			return &model.User{ID: id, Username: "user", CampusID: 1}, nil
		},
	}
}

// newService 用默认 mock 构造 Service。
func newService(repo *MockActivityRepo, campuses *MockCampusLookup, users *MockUserLookup) *activity.Service {
	if repo == nil {
		repo = newMockRepo()
	}
	if campuses == nil {
		campuses = newMockCampuses()
	}
	if users == nil {
		users = newMockUsers()
	}
	return activity.NewService(repo, campuses, users)
}

// assertAppError 断言 err 是 *AppError 且 Code 与期望匹配。
func assertAppError(t *testing.T, err error, expected *apperror.AppError) {
	t.Helper()
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr), "错误不是 *apperror.AppError: %v", err)
	assert.Equal(t, expected.Code, appErr.Code, "错误码不匹配: got %d, want %d", appErr.Code, expected.Code)
}

// ============================================================
//  List
// ============================================================

func TestActivityService_List_HQAdmin(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context) ([]model.Activity, error) {
		return []model.Activity{{Name: "活动A"}}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return []uint{2, 3}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 5, nil
	}
	svc := newService(mockRepo, nil, nil)

	activities, err := svc.List(context.Background(), 0, 0, model.RoleHQAdmin)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, "活动A", activities[0].Name)
	assert.Equal(t, []uint{2, 3}, activities[0].ContactIDs)
	assert.Equal(t, 5, activities[0].TotalExecuted)
}

func TestActivityService_List_CampusOperator(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByCampusIDFn = func(ctx context.Context, campusID uint) ([]model.Activity, error) {
		assert.Equal(t, uint(5), campusID)
		return []model.Activity{{Name: "活动B"}}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return nil, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 0, nil
	}
	svc := newService(mockRepo, nil, nil)

	activities, err := svc.List(context.Background(), 0, 5, model.RoleCampusOperator)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
}

func TestActivityService_List_ActivityContact(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByContactUserIDFn = func(ctx context.Context, userID uint) ([]model.Activity, error) {
		assert.Equal(t, uint(7), userID)
		return []model.Activity{{Name: "活动C"}}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return nil, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 0, nil
	}
	svc := newService(mockRepo, nil, nil)

	activities, err := svc.List(context.Background(), 7, 0, model.RoleActivityContact)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
}

func TestActivityService_List_InvalidRole(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.List(context.Background(), 1, 1, "invalid_role")
	assertAppError(t, err, apperror.ErrActivityPermissionDenied)
}

func TestActivityService_List_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context) ([]model.Activity, error) {
		return nil, errors.New("db error")
	}
	svc := newService(mockRepo, nil, nil)

	_, err := svc.List(context.Background(), 0, 0, model.RoleHQAdmin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查询活动列表失败")
}

// ============================================================
//  Create
// ============================================================

func TestActivityService_Create_EmptyName(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.Create(context.Background(), "", 1, nil, 10, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrActivityNameEmpty)
}

func TestActivityService_Create_NameTooLong(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.Create(context.Background(), strings.Repeat("x", 201), 1, nil, 10, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrActivityNameTooLong)
}

func TestActivityService_Create_PlannedExecInvalid(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.Create(context.Background(), "活动", 1, nil, 0, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrActivityPlannedExecInvalid)
}

func TestActivityService_Create_PermissionDenied(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.Create(context.Background(), "活动", 1, nil, 10, "2025-01-01", "2025-12-31", 1, model.RoleActivityContact, 1)
	assertAppError(t, err, apperror.ErrActivityPermissionDenied)
}

func TestActivityService_Create_CampusOperatorCampusMismatch(t *testing.T) {
	svc := newService(nil, nil, nil)
	// 校区操作员所属校区=1，但要创建活动在校区=2
	_, err := svc.Create(context.Background(), "活动", 2, nil, 10, "2025-01-01", "2025-12-31", 1, model.RoleCampusOperator, 1)
	assertAppError(t, err, apperror.ErrActivityCampusMismatch)
}

func TestActivityService_Create_CampusNotFound(t *testing.T) {
	mockCampuses := &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := newService(nil, mockCampuses, nil)
	_, err := svc.Create(context.Background(), "活动", 99, nil, 10, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityCampusNotFound)
}

func TestActivityService_Create_DateInvalid(t *testing.T) {
	svc := newService(nil, nil, nil)
	// 结束日期早于开始日期
	_, err := svc.Create(context.Background(), "活动", 1, nil, 10, "2025-12-31", "2025-01-01", 1, model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrActivityDateInvalid)
}

func TestActivityService_Create_DateParseError(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.Create(context.Background(), "活动", 1, nil, 10, "not-a-date", "2025-12-31", 1, model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrActivityDateInvalid)
}

func TestActivityService_Create_ContactNotFound(t *testing.T) {
	mockUsers := &MockUserLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := newService(nil, nil, mockUsers)
	_, err := svc.Create(context.Background(), "活动", 1, []uint{99}, 10, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrActivityContactsNotFound)
}

func TestActivityService_Create_ContactAcrossCampus(t *testing.T) {
	// 联系人可以跨校区（不再限制同一校区）
	mockRepo := newMockRepo()
	mockRepo.CreateFn = func(ctx context.Context, a *model.Activity) error {
		a.ID = 10
		return nil
	}
	mockUsers := &MockUserLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
			// 联系人属于校区 2，活动在校区 1 — 应允许
			return &model.User{ID: id, Username: "user", CampusID: 2}, nil
		},
	}
	svc := newService(mockRepo, nil, mockUsers)
	created, err := svc.Create(context.Background(), "活动", 1, []uint{5}, 10, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	require.NoError(t, err)
	assert.Equal(t, uint(10), created.ID)
}

func TestActivityService_Create_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.CreateFn = func(ctx context.Context, a *model.Activity) error {
		a.ID = 10 // 模拟数据库生成 ID
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	created, err := svc.Create(context.Background(), "新活动", 1, []uint{2, 3}, 20, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	require.NoError(t, err)
	assert.Equal(t, uint(10), created.ID)
	assert.Equal(t, "新活动", created.Name)
	assert.Equal(t, model.ActivityNotStarted, created.Status)
	assert.Equal(t, 20, created.PlannedExecutions)
}

func TestActivityService_Create_Success_NoContacts(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.CreateFn = func(ctx context.Context, a *model.Activity) error {
		a.ID = 20
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	created, err := svc.Create(context.Background(), "活动无联系人", 1, nil, 5, "2025-06-01", "2025-06-30", 1, model.RoleHQAdmin, 1)
	require.NoError(t, err)
	assert.Equal(t, uint(20), created.ID)
}

func TestActivityService_Create_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.CreateFn = func(ctx context.Context, a *model.Activity) error {
		return errors.New("db error")
	}
	svc := newService(mockRepo, nil, nil)

	_, err := svc.Create(context.Background(), "活动", 1, nil, 10, "2025-01-01", "2025-12-31", 1, model.RoleHQAdmin, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "创建活动失败")
}

// ============================================================
//  Update
// ============================================================

func TestActivityService_Update_PermissionDenied(t *testing.T) {
	svc := newService(nil, nil, nil)
	_, err := svc.Update(context.Background(), 1, "新名称", nil, 0, model.RoleActivityContact, 1)
	assertAppError(t, err, apperror.ErrActivityPermissionDenied)
}

func TestActivityService_Update_ActivityNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, nil, nil)
	_, err := svc.Update(context.Background(), 999, "新名称", nil, 0, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityNotFound)
}

func TestActivityService_Update_CampusOperatorPermission(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		// 活动属于校区 2，但操作员属于校区 1
		return &model.Activity{ID: 1, Name: "活动", CampusID: 2, Status: model.ActivityNotStarted}, nil
	}
	svc := newService(mockRepo, nil, nil)
	_, err := svc.Update(context.Background(), 1, "新名称", nil, 0, model.RoleCampusOperator, 1)
	assertAppError(t, err, apperror.ErrActivityPermissionDenied)
}

func TestActivityService_Update_ArchivedCannotModify(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, Status: model.ActivityArchived}, nil
	}
	svc := newService(mockRepo, nil, nil)
	_, err := svc.Update(context.Background(), 1, "新名称", nil, 0, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityArchivedCannotModify)
}

func TestActivityService_Update_NameTooLong(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, Status: model.ActivityNotStarted}, nil
	}
	svc := newService(mockRepo, nil, nil)
	_, err := svc.Update(context.Background(), 1, strings.Repeat("x", 201), nil, 0, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityNameTooLong)
}

func TestActivityService_Update_PlannedExecBelowExecuted(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, PlannedExecutions: 10, Status: model.ActivityNotStarted}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 8, nil // 已执行 8 次
	}
	mockRepo.UpdateFn = func(ctx context.Context, a *model.Activity) error { return nil }
	svc := newService(mockRepo, nil, nil)
	// 尝试将计划次数改为 5（< 8）
	_, err := svc.Update(context.Background(), 1, "", nil, 5, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityPlannedExecBelowExecuted)
}

func TestActivityService_Update_ContactsNotModifiedWhenNil(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, PlannedExecutions: 10, Status: model.ActivityNotStarted}, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, a *model.Activity) error { return nil }
	setContactsCalled := false
	mockRepo.SetContactsFn = func(ctx context.Context, activityID uint, userIDs []uint) error {
		setContactsCalled = true
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	// contactIDs = nil 表示不修改联系人
	updated, err := svc.Update(context.Background(), 1, "新名称", nil, 0, model.RoleHQAdmin, 0)
	require.NoError(t, err)
	assert.Equal(t, "新名称", updated.Name)
	assert.False(t, setContactsCalled, "contactIDs=nil 时不应调用 SetContacts")
}

func TestActivityService_Update_ClearContacts(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, PlannedExecutions: 10, Status: model.ActivityNotStarted}, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, a *model.Activity) error { return nil }
	setContactsCalled := false
	mockRepo.SetContactsFn = func(ctx context.Context, activityID uint, userIDs []uint) error {
		setContactsCalled = true
		assert.Empty(t, userIDs, "清空联系人时 userIDs 应为空切片")
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	// contactIDs = [] 表示清空所有联系人
	_, err := svc.Update(context.Background(), 1, "", []uint{}, 0, model.RoleHQAdmin, 0)
	require.NoError(t, err)
	assert.True(t, setContactsCalled, "contactIDs=[] 时应调用 SetContacts")
}

func TestActivityService_Update_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "旧名称", CampusID: 1, PlannedExecutions: 5, Status: model.ActivityNotStarted}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 3, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, a *model.Activity) error {
		return nil
	}
	mockRepo.SetContactsFn = func(ctx context.Context, activityID uint, userIDs []uint) error {
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	updated, err := svc.Update(context.Background(), 1, "新名称", []uint{2}, 10, model.RoleHQAdmin, 0)
	require.NoError(t, err)
	assert.Equal(t, "新名称", updated.Name)
	assert.Equal(t, 10, updated.PlannedExecutions)
}

// ============================================================
//  Detail
// ============================================================

func TestActivityService_Detail_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, nil, nil)
	_, err := svc.Detail(context.Background(), 999)
	assertAppError(t, err, apperror.ErrActivityNotFound)
}

func TestActivityService_Detail_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{
			ID: 1, Name: "活动详情", CampusID: 1, PlannedExecutions: 10,
			StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			Status:    model.ActivityNotStarted, CreatedBy: 1,
		}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return []uint{2}, nil
	}
	mockRepo.FindExecutionsFn = func(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error) {
		return []model.ExecutionRecord{{ID: 1, ActivityID: 1, Count: 3, RecordedBy: 2}}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 3, nil
	}
	mockUsers := &MockUserLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
			return &model.User{ID: id, Username: fmt.Sprintf("user%d", id), Phone: "138", Role: model.RoleActivityContact, CampusID: 1}, nil
		},
	}
	svc := newService(mockRepo, nil, mockUsers)

	detail, err := svc.Detail(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "活动详情", detail.Name)
	assert.Len(t, detail.Contacts, 1)
	assert.Len(t, detail.Executions, 1)
	assert.Equal(t, 3, detail.TotalExecuted)
}

func TestActivityService_Detail_SkipDeletedContact(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{
			ID: 1, Name: "活动", CampusID: 1, PlannedExecutions: 5,
			StartDate: time.Now().AddDate(0, 0, -10),
			EndDate:   time.Now().AddDate(0, 0, 10),
			Status:    model.ActivityInProgress, CreatedBy: 1,
		}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return []uint{2, 3, 4}, nil
	}
	mockRepo.FindExecutionsFn = func(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error) {
		return nil, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 0, nil
	}
	// 用户 3 找不到（已删除），用户 2 和 4 存在
	mockUsers := &MockUserLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
			if id == 3 {
				return nil, gorm.ErrRecordNotFound
			}
			return &model.User{ID: id, Username: fmt.Sprintf("user%d", id)}, nil
		},
	}
	svc := newService(mockRepo, nil, mockUsers)

	detail, err := svc.Detail(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, detail.Contacts, 2, "应跳过已删除的联系人")
}

// ============================================================
//  AddExecution
// ============================================================

func TestActivityService_AddExecution_CountInvalid(t *testing.T) {
	svc := newService(nil, nil, nil)
	err := svc.AddExecution(context.Background(), 1, 0, 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrActivityExecCountInvalid)
}

func TestActivityService_AddExecution_ActivityNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, nil, nil)
	err := svc.AddExecution(context.Background(), 999, 3, 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrActivityNotFound)
}

func TestActivityService_AddExecution_StatusNotAllowed(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "已结束活动", PlannedExecutions: 10, Status: model.ActivityEnded}, nil
	}
	svc := newService(mockRepo, nil, nil)
	err := svc.AddExecution(context.Background(), 1, 3, 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrActivityStatusNoExec)
}

func TestActivityService_AddExecution_NotContactPerson(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", PlannedExecutions: 10, Status: model.ActivityInProgress}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return []uint{2, 3}, nil // 联系人只有 2 和 3
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 0, nil
	}
	svc := newService(mockRepo, nil, nil)
	// recordedBy=5 不是联系人，也不是 hq_admin
	err := svc.AddExecution(context.Background(), 1, 3, 5, model.RoleActivityContact)
	assertAppError(t, err, apperror.ErrActivityNotContactPerson)
}

func TestActivityService_AddExecution_ExceedPlanned(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", PlannedExecutions: 10, Status: model.ActivityInProgress}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 8, nil // 已执行 8 次
	}
	svc := newService(mockRepo, nil, nil)
	// 再加 3 次 = 11 > 10
	err := svc.AddExecution(context.Background(), 1, 3, 1, model.RoleHQAdmin)
	assertAppError(t, err, apperror.ErrActivityExecExceedPlanned)
}

func TestActivityService_AddExecution_Success_HQAdmin(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", PlannedExecutions: 10, Status: model.ActivityNotStarted}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 0, nil
	}
	execCreated := false
	mockRepo.CreateExecutionFn = func(ctx context.Context, exec *model.ExecutionRecord) error {
		execCreated = true
		assert.Equal(t, uint(1), exec.ActivityID)
		assert.Equal(t, 5, exec.Count)
		assert.Equal(t, uint(1), exec.RecordedBy)
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	err := svc.AddExecution(context.Background(), 1, 5, 1, model.RoleHQAdmin)
	require.NoError(t, err)
	assert.True(t, execCreated)
}

func TestActivityService_AddExecution_Success_ContactPerson(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", PlannedExecutions: 10, Status: model.ActivityInProgress}, nil
	}
	mockRepo.FindContactIDsFn = func(ctx context.Context, activityID uint) ([]uint, error) {
		return []uint{2, 3}, nil
	}
	mockRepo.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 2, nil
	}
	mockRepo.CreateExecutionFn = func(ctx context.Context, exec *model.ExecutionRecord) error {
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	// recordedBy=2 是本活动联系人
	err := svc.AddExecution(context.Background(), 1, 3, 2, model.RoleActivityContact)
	require.NoError(t, err)
}

// ============================================================
//  Archive
// ============================================================

func TestActivityService_Archive_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newService(mockRepo, nil, nil)
	err := svc.Archive(context.Background(), 999, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityNotFound)
}

func TestActivityService_Archive_PermissionDenied(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, Status: model.ActivitySettled}, nil
	}
	svc := newService(mockRepo, nil, nil)
	err := svc.Archive(context.Background(), 1, model.RoleActivityContact, 1)
	assertAppError(t, err, apperror.ErrActivityPermissionDenied)
}

func TestActivityService_Archive_CampusOperatorMismatch(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 2, Status: model.ActivitySettled}, nil
	}
	svc := newService(mockRepo, nil, nil)
	// operatorCampusID=1 != activity.CampusID=2
	err := svc.Archive(context.Background(), 1, model.RoleCampusOperator, 1)
	assertAppError(t, err, apperror.ErrActivityPermissionDenied)
}

func TestActivityService_Archive_NotSettled(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", Status: model.ActivityEnded}, nil
	}
	svc := newService(mockRepo, nil, nil)
	err := svc.Archive(context.Background(), 1, model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrActivityNotSettled)
}

func TestActivityService_Archive_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return &model.Activity{ID: 1, Name: "活动", CampusID: 1, Status: model.ActivitySettled}, nil
	}
	updateCalled := false
	mockRepo.UpdateFn = func(ctx context.Context, a *model.Activity) error {
		updateCalled = true
		assert.Equal(t, model.ActivityArchived, a.Status)
		return nil
	}
	svc := newService(mockRepo, nil, nil)

	err := svc.Archive(context.Background(), 1, model.RoleHQAdmin, 0)
	require.NoError(t, err)
	assert.True(t, updateCalled)
}
