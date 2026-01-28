/**
 * 支付通道枚举
 *
 * 每个枚举成员代表一种支付接入方式，字段含义统一为：
 * - channel: 三方或官方通道标识（与后端/配置保持一致）
 * - type: 接入类型，枚举值：'string' | 'direct' | 'all'
 * - signMode: 签约/免签，枚举值：'免签' | '直连支付宝官方' | '直连微信官方' | '富友间连支付宝' | '富友间连微信' | '银联签约' | '杉德签约支付宝微信' | '汇付支付' | '直连+间连支付宝签约' | '直连+间连微信签约'
 * - accountMode: 账号分配策略，枚举值：'收款号随机或轮训' | '商户号轮训'
 * - platforms: 支持的终端，枚举值：'PC' | 'WAP' | 'PC&WAP'
 * - note: 附加说明（如适用范围、限制、模式差异等）
 *
 * 表中特殊值说明：
 * - alipay-pc-all包含:alipaysign,alipay.direct.pc,alipay-facetoface, fuyou-aliqr,sandpayh5
 * - alipay-wap-all包含:alipaysign,alipay.direct.wap,alipay-facetoface, fuyou-aliqr,sandpayh5
 * - wechat-pc-all包含:wxpaynative,wxpayjsapi,fuyou-wxqr,sandpayh5
 * - wechat-wap-all包含:wxpayh5,wxpayjsapi,fuyou-wxqr,sandpayh5
 * - 主要用于签约类型的跨通道商户号的合理轮询分派收单使用，配置后启用状态的收款号就会参与轮训。例如：传入 alipay-pc-all调用接口时候，将轮训分派所有可以在电脑网站使用的商户号进行收单，。
 */
export enum ZhuFuFmPayType {
  /** 微信收款码 - 免签，收款号随机或轮训，支持 PC&WAP */
  WeChatQrCode = 'wechat',
  /** 支付宝收款码 - 免签，收款号随机或轮训，支持 PC&WAP（含上传模式与动态码模式） */
  AlipayQrCode = 'alipay',
  /** 云闪付收款码 - 免签，收款号随机或轮训，支持 PC&WAP */
  UnionPayQrCode = 'unipay',
  /** 三方聚合收款码 - 免签，收款号随机或轮训，支持 PC&WAP
   *  支持：付呗、收钱吧、钱到啦、京东收银哆啦宝、盛意旺、拉卡拉、商户数字钱宝、微商相册、银盛小Y、度小满、易生收款啦APP
   */
  AggregatedQrCode = 'qujie.qrcode',
  /** 农商行收银宝收款码 - 免签，收款号随机或轮训，支持 PC&WAP */
  JsNxQrCode = 'jsnx.qrcode',
  /** 网银APP收款 - 免签，收款号随机或轮训，支持 PC&WAP */
  BankApp = 'bankapp',

  /** 支付宝PC网站支付 - 直连支付宝官方，商户号轮训，支持 PC */
  AlipayPcDirect = 'alipay.direct.pc',
  /** 支付宝手机网站支付 - 直连支付宝官方，商户号轮训，支持 WAP */
  AlipayWapDirect = 'alipay.direct.wap',
  /** 支付宝当面付 - 直连支付宝官方，商户号轮训，支持 PC&WAP（跨地区收款不推荐） */
  AlipayFaceToFace = 'alipay-facetoface',
  /** 支付宝网站支付接口（自适应PC/手机） - 直连支付宝官方，商户号轮训，支持 PC&WAP */
  AlipaySite = 'alipaysign',

  /** 微信支付H5 - 直连微信官方签约，商户号轮训，支持 WAP */
  WeChatH5 = 'wxpayh5',
  /** 微信支付Native - 直连微信官方，商户号轮训，支持 PC */
  WeChatNative = 'wxpaynative',
  /** 微信支付JSAPI - 直连微信官方，商户号轮训，支持 PC&WAP */
  WeChatJsApi = 'wxpayjsapi',

  /** 富友支付宝服务窗 - 间连，商户号轮训，支持 PC&WAP */
  FuYouAliQr = 'fuyou-aliqr',
  /** 富友微信支付 - 间连，商户号轮训，支持 PC&WAP */
  FuYouWxQr = 'fuyou-wxqr',
  /** 富友银联扫码 - 间连，商户号轮训，支持 PC&WAP */
  FuYouBankQr = 'fuyou-bankqr',
  /** 富友收银台 - 间连，商户号轮训，支持 PC */
  FuYouPcQr = 'fuyou-pcqr',

  /** 杉德收银台 - 间连，商户号轮训，支持 PC&WAP */
  SandPayH5 = 'sandpayh5',
  /** 
   * 杉德银联扫码 - 间连，商户号轮训，支持 PC&WAP
   * 
   * （2025-10-10新增）
   */
  SandHmBankQr = 'sandhm-bankqr',
  /** 杉德支付宝 - 间连，商户号轮训，支持 PC */
  SandAlipay = 'sand-alipay',
  /** 杉德微信公众号 - 间连，商户号轮训，支持 PC&WAP */
  SandWxPay = 'sand-wxpay',

  /** 汇付快捷支付 - 间连，商户号轮训，支持 PC&WAP */
  HuiFuQuick = 'huifu-qkpay',

  /**
   * 支付宝PC签约轮训 - 直连+间连签约，商户号轮训，支持 PC
   * 包含:alipaysign,alipay.direct.pc,alipay-facetoface, fuyou-aliqr,sandpayh5
   */
  AlipayPcAll = 'alipay-pc-all',
  /**
   * 支付宝WAP签约轮训 - 直连+间连签约，商户号轮训，支持 WAP
   * 包含:alipaysign,alipay.direct.wap,alipay-facetoface, fuyou-aliqr,sandpayh5
   *
   * @deprecated 2025-10-17废弃，请使用 AlipayLoop
   */
  AlipayWapAll = 'alipay-wap-all',
  /**
   * 微信PC签约轮训 - 直连+间连签约，商户号轮训，支持 PC
   * 包含:wxpaynative,wxpayjsapi,fuyou-wxqr,sandpayh5
   *
   * @deprecated 2025-10-17废弃，请使用 WeChatLoop
   */
  WeChatPcAll = 'wechat-pc-all',
  /**
   * 微信WAP签约轮训 - 直连+间连签约，商户号轮训，支持 WAP
   * 包含:wxpayh5,wxpayjsapi,fuyou-wxqr,sandpayh5
   *
   * @deprecated 2025-10-17废弃，请使用 WeChatLoop
   */
  WeChatWapAll = 'wechat-wap-all',

  /**
   * 支付宝轮循池 - 池中通道按权重分配，支持 PC&WAP
   *
   * （2025-10-17新增）
   */
  AlipayLoop = 'aloop',
  /**
   * 微信轮循池 - 池中通道按权重分配，支持 PC&WAP
   *
   * （2025-10-17新增）
   */
  WeChatLoop = 'tloop',
  /** 
   * 网银系轮循池 - 池中通道按权重分配，支持 PC&WAP
   * 
   * （2025-10-17新增）
   */
  BankLoop = 'bloop'
}
