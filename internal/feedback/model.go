package feedback

import (
	"med-platform/internal/user"
	"time"

	"gorm.io/datatypes"
)

// PlatformFeedback 平台意见/Bug反馈表
type PlatformFeedback struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index" json:"user_id"`
	
	// 反馈类型：功能异常(Bug)、产品建议、账号问题、充值问题、其他
	Type       string         `gorm:"type:varchar(50)" json:"type"`
	
	// 反馈内容
	Content    string         `gorm:"type:text" json:"content"`
	
	// 🔥 图片凭证 (JSON数组存储多个URL: ["/uploads/1.jpg", "/uploads/2.jpg"])
	Images     datatypes.JSON `gorm:"type:json" json:"images"`
	
	// 联系方式 (选填，手机或邮箱)
	Contact    string         `gorm:"type:varchar(100)" json:"contact"`

	// 处理状态
	Status     int            `gorm:"default:0" json:"status"`      // 0:待处理, 1:处理中, 2:已解决, 3:已驳回
	AdminReply string         `gorm:"type:text" json:"admin_reply"` // 管理员回复
	
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`

	// 关联用户
	User       user.User      `gorm:"foreignKey:UserID" json:"user"`
}

func (PlatformFeedback) TableName() string {
	return "platform_feedbacks"
}