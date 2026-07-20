package settlement

import (
	"context"
	"encoding/json"
	"math"
	"school-system/internal/model"
	"school-system/pkg/apperror"
	"time"

	"gorm.io/gorm"
)

// Repository 结算数据访问接口
type Repository interface {
	// 已有方法
	FindByActivityID(ctx context.Context, activityID uint) ([]model.Settlement, error)
	CreateSettlement(ctx context.Context, s *model.Settlement) error
	UpdateSettlement(ctx context.Context, s *model.Settlement) error
	CreateRecoveryItems(ctx context.Context, items []model.RecoveryItem) error

	// 活动
	FindActivityByID(ctx context.Context, id uint) (*model.Activity, error)
	UpdateActivityStatus(ctx context.Context, id uint, status string) error

	// 配发
	FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindStockByID(ctx context.Context, id uint) (*model.Stock, error)

	// 执行
	SumExecutions(ctx context.Context, activityID uint) (int, error)
	FindExecutionsByDateRange(ctx context.Context, activityID uint, start, end time.Time) ([]model.ExecutionRecord, error)

	// 库存
	CreateStock(ctx context.Context, stock *model.Stock) error
	DeleteStock(ctx context.Context, id uint) error

	// 摊销快照
	DeleteSnapshotsByActivity(ctx context.Context, activityID uint) error
	UpsertSnapshot(ctx context.Context, snapshot *model.AmortizationSnapshot) error

	// 审计日志
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error

	// 结算查询
	FindSettlementByID(ctx context.Context, id uint) (*model.Settlement, error)
	FindRecoveryItemsBySettlement(ctx context.Context, settlementID uint) ([]model.RecoveryItem, error)
}

// PreviewItem 结算预览项
type PreviewItem struct {
	StockID       uint    `json:"stock_id"`
	MaterialName  string  `json:"material_name"`
	DistributedQty int    `json:"distributed_qty"`
	UsedQty       int     `json:"used_qty"`
	RecoveryQty   int     `json:"recovery_qty"`
	UnitPrice     float64 `json:"unit_price"`
	CostDeduction float64 `json:"cost_deduction"`
}

// PreviewResult 结算预览结果
type PreviewResult struct {
	Items              []PreviewItem `json:"items"`
	TotalReturnedAmount float64      `json:"total_returned_amount"`
	ActivityName       string        `json:"activity_name"`
	TotalExecuted      int           `json:"total_executed"`
	PlannedExecutions  int           `json:"planned_executions"`
}

// ListByActivity 查询活动的所有结算记录
func (s *Service) ListByActivity(ctx context.Context, activityID uint) ([]model.Settlement, error) {
	return s.repo.FindByActivityID(ctx, activityID)
}

// Service 结算业务逻辑
type Service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// Preview 结算预览 — 计算回收物资和成本扣减
func (s *Service) Preview(ctx context.Context, activityID uint) (*PreviewResult, error) {
	activity, err := s.repo.FindActivityByID(ctx, activityID)
	if err != nil {
		return nil, apperror.ErrSettlementActivityNotFound
	}

	if activity.Status != model.ActivityEnded {
		return nil, apperror.ErrSettlementActivityNotEnded
	}

	// 检查是否已有有效结算记录
	existing, err := s.repo.FindByActivityID(ctx, activityID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询结算记录失败: %v", err)
	}
	for _, st := range existing {
		if st.Status == model.SettlementSettled {
			return nil, apperror.ErrSettlementActiveExists
		}
	}

	items, totalReturned, err := s.computeRecoveryItems(ctx, activity)
	if err != nil {
		return nil, err
	}

	totalExec, err := s.repo.SumExecutions(ctx, activityID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询执行记录失败: %v", err)
	}

	return &PreviewResult{
		Items:              items,
		TotalReturnedAmount: totalReturned,
		ActivityName:       activity.Name,
		TotalExecuted:      totalExec,
		PlannedExecutions:  activity.PlannedExecutions,
	}, nil
}

// Execute 执行结算 — 单事务完成回收入库、结算记录、摊销重算、审计日志
func (s *Service) Execute(ctx context.Context, activityID uint, operatorID uint) (*model.Settlement, error) {
	activity, err := s.repo.FindActivityByID(ctx, activityID)
	if err != nil {
		return nil, apperror.ErrSettlementActivityNotFound
	}

	if activity.Status != model.ActivityEnded {
		return nil, apperror.ErrSettlementActivityNotEnded
	}

	// 检查是否已有有效结算记录
	existing, err := s.repo.FindByActivityID(ctx, activityID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询结算记录失败: %v", err)
	}
	for _, st := range existing {
		if st.Status == model.SettlementSettled {
			return nil, apperror.ErrSettlementActiveExists
		}
	}

	items, totalReturned, err := s.computeRecoveryItems(ctx, activity)
	if err != nil {
		return nil, err
	}

	var settlement *model.Settlement

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 用事务 db 创建临时 repo
		txRepo := NewRepository(tx)

		// 1. 创建回收 Stock 记录
		for _, item := range items {
			if item.RecoveryQty > 0 {
				originalStock, stockErr := txRepo.FindStockByID(ctx, item.StockID)
				if stockErr != nil {
					return apperror.Newf(apperror.ErrInternal.Code, "查询原库存失败: %v", stockErr)
				}
				newStock := &model.Stock{
					PurchaseOrderID: 0, // 回收无采购单
					CategoryID:      originalStock.CategoryID,
					MaterialName:    originalStock.MaterialName,
					TotalQuantity:   item.RecoveryQty,
					RemainingQty:    item.RecoveryQty,
					UnitPrice:       originalStock.UnitPrice,
					Source:          "return",
				}
				if createErr := txRepo.CreateStock(ctx, newStock); createErr != nil {
					return apperror.Newf(apperror.ErrInternal.Code, "创建回收库存失败: %v", createErr)
				}
				// 更新 item 的 StockID 为新建的回收库存 ID
				item.StockID = newStock.ID
			}
		}

		// 2. 创建结算记录
		settlement = &model.Settlement{
			ActivityID:          activityID,
			Status:              model.SettlementSettled,
			OperatorID:          operatorID,
			TotalReturnedAmount: totalReturned,
		}
		if createErr := txRepo.CreateSettlement(ctx, settlement); createErr != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "创建结算记录失败: %v", createErr)
		}

		// 3. 创建回收明细
		recoveryItems := make([]model.RecoveryItem, 0, len(items))
		for _, item := range items {
			if item.RecoveryQty > 0 {
				recoveryItems = append(recoveryItems, model.RecoveryItem{
					SettlementID:  settlement.ID,
					StockID:       item.StockID,
					Quantity:      item.RecoveryQty,
					CostDeduction: item.CostDeduction,
				})
			}
		}
		if len(recoveryItems) > 0 {
			if createErr := txRepo.CreateRecoveryItems(ctx, recoveryItems); createErr != nil {
				return apperror.Newf(apperror.ErrInternal.Code, "创建回收明细失败: %v", createErr)
			}
		}

		// 4. 更新活动状态
		if updateErr := txRepo.UpdateActivityStatus(ctx, activityID, model.ActivitySettled); updateErr != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "更新活动状态失败: %v", updateErr)
		}

		// 5. 重算摊销快照
		if recalcErr := s.recalculateSnapshotsTx(ctx, txRepo, activity); recalcErr != nil {
			return recalcErr
		}

		// 6. 写审计日志
		beforeVal, _ := json.Marshal(map[string]interface{}{
			"status":     model.ActivityEnded,
			"activity_id": activityID,
		})
		afterVal, _ := json.Marshal(map[string]interface{}{
			"status":              model.ActivitySettled,
			"total_returned":      totalReturned,
			"recovery_item_count": len(recoveryItems),
		})
		auditLog := &model.AuditLog{
			OperatorID:    operatorID,
			OperationType: model.AuditOpSettle,
			EntityType:    "settlement",
			EntityID:      settlement.ID,
			BeforeValue:   string(beforeVal),
			AfterValue:    string(afterVal),
			ImpactAmount:  totalReturned,
		}
		return txRepo.CreateAuditLog(ctx, auditLog)
	})

	if err != nil {
		return nil, err
	}

	return settlement, nil
}

// Reverse 回撤结算 — 单事务恢复库存、活动状态、摊销快照
func (s *Service) Reverse(ctx context.Context, settlementID uint, operatorID uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewRepository(tx)

		settlement, err := txRepo.FindSettlementByID(ctx, settlementID)
		if err != nil {
			return apperror.ErrSettlementNotFound
		}

		if settlement.Status == model.SettlementReversed {
			return apperror.ErrSettlementAlreadyReversed
		}

		// 1. 删除回收产生的库存记录
		recoveryItems, err := txRepo.FindRecoveryItemsBySettlement(ctx, settlementID)
		if err != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "查询回收明细失败: %v", err)
		}
		for _, item := range recoveryItems {
			if delErr := txRepo.DeleteStock(ctx, item.StockID); delErr != nil {
				return apperror.Newf(apperror.ErrInternal.Code, "删除回收库存失败: %v", delErr)
			}
		}

		// 2. 更新结算状态
		settlement.Status = model.SettlementReversed
		if updateErr := txRepo.UpdateSettlement(ctx, settlement); updateErr != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "更新结算状态失败: %v", updateErr)
		}

		// 3. 恢复活动状态
		if updateErr := txRepo.UpdateActivityStatus(ctx, settlement.ActivityID, model.ActivityEnded); updateErr != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "恢复活动状态失败: %v", updateErr)
		}

		// 4. 重算摊销快照
		activity, err := txRepo.FindActivityByID(ctx, settlement.ActivityID)
		if err != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "查询活动失败: %v", err)
		}
		if recalcErr := s.recalculateSnapshotsTx(ctx, txRepo, activity); recalcErr != nil {
			return recalcErr
		}

		// 5. 写审计日志
		beforeVal, _ := json.Marshal(map[string]interface{}{
			"status":              model.SettlementSettled,
			"total_returned":      settlement.TotalReturnedAmount,
		})
		afterVal, _ := json.Marshal(map[string]interface{}{
			"status": model.SettlementReversed,
		})
		auditLog := &model.AuditLog{
			OperatorID:    operatorID,
			OperationType: model.AuditOpReverseSettle,
			EntityType:    "settlement",
			EntityID:      settlementID,
			BeforeValue:   string(beforeVal),
			AfterValue:    string(afterVal),
			ImpactAmount:  -settlement.TotalReturnedAmount, // 回撤，金额取反
		}
		return txRepo.CreateAuditLog(ctx, auditLog)
	})
}

// ---- 内部辅助方法 ----

// computeRecoveryItems 计算回收物资明细
func (s *Service) computeRecoveryItems(ctx context.Context, activity *model.Activity) ([]PreviewItem, float64, error) {
	distributions, err := s.repo.FindDistributionsByActivity(ctx, activity.ID)
	if err != nil {
		return nil, 0, apperror.Newf(apperror.ErrInternal.Code, "查询配发记录失败: %v", err)
	}

	if len(distributions) == 0 {
		return nil, 0, apperror.ErrSettlementNoDistribution
	}

	totalExecuted, err := s.repo.SumExecutions(ctx, activity.ID)
	if err != nil {
		return nil, 0, apperror.Newf(apperror.ErrInternal.Code, "查询执行记录失败: %v", err)
	}

	planned := activity.PlannedExecutions
	if planned == 0 {
		return nil, 0, apperror.Newf(apperror.ErrInternal.Code, "活动计划执行次数为0")
	}

	items := make([]PreviewItem, 0, len(distributions))
	var totalReturned float64

	for _, d := range distributions {
		stock, err := s.repo.FindStockByID(ctx, d.StockID)
		if err != nil {
			return nil, 0, apperror.Newf(apperror.ErrInternal.Code, "查询库存记录失败: %v", err)
		}

		// 按执行进度比例计算已使用量（向下取整）
		usedQty := int(math.Floor(float64(d.Quantity) * float64(totalExecuted) / float64(planned)))

		// 数据异常校验
		if usedQty > d.Quantity {
			return nil, 0, apperror.ErrSettlementAnomaly
		}

		recoveryQty := d.Quantity - usedQty
		costDeduction := float64(recoveryQty) * stock.UnitPrice

		items = append(items, PreviewItem{
			StockID:        d.StockID,
			MaterialName:   stock.MaterialName,
			DistributedQty: d.Quantity,
			UsedQty:        usedQty,
			RecoveryQty:    recoveryQty,
			UnitPrice:      stock.UnitPrice,
			CostDeduction:  costDeduction,
		})

		totalReturned += costDeduction
	}

	// 四舍五入保留两位小数
	totalReturned = math.Round(totalReturned*100) / 100

	return items, totalReturned, nil
}

// recalculateSnapshotsTx 在事务中重算摊销快照
func (s *Service) recalculateSnapshotsTx(ctx context.Context, txRepo Repository, activity *model.Activity) error {
	// 1. 删除该活动全部现有快照
	if err := txRepo.DeleteSnapshotsByActivity(ctx, activity.ID); err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "删除旧摊销快照失败: %v", err)
	}

	// 2. 计算配发总金额
	distributions, err := txRepo.FindDistributionsByActivity(ctx, activity.ID)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "查询配发记录失败: %v", err)
	}

	var totalDistValue float64
	for _, d := range distributions {
		stock, err := txRepo.FindStockByID(ctx, d.StockID)
		if err != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "查询库存失败: %v", err)
		}
		totalDistValue += float64(d.Quantity) * stock.UnitPrice
	}

	// 3. 计算已结算回收总额（status="settled" 的结算记录）
	settlements, err := txRepo.FindByActivityID(ctx, activity.ID)
	if err != nil {
		return apperror.Newf(apperror.ErrInternal.Code, "查询结算记录失败: %v", err)
	}

	var totalRecovered float64
	for _, st := range settlements {
		if st.Status == model.SettlementSettled {
			recoveryItems, err := txRepo.FindRecoveryItemsBySettlement(ctx, st.ID)
			if err != nil {
				return apperror.Newf(apperror.ErrInternal.Code, "查询回收明细失败: %v", err)
			}
			for _, ri := range recoveryItems {
				totalRecovered += ri.CostDeduction
			}
		}
	}

	// 4. 摊销基数 = 配发总额 - 已回收总额
	amortizationBase := totalDistValue - totalRecovered

	// 5. 按日统计执行次数
	dailyCounts, err := s.aggregateDailyExecutions(ctx, txRepo, activity)
	if err != nil {
		return err
	}

	// 6. 逐日生成摊销快照
	planned := activity.PlannedExecutions
	currentDate := activity.StartDate
	endDate := activity.EndDate

	for !currentDate.After(endDate) {
		dateKey := currentDate.Format("2006-01-02")
		dayCount := dailyCounts[dateKey]

		var dailyAmount float64
		if planned > 0 {
			dailyAmount = amortizationBase * float64(dayCount) / float64(planned)
		}
		dailyAmount = math.Round(dailyAmount*100) / 100

		snapshot := &model.AmortizationSnapshot{
			ActivityID:       activity.ID,
			Date:             currentDate,
			ExecutionCount:   dayCount,
			AmortizationBase: math.Round(amortizationBase*100) / 100,
			DailyAmount:      dailyAmount,
		}

		if err := txRepo.UpsertSnapshot(ctx, snapshot); err != nil {
			return apperror.Newf(apperror.ErrInternal.Code, "写入摊销快照失败: %v", err)
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return nil
}

// aggregateDailyExecutions 按日聚合执行次数
func (s *Service) aggregateDailyExecutions(ctx context.Context, txRepo Repository, activity *model.Activity) (map[string]int, error) {
	executions, err := txRepo.FindExecutionsByDateRange(ctx, activity.ID, activity.StartDate, activity.EndDate)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询执行记录失败: %v", err)
	}

	dailyCounts := make(map[string]int)
	for _, exec := range executions {
		dateKey := exec.CreatedAt.Format("2006-01-02")
		dailyCounts[dateKey] += exec.Count
	}

	return dailyCounts, nil
}
