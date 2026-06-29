package activity

import "github.com/gin-gonic/gin"

// Handler 活动 HTTP 处理
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(c *gin.Context)        {}
func (h *Handler) Create(c *gin.Context)      {}
func (h *Handler) Update(c *gin.Context)      {}
func (h *Handler) Detail(c *gin.Context)      {}
func (h *Handler) AddExecution(c *gin.Context) {}
func (h *Handler) Archive(c *gin.Context)     {}
