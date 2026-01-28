package product

import (
	"time"
	"gorm.io/gorm"
)

// 1. 商品定义 (Product) - “壳”
type Product struct {
	gorm.Model
	Name        string `gorm:"unique;not null" json:"name"` // 商品名
	Description string `json:"description"`                 // 描述
	IsOnShelf   bool   `gorm:"default:true" json:"is_on_shelf"` // 上架状态
}

// 2. 商品内容绑定 (ProductContent) - “肉”
type ProductContent struct {
	gorm.Model
	ProductID uint   `gorm:"index;not null" json:"product_id"`
	Source    string `gorm:"index;not null" json:"source"`   // 题库源
	Category  string `gorm:"index;not null" json:"category"` // 科目名 (一级目录)
}

// 3. 用户持有记录 (UserProduct) - “凭证”
type UserProduct struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	
	// 快照
	ProductName string    `json:"product_name"` 

	ExpireAt  time.Time `json:"expire_at"` 
	
	// 🔥🔥🔥 核心修改：加上 constraint:- 🔥🔥🔥
	// 含义：禁止 GORM 在数据库层面创建外键约束。
	// 这样我们就可以物理删除 Product，而保留 UserProduct (即使它指向一个不存在的 ProductID 也没关系，因为我们有快照)
	Product   Product   `gorm:"foreignKey:ProductID;constraint:-" json:"product,omitempty"`
}

// 4. 授权审计日志 (ProductAuthLog) - “黑匣子”
type ProductAuthLog struct {
	gorm.Model
	// 操作员信息
	OperatorID   uint   `gorm:"index;not null" json:"operator_id"`   
	OperatorName string `json:"operator_name"`                       

	// 🔥🔥🔥 [新增] 被操作用户信息 🔥🔥🔥
	TargetUserID   uint   `gorm:"index;not null" json:"target_user_id"`
	TargetUserName string `json:"target_user_name"` // 新增：存客户用户名的快照

	// 业务详情
	Action      string `gorm:"size:20;not null" json:"action"` 
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"` 
	DurationDays int       `json:"duration_days"` 
	ExpireAt     time.Time `json:"expire_at"`     
	Memo         string    `json:"memo"`          
}