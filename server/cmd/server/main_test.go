package main

import (
	"testing"

	"school-system/internal/model"
	"school-system/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrateAmountToCents 验证金额单位迁移：float 元 → int 分（×100），且幂等
func TestMigrateAmountToCents(t *testing.T) {
	db := testutil.NewTestDB(t)

	// 预置旧格式数据（元）——通过原始 SQL 写入浮点值
	require.NoError(t, db.Exec(
		"INSERT INTO purchase_orders (material_name, category_id, quantity, total_amount, unit_price, purchaser_id, created_at) VALUES ('教材', 1, 10, 110.0, 11.0, 1, '2024-01-01 00:00:00')",
	).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO stocks (purchase_order_id, category_id, material_name, total_quantity, remaining_qty, unit_price, source, created_at, updated_at) VALUES (1, 1, '教材', 10, 10, 11.0, 'purchase', '2024-01-01 00:00:00', '2024-01-01 00:00:00')",
	).Error)

	// 执行迁移
	require.NoError(t, migrateAmountToCents(db))

	var po model.PurchaseOrder
	require.NoError(t, db.First(&po, 1).Error)
	assert.Equal(t, int64(11000), po.TotalAmount, "110.0 元 → 11000 分")
	assert.Equal(t, int64(1100), po.UnitPrice, "11.0 元 → 1100 分")

	var stock model.Stock
	require.NoError(t, db.First(&stock, 1).Error)
	assert.Equal(t, int64(1100), stock.UnitPrice)

	// 幂等：再次执行不应重复 ×100
	require.NoError(t, migrateAmountToCents(db))
	var po2 model.PurchaseOrder
	require.NoError(t, db.First(&po2, 1).Error)
	assert.Equal(t, int64(11000), po2.TotalAmount, "幂等：二次迁移不重复乘 100")
}
