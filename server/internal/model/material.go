package model

import (
	"time"

	"gorm.io/gorm"
)

// MaterialCategory 物资分类
type MaterialCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`  // 分类名称
	Note      string    `gorm:"size:200" json:"note"`          // 备注
	CreatedAt time.Time `json:"created_at"`
}

func (MaterialCategory) TableName() string { return "material_categories" }

// PurchaseOrder 采购单（一单一品）
type PurchaseOrder struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	MaterialName string    `gorm:"size:200;not null" json:"material_name"` // 物资名称
	CategoryID   uint      `gorm:"not null;index" json:"category_id"`      // 物资分类ID
	Quantity     int       `gorm:"not null" json:"quantity"`               // 采购数量
	TotalAmount  int64     `gorm:"not null" json:"total_amount"`       // 总金额（单位：分）
	UnitPrice    int64     `gorm:"not null" json:"unit_price"`         // 单价=总金额÷数量（单位：分）
	Notes        string    `gorm:"size:500" json:"notes"`                  // 备注
	PurchaserID  uint      `gorm:"not null" json:"purchaser_id"`           // 采购人
	CreatedAt    time.Time `json:"created_at"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

// Stock 总部库存
type Stock struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PurchaseOrderID uint      `gorm:"uniqueIndex;not null" json:"purchase_order_id"`
	CategoryID      uint      `gorm:"not null;index" json:"category_id"`
	MaterialName    string    `gorm:"size:200;not null" json:"material_name"`
	TotalQuantity   int       `gorm:"not null" json:"total_quantity"`   // 采购入库总量
	RemainingQty    int       `gorm:"not null" json:"remaining_qty"`    // 当前剩余可派发量
	UnitPrice       int64     `gorm:"not null" json:"unit_price"`
	Source          string         `gorm:"size:20;not null;default:'purchase'" json:"source"` // purchase=采购入库 return=结算回收
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"` // 软删除：保留结算历史可追溯
}

func (Stock) TableName() string { return "stocks" }
