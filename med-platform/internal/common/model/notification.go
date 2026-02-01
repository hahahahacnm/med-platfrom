package model

import (
	"time"

	"gorm.io/gorm"
	"med-platform/internal/user"
)

// Notification 全局通知表
type Notification struct {
	// 🔥🔥🔥 关键修改：加上 json:"..." 标签 🔥🔥🔥
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"index:idx_read_time" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 不返回给前端

	UserID    uint   `json:"user_id" gorm:"index"`
	SenderID  uint   `json:"sender_id"`
	
	SourceType string `json:"source_type"`
	SourceID   uint   `json:"source_id"`
	
	Content    string `json:"content"`
	Title      string `json:"title"`
	
	IsRead     bool   `json:"is_read" gorm:"default:false;index:idx_read_time"`

	Sender user.User `json:"sender" gorm:"foreignKey:SenderID"`
}