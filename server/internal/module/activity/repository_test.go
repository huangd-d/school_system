package activity_test

import (
	"context"
	"testing"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/activity"
	"school-system/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- 辅助：快速创建测试活动 ----

func createTestActivity(t *testing.T, db *gorm.DB, name string, campusID uint, plannedExec int, status string) *model.Activity {
	t.Helper()
	a := &model.Activity{
		Name:              name,
		CampusID:          campusID,
		PlannedExecutions: plannedExec,
		StartDate:         time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:           time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:            status,
		CreatedBy:         1,
	}
	require.NoError(t, db.Create(a).Error, "创建测试活动失败")
	return a
}

func createTestContact(t *testing.T, db *gorm.DB, activityID, userID uint) {
	t.Helper()
	require.NoError(t, db.Create(&model.ActivityContact{ActivityID: activityID, UserID: userID}).Error, "创建测试联系人失败")
}

// ============================================================
//  FindAll
// ============================================================

func TestActivityRepo_FindAll(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)
	ctx := context.Background()

	c1 := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	c2 := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)
	createTestActivity(t, db, "活动1", c1.ID, 10, model.ActivityNotStarted)
	createTestActivity(t, db, "活动2", c2.ID, 5, model.ActivityInProgress)

	activities, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, activities, 2)
}

func TestActivityRepo_FindAll_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	activities, err := repo.FindAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, activities)
}

// ============================================================
//  FindByCampusID
// ============================================================

func TestActivityRepo_FindByCampusID(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)
	ctx := context.Background()

	c1 := testutil.CreateTestCampus(t, db, "总部", model.CampusTypeHQ)
	c2 := testutil.CreateTestCampus(t, db, "校区A", model.CampusTypeNormal)
	createTestActivity(t, db, "活动1", c1.ID, 10, model.ActivityNotStarted)
	createTestActivity(t, db, "活动2", c2.ID, 5, model.ActivityInProgress)

	activities, err := repo.FindByCampusID(ctx, c1.ID)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, "活动1", activities[0].Name)
}

func TestActivityRepo_FindByCampusID_NoMatch(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	activities, err := repo.FindByCampusID(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, activities)
}

// ============================================================
//  FindByContactUserID
// ============================================================

func TestActivityRepo_FindByContactUserID(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)
	ctx := context.Background()

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	a1 := createTestActivity(t, db, "活动A", c.ID, 10, model.ActivityNotStarted)
	a2 := createTestActivity(t, db, "活动B", c.ID, 5, model.ActivityInProgress)
	// 用户 2 是两个活动的联系人，用户 3 只是活动B的联系人
	createTestContact(t, db, a1.ID, 2)
	createTestContact(t, db, a2.ID, 2)
	createTestContact(t, db, a2.ID, 3)

	// 用户 2 关联 2 个活动
	activities, err := repo.FindByContactUserID(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, activities, 2)

	// 用户 3 关联 1 个活动
	activities, err = repo.FindByContactUserID(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, activities, 1)
	assert.Equal(t, "活动B", activities[0].Name)
}

func TestActivityRepo_FindByContactUserID_NoMatch(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	activities, err := repo.FindByContactUserID(context.Background(), 999)
	require.NoError(t, err)
	assert.Empty(t, activities)
}

// ============================================================
//  FindByID
// ============================================================

func TestActivityRepo_FindByID_Exists(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	created := createTestActivity(t, db, "目标活动", c.ID, 10, model.ActivityNotStarted)

	found, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "目标活动", found.Name)
	assert.Equal(t, c.ID, found.CampusID)
	assert.Equal(t, 10, found.PlannedExecutions)
}

func TestActivityRepo_FindByID_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	_, err := repo.FindByID(context.Background(), 99999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// ============================================================
//  Create
// ============================================================

func TestActivityRepo_Create_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	a := &model.Activity{
		Name:              "新建活动",
		CampusID:          1,
		PlannedExecutions: 20,
		StartDate:         time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:           time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		Status:            model.ActivityNotStarted,
		CreatedBy:         1,
	}
	err := repo.Create(context.Background(), a)
	require.NoError(t, err)
	assert.NotZero(t, a.ID)

	// 确认可再查询到
	found, err := repo.FindByID(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Equal(t, "新建活动", found.Name)
}

// ============================================================
//  Update
// ============================================================

func TestActivityRepo_Update_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	created := createTestActivity(t, db, "旧名称", c.ID, 10, model.ActivityNotStarted)

	created.Name = "已更新"
	created.PlannedExecutions = 15
	created.Status = model.ActivityInProgress
	err := repo.Update(context.Background(), created)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "已更新", found.Name)
	assert.Equal(t, 15, found.PlannedExecutions)
	assert.Equal(t, model.ActivityInProgress, found.Status)
}

// ============================================================
//  Contacts (SetContacts / FindContactIDs)
// ============================================================

func TestActivityRepo_SetContacts_Add(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	a := createTestActivity(t, db, "活动", c.ID, 10, model.ActivityNotStarted)

	// 添加联系人
	err := repo.SetContacts(context.Background(), a.ID, []uint{2, 3})
	require.NoError(t, err)

	ids, err := repo.FindContactIDs(context.Background(), a.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []uint{2, 3}, ids)
}

func TestActivityRepo_SetContacts_Replace(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	a := createTestActivity(t, db, "活动", c.ID, 10, model.ActivityNotStarted)

	// 先添加联系人 2,3
	require.NoError(t, repo.SetContacts(context.Background(), a.ID, []uint{2, 3}))
	// 再替换为 4
	require.NoError(t, repo.SetContacts(context.Background(), a.ID, []uint{4}))

	ids, err := repo.FindContactIDs(context.Background(), a.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []uint{4}, ids)
}

func TestActivityRepo_SetContacts_Clear(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	a := createTestActivity(t, db, "活动", c.ID, 10, model.ActivityNotStarted)

	// 先添加联系人
	require.NoError(t, repo.SetContacts(context.Background(), a.ID, []uint{2, 3}))
	// 清空
	require.NoError(t, repo.SetContacts(context.Background(), a.ID, nil))

	ids, err := repo.FindContactIDs(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestActivityRepo_FindContactIDs_NoContacts(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	c := testutil.CreateTestCampus(t, db, "校区", model.CampusTypeNormal)
	a := createTestActivity(t, db, "活动", c.ID, 10, model.ActivityNotStarted)

	ids, err := repo.FindContactIDs(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

// ============================================================
//  Executions (CreateExecution / FindExecutions / SumExecutions)
// ============================================================

func TestActivityRepo_CreateExecution_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	e := &model.ExecutionRecord{
		ActivityID: 1,
		Count:      5,
		RecordedBy: 2,
	}
	err := repo.CreateExecution(context.Background(), e)
	require.NoError(t, err)
	assert.NotZero(t, e.ID)
}

func TestActivityRepo_FindExecutions(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	// 录制两次执行记录
	require.NoError(t, repo.CreateExecution(context.Background(), &model.ExecutionRecord{ActivityID: 1, Count: 3, RecordedBy: 1}))
	require.NoError(t, repo.CreateExecution(context.Background(), &model.ExecutionRecord{ActivityID: 1, Count: 2, RecordedBy: 2}))
	// 另一个活动的执行记录
	require.NoError(t, repo.CreateExecution(context.Background(), &model.ExecutionRecord{ActivityID: 2, Count: 5, RecordedBy: 1}))

	records, err := repo.FindExecutions(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, records, 2)
}

func TestActivityRepo_FindExecutions_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	records, err := repo.FindExecutions(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, records)
}

func TestActivityRepo_SumExecutions(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	require.NoError(t, repo.CreateExecution(context.Background(), &model.ExecutionRecord{ActivityID: 1, Count: 3, RecordedBy: 1}))
	require.NoError(t, repo.CreateExecution(context.Background(), &model.ExecutionRecord{ActivityID: 1, Count: 7, RecordedBy: 2}))

	total, err := repo.SumExecutions(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 10, total)
}

func TestActivityRepo_SumExecutions_Zero(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := activity.NewRepository(db)

	total, err := repo.SumExecutions(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
}
