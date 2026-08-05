package report_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/report"
	"school-system/internal/testutil"
	"school-system/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertAppError(t *testing.T, err error, expected *apperror.AppError) {
	t.Helper()
	require.Error(t, err)
	var appErr *apperror.AppError
	require.True(t, errors.As(err, &appErr), "错误不是 *apperror.AppError: %v", err)
	assert.Equal(t, expected.Code, appErr.Code, "错误码不匹配")
}

// ============================================================
//  ByActivity 测试
// ============================================================

func TestReportByActivity_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := &MockReportRepo{}
	svc := report.NewService(repo, db)

	_, err := svc.ByActivity(context.Background(), 999)
	assertAppError(t, err, apperror.ErrReportActivityNotFound)
}

func TestReportByActivity_Success(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 准备数据
	campus := &model.Campus{Name: "测试校区"}
	db.Create(campus)
	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-05")
	activity := &model.Activity{Name: "活动A", CampusID: campus.ID, PlannedExecutions: 10, StartDate: start, EndDate: end, Status: model.ActivityEnded}
	db.Create(activity)

	cat := &model.MaterialCategory{Name: "教材"}
	db.Create(cat)
	stock := &model.Stock{CategoryID: cat.ID, MaterialName: "数学教材", TotalQuantity: 100, RemainingQty: 50, UnitPrice: 1000, Source: "purchase"}
	db.Create(stock)
	dist := &model.Distribution{StockID: stock.ID, ActivityID: activity.ID, Quantity: 50, OperatorID: 1}
	db.Create(dist)

	snap := &model.AmortizationSnapshot{ActivityID: activity.ID, Date: start, ExecutionCount: 5, AmortizationBase: 50000, DailyAmount: 25000}
	db.Create(snap)

	repo := &MockReportRepo{
		FindDistributionsByActivityFn: func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
			return []model.Distribution{*dist}, nil
		},
		FindStockByIDFn: func(ctx context.Context, id uint) (*model.Stock, error) {
			return stock, nil
		},
		FindSnapshotsFn: func(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
			return []model.AmortizationSnapshot{*snap}, nil
		},
	}

	svc := report.NewService(repo, db)
	result, err := svc.ByActivity(context.Background(), activity.ID)
	require.NoError(t, err)
	assert.Equal(t, "活动A", result.ActivityName)
	assert.Equal(t, "测试校区", result.CampusName)
	assert.Equal(t, int64(50000), result.TotalInvestment) // 50 * 1000 分
	assert.Equal(t, int64(25000), result.TotalAmortization)
}

// ============================================================
//  ByDateRange 测试
// ============================================================

func TestReportByDateRange_InvalidRange(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := &MockReportRepo{}
	svc := report.NewService(repo, db)

	end := time.Now()
	start := end.Add(24 * time.Hour) // start > end
	_, err := svc.ByDateRange(context.Background(), start, end)
	assertAppError(t, err, apperror.ErrReportDateInvalid)
}

func TestReportByDateRange_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-03")

	repo := &MockReportRepo{
		FindSnapshotsByDateRangeFn: func(ctx context.Context, s, e time.Time) ([]model.AmortizationSnapshot, error) {
			return []model.AmortizationSnapshot{
				{ActivityID: 1, Date: start, ExecutionCount: 3, DailyAmount: 15000},
				{ActivityID: 2, Date: start, ExecutionCount: 2, DailyAmount: 10000},
				{ActivityID: 1, Date: start.AddDate(0, 0, 1), ExecutionCount: 4, DailyAmount: 20000},
			}, nil
		},
	}

	svc := report.NewService(repo, db)
	result, err := svc.ByDateRange(context.Background(), start, end)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3) // 3 dates in range
	// 第一天聚合: 3+2=5 execution, 15000+10000=25000 amount
	assert.Equal(t, 5, result[0].ExecutionCount)
	assert.Equal(t, int64(25000), result[0].DailyAmount)
}

// ============================================================
//  ByCampus 测试
// ============================================================

func TestReportByCampus_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := &MockReportRepo{}
	svc := report.NewService(repo, db)

	start := time.Now().AddDate(0, 0, -30)
	end := time.Now()
	_, err := svc.ByCampus(context.Background(), 999, start, end)
	assertAppError(t, err, apperror.ErrReportCampusNotFound)
}

func TestReportByCampus_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	campus := &model.Campus{Name: "校区A"}
	db.Create(campus)

	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-05")
	activity := &model.Activity{Name: "活动A", CampusID: campus.ID, PlannedExecutions: 10, StartDate: start, EndDate: end}
	db.Create(activity)

	repo := &MockReportRepo{
		FindActivitiesByCampusIDFn: func(ctx context.Context, cid uint) ([]model.Activity, error) {
			return []model.Activity{*activity}, nil
		},
		FindDistributionsByActivityFn: func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
			return nil, nil
		},
		FindSnapshotsByCampusFn: func(ctx context.Context, campusID uint, s, e time.Time) ([]model.AmortizationSnapshot, error) {
			return nil, nil
		},
	}

	svc := report.NewService(repo, db)
	result, err := svc.ByCampus(context.Background(), campus.ID, start, end)
	require.NoError(t, err)
	assert.Equal(t, "校区A", result.CampusName)
	assert.Equal(t, 1, result.ActivityCount)
}

// ============================================================
//  ByCategory 测试
// ============================================================

func TestReportByCategory_Success(t *testing.T) {
	db := testutil.NewTestDB(t)
	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-06-01")

	repo := &MockReportRepo{
		FindDistributionAggByCategoryFn: func(ctx context.Context, s, e time.Time) ([]report.CategoryAggRow, error) {
			return []report.CategoryAggRow{
				{CategoryID: 1, CategoryName: "教材", TotalQuantity: 200, TotalAmount: 200000},
				{CategoryID: 2, CategoryName: "文具", TotalQuantity: 500, TotalAmount: 150000},
			}, nil
		},
	}

	svc := report.NewService(repo, db)
	result, err := svc.ByCategory(context.Background(), start, end)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "教材", result[0].CategoryName)
	assert.Equal(t, "文具", result[1].CategoryName)
	// 按总金额降序：200000 > 150000
	assert.Greater(t, result[0].TotalAmount, result[1].TotalAmount)
}

// TestReportByActivity_NoFloatTail 验证整数分下 100 件 × 1.1 元无浮点尾差
// （原 float64 场景 100 × 1.1 = 110.00000000000001，分单位整数为 100 × 110 = 11000 分）
func TestReportByActivity_NoFloatTail(t *testing.T) {
	db := testutil.NewTestDB(t)

	campus := &model.Campus{Name: "校区"}
	db.Create(campus)
	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-05")
	activity := &model.Activity{Name: "活动A", CampusID: campus.ID, PlannedExecutions: 10, StartDate: start, EndDate: end, Status: model.ActivityEnded}
	db.Create(activity)
	stock := &model.Stock{CategoryID: 1, MaterialName: "教材", TotalQuantity: 100, RemainingQty: 0, UnitPrice: 110, Source: "purchase"} // 1.1 元 = 110 分
	db.Create(stock)

	repo := &MockReportRepo{
		FindDistributionsByActivityFn: func(ctx context.Context, activityID uint) ([]model.Distribution, error) {
			return []model.Distribution{{StockID: stock.ID, ActivityID: activity.ID, Quantity: 100}}, nil
		},
		FindStockByIDFn: func(ctx context.Context, id uint) (*model.Stock, error) {
			return stock, nil
		},
		FindSnapshotsFn: func(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
			return nil, nil
		},
	}

	svc := report.NewService(repo, db)
	result, err := svc.ByActivity(context.Background(), activity.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(11000), result.TotalInvestment, "100 × 110 分 = 11000 分（110 元），必须无浮点尾差")
}