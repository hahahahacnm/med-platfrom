import { ZhuFuFmPayType } from "./pay-type"

/**
 * 支付回调参数接口
 * 顾客支付完成后（订单状态为“已支付”），支付FM把相关支付结果和相关信息发送给商户的业务系统，商户系统需要接收处理该消息，并返回应答。 如果支付FM收到商户的业务系统应答超时或者不是返回success，支付FM认为通知异常，支付FM会通过一定的策略定期重新发起通知，尽可能提高通知的成功率，但不保证通知最终能成功。（首次实时通知异常后重发通知频率为15s/30s/3m/10m/20m/30m/60m/3h/6h/ - 总计 11h3m45s）
 * 
 * @example
 * {
    payee: '微信',
    amount: '1.2',
    orderNo: '20251019215054032523',
    actualPayAmount: '1.2',
    payTime: '2025-10-19 22:14:26',
    platformOrderNo: '574081688993415168',
    merchantNum: '562134458716012544',
    sign: '078a23438365b004d33e4581288cd04b',
    trade_type: 'fuyou-wxqr',
    state: '1',
    type: 'fuyou-wxqr'
  }
 */
export interface ZhiFuFmNotifyDTO {
    /** 商户号。用户中心查看 */
    merchantNum: string
    /** 商户订单号。原样传回 */
    orderNo: string
    /** 支付方式 */
    type: ZhuFuFmPayType
    /** 订单金额。请求的支付金额（单位：元），最多小数点后保留2位 */
    amount: string
    /** 平台订单号。平台生成的唯一订单号 */
    platformOrderNo: string
    /** 实际支付金额。最多保留小数点后2位。免签类型因浮动原因，此金额可能会不等于订单金额 */
    actualPayAmount: string
    /** 付款成功标志。'1'：付款成功 */
    state: string
    /** 订单分派的收款号标识。免签类型为“手机编号”设置的值，签约类型为“账号标识”设置的值 */
    payee: string
    /** 付款时间。日期时间格式：yyyy-MM-dd HH:mm:ss */
    payTime: string
    /** 附加信息。原样传回 */
    attch?: string
    /** 签名。通过MD5加密指定的值拼接计算得出的签名值。签名值=md5(付款成功状态state的值+商户号merchantNum的值+商户订单号orderNo的值+订单金额amount的值+接入密钥)；其中“+”表示字符串拼接。请注意拼接顺序 */
    sign: string
}
