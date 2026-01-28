/**
 * 通用响应接口
 */
interface ZhiFuFmApiResponseBase {
    /** 请求是否成功 */
    success: boolean
    /** 返回的消息 */
    msg: string
    /** 状态码 */
    code: number
    /** 时间戳 */
    timestamp: number
}
/** 请求成功 */
export interface ZhiFuFmApiResponseSuccess<T> extends ZhiFuFmApiResponseBase {
    /** 请求成功 */
    success: true
    /** 返回的数据 */
    data: T
}

/** 请求失败 */
export interface ZhiFuFmApiResponseFail extends ZhiFuFmApiResponseBase {
    /** 请求失败 */
    success: false
}

/** 请求响应 */
export type ZhiFuFmApiResponse<T> = ZhiFuFmApiResponseSuccess<T> | ZhiFuFmApiResponseFail
