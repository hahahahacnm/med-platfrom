import crypto from 'node:crypto'
import {
    SubmitOrderParams,
    ZhiFuFmApiStartOrderDTO,
    ZhiFuFmApiStartOrderVO,
    ZhuFuFmConfig,
    SubmitOrderDTO,
    ZhiFuFmApiStartOrderParams,
    ZhiFuFmNotifyDTO
} from './types'

/**
 * 支付FM
 */
export class ZhuFuFm {
    private config: ZhuFuFmConfig
    constructor(config: ZhuFuFmConfig) {
        this.config = config
    }

    /**
     * 请求接口
     * @param url 请求地址
     * @param body 请求参数
     * @returns 数据
     */
    private async fetch<T>(url: string, body: any) {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: new URLSearchParams(body).toString()
        })

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        const data: T = await response.json()
        console.log('fetch', body, data)
        return data
    }

    private md5(signStr: string) {
        const md5 = crypto.createHash('md5')
        return md5.update(signStr, 'utf8').digest('hex')
    }

    /**
     * 生成签名
     * @param {Record<string, any>} params 生成签名参数
     * @returns
     */
    private generateSign(params: Record<string, any>) {
        // 1. 过滤空值、sign和sign_type字段
        const paramsWithoutSign = Object.entries(params)
            .filter(
                ([k, value]) =>
                    value !== null && value !== undefined && value !== '' && k !== 'sign' && k !== 'sign_type'
            )
            .sort(([keyA], [keyB]) => keyA.localeCompare(keyB))

        // 2. 构建签名字符串
        const signStr = new URLSearchParams(paramsWithoutSign).toString() + this.config.merchantKey
        // paramsWithoutSign.map(([k, value]) => `${k}=${value}`).join('&') + this.config.key

        // 3. 生成MD5签名
        const md5 = crypto.createHash('md5')
        const sign = md5.update(signStr, 'utf8').digest('hex')

        return sign
    }

    /**
     * 验证回调签名
     * @param {Record<string, any>} params 验证回调签名参数
     * @returns {boolean} 验证回调签名结果
     */
    private validateNotifySignSubmitORder(params: Record<string, any>) {
        // 分离签名和其他参数
        const { sign, sign_type, ...paramsWithoutSign } = params
        // 生成签名并与回调中的签名比较
        return this.generateSign(paramsWithoutSign) === sign
    }

    /**
     * 构建订单参数
     * @param {SubmitOrderParams} params 构建订单参数
     * @returns {PayBuildSubmitOrderParamsVO} 构建订单参数结果
     */
    private buildSubmitOrderParams(params: SubmitOrderParams): SubmitOrderDTO {
        const {
            type,
            outTradeNo,
            name,
            money,
            clientip,
            device = 'pc',
            param,
            returnUrl = this.config.returnUrl,
            notifyUrl = this.config.notifyUrl
        } = params

        const paymentParams: Omit<SubmitOrderDTO, 'sign' | 'sign_type'> = {
            pid: this.config.merchantNum,
            notify_url: notifyUrl,
            return_url: returnUrl,
            out_trade_no: outTradeNo,
            type,
            name,
            money,
            clientip,
            device,
            param
        }

        // 生成签名
        const sign = this.generateSign(paymentParams)

        return {
            ...paymentParams,
            sign_type: 'MD5',
            sign
        }
    }

    /**
     * 提交订单, 使用mapi.php接口
     * @param {PaySubmitOrderDTO} params 提交订单参数
     * @returns {Promise<PaySubmitOrderVO>} 支付结果
     */
    private async submitOrder(params: SubmitOrderParams) {
        const url = this.config.baseUrl + '/mapi.php'

        // 构建订单参数
        const paymentParams = this.buildSubmitOrderParams(params)

        // 构建form-data格式的请求体
        const body = new URLSearchParams(paymentParams as any).toString()
        // for(const key in paymentParams){
        //   body.append(key, paymentParams[key])
        // }
        console.log('mapi', url, body)
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body
        })

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }

        const data: SubmitOrderDTO = await response.json()
        console.log('mapi', params, data)
        return data
    }

    async startOrder(params: ZhiFuFmApiStartOrderParams): Promise<ZhiFuFmApiStartOrderVO> {
        const url = this.config.baseUrl + '/startOrder'
        const { returnUrl = this.config.returnUrl, notifyUrl = this.config.notifyUrl } = params

        const dto: ZhiFuFmApiStartOrderDTO = {
            ...params,
            returnUrl,
            notifyUrl,
            merchantNum: this.config.merchantNum,
            // 待签名字符串=商户号+商户订单号+支付金额+异步通知地址+接入密钥；
            sign: this.md5(
                `${this.config.merchantNum}${params.orderNo}${params.amount}${notifyUrl}${this.config.merchantKey}`
            )
        }
        const data = await this.fetch<ZhiFuFmApiStartOrderVO>(url, dto)
        return data
    }

    /**
     * 验证回调签名
     * @param {ZhiFuFmNotifyDTO} params 验证回调签名参数
     * @returns {boolean} 验证回调签名结果
     */
    validateNotifySign(params: ZhiFuFmNotifyDTO) {
        // 验证商户ID
        if (params.merchantNum !== this.config.merchantNum) {
            return false // res.fail(400, '商户ID验证失败', params.merchantNum)
        }

        /** 签名。通过MD5加密指定的值拼接计算得出的签名值。
         * 签名值=md5(付款成功状态state的值+商户号merchantNum的值+商户订单号orderNo的值+订单金额amount的值+接入密钥)；
         * 其中“+”表示字符串拼接。请注意拼接顺序 */
        const sign = this.md5(
            `${params.state}${this.config.merchantNum}${params.orderNo}${params.amount}${this.config.merchantKey}`
        )

        // 生成签名并与回调中的签名比较
        return params.sign === sign
    }
}
