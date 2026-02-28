package payment

import (
	"time"
	"gorm.io/gorm"
)

// Order 赞助订单表
type Order struct {
	ID        uint           `gorm:"primaryKey"`
	OrderNo   string         `gorm:"type:varchar(64);uniqueIndex;not null"`
	TradeNo   string         `gorm:"type:varchar(64)"`
	
	UserID    uint           `gorm:"index;not null"`
	Amount    float64        `gorm:"type:decimal(10,2);not null"`
	PointsAwarded int        `gorm:"not null;default:0"`

	Status    string         `gorm:"type:varchar(20);default:'PENDING'"` // PENDING, PAID
	PayTime   *time.Time
	
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// CommissionLog 代理佣金记录
type CommissionLog struct {
	gorm.Model
	AgentID     uint    `gorm:"index;not null" json:"agent_id"`
	FromUserID  uint    `json:"from_user_id"`
	OrderNo     string  `gorm:"index;not null" json:"order_no"`
	OrderAmount float64 `json:"order_amount"`
	Profit      float64 `json:"profit"`
	// 🔥 新增：记录执行分润时的实际比例，用于财务审计
	AppliedRate float64 `gorm:"type:decimal(5,4)" json:"applied_rate"` 
	Description string  `json:"description"`
	// 🏦 提现状态: 0=未提现, 1=审核中, 2=已提现, 3=已驳回
	WithdrawStatus int `gorm:"default:0;index" json:"withdraw_status"`
}

// WithdrawRequest 提现申请单
type WithdrawRequest struct {
	gorm.Model
	AgentID      uint      `gorm:"index;not null" json:"agent_id"`
	AgentName    string    `json:"agent_name"`
	Amount       float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	PaymentImage string    `gorm:"type:varchar(255)" json:"payment_image"` // 收款码快照
	Status       string    `gorm:"type:varchar(20);default:'PENDING';index" json:"status"` // PENDING, APPROVED, REJECTED
	
	AdminComment string    `json:"admin_comment"`
	HandledBy    uint      `json:"handled_by"`
	HandledAt    time.Time `json:"handled_at"`
}

// ==========================================
// 🔥 新增：激活码（卡密）表
// ==========================================
type ActivationCode struct {
	gorm.Model
	Code     string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"code"`
	Points   int        `gorm:"not null" json:"points"`          // 包含的积分额度
	Status   int        `gorm:"default:0;index" json:"status"`   // 0=未使用, 1=已使用
	UsedByID uint       `gorm:"index" json:"used_by_id"`         // 谁使用的
	UsedAt   *time.Time `json:"used_at"`                         // 使用时间
}