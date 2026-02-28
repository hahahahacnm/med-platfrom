package model

import (
	"time"
)

// NotificationSender 这是一个“影子结构体”
// 用于在不引入 internal/user 包的情况下访问用户信息，从而彻底打破循环引用
type NotificationSender struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// TableName 明确告诉 GORM 这个结构体对应的是 users 表
func (NotificationSender) TableName() string {
	return "users"
}

type Notification struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	UserID     uint      `gorm:"index" json:"user_id"`
	SenderID   uint      `gorm:"index" json:"sender_id"`
	
	// 🔥 关键修改：使用本地定义的影子结构体，不再 import user 包
	Sender     NotificationSender `gorm:"foreignKey:SenderID" json:"sender"` 
	
	SourceType string    `json:"source_type"` // "forum", "question"
	SourceID   uint      `json:"source_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	IsRead     bool      `gorm:"default:false" json:"is_read"`
}