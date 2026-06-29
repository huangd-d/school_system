package model

import "time"

// Campus 校区
type Campus struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`                    // 校区名称
	Type      string    `gorm:"size:20;not null;default:'normal'" json:"type"`   // hq=总部 normal=普通校区
	CreatedAt time.Time `json:"created_at"`
}

const (
	CampusTypeHQ     = "hq"     // 总部
	CampusTypeNormal = "normal" // 普通校区
)

func (Campus) TableName() string { return "campuses" }
