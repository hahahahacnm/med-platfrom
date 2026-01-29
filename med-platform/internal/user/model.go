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

	// 🔑 登录凭证 (系统唯一标识)
	Username string `gorm:"uniqueIndex;not null;type:varchar(50)" json:"username"`
	Password string `gorm:"not null" json:"-"` 
	
	// 🔥🔥🔥 [重构] 社交身份 🔥🔥🔥
	// 删除了 Name，只保留 Nickname。
	// 注册时如果未填 Nickname，可以默认等于 Username。
	Nickname  string `gorm:"type:varchar(50)" json:"nickname"`  
	Avatar    string `gorm:"type:varchar(255)" json:"avatar"`   
	Gender    int    `gorm:"default:0" json:"gender"`           
	
	// --- 🔐 安全/绑定 ---
	Email    string `gorm:"type:varchar(100)" json:"email"` 
	
	// --- 🎓 学籍信息 ---
	School    string `gorm:"type:varchar(100)" json:"school"`   
	Major     string `gorm:"type:varchar(100)" json:"major"`    
	Grade     string `gorm:"type:varchar(20)" json:"grade"`     
	
	// --- 💬 联系方式 ---
	QQ        string `gorm:"type:varchar(20)" json:"qq"`
	// 强制指定数据库列名为 wechat (全小写，不带下划线)
	WeChat string `gorm:"type:varchar(50);column:wechat" json:"wechat"`

	// --- 🛡️ 权限控制 ---
	Role     string     `gorm:"default:'user'" json:"role"` 
	Status   int        `gorm:"default:1" json:"status"`    
	BanUntil *time.Time `json:"ban_until"`                  

	// 关联持仓
	UserProducts []product.UserProduct `gorm:"foreignKey:UserID" json:"user_products"`
}

func (User) TableName() string {
	return "users"
}