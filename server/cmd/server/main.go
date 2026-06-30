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
