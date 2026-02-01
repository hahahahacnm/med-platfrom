package user

import (
	"time"

	"gorm.io/gorm"
	"med-platform/internal/product" 
)

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
	
	// --- 联系方式 ---
	Email    string `gorm:"type:varchar(100)" json:"email"` 
	School    string `gorm:"type:varchar(100)" json:"school"`   
	Major     string `gorm:"type:varchar(100)" json:"major"`    
	Grade     string `gorm:"type:varchar(20)" json:"grade"`     
	QQ        string `gorm:"type:varchar(20)" json:"qq"`
	WeChat    string `gorm:"type:varchar(50);column:wechat" json:"wechat"`

	// --- 🛡️ 权限控制 ---
	Role     string     `gorm:"default:'user'" json:"role"` 
	Status   int        `gorm:"default:1" json:"status"`    
	BanUntil *time.Time `json:"ban_until"`                  

	// --- 🤝 代理体系 ---
	InvitationCode string `gorm:"uniqueIndex;size:20" json:"invitation_code"` 
	InvitedBy      uint   `gorm:"index" json:"invited_by"` // 上线代理ID

	// 代理自定义优惠配置 (0-20)
	AgentDiscountRate int `gorm:"default:0" json:"agent_discount_rate"`

	// 🔥🔥🔥 新增：固定收款码 (一次上传，长期有效) 🔥🔥🔥
	PaymentImage string `gorm:"type:varchar(255)" json:"payment_image"`

	// 关联持仓
	UserProducts []product.UserProduct `gorm:"foreignKey:UserID" json:"user_products"`
}

func (User) TableName() string {
	return "users"
}