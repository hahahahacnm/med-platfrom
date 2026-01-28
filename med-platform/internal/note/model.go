package note

import (
	"med-platform/internal/question"
	"med-platform/internal/user"
	"time"
)

type Note struct {
	ID         uint              `gorm:"primarykey" json:"id"`
	UserID     uint              `gorm:"index;not null" json:"user_id"`
	QuestionID uint              `gorm:"index;not null" json:"question_id"`
	Content    string            `gorm:"type:text;not null" json:"content"`
	IsPublic   bool              `gorm:"default:false" json:"is_public"`
	ParentID   *uint             `gorm:"index" json:"parent_id"`
	
	LikeCount  int               `gorm:"default:0" json:"like_count"`
	IsLiked    bool              `gorm:"-" json:"is_liked"`
	
	// 🔥🔥🔥 新增：是否被当前用户收藏 🔥🔥🔥
	IsCollected bool             `gorm:"-" json:"is_collected"`

	User       user.User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Question   question.Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
	Parent     *Note             `gorm:"foreignKey:ParentID" json:"parent,omitempty"`

	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func (Note) TableName() string {
	return "notes"
}

// NoteLike 点赞记录表
type NoteLike struct {
	ID        uint      `gorm:"primaryKey"`
	
	// 🔥🔥🔥 核心修改：去掉了 gorm:"uniqueIndex:..." 🔥🔥🔥
	// 允许同一用户对同一笔记多次点赞（只要不是同一天）
	UserID    uint      `gorm:"index" json:"user_id"` 
	NoteID    uint      `gorm:"index" json:"note_id"`
	
	CreatedAt time.Time `json:"created_at"`
}

func (NoteLike) TableName() string {
	return "note_likes"
}

// 🔥🔥🔥 新增：笔记收藏表 (收藏别人的神评) 🔥🔥🔥
type NoteCollect struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_note_collect" json:"user_id"`
	NoteID    uint      `gorm:"uniqueIndex:idx_user_note_collect" json:"note_id"`
	CreatedAt time.Time `json:"created_at"`
}
func (NoteCollect) TableName() string { return "note_collects" }