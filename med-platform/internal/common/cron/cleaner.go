package cron

import (
	"io/fs"
	"med-platform/internal/common/db"    // 引入 DB
	"med-platform/internal/common/logger"
	"med-platform/internal/common/model" // 引入 Model
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

const (
	TempDir         = "./uploads/temp"
	MaxFileAge      = 24 * time.Hour      // 临时文件保留 24 小时
	MaxNotifAge     = 90 * 24 * time.Hour // 🔥 通知保留 90 天 (3个月)
	CleanInterval   = 1 * time.Hour       // 检查频率 (每小时唤醒一次)
)

// StartBackgroundTasks 启动所有后台清理任务 (非阻塞)
func StartBackgroundTasks() {
	go func() {
		// 1. 启动缓冲，避免跟 Server 启动抢资源
		time.Sleep(10 * time.Second)
		logger.Log.Info("🕒 后台清理守护进程已启动...")

		ticker := time.NewTicker(CleanInterval)
		defer ticker.Stop()

		// 立即执行一次
		runTasks()

		// 循环等待
		for range ticker.C {
			runTasks()
		}
	}()
}

// runTasks 统一执行所有子任务
func runTasks() {
	cleanTempFiles()
	cleanExpiredNotifications()
}

// 任务1：清理临时文件
func cleanTempFiles() {
	if _, err := os.Stat(TempDir); os.IsNotExist(err) {
		return
	}

	deletedCount := 0
	errorsCount := 0

	err := filepath.WalkDir(TempDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// 超过 24 小时未移动到正式目录的图片，视为垃圾
		if time.Since(info.ModTime()) > MaxFileAge {
			if err := os.Remove(path); err == nil {
				deletedCount++
			} else {
				errorsCount++
			}
		}
		return nil
	})

	if err != nil {
		logger.Log.Error("清理临时目录出错", zap.Error(err))
	} else if deletedCount > 0 {
		logger.Log.Info("🧹 临时文件清理完成", zap.Int("已删除", deletedCount), zap.Int("失败", errorsCount))
	}
}

// 🔥 任务2：清理过期通知 (新增)
func cleanExpiredNotifications() {
	// 计算截止时间：当前时间 - 90天
	deadline := time.Now().Add(-MaxNotifAge)

	// SQL 逻辑：删除所有 (已读 AND 创建时间早于截止时间) 的记录
	// Unscoped() 表示硬删除，彻底释放空间，而不是软删除(DeletedAt)
	result := db.DB.Unscoped().
		Where("is_read = ? AND created_at < ?", true, deadline).
		Delete(&model.Notification{})

	if result.Error != nil {
		logger.Log.Error("清理过期通知失败", zap.Error(result.Error))
	} else if result.RowsAffected > 0 {
		logger.Log.Info("📭 过期通知清理完成", 
			zap.Int64("已清理条数", result.RowsAffected), 
			zap.String("截止日期", deadline.Format("2006-01-02")),
		)
	}
}