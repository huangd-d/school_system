package model

import "time"

// AuditLog 审计日志
type AuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OperatorID    uint      `gorm:"not null;index" json:"operator_id"`
	OperationType string    `gorm:"size:30;not null;index" json:"operation_type"` // 操作类型
	EntityType    string    `gorm:"size:50;not null" json:"entity_type"`          // 操作对象类型
	EntityID      uint      `gorm:"not null" json:"entity_id"`                   // 操作对象ID
	BeforeValue   string    `gorm:"type:text" json:"before_value"`               // 变更前值
	AfterValue    string    `gorm:"type:text" json:"after_value"`                 // 变更后值
	ImpactAmount  int64     `json:"impact_amount"` // 影响金额（单位：分）
	CreatedAt     time.Time `json:"created_at"`
}

const (
	AuditOpPurchase       = "purchase"         // 采购入库
	AuditOpDistribute     = "distribute"       // 物资派发
	AuditOpAdjustDist     = "adjust_distribution" // 调整派发
	AuditOpSettle         = "settle"           // 结算回收
	AuditOpReverseSettle  = "reverse_settle"   // 回撤结算
	AuditOpModifyExec     = "modify_execution" // 修改执行次数
)

func (AuditLog) TableName() string { return "audit_logs" }
