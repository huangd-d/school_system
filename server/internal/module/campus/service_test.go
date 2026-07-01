package campus_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/campus"
	"school-system/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- 辅助：快速构造 mock 依赖 ----

// newMockRepo 返回一个所有方法默认返回 nil 的 MockCampusRepo（安全空实现）。
func newMockRepo() *MockCampusRepo {
	return &MockCampusRepo{}
}

// ============================================================
//  List
// ============================================================

func TestCampusService_List_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context) ([]model.Campus, error) {
		return []model.Campus{
			{ID: 1, Name: "总部", Type: model.CampusTypeHQ},
			{ID: 2, Name: "校区A", Type: model.CampusTypeNormal},
		}, nil
	}
	svc := campus.NewService(mockRepo)

	campuses, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, campuses, 2)
	assert.Equal(t, "总部", campuses[0].Name)
	assert.Equal(t, "校区A", campuses[1].Name)
}

func TestCampusService_List_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context) ([]model.Campus, error) {
		return nil, errors.New("db error")
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.List(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查询校区列表失败")
}

// ============================================================
//  Create
// ============================================================

func TestCampusService_Create_EmptyName(t *testing.T) {
	svc := campus.NewService(newMockRepo())
	_, err := svc.Create(context.Background(), "", model.CampusTypeNormal)
	assertAppError(t, err, apperror.ErrCampusNameEmpty)
}

func TestCampusService_Create_NameTooLong(t *testing.T) {
	svc := campus.NewService(newMockRepo())
	_, err := svc.Create(context.Background(), strings.Repeat("x", 101), model.CampusTypeNormal)
	assertAppError(t, err, apperror.ErrCampusNameTooLong)
}

func TestCampusService_Create_InvalidType(t *testing.T) {
	svc := campus.NewService(newMockRepo())
	_, err := svc.Create(context.Background(), "新校区", "bad_type")
	assertAppError(t, err, apperror.ErrCampusTypeInvalid)
}

func TestCampusService_Create_HQExists(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByTypeFn = func(ctx context.Context, campusType string) ([]model.Campus, error) {
		if campusType == model.CampusTypeHQ {
			return []model.Campus{{ID: 1, Name: "总部", Type: model.CampusTypeHQ}}, nil
		}
		return nil, nil
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.Create(context.Background(), "总部2", model.CampusTypeHQ)
	assertAppError(t, err, apperror.ErrCampusHQExists)
}

func TestCampusService_Create_NameDuplicate(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByTypeFn = func(ctx context.Context, campusType string) ([]model.Campus, error) {
		return nil, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return &model.Campus{ID: 1, Name: "已存在", Type: model.CampusTypeNormal}, nil
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.Create(context.Background(), "已存在", model.CampusTypeNormal)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrCampusNameDup.Code, appErr.Code)
	assert.Contains(t, err.Error(), "已存在")
}

func TestCampusService_Create_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByTypeFn = func(ctx context.Context, campusType string) ([]model.Campus, error) {
		return nil, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.CreateFn = func(ctx context.Context, c *model.Campus) error {
		c.ID = 10
		return nil
	}
	svc := campus.NewService(mockRepo)

	created, err := svc.Create(context.Background(), "新校区", model.CampusTypeNormal)
	require.NoError(t, err)
	assert.Equal(t, uint(10), created.ID)
	assert.Equal(t, "新校区", created.Name)
	assert.Equal(t, model.CampusTypeNormal, created.Type)
}

func TestCampusService_Create_SuccessHQ(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByTypeFn = func(ctx context.Context, campusType string) ([]model.Campus, error) {
		return nil, nil // 暂无总部
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.CreateFn = func(ctx context.Context, c *model.Campus) error {
		c.ID = 1
		return nil
	}
	svc := campus.NewService(mockRepo)

	created, err := svc.Create(context.Background(), "总部", model.CampusTypeHQ)
	require.NoError(t, err)
	assert.Equal(t, uint(1), created.ID)
	assert.Equal(t, model.CampusTypeHQ, created.Type)
}

func TestCampusService_Create_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByTypeFn = func(ctx context.Context, campusType string) ([]model.Campus, error) {
		return nil, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.CreateFn = func(ctx context.Context, c *model.Campus) error {
		return errors.New("db error")
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.Create(context.Background(), "新校区", model.CampusTypeNormal)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "创建校区失败")
}

// ============================================================
//  Update
// ============================================================

func TestCampusService_Update_EmptyName(t *testing.T) {
	svc := campus.NewService(newMockRepo())
	_, err := svc.Update(context.Background(), 1, "")
	assertAppError(t, err, apperror.ErrCampusNameEmpty)
}

func TestCampusService_Update_NameTooLong(t *testing.T) {
	svc := campus.NewService(newMockRepo())
	_, err := svc.Update(context.Background(), 1, strings.Repeat("x", 101))
	assertAppError(t, err, apperror.ErrCampusNameTooLong)
}

func TestCampusService_Update_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.Update(context.Background(), 999, "新名称")
	assertAppError(t, err, apperror.ErrCampusNotFound)
}

func TestCampusService_Update_NameDuplicate_OtherCampus(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 1, Name: "旧名称", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "已存在", Type: model.CampusTypeNormal}, nil // 不同 ID
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.Update(context.Background(), 1, "已存在")
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrCampusNameDup.Code, appErr.Code)
}

func TestCampusService_Update_NameSameCampus_OK(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 1, Name: "旧名称", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		// 查到的是自身，不是重复
		return &model.Campus{ID: 1, Name: "同名称", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, c *model.Campus) error {
		return nil
	}
	svc := campus.NewService(mockRepo)

	updated, err := svc.Update(context.Background(), 1, "同名称")
	require.NoError(t, err)
	assert.Equal(t, "同名称", updated.Name)
}

func TestCampusService_Update_Success(t *testing.T) {
	existing := &model.Campus{ID: 1, Name: "旧名称", Type: model.CampusTypeNormal}
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return existing, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.UpdateFn = func(ctx context.Context, c *model.Campus) error {
		return nil
	}
	svc := campus.NewService(mockRepo)

	updated, err := svc.Update(context.Background(), 1, "新名称")
	require.NoError(t, err)
	assert.Equal(t, "新名称", updated.Name)
}

func TestCampusService_Update_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 1, Name: "旧名称", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.FindByNameFn = func(ctx context.Context, name string) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.UpdateFn = func(ctx context.Context, c *model.Campus) error {
		return errors.New("db error")
	}
	svc := campus.NewService(mockRepo)

	_, err := svc.Update(context.Background(), 1, "新名称")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "更新校区失败")
}

// ============================================================
//  Delete
// ============================================================

func TestCampusService_Delete_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 999)
	assertAppError(t, err, apperror.ErrCampusNotFound)
}

func TestCampusService_Delete_HQDelete(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 1, Name: "总部", Type: model.CampusTypeHQ}, nil
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 1)
	assertAppError(t, err, apperror.ErrCampusHQDelete)
}

func TestCampusService_Delete_HasUsers(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "校区A", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.CountUsersFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 3, nil
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 2)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrCampusHasUsers.Code, appErr.Code)
	assert.Contains(t, err.Error(), "3")
}

func TestCampusService_Delete_HasActivities(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "校区A", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.CountUsersFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.CountActivitiesFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 5, nil
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 2)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrCampusHasActivities.Code, appErr.Code)
	assert.Contains(t, err.Error(), "5")
}

func TestCampusService_Delete_CountUsersError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "校区A", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.CountUsersFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, errors.New("db error")
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查询校区关联账户失败")
}

func TestCampusService_Delete_CountActivitiesError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "校区A", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.CountUsersFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.CountActivitiesFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, errors.New("db error")
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查询校区关联活动失败")
}

func TestCampusService_Delete_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "校区A", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.CountUsersFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.CountActivitiesFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.DeleteFn = func(ctx context.Context, id uint) error {
		return nil
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 2)
	require.NoError(t, err)
}

func TestCampusService_Delete_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.Campus, error) {
		return &model.Campus{ID: 2, Name: "校区A", Type: model.CampusTypeNormal}, nil
	}
	mockRepo.CountUsersFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.CountActivitiesFn = func(ctx context.Context, campusID uint) (int64, error) {
		return 0, nil
	}
	mockRepo.DeleteFn = func(ctx context.Context, id uint) error {
		return errors.New("db error")
	}
	svc := campus.NewService(mockRepo)

	err := svc.Delete(context.Background(), 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "删除校区失败")
}

// ============================================================
//  辅助
// ============================================================

// assertAppError 断言 err 是 *AppError 且 Code 与期望匹配。
func assertAppError(t *testing.T, err error, expected *apperror.AppError) {
	t.Helper()
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr), "错误不是 *apperror.AppError: %v", err)
	assert.Equal(t, expected.Code, appErr.Code, "错误码不匹配: got %d, want %d", appErr.Code, expected.Code)
}
