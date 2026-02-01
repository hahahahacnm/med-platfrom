package payment

import (
	"errors"
	"med-platform/internal/common/config"
	"med-platform/internal/common/db"
	"med-platform/internal/payment/strategy"
	"med-platform/internal/product"
	"med-platform/internal/user" // 🔥 引入 user 包以查询邀请关系
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

// SettleOrder 结算发货逻辑 (带代理分润记录)
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
		
		// A. 先查商品信息
		var prod product.Product
		if err := tx.First(&prod, order.ProductID).Error; err != nil {
			return err
		}

		// B. 查 SKU 信息 (为了获取准确时长 DurationDays)
		var sku product.ProductSku
		// 兼容旧数据: 如果 order.SkuID > 0 才查，否则可能 panic 或查不到
		if order.SkuID > 0 {
			tx.First(&sku, order.SkuID)
		}
		
		// 默认时长：如果 SKU 没查到 (比如旧代码生成的订单)，默认给 30 天防崩
		durationDays := 30 
		if sku.ID > 0 {
			durationDays = sku.DurationDays
		}

		// C. 检查用户是否已经持有该商品
		var existingUserProd product.UserProduct
		err := tx.Where("user_id = ? AND product_id = ?", order.UserID, order.ProductID).
			Order("expire_at desc").
			First(&existingUserProd).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// === 情况一：新购入 ===
			var newExpire time.Time
			if durationDays == -1 {
				newExpire = time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
			} else {
				newExpire = time.Now().AddDate(0, 0, durationDays)
			}

			newUserProd := product.UserProduct{
				UserID:      order.UserID,
				ProductID:   order.ProductID,
				ProductName: prod.Name,
				ExpireAt:    newExpire,
			}
			if err := tx.Create(&newUserProd).Error; err != nil {
				return err
			}
		} else if err == nil {
			// === 情况二：续费 ===
			var newExpireAt time.Time
			
			// 如果买的是永久版 (-1)，直接设为永久
			if durationDays == -1 {
				newExpireAt = time.Date(2099, 12, 31, 23, 59, 59, 0, time.Local)
			} else {
				// 普通续费
				if existingUserProd.ExpireAt.After(time.Now()) {
					// 还没过期：顺延
					newExpireAt = existingUserProd.ExpireAt.AddDate(0, 0, durationDays)
				} else {
					// 已过期：重新计算
					newExpireAt = time.Now().AddDate(0, 0, durationDays)
				}
			}

			// 更新数据库
			if err := tx.Model(&existingUserProd).Update("expire_at", newExpireAt).Error; err != nil {
				return err
			}
		} else {
			return err
		}

		// 5. 🔥🔥🔥 [核心] 代理业绩精准记账 🔥🔥🔥
		
		// A. 查询买家是否有上线 (InvitedBy)
		var buyer user.User
		// 注意：这里用 Select 优化查询性能
		if err := tx.Select("invited_by").First(&buyer, order.UserID).Error; err == nil {
			if buyer.InvitedBy > 0 {
				// B. 计算利润 (根据你的公式)
				// 平台总释放利润 = 原价 * 20%
				// 代理实际到手 = 平台总释放利润 - 已给用户的优惠
				
				baseCommission := order.OriginalAmount * 0.20 // 总盘子 20%
				agentProfit := baseCommission - order.DiscountAmount
				
				// 兜底：如果算出来是负数（比如代理设置了超过20%的优惠，虽然前端限制了，但后端要防住），则利润为0
				if agentProfit < 0 {
					agentProfit = 0
				}

				// C. 创建销售记录
				record := product.SalesRecord{
					AgentID:        buyer.InvitedBy,
					UserID:         order.UserID,
					OrderID:        order.OrderNo,
					
					OriginalPrice:  order.OriginalAmount,
					DiscountAmount: order.DiscountAmount,
					FinalAmount:    order.Amount,
					
					AgentProfit:    agentProfit, // 🔥 这是代理真正赚到的钱
					
					Description:    "用户自助购买-" + sku.Name,
					WithdrawStatus: 0, // 未提现
				}
				
				if err := tx.Create(&record).Error; err != nil {
					// 记账失败不应阻断发货，记录日志即可
					// fmt.Println("记账失败:", err)
					return err 
				}
			}
		}

		return nil
	})
}