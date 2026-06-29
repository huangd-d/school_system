package model

import "time"

// Settlement 结算记录
type Settlement struct {
	ID                  uint      `gorm:"primaryKey" json:"id"`
	ActivityID          uint      `gorm:"not null;index" json:"activity_id"`
	Status              string    `gorm:"size:20;not null;default:'settled'" json:"status"` // settled=已结算 reversed=已回撤
	OperatorID          uint      `gorm:"not null" json:"operator_id"`
	TotalReturnedAmount float64   `gorm:"not null" json:"total_returned_amount"` // 本次回收总金额
	CreatedAt           time.Time `json:"created_at"`
}

const (
	SettlementSettled  = "settled"  // 已结算
	SettlementReversed = "reversed" // 已回撤
)

func (Settlement) TableName() string { return "settlements" }

// RecoveryItem 回收明细
type RecoveryItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	SettlementID  uint      `gorm:"not null;index" json:"settlement_id"`
	StockID       uint      `gorm:"not null" json:"stock_id"`        // 回收物资ID
	Quantity      int       `gorm:"not null" json:"quantity"`        // 回收数量
	CostDeduction float64   `gorm:"not null" json:"cost_deduction"`  // 成本扣减金额
	CreatedAt     time.Time `json:"created_at"`
}

func (RecoveryItem) TableName() string { return "recovery_items" }
