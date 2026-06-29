package report

import (
	"context"
	"school-system/internal/model"
	"time"

	"gorm.io/gorm"
)

// Repository 报表数据访问接口
type Repository interface {
	FindSnapshots(ctx context.Context, activityID uint, start, end time.Time) ([]model.AmortizationSnapshot, error)
	FindSnapshotsByCampus(ctx context.Context, campusID uint, start, end time.Time) ([]model.AmortizationSnapshot, error)
}

// Service 报表/摊销业务逻辑
type Service struct {
	repo Repository
	db   *gorm.DB // 摊销重算等复杂查询直接使用 db
}

func NewService(repo Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

func (s *Service) ByActivity(ctx context.Context, activityID uint) (interface{}, error) {
	return nil, nil
}

func (s *Service) ByDateRange(ctx context.Context, start, end time.Time) (interface{}, error) {
	return nil, nil
}

func (s *Service) ByCampus(ctx context.Context, campusID uint, start, end time.Time) (interface{}, error) {
	return nil, nil
}

func (s *Service) ByCategory(ctx context.Context, start, end time.Time) (interface{}, error) {
	return nil, nil
}
