package report_test

import (
	"context"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/report"
)

// MockReportRepo 实现 report.Repository 接口
type MockReportRepo struct {
	FindSnapshotsFn                 func(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error)
	FindSnapshotsByCampusFn         func(ctx context.Context, campusID uint, start, end time.Time) ([]model.AmortizationSnapshot, error)
	FindSnapshotsByDateRangeFn      func(ctx context.Context, start, end time.Time) ([]model.AmortizationSnapshot, error)
	FindActivitiesByCampusIDFn      func(ctx context.Context, campusID uint) ([]model.Activity, error)
	FindDistributionsByActivityFn   func(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindStockByIDFn                 func(ctx context.Context, id uint) (*model.Stock, error)
	FindAllActivitiesFn             func(ctx context.Context) ([]model.Activity, error)
	FindDistributionAggByCategoryFn func(ctx context.Context, start, end time.Time) ([]report.CategoryAggRow, error)
}

func (m *MockReportRepo) FindSnapshots(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	if m.FindSnapshotsFn != nil { return m.FindSnapshotsFn(ctx, activityID, start, end) }
	return nil, nil
}
func (m *MockReportRepo) FindSnapshotsByCampus(ctx context.Context, campusID uint, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	if m.FindSnapshotsByCampusFn != nil { return m.FindSnapshotsByCampusFn(ctx, campusID, start, end) }
	return nil, nil
}
func (m *MockReportRepo) FindSnapshotsByDateRange(ctx context.Context, start, end time.Time) ([]model.AmortizationSnapshot, error) {
	if m.FindSnapshotsByDateRangeFn != nil { return m.FindSnapshotsByDateRangeFn(ctx, start, end) }
	return nil, nil
}
func (m *MockReportRepo) FindActivitiesByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error) {
	if m.FindActivitiesByCampusIDFn != nil { return m.FindActivitiesByCampusIDFn(ctx, campusID) }
	return nil, nil
}
func (m *MockReportRepo) FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error) {
	if m.FindDistributionsByActivityFn != nil { return m.FindDistributionsByActivityFn(ctx, activityID) }
	return nil, nil
}
func (m *MockReportRepo) FindStockByID(ctx context.Context, id uint) (*model.Stock, error) {
	if m.FindStockByIDFn != nil { return m.FindStockByIDFn(ctx, id) }
	return nil, nil
}
func (m *MockReportRepo) FindAllActivities(ctx context.Context) ([]model.Activity, error) {
	if m.FindAllActivitiesFn != nil { return m.FindAllActivitiesFn(ctx) }
	return nil, nil
}
func (m *MockReportRepo) FindDistributionAggByCategory(ctx context.Context, start, end time.Time) ([]report.CategoryAggRow, error) {
	if m.FindDistributionAggByCategoryFn != nil { return m.FindDistributionAggByCategoryFn(ctx, start, end) }
	return nil, nil
}

var _ report.Repository = (*MockReportRepo)(nil)
