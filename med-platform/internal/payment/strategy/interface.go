package strategy

import "net/http"

// PaymentStrategy 定义所有支付方式必须遵守的标准
type PaymentStrategy interface {
    // Pay 发起支付
    // orderNo: 订单号
    // amount: 金额 (元)
    // subject: 商品标题
    // return: 支付跳转链接/参数, 错误
    Pay(orderNo string, amount float64, subject string) (string, error)

    // HandleNotify 处理回调
    // req: 原始 HTTP 请求 (用于验签)
    // return: (解析出的订单号, 支付宝/微信的交易号, 是否成功, 错误)
    HandleNotify(req *http.Request) (string, string, bool, error)
}