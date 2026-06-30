package model

import "time"

// User 账户
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"` // 不返回给前端
	Phone        string    `gorm:"size:20;not null;default:''" json:"phone"`
	Role         string    `gorm:"size:30;not null" json:"role"` // hq_admin=总部管理员 campus_operator=校区操作员 activity_contact=活动联系人
	CampusID     uint      `gorm:"not null;index" json:"campus_id"`
	Status       string    `gorm:"size:20;not null;default:'active'" json:"status"` // active=正常 disabled=已禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

const (
	RoleHQAdmin         = "hq_admin"          // 总部管理员
	RoleCampusOperator  = "campus_operator"   // 校区操作员
	RoleActivityContact = "activity_contact"  // 活动联系人
)

const (
	UserStatusActive   = "active"   // 正常
	UserStatusDisabled = "disabled" // 已禁用
)

func (User) TableName() string { return "users" }
