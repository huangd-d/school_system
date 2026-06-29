package model

import "time"

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ActivityID uint      `gorm:"not null;index" json:"activity_id"`  // 活动ID
	Count      int       `gorm:"not null" json:"count"`              // 执行次数
	RecordedBy uint      `gorm:"not null" json:"recorded_by"`        // 录入人
	CreatedAt  time.Time `json:"created_at"`
}

func (ExecutionRecord) TableName() string { return "execution_records" }
