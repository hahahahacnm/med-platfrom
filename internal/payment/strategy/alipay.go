package strategy

import (
	"fmt"
	"net/http" // 🔥 新增：HandleNotify 需要用到这个包

	"med-platform/internal/common/config"

	"github.com/smartwalle/alipay/v3"
)

type AlipayStrategy struct {
	client *alipay.Client
}

func NewAlipayStrategy() *AlipayStrategy {
	cfg := config.GlobalConfig.Payment.Alipay
	// 这里等你有证书了，填真的，现在先留空或者注释掉防止报错
	// 注意：如果没有配置 APPID，这里可能会报错，但在 Mock 模式下我们暂时忽略它
	client, _ := alipay.New(cfg.AppID, cfg.PrivateKey, false)
	// client.LoadAliPayPublicKey(cfg.PublicKey)

	return &AlipayStrategy{client: client}
}

// Pay 发起支付
func (s *AlipayStrategy) Pay(orderID string, amount float64, subject string) (string, error) {
	cfg := config.GlobalConfig.Payment.Alipay

	p := alipay.TradePagePay{}
	p.NotifyURL = cfg.NotifyURL
	p.ReturnURL = cfg.ReturnURL
	p.Subject = subject
	p.OutTradeNo = orderID
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, err := s.client.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

// HandleNotify 处理回调
// 🔥🔥🔥 核心修复：必须补上这个方法，否则接口检查不通过 🔥🔥🔥
func (s *AlipayStrategy) HandleNotify(req *http.Request) (string, string, bool, error) {
	// 暂时返回错误，因为还没有配置证书，无法进行真实的验签
	return "", "", false, fmt.Errorf("支付宝回调逻辑暂未配置证书，请先申请执照并配置 configs/config.yaml")
}