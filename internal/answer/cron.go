package answer

import (
	"fmt"
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/question"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StartArchivingTask 启动后台归档任务
// 请在 main.go 中调用: go answer.StartArchivingTask()
func StartArchivingTask() {
	// 创建一个每 24 小时触发一次的定时器
	ticker := time.NewTicker(24 * time.Hour)
	
	// 使用协程在后台运行，不阻塞主程序
	go func() {
		// (可选) 启动时立即检查一次，防止服务重启导致错过当天的归档
		// PerformArchive() 

		for {
			// 阻塞等待下一个 tick
			<-ticker.C
			PerformArchive()
		}
	}()
}

// PerformArchive 执行核心归档逻辑
// 逻辑：将 365 天前的每日统计数据 -> 汇总累加到历史总表 -> 物理删除旧数据
func PerformArchive() {
	// 定义过期时间线：365天前
	cutoffDate := time.Now().AddDate(0, 0, -365).Format("2006-01-02")
	fmt.Printf("[Archive] 开始执行数据归档，清理 %s 之前的记录...\n", cutoffDate)

	tx := db.DB.Begin()

	// 1. 统计那些即将被删除的数据，按 UserID 分组求和
	type UserSum struct {
		UserID uint
		Total  int64
	}
	var sums []UserSum
	
	// 查询: SELECT user_id, SUM(count) FROM user_daily_stats WHERE date_str < '...' GROUP BY user_id
	if err := tx.Model(&question.UserDailyStat{}).
		Select("user_id, SUM(count) as total").
		Where("date_str < ?", cutoffDate).
		Group("user_id").
		Scan(&sums).Error; err != nil {
		tx.Rollback()
		fmt.Println("[Archive] 统计过期数据失败:", err)
		return
	}

	if len(sums) == 0 {
		tx.Rollback()
		fmt.Println("[Archive] 没有发现需要归档的数据，任务结束。")
		return
	}

	// 2. 将统计结果累加到 user_archived_stats 表 (冷数据表)
	for _, s := range sums {
		// 使用 Upsert (Insert On Conflict Update)
		// 如果该用户在归档表里没记录 -> 插入
		// 如果已有记录 -> 累加 TotalCount
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}}, // 冲突检测字段
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_count": gorm.Expr("user_archived_stats.total_count + ?", s.Total), // 累加逻辑
			}),
		}).Create(&question.UserArchivedStat{
			UserID:     s.UserID,
			TotalCount: s.Total,
		}).Error
		
		if err != nil {
			tx.Rollback()
			fmt.Printf("[Archive] 用户 %d 数据归档写入失败: %v\n", s.UserID, err)
			return
		}
	}

	// 3. 物理删除每日统计表中的老旧数据 (瘦身)
	if err := tx.Where("date_str < ?", cutoffDate).Delete(&question.UserDailyStat{}).Error; err != nil {
		tx.Rollback()
		fmt.Println("[Archive] 删除过期数据失败:", err)
		return
	}

	// 提交事务
	tx.Commit()
	fmt.Printf("[Archive] 归档完成！成功迁移了 %d 位用户的历史数据。\n", len(sums))
}