package settlement

import (
	"context"
	"school-system/internal/model"
)

// Repository 结算数据访问接口
type Repository interface {
	FindByActivityID(ctx context.Context, activityID uint) ([]model.Settlement, error)
	CreateSettlement(ctx context.Context, s *model.Settlement) error
	UpdateSettlement(ctx context.Context, s *model.Settlement) error
	CreateRecoveryItems(ctx context.Context, items []model.RecoveryItem) error
}

// Service 结算业务逻辑
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Preview(ctx context.Context, activityID uint) ([]model.RecoveryItem, error) {
	return nil, nil
}

func (s *Service) Execute(ctx context.Context, activityID uint, operatorID uint) (*model.Settlement, error) {
	return nil, nil
}

func (s *Service) Reverse(ctx context.Context, settlementID uint, operatorID uint) error {
	return nil
}
