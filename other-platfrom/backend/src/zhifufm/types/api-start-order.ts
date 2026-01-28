import { ZhuFuFmPayType } from './pay-type'
import { ZhiFuFmApiResponse } from './response'

export enum ZhiFuFmApiStartOrderReturnType {
    /** 接口返回json数据 */
    Json = 'json',
    /** 接口直接重定向到支付页面，部分开发语言非页面form表单直接提交的有可能无法重定向，请通过json获取支付链接自行跳转。 */
    Page = 'page'
}

/**
 * StartOrder请求参数
 */
export interface ZhiFuFmApiStartOrderDTO {
    /** 商户号。在支付FM商户后台【用户中心】处可查看，该值不是用户名 */
    merchantNum: string
    /** 商户订单号。仅允许字母或纯数字，建议不超过32字符，不能有中文 */
    orderNo: string
    /** 订单金额。请求的支付金额（单位：元），最多小数点后保留2位 */
    amount: number
    /** 支付结果通知网址（异步回调网址）。200字符以内，http(s)开头的网址 */
    notifyUrl: string
    /** 支付完成后展示网址（同步跳转地址）。200字符以内，http(s)开头的公网地址 */
    returnUrl?: string
    /** 支付方式。请根据所需对接的支付方式正确传值 */
    payType: ZhuFuFmPayType
    /**
     * 签名
     * - 待签名字符串进行MD5加密得出的32位签名值，小写。
     * - 待签名字符串=商户号+商户订单号+支付金额+异步通知地址+接入密钥；
     * - 其中“+”表示字符串拼接,请注意拼接顺序。
     * - 接入密钥在支付FM商户后台【用户中心】处可查看。
     */
    sign: string
    /**
     * 接口内容返回类型
     * @default 'json'
     */
    returnType?: ZhiFuFmApiStartOrderReturnType
    /** 指定收款号。该值为在本平台设置的免签类型的手机编号或签约类型的账号标识 */
    payee?: string
    /** 附加信息，回调时候原样回传。空内容会被忽略不再回传 */
    attch?: string
    /** 商品标题。100字符以内，签约类型会原样传到支付平台 */
    subject?: string
    /** 商品描述。200字符以内，签约类型会原样传到支付平台 */
    body?: string
    /**
     * 订单支付有效期，单位：分钟；默认值5，最大值15
     * @default 5
     */
    payDuration?: number
    /** 支付结果通知notifyUrl请求方式。默认/不传值：GET请求方式回调通知；post_form：POST请求方式回调通知 */
    apiMode?: 'post_form'
}

/**
 * StartOrder方法参数
 */
export interface ZhiFuFmApiStartOrderParams
    extends Omit<ZhiFuFmApiStartOrderDTO, 'sign' | 'notifyUrl' | 'merchantNum'> {
    /** 支付结果通知网址（异步回调网址）。200字符以内，http(s)开头的网址 */
    notifyUrl?: string
}

/**
 * StartOrder请求响应
 */
export type ZhiFuFmApiStartOrderVO = ZhiFuFmApiResponse<{
    /** 订单ID */
    id: string
    /** 支付链接 */
    payUrl: string
}>
