package strategy

import (
	"fmt"
	"net/http"

	"med-platform/internal/common/config"
)

type MockStrategy struct{}

// NewMockStrategy 构造函数
func NewMockStrategy() *MockStrategy {
	return &MockStrategy{}
}

// Pay 模拟支付下单
// 规范点：从配置中读取 Domain，不再硬编码 localhost
func (s *MockStrategy) Pay(orderNo string, amount float64, subject string) (string, error) {
	// 1. 获取配置中的域名 (例如 http://localhost:8080 或 http://abc.cpolar.cn)
	domain := config.GlobalConfig.Payment.Domain

	// 2. 拼接回调地址
	// 注意：这里的路径 /api/v1/payment/mock/callback 必须与 router.go 中定义的一致
	mockURL := fmt.Sprintf("%s/api/v1/payment/mock/callback?out_trade_no=%s", domain, orderNo)

	return mockURL, nil
}

// HandleNotify 模拟支付回调处理
// 规范点：必须实现接口中定义的这个方法，否则无法赋值给 PaymentStrategy 类型
func (s *MockStrategy) HandleNotify(req *http.Request) (string, string, bool, error) {
	// Mock 模式下，我们简单粗暴地认为只要回调了就是成功
	orderNo := req.URL.Query().Get("out_trade_no")
	
	// 构造一个假的流水号
	tradeNo := fmt.Sprintf("MOCK_TRADE_%s", orderNo)

	// 返回: 订单号, 交易流水号, 是否成功(true), 错误(nil)
	return orderNo, tradeNo, true, nil
}