package payment

import (
	"errors"
	"med-platform/internal/common/config"
	"med-platform/internal/common/db"
	"med-platform/internal/payment/strategy"
	"med-platform/internal/product"
	"time"

	"gorm.io/gorm"
)

// GetPaymentStrategy 工厂方法
func GetPaymentStrategy() strategy.PaymentStrategy {
	driver := config.GlobalConfig.Payment.Driver
	if driver == "alipay" {
		return strategy.NewAlipayStrategy()
	}
	// 默认 Mock
	return strategy.NewMockStrategy()
}

// SettleOrder 结算发货逻辑
func SettleOrder(orderNo string, tradeNo string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查订单
		var order Order
		if err := tx.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}

		// 2. 幂等性检查
		if order.Status == "PAID" {
			return nil
		}

		// 3. 更新订单状态
		now := time.Now()
		order.Status = "PAID"
		order.TradeNo = tradeNo
		order.PayTime = &now
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// 4. 🔥🔥🔥 发货逻辑升级：支持时长叠加 🔥🔥🔥
		
		// A. 先查商品信息（为了获取商品名做快照）
		var prod product.Product
		if err := tx.First(&prod, order.ProductID).Error; err != nil {
			return err
		}

		// B. 检查用户是否已经持有该商品
		var existingUserProd product.UserProduct
		err := tx.Where("user_id = ? AND product_id = ?", order.UserID, order.ProductID).
			Order("expire_at desc"). // 如果有脏数据（多条），取过期时间最晚的那条
			First(&existingUserProd).Error

		// 假设商品时长固定为 1 年 (实际项目中应读取 prod.DurationDays)
		durationYears := 1 
		
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// === 情况一：新购入 (以前没买过) ===
			newUserProd := product.UserProduct{
				UserID:      order.UserID,
				ProductID:   order.ProductID,
				ProductName: prod.Name,
				ExpireAt:    time.Now().AddDate(durationYears, 0, 0),
			}
			if err := tx.Create(&newUserProd).Error; err != nil {
				return err
			}
		} else if err == nil {
			// === 情况二：续费 (以前买过) ===
			var newExpireAt time.Time
			
			if existingUserProd.ExpireAt.After(time.Now()) {
				// 2.1 还没过期：在“原过期时间”基础上顺延
				// 例如：原到期 2026-05-01，现在买，新到期 2027-05-01
				newExpireAt = existingUserProd.ExpireAt.AddDate(durationYears, 0, 0)
			} else {
				// 2.2 已经过期：从“现在”开始重新计算
				// 例如：原到期 2020-01-01，现在买，新到期 = 现在 + 1年
				newExpireAt = time.Now().AddDate(durationYears, 0, 0)
			}

			// 更新数据库
			if err := tx.Model(&existingUserProd).Update("expire_at", newExpireAt).Error; err != nil {
				return err
			}
		} else {
			// 数据库查询出错
			return err
		}

		return nil
	})
}