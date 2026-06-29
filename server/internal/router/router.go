package router

import (
	"school-system/internal/config"
	"school-system/internal/middleware"
	"school-system/internal/module/activity"
	"school-system/internal/module/auth"
	"school-system/internal/module/campus"
	"school-system/internal/module/material"
	"school-system/internal/module/report"
	"school-system/internal/module/settlement"
	"school-system/internal/module/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Register 注册所有路由和中间件
func Register(
	r *gin.Engine,
	db *gorm.DB,
	cfg *config.Config,
	logger *zap.Logger,
) {
	// ---- 初始化各模块 ----
	// 认证（无需 repository）
	authSvc := auth.NewService(db, &cfg.JWT)
	authH := auth.NewHandler(authSvc)

	// 校区
	campusRepo := campus.NewRepository(db)
	campusSvc := campus.NewService(campusRepo)
	campusH := campus.NewHandler(campusSvc)

	// 账户（注入 campusRepo 用于校区校验）
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, campusRepo)
	userH := user.NewHandler(userSvc)

	// 物资
	materialRepo := material.NewRepository(db)
	materialSvc := material.NewService(materialRepo)
	materialH := material.NewHandler(materialSvc)

	// 活动
	activityRepo := activity.NewRepository(db)
	activitySvc := activity.NewService(activityRepo)
	activityH := activity.NewHandler(activitySvc)

	// 结算
	settlementRepo := settlement.NewRepository(db)
	settlementSvc := settlement.NewService(settlementRepo)
	settlementH := settlement.NewHandler(settlementSvc)

	// 报表
	reportRepo := report.NewRepository(db)
	reportSvc := report.NewService(reportRepo, db)
	reportH := report.NewHandler(reportSvc)

	// ---- 全局中间件链 ----
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// ---- 公开路由（无需登录）----
	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", authH.Login)
		api.POST("/auth/refresh", authH.Refresh)
	}

	// ---- 需登录路由 ----
	authGroup := api.Group("")
	authGroup.Use(middleware.Auth(cfg.JWT.Secret))
	authGroup.Use(middleware.DataScope())
	{
		// 校区
		authGroup.GET("/campuses", campusH.List)
		authGroup.POST("/campuses", campusH.Create)
		authGroup.PUT("/campuses/:id", campusH.Update)
		authGroup.DELETE("/campuses/:id", campusH.Delete)

		// 账户
		authGroup.GET("/users", userH.List)
		authGroup.POST("/users", userH.Create)
		authGroup.PUT("/users/:id", userH.Update)
		authGroup.PUT("/users/:id/disable", userH.Disable)
		authGroup.PUT("/users/:id/reset-pwd", userH.ResetPassword)

		// 物资
		authGroup.GET("/materials/categories", materialH.ListCategories)
		authGroup.POST("/materials/categories", materialH.CreateCategory)
		authGroup.GET("/materials/stock", materialH.ListStock)
		authGroup.POST("/materials/purchase", materialH.Purchase)
		authGroup.POST("/materials/distribute", materialH.Distribute)
		authGroup.PUT("/materials/distribute/:id", materialH.AdjustDistribution)

		// 活动
		authGroup.GET("/activities", activityH.List)
		authGroup.POST("/activities", activityH.Create)
		authGroup.PUT("/activities/:id", activityH.Update)
		authGroup.GET("/activities/:id", activityH.Detail)
		authGroup.POST("/activities/:id/executions", activityH.AddExecution)
		authGroup.PUT("/activities/:id/archive", activityH.Archive)

		// 结算
		authGroup.POST("/settlements/preview/:activity_id", settlementH.Preview)
		authGroup.POST("/settlements/execute/:activity_id", settlementH.Execute)
		authGroup.POST("/settlements/reverse/:settlement_id", settlementH.Reverse)

		// 报表
		authGroup.GET("/reports/by-activity", reportH.ByActivity)
		authGroup.GET("/reports/by-date-range", reportH.ByDateRange)
		authGroup.GET("/reports/by-campus", reportH.ByCampus)
		authGroup.GET("/reports/by-category", reportH.ByCategory)
	}
}
