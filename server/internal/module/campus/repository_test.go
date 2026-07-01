package campus_test

import (
	"context"
	"testing"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/campus"
	"school-system/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- FindAll ----

func TestCampusRepo_FindAll_ReturnsAll(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)

	campuses, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, campuses, 2)
}

func TestCampusRepo_FindAll_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)

	campuses, err := repo.FindAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, campuses)
}

func TestCampusRepo_FindAll_IDOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	c1 := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	c2 := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)

	campuses, err := repo.FindAll(ctx)
	require.NoError(t, err)
	require.Len(t, campuses, 2)
	assert.Equal(t, c1.ID, campuses[0].ID)
	assert.Equal(t, c2.ID, campuses[1].ID)
}

// ---- FindByID ----

func TestCampusRepo_FindByID_Exists(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)

	created := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)

	found, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "总部", found.Name)
	assert.Equal(t, model.CampusTypeHQ, found.Type)
}

func TestCampusRepo_FindByID_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)

	_, err := repo.FindByID(context.Background(), 99999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// ---- FindByName ----

func TestCampusRepo_FindByName_Exists(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)

	testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)

	found, err := repo.FindByName(context.Background(), "总部")
	require.NoError(t, err)
	assert.Equal(t, "总部", found.Name)
}

func TestCampusRepo_FindByName_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)

	_, err := repo.FindByName(context.Background(), "不存在的校区")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// ---- FindByType ----

func TestCampusRepo_FindByType_HQ(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)

	result, err := repo.FindByType(ctx, model.CampusTypeHQ)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, model.CampusTypeHQ, result[0].Type)
}

func TestCampusRepo_FindByType_Normal(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)

	result, err := repo.FindByType(ctx, model.CampusTypeNormal)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, model.CampusTypeNormal, result[0].Type)
}

func TestCampusRepo_FindByType_NoMatch(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)

	result, err := repo.FindByType(ctx, model.CampusTypeNormal)
	require.NoError(t, err)
	assert.Empty(t, result)
}

// ---- Create ----

func TestCampusRepo_Create_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	c := &model.Campus{Name: "新校区", Type: model.CampusTypeNormal}
	err := repo.Create(ctx, c)
	require.NoError(t, err)
	assert.NotZero(t, c.ID)

	// 确认可再查询到
	found, err := repo.FindByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, "新校区", found.Name)
	assert.Equal(t, model.CampusTypeNormal, found.Type)
}

// ---- Update ----

func TestCampusRepo_Update_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	created := testutil.CreateTestCampus(t, db, "旧名称", model.CampusTypeNormal)

	created.Name = "新名称"
	err := repo.Update(ctx, created)
	require.NoError(t, err)

	// 重新查询确认
	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "新名称", found.Name)
}

// ---- Delete ----

func TestCampusRepo_Delete_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	created := testutil.CreateTestCampus(t, db, "待删除", model.CampusTypeNormal)

	err := repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	// 删除后应查询不到
	_, err = repo.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCampusRepo_Delete_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)

	// GORM Delete 在无匹配行时不报错
	err := repo.Delete(context.Background(), 99999)
	assert.NoError(t, err)
}

// ---- CountUsers ----

func TestCampusRepo_CountUsers_HasUsers(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	c := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)
	testutil.CreateTestUser(t, db, "user1", model.RoleCampusOperator, c.ID)
	testutil.CreateTestUser(t, db, "user2", model.RoleActivityContact, c.ID)

	count, err := repo.CountUsers(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCampusRepo_CountUsers_NoUsers(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	c := testutil.CreateTestCampus(t, db, "空校区", model.CampusTypeNormal)

	count, err := repo.CountUsers(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ---- CountActivities ----

func TestCampusRepo_CountActivities_HasActivities(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	c := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)

	now := time.Now()
	a1 := &model.Activity{
		Name:              "活动1",
		CampusID:          c.ID,
		PlannedExecutions: 5,
		StartDate:         now,
		EndDate:           now.AddDate(0, 1, 0),
		Status:            model.ActivityNotStarted,
		CreatedBy:         1,
	}
	a2 := &model.Activity{
		Name:              "活动2",
		CampusID:          c.ID,
		PlannedExecutions: 3,
		StartDate:         now,
		EndDate:           now.AddDate(0, 2, 0),
		Status:            model.ActivityNotStarted,
		CreatedBy:         1,
	}
	require.NoError(t, db.Create(a1).Error)
	require.NoError(t, db.Create(a2).Error)

	count, err := repo.CountActivities(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCampusRepo_CountActivities_NoActivities(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := campus.NewRepository(db)
	ctx := context.Background()

	c := testutil.CreateTestCampus(t, db, "空校区", model.CampusTypeNormal)

	count, err := repo.CountActivities(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
