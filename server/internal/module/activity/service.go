package activity

import (
	"context"
	"school-system/internal/model"
)

// Repository 活动数据访问接口
type Repository interface {
	FindAll(ctx context.Context) ([]model.Activity, error)
	FindByCampusID(ctx context.Context, campusID uint) ([]model.Activity, error)
	FindByID(ctx context.Context, id uint) (*model.Activity, error)
	Create(ctx context.Context, activity *model.Activity) error
	Update(ctx context.Context, activity *model.Activity) error
	// 联系人
	SetContacts(ctx context.Context, activityID uint, userIDs []uint) error
	FindContactIDs(ctx context.Context, activityID uint) ([]uint, error)
	// 执行记录
	CreateExecution(ctx context.Context, exec *model.ExecutionRecord) error
	FindExecutions(ctx context.Context, activityID uint) ([]model.ExecutionRecord, error)
	SumExecutions(ctx context.Context, activityID uint) (int, error)
}

// Service 活动业务逻辑
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, campusID uint, role string) ([]model.Activity, error) {
	return nil, nil
}

func (s *Service) Create(ctx context.Context, name string, campusID uint, contactIDs []uint, plannedExec int, startDate, endDate string, createdBy uint) (*model.Activity, error) {
	return nil, nil
}

func (s *Service) Update(ctx context.Context, id uint, name string, contactIDs []uint, plannedExec int) (*model.Activity, error) {
	return nil, nil
}

func (s *Service) Detail(ctx context.Context, id uint) (*model.Activity, error) {
	return nil, nil
}

func (s *Service) AddExecution(ctx context.Context, activityID uint, count int, recordedBy uint) error {
	return nil
}

func (s *Service) Archive(ctx context.Context, id uint) error { return nil }
