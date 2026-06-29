package settlement

import "github.com/gin-gonic/gin"

// Handler 结算 HTTP 处理
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Preview(c *gin.Context) {}
func (h *Handler) Execute(c *gin.Context) {}
func (h *Handler) Reverse(c *gin.Context) {}
