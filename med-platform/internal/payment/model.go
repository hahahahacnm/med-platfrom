package payment

import (
	"time"
	"gorm.io/gorm"
)

// Order 订单表
type Order struct {
	ID        uint           `gorm:"primaryKey"`
	OrderNo   string         `gorm:"type:varchar(64);uniqueIndex;not null"`
	TradeNo   string         `gorm:"type:varchar(64)"`
	
	UserID    uint           `gorm:"index;not null"`
	ProductID uint           `gorm:"index;not null"` // 冗余存一下商品ID，方便统计
	
	// 🔥🔥🔥 新增：记录买了哪个规格 (月卡/年卡) 🔥🔥🔥
	SkuID     uint           `gorm:"index;not null"` 
	
	Amount    float64        `gorm:"type:decimal(10,2);not null"`
	Status    string         `gorm:"type:varchar(20);default:'PENDING'"`
	PayTime   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}