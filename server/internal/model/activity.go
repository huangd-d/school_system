package model

import "time"

// Activity 活动
type Activity struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"size:200;not null" json:"name"`              // 活动名称
	CampusID          uint      `gorm:"not null;index" json:"campus_id"`           // 所属校区
	PlannedExecutions int       `gorm:"not null" json:"planned_executions"`        // 计划执行次数
	StartDate         time.Time `gorm:"not null" json:"start_date"`                 // 开始日期
	EndDate           time.Time `gorm:"not null" json:"end_date"`                   // 结束日期
	Status            string    `gorm:"size:20;not null;default:'not_started'" json:"status"`
	CreatedBy         uint      `gorm:"not null" json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

const (
	ActivityNotStarted = "not_started" // 未开始
	ActivityInProgress = "in_progress" // 进行中
	ActivityEnded      = "ended"       // 已结束
	ActivitySettled    = "settled"     // 已结算
	ActivityArchived   = "archived"    // 已归档
)

func (Activity) TableName() string { return "activities" }

// ActivityContact 活动联系人关联
type ActivityContact struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ActivityID uint      `gorm:"uniqueIndex:idx_activity_user;not null" json:"activity_id"`
	UserID     uint      `gorm:"uniqueIndex:idx_activity_user;not null" json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ActivityContact) TableName() string { return "activity_contacts" }
