package user_test

import (
	"context"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/user"
	"school-system/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- FindAll ----

func TestUserRepo_FindAll_NoFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	// 准备：两个不同校区的用户
	c1 := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	c2 := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)
	testutil.CreateTestUser(t, db, "admin", model.RoleHQAdmin, c1.ID)
	testutil.CreateTestUser(t, db, "user1", model.RoleCampusOperator, c2.ID)

	// campusID=0 应返回全部
	users, err := repo.FindAll(ctx, 0)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepo_FindAll_WithCampusFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	c1 := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	c2 := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)
	testutil.CreateTestUser(t, db, "admin", model.RoleHQAdmin, c1.ID)
	testutil.CreateTestUser(t, db, "user1", model.RoleCampusOperator, c2.ID)

	// 仅查校区 c2 的用户
	users, err := repo.FindAll(ctx, c2.ID)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "user1", users[0].Username)
}

func TestUserRepo_FindAll_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	users, err := repo.FindAll(context.Background(), 0)
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserRepo_FindAll_IDOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	c := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	u1 := testutil.CreateTestUser(t, db, "b", model.RoleHQAdmin, c.ID)
	u2 := testutil.CreateTestUser(t, db, "a", model.RoleCampusOperator, c.ID)

	users, err := repo.FindAll(ctx, 0)
	require.NoError(t, err)
	require.Len(t, users, 2)
	// 应按 ID ASC 排列
	assert.Equal(t, u1.ID, users[0].ID)
	assert.Equal(t, u2.ID, users[1].ID)
}

// ---- FindByID ----

func TestUserRepo_FindByID_Exists(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	created := testutil.CreateTestUser(t, db, "admin", model.RoleHQAdmin, c.ID)

	found, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", found.Username)
	assert.Equal(t, c.ID, found.CampusID)
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	_, err := repo.FindByID(context.Background(), 99999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// ---- FindByUsername ----

func TestUserRepo_FindByUsername_Exists(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	testutil.CreateTestUser(t, db, "target", model.RoleHQAdmin, c.ID)

	found, err := repo.FindByUsername(context.Background(), "target")
	require.NoError(t, err)
	assert.Equal(t, "target", found.Username)
}

func TestUserRepo_FindByUsername_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	_, err := repo.FindByUsername(context.Background(), "no_such_user")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// ---- Create ----

func TestUserRepo_Create_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	u := &model.User{
		Username:     "newuser",
		PasswordHash: testutil.FastHash("secret"),
		Phone:        "13900139000",
		Role:         model.RoleCampusOperator,
		CampusID:     1,
		Status:       model.UserStatusActive,
	}
	err := repo.Create(context.Background(), u)
	require.NoError(t, err)
	assert.NotZero(t, u.ID)

	// 确认可再查询到
	found, err := repo.FindByID(context.Background(), u.ID)
	require.NoError(t, err)
	assert.Equal(t, "newuser", found.Username)
}

func TestUserRepo_Create_DuplicateUsername(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	testutil.CreateTestUser(t, db, "dup", model.RoleHQAdmin, c.ID)

	// 重复用户名应失败（uniqueIndex）
	u := &model.User{
		Username:     "dup",
		PasswordHash: testutil.FastHash("secret"),
		Phone:        "13900139000",
		Role:         model.RoleCampusOperator,
		CampusID:     c.ID,
		Status:       model.UserStatusActive,
	}
	err := repo.Create(context.Background(), u)
	assert.Error(t, err)
}

// ---- Update ----

func TestUserRepo_Update_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := user.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	created := testutil.CreateTestUser(t, db, "user", model.RoleCampusOperator, c.ID)

	// 修改角色
	created.Role = model.RoleActivityContact
	created.Phone = "11111111111"
	err := repo.Update(context.Background(), created)
	require.NoError(t, err)

	// 重新查询确认
	found, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, model.RoleActivityContact, found.Role)
	assert.Equal(t, "11111111111", found.Phone)
}
