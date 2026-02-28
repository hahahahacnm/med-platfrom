package product

import (
	"time"

	"gorm.io/gorm"
)

// 1. 商品定义 (🔥 已升维：增加封面、分类、标签、富文本详情)
type Product struct {
	gorm.Model
	Name        string `gorm:"unique;not null;size:100" json:"name"` 
	Description string `gorm:"size:255" json:"description"` // 简短描述（用于列表副标题）

	// 👇 新增的商业化展示字段
	CoverImg    string `gorm:"size:255" json:"cover_img"`   // 商品封面图 URL
	Category    string `gorm:"index;size:50" json:"category"` // 商品分类 (如: "vip", "question_bank", "course")
	Tags        string `gorm:"size:100" json:"tags"`        // 促销标签，逗号分隔 (如: "限时特惠,爆款,官方推荐")
	Detail      string `gorm:"type:text" json:"detail"`     // 富文本详情页 (HTML 或 Markdown)

	IsOnShelf   bool   `gorm:"default:true;index" json:"is_on_shelf"` // 是否上架
	
	// 关联的规格
	Skus        []ProductSku `gorm:"foreignKey:ProductID" json:"skus"`
}

// 2. 商品规格 (积分制) (🔥 优化：明确正数约束)
type ProductSku struct {
	gorm.Model
	ProductID    uint   `gorm:"index;not null" json:"product_id"`
	Name         string `gorm:"size:50;not null" json:"name"` // 规格名称 (如: "连续包月", "永久买断")
	
	// 🔥 安全约束：积分必须大于等于 0，防范负数零元购
	Points       int    `gorm:"not null;default:0;check:points >= 0" json:"points"` 
	DurationDays int    `gorm:"not null" json:"duration_days"` // 有效期天数 (-1 表示永久)
}

// 3. 商品内容绑定 (保持不变)
type ProductContent struct {
	gorm.Model
	ProductID uint   `gorm:"index;not null" json:"product_id"`
	Source    string `gorm:"index;not null" json:"source"`   
	Category  string `gorm:"index;not null" json:"category"` 
}

// 4. 用户持有记录 (凭证)
type UserProduct struct {
	gorm.Model
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	ProductID   uint      `gorm:"index;not null" json:"product_id"`
	
	Product     Product   `gorm:"foreignKey:ProductID" json:"product"`

	ProductName string    `gorm:"size:100" json:"product_name"` 
	ExpireAt    time.Time `gorm:"index" json:"expire_at"`       
}

// 5. 权限变更审计日志 (保持不变)
type ProductAuthLog struct {
	gorm.Model
	OperatorID     uint      `json:"operator_id"`
	OperatorName   string    `json:"operator_name"`
	TargetUserID   uint      `gorm:"index;not null" json:"target_user_id"`
	TargetUserName string    `json:"target_user_name"` 
	Action         string    `gorm:"size:20;not null" json:"action"` 
	ProductID      uint      `json:"product_id"`
	ProductName    string    `json:"product_name"` 
	DurationDays   int       `json:"duration_days"` 
	ExpireAt       time.Time `json:"expire_at"`     
	Memo           string    `json:"memo"`          
}

// 6. 积分兑换记录 (保持不变)
type ExchangeRecord struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	ProductID   uint   `json:"product_id"`
	SkuID       uint   `json:"sku_id"`
	ProductName string `json:"product_name"`
	SkuName     string `json:"sku_name"`
	PointsPaid  int    `json:"points_paid"` 
}