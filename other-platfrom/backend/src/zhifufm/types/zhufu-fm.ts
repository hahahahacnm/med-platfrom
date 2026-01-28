/**
 * 支付FM配置接口
 */
export interface ZhuFuFmConfig {
    /** 接口根地址, 根地址为所有接口地址的相同部分，拼接对应的接口名称请求。 */
    baseUrl: string
    /** 商户号。在支付FM商户后台【用户中心】处可查看，该值不是用户名 */
    merchantNum: string
    /** 接入密钥。在支付FM商户后台【用户中心】处可查看 */
    merchantKey: string
    /** 支付结果通知网址（异步回调网址）。200字符以内，http(s)开头的网址 */
    notifyUrl: string
    /** 支付完成后展示网址（同步跳转地址）。200字符以内，http(s)开头的公网地址 */
    returnUrl: string

}
