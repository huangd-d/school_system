package material_test

import (
	"context"
	"testing"

	"school-system/internal/model"
	"school-system/internal/module/material"
	"school-system/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---- 辅助：创建测试数据 ----

func createTestCategory(t *testing.T, db *gorm.DB, name string) *model.MaterialCategory {
	t.Helper()
	cat := &model.MaterialCategory{Name: name}
	require.NoError(t, db.Create(cat).Error, "创建测试分类失败")
	return cat
}

// ============================================================
//  分类 CRUD
// ============================================================

func TestMaterialRepo_FindAllCategories(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	createTestCategory(t, db, "教材")
	createTestCategory(t, db, "文具")

	cats, err := repo.FindAllCategories(context.Background())
	require.NoError(t, err)
	assert.Len(t, cats, 2)
}

func TestMaterialRepo_FindAllCategories_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	cats, err := repo.FindAllCategories(context.Background())
	require.NoError(t, err)
	assert.Empty(t, cats)
}

func TestMaterialRepo_FindCategoryByID(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	created := createTestCategory(t, db, "教材")

	cat, err := repo.FindCategoryByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "教材", cat.Name)
}

func TestMaterialRepo_FindCategoryByName(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	createTestCategory(t, db, "教材")

	cat, err := repo.FindCategoryByName(context.Background(), "教材")
	require.NoError(t, err)
	assert.Equal(t, "教材", cat.Name)
}

func TestMaterialRepo_FindCategoryByName_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	_, err := repo.FindCategoryByName(context.Background(), "不存在的分类")
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestMaterialRepo_CreateCategory(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	cat := &model.MaterialCategory{Name: "教材", Note: "教学用书"}
	err := repo.CreateCategory(context.Background(), cat)
	require.NoError(t, err)
	assert.NotZero(t, cat.ID)

	// 验证持久化
	found, err := repo.FindCategoryByID(context.Background(), cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "教材", found.Name)
}

func TestMaterialRepo_UpdateCategory(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	cat := createTestCategory(t, db, "旧名称")
	cat.Name = "新名称"
	cat.Note = "新备注"

	err := repo.UpdateCategory(context.Background(), cat)
	require.NoError(t, err)

	found, err := repo.FindCategoryByID(context.Background(), cat.ID)
	require.NoError(t, err)
	assert.Equal(t, "新名称", found.Name)
	assert.Equal(t, "新备注", found.Note)
}

func TestMaterialRepo_DeleteCategory(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	cat := createTestCategory(t, db, "待删除")

	err := repo.DeleteCategory(context.Background(), cat.ID)
	require.NoError(t, err)

	_, err = repo.FindCategoryByID(context.Background(), cat.ID)
	assert.Error(t, err)
}

// ============================================================
//  采购单
// ============================================================

func TestMaterialRepo_CreatePurchaseOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	po := &model.PurchaseOrder{
		MaterialName: "语文教材",
		CategoryID:   1,
		Quantity:     100,
		TotalAmount:  5000,
		UnitPrice:    50.0,
		PurchaserID:  1,
	}
	err := repo.CreatePurchaseOrder(context.Background(), po)
	require.NoError(t, err)
	assert.NotZero(t, po.ID)
}

func TestMaterialRepo_FindPurchaseOrders_Pagination(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	// 创建3条采购单
	for i := 0; i < 3; i++ {
		po := &model.PurchaseOrder{
			MaterialName: "测试物资",
			CategoryID:   1,
			Quantity:     10,
			TotalAmount:  100,
			UnitPrice:    10.0,
			PurchaserID:  1,
		}
		require.NoError(t, db.Create(po).Error)
	}

	// 第一页
	orders, total, err := repo.FindPurchaseOrders(context.Background(), 0, 2)
	require.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, int64(3), total)

	// 第二页
	orders, total, err = repo.FindPurchaseOrders(context.Background(), 2, 2)
	require.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, int64(3), total)
}

// ============================================================
//  库存
// ============================================================

func TestMaterialRepo_CreateStock(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	stock := &model.Stock{
		PurchaseOrderID: 1,
		CategoryID:      1,
		MaterialName:    "语文教材",
		TotalQuantity:   100,
		RemainingQty:    100,
		UnitPrice:       50.0,
		Source:          "purchase",
	}
	err := repo.CreateStock(context.Background(), stock)
	require.NoError(t, err)
	assert.NotZero(t, stock.ID)
}

func TestMaterialRepo_FindStockWithFilter_Category(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	// 创建两个不同分类的库存
	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "分类1物资", TotalQuantity: 10, RemainingQty: 10, UnitPrice: 10, Source: "purchase"}
	s2 := &model.Stock{PurchaseOrderID: 2, CategoryID: 2, MaterialName: "分类2物资", TotalQuantity: 20, RemainingQty: 20, UnitPrice: 20, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)
	require.NoError(t, db.Create(s2).Error)

	stocks, total, err := repo.FindStockWithFilter(context.Background(), 1, "", 0, 10)
	require.NoError(t, err)
	assert.Len(t, stocks, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "分类1物资", stocks[0].MaterialName)
}

func TestMaterialRepo_FindStockWithFilter_Keyword(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 10, RemainingQty: 10, UnitPrice: 10, Source: "purchase"}
	s2 := &model.Stock{PurchaseOrderID: 2, CategoryID: 1, MaterialName: "数学教材", TotalQuantity: 20, RemainingQty: 20, UnitPrice: 20, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)
	require.NoError(t, db.Create(s2).Error)

	stocks, total, err := repo.FindStockWithFilter(context.Background(), 0, "语文", 0, 10)
	require.NoError(t, err)
	assert.Len(t, stocks, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "语文教材", stocks[0].MaterialName)
}

func TestMaterialRepo_UpdateStock(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	stock := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 100, UnitPrice: 50, Source: "purchase"}
	require.NoError(t, db.Create(stock).Error)

	stock.RemainingQty = 80
	err := repo.UpdateStock(context.Background(), stock)
	require.NoError(t, err)

	found, err := repo.FindStockByID(context.Background(), stock.ID)
	require.NoError(t, err)
	assert.Equal(t, 80, found.RemainingQty)
}

// ============================================================
//  派发
// ============================================================

func TestMaterialRepo_CreateDistribution(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	dist := &model.Distribution{
		StockID:    1,
		ActivityID: 1,
		Quantity:   10,
		OperatorID: 1,
		Reason:     "活动需要",
	}
	err := repo.CreateDistribution(context.Background(), dist)
	require.NoError(t, err)
	assert.NotZero(t, dist.ID)
}

func TestMaterialRepo_FindDistributionByID(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	dist := &model.Distribution{StockID: 1, ActivityID: 1, Quantity: 10, OperatorID: 1}
	require.NoError(t, db.Create(dist).Error)

	found, err := repo.FindDistributionByID(context.Background(), dist.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, found.Quantity)
}

func TestMaterialRepo_SumDistributionsByStock(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	// 同一库存创建两条派发记录
	d1 := &model.Distribution{StockID: 1, ActivityID: 1, Quantity: 10, OperatorID: 1}
	d2 := &model.Distribution{StockID: 1, ActivityID: 2, Quantity: 5, OperatorID: 1}
	require.NoError(t, db.Create(d1).Error)
	require.NoError(t, db.Create(d2).Error)

	total, err := repo.SumDistributionsByStock(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 15, total)
}

func TestMaterialRepo_UpdateDistribution(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	dist := &model.Distribution{StockID: 1, ActivityID: 1, Quantity: 10, OperatorID: 1}
	require.NoError(t, db.Create(dist).Error)

	dist.Quantity = 15
	dist.Reason = "追加配发"
	err := repo.UpdateDistribution(context.Background(), dist)
	require.NoError(t, err)

	found, err := repo.FindDistributionByID(context.Background(), dist.ID)
	require.NoError(t, err)
	assert.Equal(t, 15, found.Quantity)
	assert.Equal(t, "追加配发", found.Reason)
}

func TestMaterialRepo_FindDistributionsByActivity(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	d1 := &model.Distribution{StockID: 1, ActivityID: 1, Quantity: 10, OperatorID: 1}
	d2 := &model.Distribution{StockID: 2, ActivityID: 2, Quantity: 5, OperatorID: 1}
	require.NoError(t, db.Create(d1).Error)
	require.NoError(t, db.Create(d2).Error)

	dists, err := repo.FindDistributionsByActivity(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, dists, 1)
	assert.Equal(t, uint(1), dists[0].ActivityID)
}

// ============================================================
//  审计日志
// ============================================================

func TestMaterialRepo_CreateAuditLog(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	log := &model.AuditLog{
		OperatorID:    1,
		OperationType: model.AuditOpPurchase,
		EntityType:    "stock",
		EntityID:      1,
		AfterValue:    "name:教材,qty:100,amount:5000",
		ImpactAmount:  5000,
	}
	err := repo.CreateAuditLog(context.Background(), log)
	require.NoError(t, err)
	assert.NotZero(t, log.ID)
}

// ============================================================
//  CountPurchasesByCategory
// ============================================================

func TestMaterialRepo_CountPurchasesByCategory(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	// 同一分类创建两条采购单
	for i := 0; i < 2; i++ {
		po := &model.PurchaseOrder{
			MaterialName: "测试物资",
			CategoryID:   1,
			Quantity:     10,
			TotalAmount:  100,
			UnitPrice:    10.0,
			PurchaserID:  1,
		}
		require.NoError(t, db.Create(po).Error)
	}

	count, err := repo.CountPurchasesByCategory(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// 不存在的分类
	count, err = repo.CountPurchasesByCategory(context.Background(), 999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ============================================================
//  FindDistributionsWithFilter — 派发记录筛选查询
// ============================================================

func TestMaterialRepo_FindDistributionsWithFilter_All(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	// 创建库存
	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 100, UnitPrice: 50, Source: "purchase"}
	s2 := &model.Stock{PurchaseOrderID: 2, CategoryID: 1, MaterialName: "数学教材", TotalQuantity: 50, RemainingQty: 50, UnitPrice: 30, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)
	require.NoError(t, db.Create(s2).Error)

	// 创建派发记录
	d1 := &model.Distribution{StockID: s1.ID, ActivityID: 1, Quantity: 10, OperatorID: 1, Reason: "活动需要"}
	d2 := &model.Distribution{StockID: s2.ID, ActivityID: 2, Quantity: 5, OperatorID: 1, Reason: "教学使用"}
	require.NoError(t, db.Create(d1).Error)
	require.NoError(t, db.Create(d2).Error)

	// 无筛选条件，应返回全部
	dists, total, err := repo.FindDistributionsWithFilter(context.Background(), 0, "", "", "", 0, 10)
	require.NoError(t, err)
	assert.Len(t, dists, 2)
	assert.Equal(t, int64(2), total)
	// 验证 JOIN 出了物资名称
	assert.Equal(t, "语文教材", dists[0].MaterialName)
	assert.Equal(t, "数学教材", dists[1].MaterialName)
}

func TestMaterialRepo_FindDistributionsWithFilter_ByActivity(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 100, UnitPrice: 50, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)

	d1 := &model.Distribution{StockID: s1.ID, ActivityID: 1, Quantity: 10, OperatorID: 1}
	d2 := &model.Distribution{StockID: s1.ID, ActivityID: 2, Quantity: 5, OperatorID: 1}
	require.NoError(t, db.Create(d1).Error)
	require.NoError(t, db.Create(d2).Error)

	dists, total, err := repo.FindDistributionsWithFilter(context.Background(), 1, "", "", "", 0, 10)
	require.NoError(t, err)
	assert.Len(t, dists, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, uint(1), dists[0].ActivityID)
}

func TestMaterialRepo_FindDistributionsWithFilter_ByKeyword(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 100, UnitPrice: 50, Source: "purchase"}
	s2 := &model.Stock{PurchaseOrderID: 2, CategoryID: 1, MaterialName: "数学教材", TotalQuantity: 50, RemainingQty: 50, UnitPrice: 30, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)
	require.NoError(t, db.Create(s2).Error)

	d1 := &model.Distribution{StockID: s1.ID, ActivityID: 1, Quantity: 10, OperatorID: 1}
	d2 := &model.Distribution{StockID: s2.ID, ActivityID: 1, Quantity: 5, OperatorID: 1}
	require.NoError(t, db.Create(d1).Error)
	require.NoError(t, db.Create(d2).Error)

	dists, total, err := repo.FindDistributionsWithFilter(context.Background(), 0, "语文", "", "", 0, 10)
	require.NoError(t, err)
	assert.Len(t, dists, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "语文教材", dists[0].MaterialName)
}

func TestMaterialRepo_FindDistributionsWithFilter_ByDateRange(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 100, UnitPrice: 50, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)

	// 创建派发记录（SQLite 使用当前时间）
	d1 := &model.Distribution{StockID: s1.ID, ActivityID: 1, Quantity: 10, OperatorID: 1}
	require.NoError(t, db.Create(d1).Error)

	// 使用一个很大的日期范围应能查到
	dists, total, err := repo.FindDistributionsWithFilter(context.Background(), 0, "", "2020-01-01", "2099-12-31", 0, 10)
	require.NoError(t, err)
	assert.Len(t, dists, 1)
	assert.Equal(t, int64(1), total)

	// 使用未来的日期范围应查不到
	dists, total, err = repo.FindDistributionsWithFilter(context.Background(), 0, "", "2099-01-01", "2099-12-31", 0, 10)
	require.NoError(t, err)
	assert.Empty(t, dists)
	assert.Equal(t, int64(0), total)
}

func TestMaterialRepo_FindDistributionsWithFilter_Pagination(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	s1 := &model.Stock{PurchaseOrderID: 1, CategoryID: 1, MaterialName: "语文教材", TotalQuantity: 100, RemainingQty: 100, UnitPrice: 50, Source: "purchase"}
	require.NoError(t, db.Create(s1).Error)

	for i := 0; i < 3; i++ {
		d := &model.Distribution{StockID: s1.ID, ActivityID: 1, Quantity: 10, OperatorID: 1}
		require.NoError(t, db.Create(d).Error)
	}

	// 第一页
	dists, total, err := repo.FindDistributionsWithFilter(context.Background(), 0, "", "", "", 0, 2)
	require.NoError(t, err)
	assert.Len(t, dists, 2)
	assert.Equal(t, int64(3), total)

	// 第二页
	dists, total, err = repo.FindDistributionsWithFilter(context.Background(), 0, "", "", "", 2, 2)
	require.NoError(t, err)
	assert.Len(t, dists, 1)
	assert.Equal(t, int64(3), total)
}

func TestMaterialRepo_FindDistributionsWithFilter_Empty(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := material.NewRepository(db)

	dists, total, err := repo.FindDistributionsWithFilter(context.Background(), 0, "", "", "", 0, 10)
	require.NoError(t, err)
	assert.Empty(t, dists)
	assert.Equal(t, int64(0), total)
}
