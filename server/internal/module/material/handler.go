package material

import "github.com/gin-gonic/gin"

// Handler 物资 HTTP 处理
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// 分类管理
func (h *Handler) ListCategories(c *gin.Context)  {}
func (h *Handler) CreateCategory(c *gin.Context)  {}

// 采购与库存
func (h *Handler) Purchase(c *gin.Context)  {}
func (h *Handler) ListStock(c *gin.Context) {}

// 派发与调整
func (h *Handler) Distribute(c *gin.Context)       {}
func (h *Handler) AdjustDistribution(c *gin.Context) {}
