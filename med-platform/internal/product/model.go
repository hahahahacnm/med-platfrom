package product

import (
	"time"
	"gorm.io/gorm"
)

// 1. 商品定义 (Product) - “壳” (SPU)
// 只定义商品是什么，不再定义多少钱
type Product struct {
	gorm.Model
	Name        string `gorm:"unique;not null" json:"name"` // 商品名 (例如：高考数学题库)
	Description string `json:"description"`                 // 描述
	IsOnShelf   bool   `gorm:"default:true" json:"is_on_shelf"` // 上架状态
	
	// ❌ 删除：Price float64 (价格已移至 SKU)
	
	// 🔥🔥🔥 新增：关联规格 (一个商品对应多个规格) 🔥🔥🔥
	// 例如：[月卡, 年卡, 永久卡]
	Skus []ProductSku `gorm:"foreignKey:ProductID" json:"skus"`
}

// 🔥🔥🔥 [新增] 2. 商品规格 (ProductSku) - “实际售卖单元” (SKU) 🔥🔥🔥
// 定义“怎么卖”：多少钱、多久
type ProductSku struct {
	gorm.Model
	ProductID    uint    `gorm:"index;not null" json:"product_id"`
	
	Name         string  `gorm:"size:50;not null" json:"name"` // 规格名 (例如：冲刺月卡 / 至尊永久版)
	Price        float64 `gorm:"type:decimal(10,2);not null" json:"price"` // 价格
	
	// 核心字段：有效期时长 (天)
	// 7 = 一周
	// 30 = 一月
	// 365 = 一年
	// -1 = 永久有效 (百年好合)
	DurationDays int     `gorm:"not null" json:"duration_days"` 
}

// 3. 商品内容绑定 (ProductContent) - “肉”
type ProductContent struct {
	gorm.Model
	ProductID uint   `gorm:"index;not null" json:"product_id"`
	Source    string `gorm:"index;not null" json:"source"`   // 题库源
	Category  string `gorm:"index;not null" json:"category"` // 科目名 (一级目录)
}

// 4. 用户持有记录 (UserProduct) - “凭证”
type UserProduct struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	
	// 快照
	ProductName string    `json:"product_name"` 

	// 最终计算出的过期时间点 (由 Sku.DurationDays + 当前时间/原过期时间 算出)
	ExpireAt  time.Time `json:"expire_at"` 
	
	// 禁止外键约束，允许 Product 物理删除后保留记录
	Product   Product   `gorm:"foreignKey:ProductID;constraint:-" json:"product,omitempty"`
}

// 5. 授权审计日志 (ProductAuthLog) - “黑匣子”
type ProductAuthLog struct {
	gorm.Model
	// 操作员信息
	OperatorID   uint   `gorm:"index;not null" json:"operator_id"`   
	OperatorName string `json:"operator_name"`                       

	// 被操作用户信息
	TargetUserID   uint   `gorm:"index;not null" json:"target_user_id"`
	TargetUserName string `json:"target_user_name"` 

	// 业务详情
	Action      string `gorm:"size:20;not null" json:"action"` 
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"` 
	
	// 这里可以记录当时的 SKU 信息作为快照
	DurationDays int       `json:"duration_days"` 
	ExpireAt     time.Time `json:"expire_at"`     
	Memo         string    `json:"memo"`          
}