package answer

import (
	"time"

	"med-platform/internal/question"

	"gorm.io/gorm"
)

// AnswerRecord 用户作答流水表
type AnswerRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index" json:"user_id"`
	QuestionID uint          `gorm:"index" json:"question_id"`
	Choice    string         `json:"choice"`
	IsCorrect bool           `json:"is_correct"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 🔥 新增：加上关联，方便后续统计时通过作答记录反查题目信息（如分类）
	Question question.Question `gorm:"foreignKey:QuestionID" json:"-"`
}

func (AnswerRecord) TableName() string {
	return "answer_records"
}

// ---------------------------------------------------------

// UserMistake 错题本表
type UserMistake struct {
	ID uint `gorm:"primarykey" json:"id"`

	// 联合唯一索引
	UserID     uint `gorm:"index:idx_user_question,unique;not null" json:"user_id"`
	QuestionID uint `gorm:"index:idx_user_question,unique;not null" json:"question_id"`

	Choice string `json:"choice"`

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