package payment

import (
	"med-platform/internal/common/config"
	"med-platform/internal/common/db"
	"med-platform/internal/payment/strategy"
	"med-platform/internal/sysconfig" // 🔥 引入配置包
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetPaymentStrategy() strategy.PaymentStrategy {
	if config.GlobalConfig.Payment.Driver == "alipay" {
		return strategy.NewAlipayStrategy()
	}
	return strategy.NewMockStrategy()
}

// SettleOrder 结算逻辑：加积分 + 记账 + 代理分润
func SettleOrder(orderNo string, tradeNo string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查订单 (加锁)
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
			return err
		}

		if order.Status == "PAID" {
			return nil
		}

		// 2. 标记支付成功
		now := time.Now()
		order.Status = "PAID"
		order.TradeNo = tradeNo
		order.PayTime = &now
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		// 3. 给用户加积分
		if err := tx.Table("users").Where("id = ?", order.UserID).
			Update("points", gorm.Expr("points + ?", order.PointsAwarded)).Error; err != nil {
			return err
		}

		// 4. 🔥🔥🔥 代理分润 (已修复变量错误并接入强类型配置) 🔥🔥🔥
		type SimpleUser struct {
			ID        uint
			InvitedBy uint
		}
		var currentUser SimpleUser
		
		// 修正：将 orderToSettle 改为 order
		if err := tx.Table("users").Select("id, invited_by").Where("id = ?", order.UserID).Scan(&currentUser).Error; err == nil {
			if currentUser.InvitedBy > 0 {
				var agentID uint
				if err := tx.Table("users").Select("id").Where("id = ?", currentUser.InvitedBy).Scan(&agentID).Error; err == nil && agentID > 0 {
					
					// 🔥 使用 sysconfig 强类型函数，传入常量 Key 和 兜底值
					rate := sysconfig.GetFloat(sysconfig.KeyAgentRateDirect, 0.20)

					profit := order.Amount * rate // 修正：orderToSettle -> order

					if profit > 0 {
						commLog := CommissionLog{
							AgentID:        agentID,
							FromUserID:     order.UserID,
							OrderNo:        order.OrderNo,
							OrderAmount:    order.Amount,
							Profit:         profit,
							AppliedRate:    rate, // 📸 记录快照
							Description:    "下线用户在线支付分润",
							WithdrawStatus: 0,
						}
						tx.Create(&commLog)
					}
				}
			}
		}
		return nil
	})
}