package testutil

import (
	"testing"

	"school-system/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB 创建内存 SQLite 数据库，自动迁移全部表。
// 每个测试调用一次，各测试之间数据库完全隔离。
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "无法打开内存数据库")

	err = db.AutoMigrate(
		&model.Campus{},
		&model.User{},
		&model.MaterialCategory{},
		&model.PurchaseOrder{},
		&model.Stock{},
		&model.Activity{},
		&model.ActivityContact{},
		&model.Distribution{},
		&model.ExecutionRecord{},
		&model.AmortizationSnapshot{},
		&model.Settlement{},
		&model.RecoveryItem{},
		&model.AuditLog{},
	)
	require.NoError(t, err, "数据库迁移失败")

	// 启用外键约束（SQLite 默认不启用）
	db.Exec("PRAGMA foreign_keys = ON")

	return db
}
