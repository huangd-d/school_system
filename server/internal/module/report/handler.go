package report

import "github.com/gin-gonic/gin"

// Handler 报表 HTTP 处理
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) ByActivity(c *gin.Context)  {}
func (h *Handler) ByDateRange(c *gin.Context) {}
func (h *Handler) ByCampus(c *gin.Context)    {}
func (h *Handler) ByCategory(c *gin.Context)  {}
