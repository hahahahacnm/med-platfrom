package product

import (
	"time"
	"gorm.io/gorm"
)

// 1. 商品定义
type Product struct {
	gorm.Model
	Name        string `gorm:"unique;not null" json:"name"` 
	Description string `json:"description"`                 
	IsOnShelf   bool   `gorm:"default:true" json:"is_on_shelf"` 
	Skus        []ProductSku `gorm:"foreignKey:ProductID" json:"skus"`
}

// 2. 商品规格
type ProductSku struct {
	gorm.Model
	ProductID    uint    `gorm:"index;not null" json:"product_id"`
	Name         string  `gorm:"size:50;not null" json:"name"` 
	Price        float64 `gorm:"type:decimal(10,2);not null" json:"price"` 
	DurationDays int     `gorm:"not null" json:"duration_days"` 
}

// 3. 商品内容绑定
type ProductContent struct {
	gorm.Model
	ProductID uint   `gorm:"index;not null" json:"product_id"`
	Source    string `gorm:"index;not null" json:"source"`   
	Category  string `gorm:"index;not null" json:"category"` 
}

// 4. 用户持有记录
type UserProduct struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	ProductName string    `json:"product_name"` 
	ExpireAt    time.Time `json:"expire_at"` 
	Product     Product   `gorm:"foreignKey:ProductID;constraint:-" json:"product,omitempty"`
}

// 5. 授权审计日志
type ProductAuthLog struct {
	gorm.Model
	OperatorID   uint   `gorm:"index;not null" json:"operator_id"`   
	OperatorName string `json:"operator_name"`                       
	TargetUserID   uint   `gorm:"index;not null" json:"target_user_id"`
	TargetUserName string `json:"target_user_name"` 
	Action      string `gorm:"size:20;not null" json:"action"` 
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"` 
	DurationDays int       `json:"duration_days"` 
	ExpireAt     time.Time `json:"expire_at"`     
	Memo         string    `json:"memo"`          
}

// 6. 销售/推广记录表 (账本)
type SalesRecord struct {
	gorm.Model
	AgentID     uint    `json:"agent_id" gorm:"index"` 
	UserID      uint    `json:"user_id"`               
	OrderID     string  `json:"order_id" gorm:"index"` 
	
	OriginalPrice  float64 `json:"original_price"`  // 原价
	DiscountAmount float64 `json:"discount_amount"` // 优惠
	FinalAmount    float64 `json:"final_amount"`    // 实付
	
	AgentProfit    float64 `json:"agent_profit"`    // 代理利润
	
	Description string  `json:"description"`
	
	// 🏦 提现状态: 0=未提现(Available), 1=审核中(Frozen), 2=已提现(Paid)
	WithdrawStatus int `gorm:"default:0;index" json:"withdraw_status"` 
}

// 7. 提现申请表
type WithdrawRequest struct {
	gorm.Model
	AgentID        uint    `gorm:"index;not null" json:"agent_id"`
	// 为了方便前端显示，我们存一下代理名字
	AgentName      string  `json:"agent_name"` 
	
	Amount         float64 `json:"amount"`          // 提现金额
	PaymentImage   string  `json:"payment_image"`   // 收款码图片地址
	
	// 状态: PENDING(待审核), APPROVED(已打款), REJECTED(已驳回)
	Status         string  `gorm:"default:'PENDING';index" json:"status"` 
	AdminComment   string  `json:"admin_comment"`   // 管理员备注
}