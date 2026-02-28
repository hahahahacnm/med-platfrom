package model

import (
	"time"
)

// ForumReportUser 影子结构体：用于打破与 internal/user 的循环引用
// 它只包含我们需要展示的基础用户信息
type ForumReportUser struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// TableName 明确映射到数据库中的 users 表
func (ForumReportUser) TableName() string {
	return "users"
}

// ForumReport 举报模型
type ForumReport struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`

	TargetID   uint      `json:"target_id"`
	TargetType string    `json:"target_type"` // "post" 或 "comment"
	Reason     string    `json:"reason"`
	
	ReporterID uint            `json:"reporter_id"`
	Reporter   ForumReportUser `gorm:"foreignKey:ReporterID" json:"reporter"`
	
	Status     int       `gorm:"default:0" json:"status"` // 0: 待处理, 1: 已处理
}