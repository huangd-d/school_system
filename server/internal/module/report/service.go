package report

import (
	"context"
	"school-system/internal/model"
	"school-system/pkg/apperror"
	"time"

	"gorm.io/gorm"
)

// Repository 报表数据访问接口
type Repository interface {
	// 已有方法
	FindSnapshots(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error)
	FindSnapshotsByCampus(ctx context.Context, campusID uint, start, end time.Time) ([]model.AmortizationSnapshot, error)

	// 新增方法
	FindSnapshotsByDateRange(ctx context.Context, start, end time.Time) ([]model.AmortizationSnapshot, error)
	FindActivitiesByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error)
	FindDistributionsByActivity(ctx context.Context, activityID uint) ([]model.Distribution, error)
	FindStockByID(ctx context.Context, id uint) (*model.Stock, error)
	FindAllActivities(ctx context.Context) ([]model.Activity, error)
	FindDistributionAggByCategory(ctx context.Context, start, end time.Time) ([]CategoryAggRow, error)
}

// CategoryAggRow 品类聚合查询结果行
type CategoryAggRow struct {
	CategoryID   uint    `gorm:"column:category_id"`
	CategoryName string  `gorm:"column:category_name"`
	TotalQuantity int    `gorm:"column:total_quantity"`
	TotalAmount  float64 `gorm:"column:total_amount"`
}

// ---- 报表返回结构体 ----

// ActivityReport 按活动报表
type ActivityReport struct {
	ActivityID       uint    `json:"activity_id"`
	ActivityName     string  `json:"activity_name"`
	CampusName       string  `json:"campus_name"`
	TotalInvestment  float64 `json:"total_investment"`
	TotalAmortization float64 `json:"total_amortization"`
	PlannedExecutions int    `json:"planned_executions"`
	TotalExecuted    int     `json:"total_executed"`
}

// DateRangeItem 按日期范围报表项
type DateRangeItem struct {
	Date           string  `json:"date"`
	ExecutionCount int     `json:"execution_count"`
	DailyAmount    float64 `json:"daily_amount"`
}

// CampusReport 按校区报表
type CampusReport struct {
	CampusID          uint    `json:"campus_id"`
	CampusName        string  `json:"campus_name"`
	ActivityCount     int     `json:"activity_count"`
	TotalInvestment   float64 `json:"total_investment"`
	TotalAmortization float64 `json:"total_amortization"`
}

// CategoryReportItem 按品类报表项
type CategoryReportItem struct {
	CategoryID    uint    `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	TotalQuantity int     `json:"total_quantity"`
	TotalAmount   float64 `json:"total_amount"`
}

// Service 报表/摊销业务逻辑
type Service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// ByActivity 按活动维度统计
func (s *Service) ByActivity(ctx context.Context, activityID uint) (*ActivityReport, error) {
	var activity model.Activity
	if err := s.db.WithContext(ctx).First(&activity, activityID).Error; err != nil {
		return nil, apperror.ErrReportActivityNotFound
	}

	// 查询校区名称
	var campus model.Campus
	if err := s.db.WithContext(ctx).First(&campus, activity.CampusID).Error; err != nil {
		campus.Name = "未知校区"
	}

	// 总投资 = Σ(配发量 × 单价)
	distributions, err := s.repo.FindDistributionsByActivity(ctx, activityID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询配发记录失败: %v", err)
	}
	var totalInvestment float64
	for _, d := range distributions {
		stock, err := s.repo.FindStockByID(ctx, d.StockID)
		if err != nil {
			continue // 静默跳过已删除的库存
		}
		totalInvestment += float64(d.Quantity) * stock.UnitPrice
	}

	// 总摊销 = Σ(快照.DailyAmount)
	var totalAmortization float64
	snapshots, err := s.repo.FindSnapshots(ctx, activityID, time.Time{}, time.Now())
	if err == nil {
		for _, sn := range snapshots {
			totalAmortization += sn.DailyAmount
		}
	}

	// 累计执行次数
	var totalExecuted int64
	s.db.WithContext(ctx).Model(&model.ExecutionRecord{}).
		Where("activity_id = ?", activityID).
		Select("COALESCE(SUM(count), 0)").
		Scan(&totalExecuted)

	return &ActivityReport{
		ActivityID:        activity.ID,
		ActivityName:      activity.Name,
		CampusName:        campus.Name,
		TotalInvestment:   totalInvestment,
		TotalAmortization: totalAmortization,
		PlannedExecutions: activity.PlannedExecutions,
		TotalExecuted:     int(totalExecuted),
	}, nil
}

// ByDateRange 按日期范围维度统计
func (s *Service) ByDateRange(ctx context.Context, start, end time.Time) ([]DateRangeItem, error) {
	if !end.After(start) {
		return nil, apperror.ErrReportDateInvalid
	}

	snapshots, err := s.repo.FindSnapshotsByDateRange(ctx, start, end)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询摊销快照失败: %v", err)
	}

	// 按日期聚合
	dateMap := make(map[string]*DateRangeItem)
	for _, sn := range snapshots {
		dateKey := sn.Date.Format("2006-01-02")
		if existing, ok := dateMap[dateKey]; ok {
			existing.ExecutionCount += sn.ExecutionCount
			existing.DailyAmount += sn.DailyAmount
		} else {
			dateMap[dateKey] = &DateRangeItem{
				Date:           dateKey,
				ExecutionCount: sn.ExecutionCount,
				DailyAmount:    sn.DailyAmount,
			}
		}
	}

	// 转为有序切片
	result := make([]DateRangeItem, 0, len(dateMap))
	currentDate := start
	for !currentDate.After(end) {
		dateKey := currentDate.Format("2006-01-02")
		if item, ok := dateMap[dateKey]; ok {
			result = append(result, *item)
		} else {
			result = append(result, DateRangeItem{Date: dateKey, ExecutionCount: 0, DailyAmount: 0})
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return result, nil
}

// ByCampus 按校区维度统计
func (s *Service) ByCampus(ctx context.Context, campusID uint, start, end time.Time) (*CampusReport, error) {
	var campus model.Campus
	if campusID != 0 {
		if err := s.db.WithContext(ctx).First(&campus, campusID).Error; err != nil {
			return nil, apperror.ErrReportCampusNotFound
		}
	} else {
		campus = model.Campus{ID: 0, Name: "全部校区"}
	}

	// 查询该校区所有活动
	activities, err := s.repo.FindActivitiesByCampusID(ctx, campusID)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询活动失败: %v", err)
	}

	var totalInvestment float64
	var totalAmortization float64
	activityIDs := make([]uint, 0, len(activities))

	for _, a := range activities {
		activityIDs = append(activityIDs, a.ID)

		// 投资
		distributions, _ := s.repo.FindDistributionsByActivity(ctx, a.ID)
		for _, d := range distributions {
			stock, _ := s.repo.FindStockByID(ctx, d.StockID)
			if stock != nil {
				totalInvestment += float64(d.Quantity) * stock.UnitPrice
			}
		}
	}

	// 摊销（利用 campus 查询方法）
	if campusID == 0 {
		// 全部校区，直接从快照表查询
		snapshots, err := s.repo.FindSnapshotsByDateRange(ctx, start, end)
		if err == nil {
			for _, sn := range snapshots {
				totalAmortization += sn.DailyAmount
			}
		}
	} else {
		snapshots, err := s.repo.FindSnapshotsByCampus(ctx, campusID, start, end)
		if err == nil {
			for _, sn := range snapshots {
				totalAmortization += sn.DailyAmount
			}
		}
	}

	return &CampusReport{
		CampusID:          campus.ID,
		CampusName:        campus.Name,
		ActivityCount:     len(activities),
		TotalInvestment:   totalInvestment,
		TotalAmortization: totalAmortization,
	}, nil
}

// ByCategory 按品类维度统计
func (s *Service) ByCategory(ctx context.Context, start, end time.Time) ([]CategoryReportItem, error) {
	rows, err := s.repo.FindDistributionAggByCategory(ctx, start, end)
	if err != nil {
		return nil, apperror.Newf(apperror.ErrInternal.Code, "查询品类统计失败: %v", err)
	}

	result := make([]CategoryReportItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, CategoryReportItem{
			CategoryID:    row.CategoryID,
			CategoryName:  row.CategoryName,
			TotalQuantity: row.TotalQuantity,
			TotalAmount:   row.TotalAmount,
		})
	}

	// 结果可以为空（无数据时返回空数组）
	if result == nil {
		result = []CategoryReportItem{}
	}

	return result, nil
}
