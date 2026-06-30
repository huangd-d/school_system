package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/user"
	"school-system/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- 辅助：快速构造 mock 依赖 ----

// newMockRepo 返回一个所有方法默认返回 nil 的 MockUserRepo（安全空实现）。
func newMockRepo() *MockUserRepo {
	return &MockUserRepo{}
}

// newMockCampuses 返回一个默认 FindByID 返回总部校区的 MockCampusLookup。
func newMockCampuses() *MockCampusLookup {
	return &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return &model.Campus{ID: id, Name: "总部", Type: model.CampusTypeHQ}, nil
		},
	}
}

// ============================================================
//  List
// ============================================================

func TestUserService_List_HQAdmin(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context, campusID uint) ([]model.User, error) {
		// hq_admin 应传 campusID=0
		assert.Equal(t, uint(0), campusID)
		return []model.User{{Username: "admin"}}, nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	users, err := svc.List(context.Background(), model.RoleHQAdmin, 1)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestUserService_List_CampusOperator(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context, campusID uint) ([]model.User, error) {
		// campus_operator 应传入自己的 campusID
		assert.Equal(t, uint(5), campusID)
		return []model.User{{Username: "user1"}}, nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	users, err := svc.List(context.Background(), model.RoleCampusOperator, 5)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestUserService_List_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context, campusID uint) ([]model.User, error) {
		return nil, errors.New("db error")
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	_, err := svc.List(context.Background(), model.RoleHQAdmin, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查询账户列表失败")
}

// ============================================================
//  Create
// ============================================================

func TestUserService_Create_EmptyUsername(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), "", "pass", "123", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserUsernameEmpty)
}

func TestUserService_Create_UsernameTooLong(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), strings.Repeat("x", 51), "pass", "123", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserUsernameTooLong)
}

func TestUserService_Create_EmptyPassword(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), "user", "", "123", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserPasswordEmpty)
}

func TestUserService_Create_EmptyPhone(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), "user", "pass", "", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserPhoneEmpty)
}

func TestUserService_Create_PhoneTooLong(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), "user", "pass", strings.Repeat("1", 21), model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserPhoneTooLong)
}

func TestUserService_Create_InvalidRole(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), "user", "pass", "123", "super_admin", 1)
	assertAppError(t, err, apperror.ErrUserRoleInvalid)
}

func TestUserService_Create_CampusIDZero(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Create(context.Background(), "user", "pass", "123", model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrUserCampusRequired)
}

func TestUserService_Create_CampusNotFound(t *testing.T) {
	mockCampuses := &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := user.NewService(newMockRepo(), mockCampuses)
	_, err := svc.Create(context.Background(), "user", "pass", "123", model.RoleHQAdmin, 99)
	assertAppError(t, err, apperror.ErrUserCampusNotFound)
}

func TestUserService_Create_HQAdminOnNormalCampus(t *testing.T) {
	mockCampuses := &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return &model.Campus{ID: id, Name: "普通校区", Type: model.CampusTypeNormal}, nil
		},
	}
	svc := user.NewService(newMockRepo(), mockCampuses)
	_, err := svc.Create(context.Background(), "user", "pass", "123", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserHQAdminCampus)
}

func TestUserService_Create_NormalRoleOnHQCampus(t *testing.T) {
	mockCampuses := &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return &model.Campus{ID: id, Name: "总部", Type: model.CampusTypeHQ}, nil
		},
	}
	svc := user.NewService(newMockRepo(), mockCampuses)
	_, err := svc.Create(context.Background(), "user", "pass", "123", model.RoleCampusOperator, 1)
	assertAppError(t, err, apperror.ErrUserNormalCampus)
}

func TestUserService_Create_DuplicateUsername(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByUsernameFn = func(ctx context.Context, username string) (*model.User, error) {
		return &model.User{ID: 1, Username: "dup"}, nil // 已存在
	}
	mockRepo.CreateFn = func(ctx context.Context, u *model.User) error {
		return nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	_, err := svc.Create(context.Background(), "dup", "pass", "123", model.RoleHQAdmin, 1)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrUserUsernameDup.Code, appErr.Code)
}

func TestUserService_Create_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByUsernameFn = func(ctx context.Context, username string) (*model.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	mockRepo.CreateFn = func(ctx context.Context, u *model.User) error {
		u.ID = 10
		return nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	created, err := svc.Create(context.Background(), "newuser", "password123", "13800138000", model.RoleHQAdmin, 1)
	require.NoError(t, err)
	assert.Equal(t, uint(10), created.ID)
	assert.Equal(t, "newuser", created.Username)
	assert.Equal(t, model.UserStatusActive, created.Status)
	// 密码应为哈希值（不是明文）
	assert.NotEqual(t, "password123", created.PasswordHash)
	assert.NotEmpty(t, created.PasswordHash)
}

// ============================================================
//  Update
// ============================================================

func TestUserService_Update_InvalidRole(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Update(context.Background(), 1, "", "", "bad_role", 1)
	assertAppError(t, err, apperror.ErrUserRoleInvalid)
}

func TestUserService_Update_CampusIDZero(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	_, err := svc.Update(context.Background(), 1, "", "", model.RoleHQAdmin, 0)
	assertAppError(t, err, apperror.ErrUserCampusRequired)
}

func TestUserService_Update_UserNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	_, err := svc.Update(context.Background(), 999, "", "", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserNotFound)
}

func TestUserService_Update_CampusNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return &model.User{ID: 1, Role: model.RoleCampusOperator, CampusID: 1}, nil
	}
	mockCampuses := &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := user.NewService(mockRepo, mockCampuses)
	_, err := svc.Update(context.Background(), 1, "", "", model.RoleHQAdmin, 2)
	assertAppError(t, err, apperror.ErrUserCampusNotFound)
}

func TestUserService_Update_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return &model.User{ID: 1, Username: "user", Role: model.RoleCampusOperator, CampusID: 1, Status: model.UserStatusActive}, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, u *model.User) error {
		return nil
	}
	// 目标校区为普通校区
	mockCampuses := &MockCampusLookup{
		FindByIDFn: func(ctx context.Context, id uint) (*model.Campus, error) {
			return &model.Campus{ID: 2, Name: "校区B", Type: model.CampusTypeNormal}, nil
		},
	}
	svc := user.NewService(mockRepo, mockCampuses)

	updated, err := svc.Update(context.Background(), 1, "newuser", "13900139000", model.RoleCampusOperator, 2)
	require.NoError(t, err)
	assert.Equal(t, "newuser", updated.Username)
	assert.Equal(t, "13900139000", updated.Phone)
	assert.Equal(t, model.RoleCampusOperator, updated.Role)
	assert.Equal(t, uint(2), updated.CampusID)
}

func TestUserService_Update_UsernameTooLong(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return &model.User{ID: 1, Username: "user"}, nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	_, err := svc.Update(context.Background(), 1, strings.Repeat("x", 51), "", model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserUsernameTooLong)
}

func TestUserService_Update_DuplicateUsername(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return &model.User{ID: 1, Username: "user1"}, nil
	}
	mockRepo.FindByUsernameFn = func(ctx context.Context, username string) (*model.User, error) {
		return &model.User{ID: 2, Username: "existing"}, nil // 返回其他用户
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	_, err := svc.Update(context.Background(), 1, "existing", "", model.RoleHQAdmin, 1)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.ErrUserUsernameDup.Code, appErr.Code)
}

func TestUserService_Update_PhoneTooLong(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return &model.User{ID: 1, Username: "user"}, nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	_, err := svc.Update(context.Background(), 1, "", strings.Repeat("1", 21), model.RoleHQAdmin, 1)
	assertAppError(t, err, apperror.ErrUserPhoneTooLong)
}

// ============================================================
//  Disable
// ============================================================

func TestUserService_Disable_Self(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	err := svc.Disable(context.Background(), 1, 1) // id == operatorID
	assertAppError(t, err, apperror.ErrUserDisableSelf)
}

func TestUserService_Disable_UserNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	err := svc.Disable(context.Background(), 999, 1)
	assertAppError(t, err, apperror.ErrUserNotFound)
}

func TestUserService_Disable_Success(t *testing.T) {
	existingUser := &model.User{ID: 2, Username: "target", Status: model.UserStatusActive}
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return existingUser, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, u *model.User) error {
		return nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	err := svc.Disable(context.Background(), 2, 1) // operator(1) != target(2)
	require.NoError(t, err)
	assert.Equal(t, model.UserStatusDisabled, existingUser.Status)
}

// ============================================================
//  Enable
// ============================================================

func TestUserService_Enable_UserNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	err := svc.Enable(context.Background(), 999)
	assertAppError(t, err, apperror.ErrUserNotFound)
}

func TestUserService_Enable_Success(t *testing.T) {
	existingUser := &model.User{ID: 2, Username: "disabled_user", Status: model.UserStatusDisabled}
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return existingUser, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, u *model.User) error {
		return nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	err := svc.Enable(context.Background(), 2)
	require.NoError(t, err)
	assert.Equal(t, model.UserStatusActive, existingUser.Status)
}

// ============================================================
//  ResetPassword
// ============================================================

func TestUserService_ResetPassword_Empty(t *testing.T) {
	svc := user.NewService(newMockRepo(), newMockCampuses())
	err := svc.ResetPassword(context.Background(), 1, "")
	assertAppError(t, err, apperror.ErrUserPasswordEmpty)
}

func TestUserService_ResetPassword_UserNotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := user.NewService(mockRepo, newMockCampuses())
	err := svc.ResetPassword(context.Background(), 999, "newpass")
	assertAppError(t, err, apperror.ErrUserNotFound)
}

func TestUserService_ResetPassword_Success(t *testing.T) {
	existingUser := &model.User{ID: 1, Username: "user", PasswordHash: "oldhash"}
	hashUpdated := false
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id uint) (*model.User, error) {
		return existingUser, nil
	}
	mockRepo.UpdateFn = func(ctx context.Context, u *model.User) error {
		hashUpdated = true
		return nil
	}
	svc := user.NewService(mockRepo, newMockCampuses())

	err := svc.ResetPassword(context.Background(), 1, "new_password")
	require.NoError(t, err)
	assert.True(t, hashUpdated)
	// 密码不应是明文
	assert.NotEqual(t, "new_password", existingUser.PasswordHash)
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
