package model

import "time"

// AmortizationSnapshot 摊销快照（按天粒度）
type AmortizationSnapshot struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ActivityID       uint      `gorm:"uniqueIndex:idx_activity_date;not null" json:"activity_id"`
	Date             time.Time `gorm:"uniqueIndex:idx_activity_date;not null" json:"date"` // 快照日期
	ExecutionCount   int       `gorm:"not null" json:"execution_count"`                   // 当日执行次数
	AmortizationBase int64     `gorm:"not null" json:"amortization_base"` // 有效摊销基数（单位：分）
	DailyAmount      int64     `gorm:"not null" json:"daily_amount"`      // 当日摊销金额（单位：分）
	CreatedAt        time.Time `json:"created_at"`
}

func (AmortizationSnapshot) TableName() string { return "amortization_snapshots" }
