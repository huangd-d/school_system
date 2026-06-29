package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// StartBackupScheduler 启动定时备份（每天凌晨3点执行）
func StartBackupScheduler(dbPath string, logger *zap.Logger) *cron.Cron {
	c := cron.New()

	c.AddFunc("0 3 * * *", func() {
		if err := backup(dbPath); err != nil {
			logger.Error("数据库备份失败", zap.Error(err))
		} else {
			logger.Info("数据库备份成功")
		}
	})

	c.Start()
	return c
}

// backup 将数据库文件复制到 backups 目录
func backup(dbPath string) error {
	src, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库文件失败: %w", err)
	}
	defer src.Close()

	backupDir := filepath.Join(filepath.Dir(dbPath), "backups")
	os.MkdirAll(backupDir, 0755)

	backupFile := filepath.Join(backupDir,
		fmt.Sprintf("school_%s.db", time.Now().Format("20060102_150405")))

	dst, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("创建备份文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(src, dst); err != nil {
		return fmt.Errorf("复制数据库失败: %w", err)
	}

	return nil
}
