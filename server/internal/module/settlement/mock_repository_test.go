package settlement_test

import (
	"context"
	"time"

	"school-system/internal/model"
	"school-system/internal/module/settlement"
)

// MockSettlementRepo 实现 settlement.Repository 接口
type MockSettlementRepo struct {
	FindByActivityIDFn                func(ctx context.Context, activityID uint) ([]model.Settlement, error)
	CreateSettlementFn                func(ctx context.Context, s *model.Settlement) error
	UpdateSettlementFn                func(ctx context.Context, s *model.Settlement) error
	CreateRecoveryItemsFn             func(ctx context.Context, items []model.RecoveryItem) error
	FindActivityByIDFn                func(ctx context.Context, id uint) (*model.Activity, error)
	UpdateActivityStatusFn            func(ctx context.Context, id uint, status string) error
	FindDistributionsByActivityFn     func(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindStockByIDFn                   func(ctx context.Context, id uint) (*model.Stock, error)
	SumExecutionsFn                   func(ctx context.Context, activityID uint) (int, error)
	FindExecutionsByDateRangeFn       func(ctx context.Context, activityID uint, start, end time.Time) ([]model.ExecutionRecord, error)
	CreateStockFn                     func(ctx context.Context, stock *model.Stock) error
	DeleteStockFn                     func(ctx context.Context, id uint) error
	DeleteSnapshotsByActivityFn       func(ctx context.Context, activityID uint) error
	UpsertSnapshotFn                  func(ctx context.Context, snapshot *model.AmortizationSnapshot) error
	CreateAuditLogFn                  func(ctx context.Context, log *model.AuditLog) error
	FindSettlementByIDFn              func(ctx context.Context, id uint) (*model.Settlement, error)
	FindRecoveryItemsBySettlementFn   func(ctx context.Context, settlementID uint) ([]model.RecoveryItem, error)
}

func (m *MockSettlementRepo) FindByActivityID(ctx context.Context, activityID uint) ([]model.Settlement, error) {
	if m.FindByActivityIDFn != nil { return m.FindByActivityIDFn(ctx, activityID) }
	return nil, nil
}
func (m *MockSettlementRepo) CreateSettlement(ctx context.Context, s *model.Settlement) error {
	if m.CreateSettlementFn != nil { return m.CreateSettlementFn(ctx, s) }
	return nil
}
func (m *MockSettlementRepo) UpdateSettlement(ctx context.Context, s *model.Settlement) error {
	if m.UpdateSettlementFn != nil { return m.UpdateSettlementFn(ctx, s) }
	return nil
}
func (m *MockSettlementRepo) CreateRecoveryItems(ctx context.Context, items []model.RecoveryItem) error {
	if m.CreateRecoveryItemsFn != nil { return m.CreateRecoveryItemsFn(ctx, items) }
	return nil
}
func (m *MockSettlementRepo) FindActivityByID(ctx context.Context, id uint) (*model.Activity, error) {
	if m.FindActivityByIDFn != nil { return m.FindActivityByIDFn(ctx, id) }
	return nil, nil
}
func (m *MockSettlementRepo) UpdateActivityStatus(ctx context.Context, id uint, status string) error {
	if m.UpdateActivityStatusFn != nil { return m.UpdateActivityStatusFn(ctx, id, status) }
	return nil
}
func (m *MockSettlementRepo) FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error) {
	if m.FindDistributionsByActivityFn != nil { return m.FindDistributionsByActivityFn(ctx, activityID) }
	return nil, nil
}
func (m *MockSettlementRepo) FindStockByID(ctx context.Context, id uint) (*model.Stock, error) {
	if m.FindStockByIDFn != nil { return m.FindStockByIDFn(ctx, id) }
	return nil, nil
}
func (m *MockSettlementRepo) SumExecutions(ctx context.Context, activityID uint) (int, error) {
	if m.SumExecutionsFn != nil { return m.SumExecutionsFn(ctx, activityID) }
	return 0, nil
}
func (m *MockSettlementRepo) FindExecutionsByDateRange(ctx context.Context, activityID uint, start, end time.Time) ([]model.ExecutionRecord, error) {
	if m.FindExecutionsByDateRangeFn != nil { return m.FindExecutionsByDateRangeFn(ctx, activityID, start, end) }
	return nil, nil
}
func (m *MockSettlementRepo) CreateStock(ctx context.Context, stock *model.Stock) error {
	if m.CreateStockFn != nil { return m.CreateStockFn(ctx, stock) }
	return nil
}
func (m *MockSettlementRepo) DeleteStock(ctx context.Context, id uint) error {
	if m.DeleteStockFn != nil { return m.DeleteStockFn(ctx, id) }
	return nil
}
func (m *MockSettlementRepo) DeleteSnapshotsByActivity(ctx context.Context, activityID uint) error {
	if m.DeleteSnapshotsByActivityFn != nil { return m.DeleteSnapshotsByActivityFn(ctx, activityID) }
	return nil
}
func (m *MockSettlementRepo) UpsertSnapshot(ctx context.Context, snapshot *model.AmortizationSnapshot) error {
	if m.UpsertSnapshotFn != nil { return m.UpsertSnapshotFn(ctx, snapshot) }
	return nil
}
func (m *MockSettlementRepo) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	if m.CreateAuditLogFn != nil { return m.CreateAuditLogFn(ctx, log) }
	return nil
}
func (m *MockSettlementRepo) FindSettlementByID(ctx context.Context, id uint) (*model.Settlement, error) {
	if m.FindSettlementByIDFn != nil { return m.FindSettlementByIDFn(ctx, id) }
	return nil, nil
}
func (m *MockSettlementRepo) FindRecoveryItemsBySettlement(ctx context.Context, settlementID uint) ([]model.RecoveryItem, error) {
	if m.FindRecoveryItemsBySettlementFn != nil { return m.FindRecoveryItemsBySettlementFn(ctx, settlementID) }
	return nil, nil
}

// 编译期校验接口实现
var _ settlement.Repository = (*MockSettlementRepo)(nil)
