package answer

import (
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/question"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// BatchCreateOrUpdate 批量保存或更新作答流水（支持组合大题一次性提交）
// 🔥 核心优化：
// 1. 批量处理，极大减少数据库往返次数 (RTT)
// 2. 将 N 次的每日统计表事务锁竞争，合并为 1 次批量加 N
// 3. 同步写入 AnswerHistory（历史轨迹），为后续学习曲线分析做准备
func (r *Repository) BatchCreateOrUpdate(records []*AnswerRecord) error {
	if len(records) == 0 {
		return nil
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		userID := records[0].UserID
		today := time.Now().Format("2006-01-02")
		
		for _, record := range records {
			// -------------------------------------------------------
			// 1. 更新当前状态表 (AnswerRecord) - 决定答题卡的颜色
			// -------------------------------------------------------
			var existing AnswerRecord
			err := tx.Where("user_id = ? AND question_id = ?", record.UserID, record.QuestionID).First(&existing).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 没做过 -> 插入新记录
					if err := tx.Create(record).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				// 做过 -> 覆盖旧的选项、对错状态
				existing.Choice = record.Choice
				existing.IsCorrect = record.IsCorrect
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			}

			// -------------------------------------------------------
			// 2. 追加历史轨迹表 (AnswerHistory) - 记录用户的每一次手跳
			// -------------------------------------------------------
			// 历史表是 Append-Only（只增不改）的，所以直接 Create
			history := AnswerHistory{
				UserID:     record.UserID,
				QuestionID: record.QuestionID,
				Choice:     record.Choice,
				IsCorrect:  record.IsCorrect,
			}
			if err := tx.Create(&history).Error; err != nil {
				return err
			}
		}

		// -------------------------------------------------------
		// 3. 批量更新每日刷题统计 (user_daily_stats)
		// -------------------------------------------------------
		// 逻辑：直接增加本次提交的题目总数 (len)
		stat := question.UserDailyStat{
			UserID:  userID,
			DateStr: today,
			Count:   len(records), 
		}
		
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "date_str"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"count": gorm.Expr("user_daily_stats.count + ?", len(records)), // 🔥 一次性加 N
			}),
		}).Create(&stat).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete 物理删除单条作答当前记录 (用于重做单题)
// 💡 优化：重做只删除"当前状态表(Record)"，"历史轨迹(History)"和"每日统计(Stats)"将永久保留
func (r *Repository) Delete(userID, questionID uint) error {
	return db.DB.Unscoped().
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&AnswerRecord{}).Error
}

// ResetCategory 物理删除某章节下的所有当前记录 (用于重做本章)
func (r *Repository) ResetCategory(userID uint, categoryPath string) error {
	var qIDs []uint
	err := db.DB.Table("questions").
		Where("category_path LIKE ?", categoryPath+"%").
		Pluck("id", &qIDs).Error
	
	if err != nil {
		return err
	}
	
	if len(qIDs) == 0 {
		return nil 
	}

	return db.DB.Unscoped().
		Where("user_id = ? AND question_id IN ?", userID, qIDs).
		Delete(&AnswerRecord{}).Error
}