package material

import (
	"context"
	"strconv"

	"school-system/internal/model"
	"school-system/pkg/apperror"
	"school-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// ---- 请求结构体 ----

// CreateCategoryReq 创建分类请求
type CreateCategoryReq struct {
	Name string `json:"name" binding:"required"`
	Note string `json:"note"`
}

// UpdateCategoryReq 更新分类请求
type UpdateCategoryReq struct {
	Name string `json:"name" binding:"required"`
	Note string `json:"note"`
}

// PurchaseReq 采购入库请求
type PurchaseReq struct {
	MaterialName string  `json:"material_name" binding:"required"`
	CategoryID   uint    `json:"category_id" binding:"required,min=1"`
	Quantity     int     `json:"quantity" binding:"required,min=1"`
	TotalAmount  float64 `json:"total_amount" binding:"required,min=0.01"`
	Notes        string  `json:"notes"`
}

// DistributeReq 派发物资请求
type DistributeReq struct {
	StockID    uint   `json:"stock_id" binding:"required,min=1"`
	ActivityID uint   `json:"activity_id" binding:"required,min=1"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	Reason     string `json:"reason"`
}

// AdjustDistributionReq 调整派发请求
type AdjustDistributionReq struct {
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Reason   string `json:"reason"`
}

// ---- 响应结构体 ----

// CategoryResp 分类响应
type CategoryResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

// StockResp 库存响应
type StockResp struct {
	ID              uint    `json:"id"`
	PurchaseOrderID uint    `json:"purchase_order_id"`
	CategoryID      uint    `json:"category_id"`
	MaterialName    string  `json:"material_name"`
	TotalQuantity   int     `json:"total_quantity"`
	RemainingQty    int     `json:"remaining_qty"`
	UnitPrice       float64 `json:"unit_price"`
	Source          string  `json:"source"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// DistributionResp 派发记录响应
type DistributionResp struct {
	ID         uint   `json:"id"`
	StockID    uint   `json:"stock_id"`
	ActivityID uint   `json:"activity_id"`
	Quantity   int    `json:"quantity"`
	OperatorID uint   `json:"operator_id"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
}

// PurchaseOrderResp 采购单响应
type PurchaseOrderResp struct {
	ID           uint    `json:"id"`
	MaterialName string  `json:"material_name"`
	CategoryID   uint    `json:"category_id"`
	Quantity     int     `json:"quantity"`
	TotalAmount  float64 `json:"total_amount"`
	UnitPrice    float64 `json:"unit_price"`
	Notes        string  `json:"notes"`
	PurchaserID  uint    `json:"purchaser_id"`
	CreatedAt    string  `json:"created_at"`
}

// StockListResp 库存列表响应
type StockListResp struct {
	List  []StockResp `json:"list"`
	Total int64       `json:"total"`
}

// PurchaseOrderListResp 采购单列表响应
type PurchaseOrderListResp struct {
	List  []PurchaseOrderResp `json:"list"`
	Total int64               `json:"total"`
}

// ---- 辅助函数 ----

// getOperator 从 context 获取当前操作人信息（由 auth 中间件注入）
func getOperator(c *gin.Context) (userID, campusID uint, role string) {
	if v, ok := c.Get("user_id"); ok {
		userID = v.(uint)
	}
	if v, ok := c.Get("campus_id"); ok {
		campusID = v.(uint)
	}
	if v, ok := c.Get("role"); ok {
		role = v.(string)
	}
	return
}

func toCategoryResp(cat model.MaterialCategory) CategoryResp {
	return CategoryResp{
		ID:        cat.ID,
		Name:      cat.Name,
		Note:      cat.Note,
		CreatedAt: cat.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toStockResp(s model.Stock) StockResp {
	return StockResp{
		ID:              s.ID,
		PurchaseOrderID: s.PurchaseOrderID,
		CategoryID:      s.CategoryID,
		MaterialName:    s.MaterialName,
		TotalQuantity:   s.TotalQuantity,
		RemainingQty:    s.RemainingQty,
		UnitPrice:       s.UnitPrice,
		Source:          s.Source,
		CreatedAt:       s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toDistributionResp(d model.Distribution) DistributionResp {
	return DistributionResp{
		ID:         d.ID,
		StockID:    d.StockID,
		ActivityID: d.ActivityID,
		Quantity:   d.Quantity,
		OperatorID: d.OperatorID,
		Reason:     d.Reason,
		CreatedAt:  d.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toPurchaseOrderResp(po model.PurchaseOrder) PurchaseOrderResp {
	return PurchaseOrderResp{
		ID:           po.ID,
		MaterialName: po.MaterialName,
		CategoryID:   po.CategoryID,
		Quantity:     po.Quantity,
		TotalAmount:  po.TotalAmount,
		UnitPrice:    po.UnitPrice,
		Notes:        po.Notes,
		PurchaserID:  po.PurchaserID,
		CreatedAt:    po.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ---- ServiceInterface（handler 依赖的服务接口，便于测试时注入 mock）----

type ServiceInterface interface {
	ListCategories(ctx context.Context) ([]model.MaterialCategory, error)
	CreateCategory(ctx context.Context, name, note string, operatorRole string) (*model.MaterialCategory, error)
	UpdateCategory(ctx context.Context, id uint, name, note string, operatorRole string) (*model.MaterialCategory, error)
	DeleteCategory(ctx context.Context, id uint, operatorRole string) error
	Purchase(ctx context.Context, materialName string, categoryID uint, quantity int, totalAmount float64, notes string, purchaserID uint, operatorRole string) (*model.Stock, error)
	ListPurchaseOrders(ctx context.Context, page, pageSize int) (*PurchaseOrderListResult, error)
	ListStock(ctx context.Context, categoryID uint, keyword string, page, pageSize int) (*StockListResult, error)
	GetStock(ctx context.Context, id uint) (*model.Stock, error)
	ListStockDistributions(ctx context.Context, stockID uint) ([]model.Distribution, error)
	Distribute(ctx context.Context, stockID uint, activityID uint, quantity int, operatorID uint, reason string, operatorRole string) (*model.Distribution, error)
	AdjustDistribution(ctx context.Context, distributionID uint, newQuantity int, operatorID uint, reason string, operatorRole string) error
}

// ---- Handler ----

// Handler 物资 HTTP 处理
type Handler struct {
	svc ServiceInterface
}

func NewHandler(svc ServiceInterface) *Handler { return &Handler{svc: svc} }

// ---- 分类管理 ----

// ListCategories 获取分类列表
func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]CategoryResp, 0, len(cats))
	for _, cat := range cats {
		resp = append(resp, toCategoryResp(cat))
	}
	response.OK(c, resp)
}

// CreateCategory 创建分类
func (h *Handler) CreateCategory(c *gin.Context) {
	var req CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	_, _, role := getOperator(c)

	cat, err := h.svc.CreateCategory(c.Request.Context(), req.Name, req.Note, role)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toCategoryResp(*cat))
}

// UpdateCategory 更新分类
func (h *Handler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	_, _, role := getOperator(c)

	cat, err := h.svc.UpdateCategory(c.Request.Context(), uint(id), req.Name, req.Note, role)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toCategoryResp(*cat))
}

// DeleteCategory 删除分类
func (h *Handler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	_, _, role := getOperator(c)

	if err := h.svc.DeleteCategory(c.Request.Context(), uint(id), role); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}

// ---- 采购与库存 ----

// Purchase 采购入库
func (h *Handler) Purchase(c *gin.Context) {
	var req PurchaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	userID, _, role := getOperator(c)

	stock, err := h.svc.Purchase(c.Request.Context(), req.MaterialName, req.CategoryID, req.Quantity, req.TotalAmount, req.Notes, userID, role)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toStockResp(*stock))
}

// ListPurchaseOrders 获取采购单列表
func (h *Handler) ListPurchaseOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListPurchaseOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]PurchaseOrderResp, 0, len(result.Orders))
	for _, po := range result.Orders {
		resp = append(resp, toPurchaseOrderResp(po))
	}

	response.OK(c, PurchaseOrderListResp{List: resp, Total: result.Total})
}

// ListStock 获取库存列表
func (h *Handler) ListStock(c *gin.Context) {
	categoryID, _ := strconv.ParseUint(c.DefaultQuery("category_id", "0"), 10, 64)
	keyword := c.DefaultQuery("keyword", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.svc.ListStock(c.Request.Context(), uint(categoryID), keyword, page, pageSize)
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]StockResp, 0, len(result.Stocks))
	for _, s := range result.Stocks {
		resp = append(resp, toStockResp(s))
	}

	response.OK(c, StockListResp{List: resp, Total: result.Total})
}

// GetStock 获取库存详情
func (h *Handler) GetStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	stock, err := h.svc.GetStock(c.Request.Context(), uint(id))
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toStockResp(*stock))
}

// ListStockDistributions 获取某库存的所有派发记录
func (h *Handler) ListStockDistributions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	dists, err := h.svc.ListStockDistributions(c.Request.Context(), uint(id))
	if err != nil {
		response.Err(c, err)
		return
	}

	resp := make([]DistributionResp, 0, len(dists))
	for _, d := range dists {
		resp = append(resp, toDistributionResp(d))
	}
	response.OK(c, resp)
}

// ---- 派发与调整 ----

// Distribute 派发物资到活动
func (h *Handler) Distribute(c *gin.Context) {
	var req DistributeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	userID, _, role := getOperator(c)

	dist, err := h.svc.Distribute(c.Request.Context(), req.StockID, req.ActivityID, req.Quantity, userID, req.Reason, role)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, toDistributionResp(*dist))
}

// AdjustDistribution 调整派发数量
func (h *Handler) AdjustDistribution(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	var req AdjustDistributionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperror.ErrInvalidParam)
		return
	}

	userID, _, role := getOperator(c)

	if err := h.svc.AdjustDistribution(c.Request.Context(), uint(id), req.Quantity, userID, req.Reason, role); err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, nil)
}
