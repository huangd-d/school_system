package settlement_test

import (
	"context"
	"errors"
	"school-system/internal/model"
	"school-system/internal/module/settlement"
	"school-system/internal/testutil"
	"school-system/pkg/apperror"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newMockRepo() *MockSettlementRepo { return &MockSettlementRepo{} }

func newTestService(repo settlement.Repository, db *gorm.DB) *settlement.Service {
	return settlement.NewService(repo, db)
}

func assertAppError(t *testing.T, err error, expected *apperror.AppError) {
	t.Helper()
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr), "错误不是 *apperror.AppError: %v", err)
	assert.Equal(t, expected.Code, appErr.Code, "错误码不匹配")
}

func testActivity() *model.Activity {
	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-05")
	return &model.Activity{
		ID: 1, Name: "测试活动", CampusID: 1,
		PlannedExecutions: 10, StartDate: start, EndDate: end,
		Status: model.ActivityEnded,
	}
}

func testStock() *model.Stock {
	return &model.Stock{ID: 1, CategoryID: 1, MaterialName: "教材", TotalQuantity: 100, RemainingQty: 50, UnitPrice: 1000, Source: "purchase"}
}

// ============================================================
//  Preview 测试
// ============================================================

func TestSettlementPreview_ActivityNotFound(t *testing.T) {
	mock := newMockRepo()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newTestService(mock, nil)
	_, err := svc.Preview(context.Background(), 1)
	assertAppError(t, err, apperror.ErrSettlementActivityNotFound)
}

func TestSettlementPreview_ActivityNotEnded(t *testing.T) {
	mock := newMockRepo()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		a := testActivity()
		a.Status = model.ActivityInProgress
		return a, nil
	}
	svc := newTestService(mock, nil)
	_, err := svc.Preview(context.Background(), 1)
	assertAppError(t, err, apperror.ErrSettlementActivityNotEnded)
}

func TestSettlementPreview_ActiveSettlementExists(t *testing.T) {
	mock := newMockRepo()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return testActivity(), nil
	}
	mock.FindByActivityIDFn = func(ctx context.Context, activityID uint) ([]model.Settlement, error) {
		return []model.Settlement{{ID: 1, ActivityID: 1, Status: model.SettlementSettled}}, nil
	}
	svc := newTestService(mock, nil)
	_, err := svc.Preview(context.Background(), 1)
	assertAppError(t, err, apperror.ErrSettlementActiveExists)
}

func TestSettlementPreview_NoDistributions(t *testing.T) {
	mock := newMockRepo()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return testActivity(), nil
	}
	mock.FindByActivityIDFn = func(ctx context.Context, activityID uint) ([]model.Settlement, error) {
		return nil, nil
	}
	mock.FindDistributionsByActivityFn = func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
		return []model.Distribution{}, nil
	}
	svc := newTestService(mock, nil)
	_, err := svc.Preview(context.Background(), 1)
	assertAppError(t, err, apperror.ErrSettlementNoDistribution)
}

func TestSettlementPreview_Success(t *testing.T) {
	mock := newMockRepo()
	activity := testActivity()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return activity, nil
	}
	mock.FindByActivityIDFn = func(ctx context.Context, activityID uint) ([]model.Settlement, error) {
		return nil, nil
	}
	// 配发了 100 件教材
	mock.FindDistributionsByActivityFn = func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
		return []model.Distribution{{ID: 1, StockID: 1, ActivityID: 1, Quantity: 100}}, nil
	}
	mock.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return testStock(), nil
	}
	// 已执行 5 次（共计划 10 次）
	mock.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 5, nil
	}

	svc := newTestService(mock, nil)
	result, err := svc.Preview(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	item := result.Items[0]
	// usedQty = 100 * 5/10 = 50
	assert.Equal(t, 50, item.UsedQty)
	// recoveryQty = 100 - 50 = 50
	assert.Equal(t, 50, item.RecoveryQty)
	// costDeduction = 50 * 10 = 500
	assert.Equal(t, int64(50000), item.CostDeduction)
	assert.Equal(t, int64(50000), result.TotalReturnedAmount)
}

func TestSettlementPreview_AnomalyUsedExceedsDist(t *testing.T) {
	mock := newMockRepo()
	activity := testActivity()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return activity, nil
	}
	mock.FindByActivityIDFn = func(ctx context.Context, activityID uint) ([]model.Settlement, error) {
		return nil, nil
	}
	mock.FindDistributionsByActivityFn = func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
		return []model.Distribution{{ID: 1, StockID: 1, ActivityID: 1, Quantity: 10}}, nil
	}
	mock.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return testStock(), nil
	}
	// 已执行 20 次（超过计划 10 次）—— usedQty = 10*20/10 = 20 > 10 = 异常
	mock.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) {
		return 20, nil
	}

	svc := newTestService(mock, nil)
	_, err := svc.Preview(context.Background(), 1)
	assertAppError(t, err, apperror.ErrSettlementAnomaly)
}

// ============================================================
//  Reverse 测试（需要真实 DB 用于事务）
// ============================================================

func TestSettlementReverse_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	mock := newMockRepo()
	mock.FindSettlementByIDFn = func(ctx context.Context, id uint) (*model.Settlement, error) {
		return nil, gorm.ErrRecordNotFound
	}
	svc := newTestService(mock, db)
	err := svc.Reverse(context.Background(), 999, 1)
	assertAppError(t, err, apperror.ErrSettlementNotFound)
}

func TestSettlementReverse_AlreadyReversed(t *testing.T) {
	db := testutil.NewTestDB(t)
	// 预置一条已回撤记录：Reverse 在事务内通过真实 DB 查询，mock 的 FindSettlementByIDFn 不会生效
	require.NoError(t, db.Create(&model.Settlement{
		ID: 1, ActivityID: 1, Status: model.SettlementReversed,
	}).Error)
	mock := newMockRepo()
	svc := newTestService(mock, db)
	err := svc.Reverse(context.Background(), 1, 1)
	assertAppError(t, err, apperror.ErrSettlementAlreadyReversed)
}

// TestSettlementReverse_SoftDeleteRecoveryStock 验证回撤后回收库存软删除、结算明细可追溯
func TestSettlementReverse_SoftDeleteRecoveryStock(t *testing.T) {
	db := testutil.NewTestDB(t)

	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-01") // 单日，缩短快照重算循环
	require.NoError(t, db.Create(&model.Activity{
		ID: 1, Name: "测试活动", CampusID: 1,
		PlannedExecutions: 10, StartDate: start, EndDate: end,
		Status: model.ActivitySettled,
	}).Error)

	// 回收产生的库存（source=return）
	recoverStock := &model.Stock{
		ID: 2, PurchaseOrderID: 0, CategoryID: 1, MaterialName: "教材",
		TotalQuantity: 5, RemainingQty: 5, UnitPrice: 1000, Source: "return",
	}
	require.NoError(t, db.Create(recoverStock).Error)

	// 已结算记录 + 回收明细
	require.NoError(t, db.Create(&model.Settlement{
		ID: 1, ActivityID: 1, Status: model.SettlementSettled, OperatorID: 1, TotalReturnedAmount: 5000,
	}).Error)
	require.NoError(t, db.Create(&model.RecoveryItem{
		SettlementID: 1, StockID: 2, Quantity: 5, CostDeduction: 5000,
	}).Error)

	mock := newMockRepo()
	svc := newTestService(mock, db)
	require.NoError(t, svc.Reverse(context.Background(), 1, 1))

	// 1. 回收库存默认查询不到（软删除）
	var s model.Stock
	assert.Error(t, db.First(&s, recoverStock.ID).Error)

	// 2. Unscoped 可见（历史可追溯）
	var soft model.Stock
	require.NoError(t, db.Unscoped().First(&soft, recoverStock.ID).Error)
	assert.True(t, soft.DeletedAt.Valid, "回收库存应标记为软删除")

	// 3. recovery_items 明细保留（结算历史可追溯）
	var count int64
	require.NoError(t, db.Model(&model.RecoveryItem{}).Where("settlement_id = ?", 1).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	// 4. 结算记录状态已回撤（唯一索引释放，允许再次结算）
	var st model.Settlement
	require.NoError(t, db.First(&st, 1).Error)
	assert.Equal(t, model.SettlementReversed, st.Status)
}

// TestSettlementOverview 验证结算管理概览：一次返回 ended/settled 活动的表格数据与成本聚合
func TestSettlementOverview(t *testing.T) {
	db := testutil.NewTestDB(t)

	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-31")
	// 活动1：已结束（未结算），配发投入 100
	require.NoError(t, db.Create(&model.Activity{ID: 1, Name: "活动A", CampusID: 1, PlannedExecutions: 10, StartDate: start, EndDate: end, Status: model.ActivityEnded}).Error)
	// 活动2：已结算，配发投入 100、回收 40、执行 3
	require.NoError(t, db.Create(&model.Activity{ID: 2, Name: "活动B", CampusID: 1, PlannedExecutions: 20, StartDate: start, EndDate: end, Status: model.ActivitySettled}).Error)
	// 活动3：进行中（不应出现在概览中）
	require.NoError(t, db.Create(&model.Activity{ID: 3, Name: "活动C", CampusID: 1, PlannedExecutions: 5, StartDate: start, EndDate: end, Status: model.ActivityInProgress}).Error)

	// 库存与配发
	require.NoError(t, db.Create(&model.Stock{ID: 1, PurchaseOrderID: 1, CategoryID: 1, MaterialName: "教材", TotalQuantity: 10, RemainingQty: 0, UnitPrice: 1000, Source: "purchase"}).Error)
	require.NoError(t, db.Create(&model.Stock{ID: 2, PurchaseOrderID: 2, CategoryID: 1, MaterialName: "教具", TotalQuantity: 5, RemainingQty: 0, UnitPrice: 2000, Source: "purchase"}).Error)
	require.NoError(t, db.Create(&model.Distribution{ID: 1, StockID: 1, ActivityID: 1, Quantity: 10}).Error)
	require.NoError(t, db.Create(&model.Distribution{ID: 2, StockID: 2, ActivityID: 2, Quantity: 5}).Error)

	// 结算记录（活动2 已结算，回收 40）
	require.NoError(t, db.Create(&model.Settlement{ID: 1, ActivityID: 2, Status: model.SettlementSettled, OperatorID: 1, TotalReturnedAmount: 4000}).Error)

	// 执行记录（活动2 累计 3 次）
	require.NoError(t, db.Create(&model.ExecutionRecord{ID: 1, ActivityID: 2, Count: 3, RecordedBy: 1}).Error)

	svc := newTestService(newMockRepo(), db)
	items, err := svc.Overview(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 2, "应只返回 ended/settled 两个活动")

	byID := make(map[uint]settlement.SettlementOverviewItem)
	for _, it := range items {
		byID[it.ActivityID] = it
	}

	a := byID[1]
	assert.Equal(t, "活动A", a.ActivityName)
	assert.Equal(t, model.ActivityEnded, a.Status)
	assert.Equal(t, int64(10000), a.TotalInvestment)
	assert.Equal(t, int64(0), a.TotalReturnedAmount)
	assert.Equal(t, int64(10000), a.SettledCost, "未结算活动成本=投入")

	b := byID[2]
	assert.Equal(t, "活动B", b.ActivityName)
	assert.Equal(t, model.ActivitySettled, b.Status)
	assert.Equal(t, 20, b.PlannedExecutions)
	assert.Equal(t, 3, b.TotalExecuted)
	assert.Equal(t, int64(10000), b.TotalInvestment)
	assert.Equal(t, int64(4000), b.TotalReturnedAmount)
	assert.Equal(t, int64(6000), b.SettledCost, "结算后成本=投入−回收")
}

// TestSettlementExecute_ActiveSettlementExists_InTx 验证 Execute 在事务内复查有效结算记录
// （并发防护：即使事务外检查被绕过，事务内也会拦截重复结算）
func TestSettlementExecute_ActiveSettlementExists_InTx(t *testing.T) {
	db := testutil.NewTestDB(t)
	// 预置一条已结算记录（模拟并发下已被其他请求结算成功）
	require.NoError(t, db.Create(&model.Settlement{
		ActivityID:          1,
		Status:              model.SettlementSettled,
		TotalReturnedAmount: 10000,
	}).Error)

	mock := newMockRepo()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) {
		return testActivity(), nil
	}
	// 配发/库存/执行数据：让 computeRecoveryItems 正常走通（事务内复查应优先拦截）
	mock.FindDistributionsByActivityFn = func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
		return []model.Distribution{{ID: 1, StockID: 1, ActivityID: 1, Quantity: 10}}, nil
	}
	mock.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) {
		return testStock(), nil
	}
	mock.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) { return 5, nil }

	svc := newTestService(mock, db)
	_, err := svc.Execute(context.Background(), 1, 1)
	assertAppError(t, err, apperror.ErrSettlementActiveExists)
}

// TestSettlementExecute_AmortizationSum 验证整数分摊销尾差归末项：
// 配发 5 件 × 10000 分 = 50000 分；执行 3/10 次 → 回收 4 件（40000 分）
// 摊销基数 = 50000 − 40000 = 10000 分；Σ(每日摊销) == round(10000×3/10) = 3000 分，且每项非负
func TestSettlementExecute_AmortizationSum(t *testing.T) {
	db := testutil.NewTestDB(t)

	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-03")
	activity := &model.Activity{ID: 1, Name: "活动A", CampusID: 1, PlannedExecutions: 10, StartDate: start, EndDate: end, Status: model.ActivityEnded}
	require.NoError(t, db.Create(activity).Error)
	stock := &model.Stock{ID: 1, PurchaseOrderID: 1, CategoryID: 1, MaterialName: "教材", TotalQuantity: 5, RemainingQty: 0, UnitPrice: 10000, Source: "purchase"}
	require.NoError(t, db.Create(stock).Error)
	require.NoError(t, db.Create(&model.Distribution{ID: 1, StockID: 1, ActivityID: 1, Quantity: 5}).Error)
	// 三天各执行 1 次（CreatedAt 落在活动日期内，与摊销日期匹配）
	require.NoError(t, db.Create(&model.ExecutionRecord{ID: 1, ActivityID: 1, Count: 1, RecordedBy: 1, CreatedAt: start}).Error)
	require.NoError(t, db.Create(&model.ExecutionRecord{ID: 2, ActivityID: 1, Count: 1, RecordedBy: 1, CreatedAt: start.AddDate(0, 0, 1)}).Error)
	require.NoError(t, db.Create(&model.ExecutionRecord{ID: 3, ActivityID: 1, Count: 1, RecordedBy: 1, CreatedAt: start.AddDate(0, 0, 2)}).Error)

	mock := newMockRepo()
	mock.FindActivityByIDFn = func(ctx context.Context, id uint) (*model.Activity, error) { return activity, nil }
	mock.FindDistributionsByActivityFn = func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
		return []model.Distribution{{ID: 1, StockID: 1, ActivityID: 1, Quantity: 5}}, nil
	}
	mock.FindStockByIDFn = func(ctx context.Context, id uint) (*model.Stock, error) { return stock, nil }
	mock.SumExecutionsFn = func(ctx context.Context, activityID uint) (int, error) { return 3, nil }

	svc := newTestService(mock, db)
	_, err := svc.Execute(context.Background(), 1, 1)
	require.NoError(t, err)

	var snapshots []model.AmortizationSnapshot
	require.NoError(t, db.Where("activity_id = ?", 1).Find(&snapshots).Error)
	require.Len(t, snapshots, 3)

	var sum int64
	for _, sn := range snapshots {
		assert.GreaterOrEqual(t, sn.DailyAmount, int64(0), "每日摊销不得为负")
		sum += sn.DailyAmount
	}
	assert.Equal(t, int64(3000), sum, "Σ(每日摊销) == round(摊销基数×总执行/计划)")
}