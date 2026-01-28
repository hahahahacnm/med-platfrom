import { ZhuFuFmPayType } from "./pay-type"

/**
 * 支付参数接口
 * @interface SubmitOrderParams
 */
export interface SubmitOrderParamsBase {
    /** 支付类型 */
    type: ZhuFuFmPayType
    /** 商品名称 */
    name: string
    /** 商品金额（单位：元，最大2位小数） */
    money: number
    /** 客户端IP */
    clientip: string
    /** 设备类型 */
    device?: string
    /** 自定义参数 */
    param?: string
}

/**
 * 支付参数接口
 * @interface SubmitOrderParams
 */
export interface SubmitOrderParams extends SubmitOrderParamsBase {
    /** 商户订单号 */
    outTradeNo: string
    /** 异步通知地址 */
    notifyUrl?: string
    /** 同步回调地址 */
    returnUrl?: string
}

/**
 * 支付参数接口
 * @interface SubmitOrderDTO
 */
export interface SubmitOrderDTO extends SubmitOrderParamsBase {
    /** 商户ID */
    pid: string
    /** 商户订单号 */
    out_trade_no: string
    /** 异步通知地址 */
    notify_url?: string
    /** 同步回调地址 */
    return_url?: string
    /** 签名类型 */
    sign_type: 'MD5'
    /** 签名 */
    sign: string
}



/**
 * 支付结果接口
 * @interface SubmitOrderVO
 */
export interface SubmitOrderVO {
    /** 返回状态码 */
    code: number
    /** 返回消息 */
    msg?: string
    /** 交易号 */
    trade_no?: string
    /** 支付URL */
    payurl?: string
    /** 二维码内容 */
    qrcode?: string
}
