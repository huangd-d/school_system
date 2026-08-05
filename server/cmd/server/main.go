package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"school-system/internal/config"
	"school-system/internal/database"
	"school-system/internal/model"
	"school-system/internal/router"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
)

func main() {
	// ---- 加载配置 ----
	cfg := config.Load()

	// ---- 初始化日志 ----
	logger := initLogger()
	defer logger.Sync()

	// ---- 连接数据库 ----
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// ---- 自动创建表 ----
	if err := db.AutoMigrate(
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
	); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// ---- 部分唯一索引：一活动最多一条有效（settled）结算记录（并发兜底） ----
	if err := db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_settlements_activity_settled ON settlements(activity_id) WHERE status = 'settled'",
	).Error; err != nil {
		logger.Fatal("创建结算唯一索引失败", zap.Error(err))
	}

	// ---- 金额单位迁移：float 元 → int 分（user_version=0 → 1，幂等） ----
	if err := migrateAmountToCents(db); err != nil {
		logger.Fatal("金额单位迁移失败", zap.Error(err))
	}

	// ---- 种子数据（确保总部校区和默认管理员存在） ----
	seedDefaults(db, cfg, logger)

	// ---- 启动定时备份 ----
	backupCron := database.StartBackupScheduler(cfg.Database.Path, logger)
	defer backupCron.Stop()

	// ---- 配置 Gin ----
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// ---- 注册路由 ----
	router.Register(r, db, cfg, logger)

	// ---- 启动 HTTP 服务 ----
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		logger.Info("服务启动", zap.String("端口", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// ---- 等待退出信号 ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务关闭异常", zap.Error(err))
	}
	logger.Info("服务已关闭")
}

// initLogger 初始化日志（控制台 + 文件按大小切割）
func initLogger() *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./data/logs/server.log",
		MaxSize:    10, // 单位：MB
		MaxBackups: 5,
		MaxAge:     30, // 天
	})

	consoleWriter := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, zapcore.InfoLevel),
		zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), consoleWriter, zapcore.DebugLevel),
	)

	return zap.New(core)
}

// migrateAmountToCents 金额单位迁移：float 元 → int 分。
// 用 PRAGMA user_version 标记版本（0 → 1），幂等执行；SQLite 动态类型兼容既有 REAL 列。
func migrateAmountToCents(db *gorm.DB) error {
	var version int
	if err := db.Raw("PRAGMA user_version").Scan(&version).Error; err != nil {
		return err
	}
	if version >= 1 {
		return nil // 已迁移
	}

	stmts := []string{
		"UPDATE purchase_orders SET total_amount = ROUND(total_amount * 100), unit_price = ROUND(unit_price * 100)",
		"UPDATE stocks SET unit_price = ROUND(unit_price * 100)",
		"UPDATE settlements SET total_returned_amount = ROUND(total_returned_amount * 100)",
		"UPDATE recovery_items SET cost_deduction = ROUND(cost_deduction * 100)",
		"UPDATE amortization_snapshots SET amortization_base = ROUND(amortization_base * 100), daily_amount = ROUND(daily_amount * 100)",
		"UPDATE audit_logs SET impact_amount = ROUND(impact_amount * 100)",
	}
	for _, sql := range stmts {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return db.Exec("PRAGMA user_version = 1").Error
}

// seedDefaults 确保总部校区和默认管理员账户存在
func seedDefaults(db *gorm.DB, cfg *config.Config, logger *zap.Logger) {
	// 1. 确保总部校区存在
	var hq model.Campus
	result := db.Where("type = ?", model.CampusTypeHQ).First(&hq)
	if result.Error != nil {
		hq = model.Campus{
			Name: cfg.Seed.DefaultHQName,
			Type: model.CampusTypeHQ,
		}
		if err := db.Create(&hq).Error; err != nil {
			logger.Error("创建总部校区失败", zap.Error(err))
			return
		}
		logger.Info("总部校区已创建", zap.String("名称", hq.Name))
	}

	// 2. 确保默认总部管理员存在
	var adminCount int64
	db.Model(&model.User{}).Where("role = ?", model.RoleHQAdmin).Count(&adminCount)
	if adminCount == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Seed.DefaultAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("密码加密失败", zap.Error(err))
			return
		}

		admin := model.User{
			Username:     cfg.Seed.DefaultAdminUsername,
			PasswordHash: string(hash),
			Phone:        cfg.Seed.DefaultAdminPhone,
			Role:         model.RoleHQAdmin,
			CampusID:     hq.ID,
			Status:       model.UserStatusActive,
		}
		if err := db.Create(&admin).Error; err != nil {
			logger.Error("创建默认管理员失败", zap.Error(err))
		} else {
			logger.Info("默认管理员已创建",
				zap.String("用户名", admin.Username),
				zap.String("密码", cfg.Seed.DefaultAdminPassword),
			)
		}
	}
}
