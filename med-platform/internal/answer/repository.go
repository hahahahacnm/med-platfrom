package answer

import (
	"time"

	"med-platform/internal/common/db"
	"med-platform/internal/question" // 👈 需要引入 question 包来使用 UserDailyStat

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

// CreateOrUpdate 保存或更新作答流水
// 🔥 核心逻辑：同时维护 "AnswerRecord"(状态) 和 "UserDailyStat"(计数)
func (r *Repository) CreateOrUpdate(record *AnswerRecord) error {
	// 使用事务，确保两个表同时成功或同时失败
	return db.DB.Transaction(func(tx *gorm.DB) error {
		
		// -------------------------------------------------------
		// 1. 处理 AnswerRecord (只保留最后一次状态)
		// -------------------------------------------------------
		var existing AnswerRecord
		// 查找是否已存在记录
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
			// 做过 -> 更新 (覆盖旧的选项、对错状态、更新时间)
			existing.Choice = record.Choice
			existing.IsCorrect = record.IsCorrect
			// GORM 的 Save 会自动更新 UpdatedAt 字段
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
		}

		// -------------------------------------------------------
		// 2. 🔥🔥🔥 关键缺失修复：维护每日统计表 🔥🔥🔥
		// -------------------------------------------------------
		// 逻辑：不管你是做新题，还是重做旧题，只要提交了，就算一次"练习量"
		// 这会让今日刷题数实时 +1
		
		today := time.Now().Format("2006-01-02")
		
		// 构造统计对象
		stat := question.UserDailyStat{
			UserID:  record.UserID,
			DateStr: today,
			Count:   1, // 基础增量
		}
		
		// 使用 Upsert (不存在则插入，存在则 Count + 1)
		// SQL: INSERT ... ON CONFLICT (user_id, date_str) DO UPDATE SET count = user_daily_stats.count + 1
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "date_str"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"count": gorm.Expr("user_daily_stats.count + 1"), // 👈 这一步让数据实时更新！
			}),
		}).Create(&stat).Error; err != nil {
			return err
		}

		return nil
	})
}

// Delete 物理删除单条作答记录 (用于重做单题)
// 逻辑：只删记录表，不扣减统计表（保留工作量）
func (r *Repository) Delete(userID, questionID uint) error {
	return db.DB.Unscoped().
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&AnswerRecord{}).Error
}

// ResetCategory 物理删除某章节下的所有记录 (用于重做本章)
// 逻辑：只删记录表，不扣减统计表（保留工作量）
func (r *Repository) ResetCategory(userID uint, categoryPath string) error {
	// 1. 先查出该章节下的所有题目 ID
	// 这里直接查 "questions" 表，避免引入 questionRepo 造成循环依赖
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

	// 2. 物理删除这些题目的作答记录
	return db.DB.Unscoped().
		Where("user_id = ? AND question_id IN ?", userID, qIDs).
		Delete(&AnswerRecord{}).Error
}