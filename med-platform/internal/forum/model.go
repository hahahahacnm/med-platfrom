package forum

import (
	"med-platform/internal/user"
	"time"

	"gorm.io/gorm"
)

// ForumBoard 板块
type ForumBoard struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `json:"name" gorm:"type:varchar(100);not null"`
	Description string `json:"description" gorm:"type:varchar(255)"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
	
	IsLocked    bool   `json:"is_locked" gorm:"default:false"`
	MinRole     string `json:"min_role" gorm:"default:'user'"`
}

// ForumPost 帖子
type ForumPost struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BoardID uint       `json:"board_id"`
	Board   ForumBoard `json:"board" gorm:"foreignKey:BoardID"`

	AuthorID uint      `json:"author_id"`
	Author   user.User `json:"author" gorm:"foreignKey:AuthorID"`

	Title    string `json:"title" gorm:"type:varchar(255);not null"`
	Summary  string `json:"summary" gorm:"type:varchar(500)"`
	Content  string `json:"content" gorm:"type:text"`
	
	ViewCount    int `json:"view_count" gorm:"default:0"`
	CommentCount int `json:"comment_count" gorm:"default:0"`
	
	IsPinned bool `json:"is_pinned" gorm:"default:false"`
	IsGlobal bool `json:"is_global" gorm:"default:false"`
}

// ForumComment 评论 (支持无限层级)
type ForumComment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	// 注意：硬删除不需要 DeletedAt，但为了保持 GORM 习惯可以留着，我们在 Handler 里用 Unscoped 忽略它
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PostID   uint      `json:"post_id"`
	
	AuthorID uint      `json:"author_id"`
	Author   user.User `json:"author" gorm:"foreignKey:AuthorID"`
	
	// 🔥 新增：父评论 ID
	ParentID *uint `json:"parent_id" gorm:"index"` 
	// 关联子评论 (仅用于预加载，不做数据库列)
	Children []ForumComment `json:"children" gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Content  string    `json:"content" gorm:"type:text"`
}

// ForumReport 举报
type ForumReport struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	TargetID   uint      `json:"target_id"`
	TargetType string    `json:"target_type"`
	Reason     string    `json:"reason"`
	
	ReporterID uint      `json:"reporter_id"`
	Reporter   user.User `json:"reporter" gorm:"foreignKey:ReporterID"`
	
	Status     int       `json:"status" gorm:"default:0"`
}