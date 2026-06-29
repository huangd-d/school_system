package model

import "time"

// Distribution 派发记录（物资配发到活动）
type Distribution struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	StockID    uint      `gorm:"not null;index" json:"stock_id"`       // 库存物资ID
	ActivityID uint      `gorm:"not null;index" json:"activity_id"`    // 目标活动ID
	Quantity   int       `gorm:"not null" json:"quantity"`             // 派发数量
	OperatorID uint      `gorm:"not null" json:"operator_id"`          // 操作人
	Reason     string    `gorm:"size:500" json:"reason"`               // 操作原因
	CreatedAt  time.Time `json:"created_at"`
}

func (Distribution) TableName() string { return "distributions" }
