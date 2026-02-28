package answer

import (
	"time"

	"med-platform/internal/question"

	"gorm.io/gorm"
)

// AnswerRecord 用户作答流水表 (当前最新状态，决定答题卡颜色)
type AnswerRecord struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	UserID     uint           `gorm:"index" json:"user_id"`
	QuestionID uint           `gorm:"index" json:"question_id"`
	
	// 🔥 性能优化：冗余题目所属的分类ID，后续统计仪表盘时彻底告别 JOIN
	CategoryID uint           `gorm:"index;default:0" json:"category_id"` 
	
	Choice     string         `json:"choice"`
	IsCorrect  bool           `json:"is_correct"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Question question.Question `gorm:"foreignKey:QuestionID" json:"-"`
}

func (AnswerRecord) TableName() string {
	return "answer_records"
}

// ---------------------------------------------------------
// 🔥 新增：AnswerHistory 用户作答历史轨迹表 (Append-Only)
// 作用：无论用户重做多少次，每一次的选项都会被记录，用于生成学习曲线和遗忘曲线
// ---------------------------------------------------------
type AnswerHistory struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	QuestionID uint      `gorm:"index" json:"question_id"`
	Choice     string    `json:"choice"`
	IsCorrect  bool      `json:"is_correct"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"` // 核心查询依据
}

func (AnswerHistory) TableName() string {
	return "answer_histories"
}

// ---------------------------------------------------------

// UserMistake 错题本表
type UserMistake struct {
	ID uint `gorm:"primarykey" json:"id"`

	// 联合唯一索引
	UserID     uint `gorm:"index:idx_user_question,unique;not null" json:"user_id"`
	QuestionID uint `gorm:"index:idx_user_question,unique;not null" json:"question_id"`

	Choice     string `json:"choice"`

	// 🔥 进阶优化：错题次数统计
	WrongCount int    `gorm:"default:1" json:"wrong_count"`

	// GORM 会在 save/update 时自动更新这个时间，非常适合"错题重做浮到最前"的逻辑
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Question question.Question `gorm:"foreignKey:QuestionID" json:"question"`
}

func (UserMistake) TableName() string {
	return "user_mistakes"
}

// ---------------------------------------------------------

// UserFavorite 用户收藏表
type UserFavorite struct {
	ID uint `gorm:"primarykey" json:"id"`

	UserID     uint `gorm:"index:idx_user_fav_q,unique;not null" json:"user_id"`
	QuestionID uint `gorm:"index:idx_user_fav_q,unique;not null" json:"question_id"`

	CreatedAt time.Time `json:"created_at"`

	Question question.Question `gorm:"foreignKey:QuestionID" json:"question"`
}

func (UserFavorite) TableName() string {
	return "user_favorites"
}