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
	return &model.Stock{ID: 1, CategoryID: 1, MaterialName: "教材", TotalQuantity: 100, RemainingQty: 50, UnitPrice: 10.0, Source: "purchase"}
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
	assert.Equal(t, 500.0, item.CostDeduction)
	assert.Equal(t, 500.0, result.TotalReturnedAmount)
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
	mock := newMockRepo()
	mock.FindSettlementByIDFn = func(ctx context.Context, id uint) (*model.Settlement, error) {
		return &model.Settlement{ID: 1, Status: model.SettlementReversed}, nil
	}
	svc := newTestService(mock, db)
	err := svc.Reverse(context.Background(), 1, 1)
	assertAppError(t, err, apperror.ErrSettlementAlreadyReversed)
}
