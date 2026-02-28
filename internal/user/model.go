package user

import (
	"time"

	"gorm.io/gorm"
	"med-platform/internal/product" 
)

// User 用户表
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 🔑 登录凭证
	Username string `gorm:"uniqueIndex;not null;type:varchar(50)" json:"username"`
	Password string `gorm:"not null" json:"-"` 
	
	// 🔥 社交身份
	Nickname  string `gorm:"type:varchar(50)" json:"nickname"`  
	Avatar    string `gorm:"type:varchar(255)" json:"avatar"`   
	Gender    int    `gorm:"default:0" json:"gender"`           
	
	// --- 💰 资产账户 ---
	Points int `gorm:"default:0;not null" json:"points"`

	// --- 联系方式 ---
	Email     string `gorm:"type:varchar(100);index" json:"email"` 
	School    string `gorm:"type:varchar(100)" json:"school"`   
	Major     string `gorm:"type:varchar(100)" json:"major"`    
	Grade     string `gorm:"type:varchar(20)" json:"grade"`     
	QQ        string `gorm:"type:varchar(20)" json:"qq"`
	WeChat    string `gorm:"type:varchar(50);column:wechat" json:"wechat"`

	// --- 🛡️ 权限控制 ---
	Role     string     `gorm:"default:'user'" json:"role"` 
	// 🔥 核心修改：0=待激活, 1=正常, 2=封禁
	Status   int        `gorm:"default:0" json:"status"`    
	BanUntil *time.Time `json:"ban_until"`                  

	// --- 🤝 代理体系 ---
	InvitationCode *string `gorm:"uniqueIndex;size:20" json:"invitation_code"` 
	InvitedBy      uint    `gorm:"index" json:"invited_by"` 
	AgentDiscountRate int  `gorm:"default:0" json:"agent_discount_rate"`
	PaymentImage   string  `gorm:"type:varchar(255)" json:"payment_image"`

	// 关联持仓
	UserProducts []product.UserProduct `gorm:"foreignKey:UserID" json:"user_products"`
}

func (User) TableName() string {
	return "users"
}

// ==========================================
// 🔥 新增：魔法链接验证令牌表
// ==========================================
type VerificationToken struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null"`
	Email     string    `gorm:"type:varchar(100);not null;index"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	Type      string    `gorm:"type:varchar(20);not null"` // "register" 或 "change_email"
	ExpiresAt time.Time `gorm:"not null"`
}