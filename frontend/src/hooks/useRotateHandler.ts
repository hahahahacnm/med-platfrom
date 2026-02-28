import { reactive, onMounted } from 'vue'
import request from '../utils/request'
import { useMessage } from 'naive-ui'

export const useHandler = (domRef: any, submitCallback: Function) => {
  const message = useMessage()
  
  const cData = reactive({
    image: "",    
    thumb: "",    
    thumbSize: 0, 
    captKey: ""   
  })

  const requestCaptchaData = () => {
    // 保护性清理
    if (domRef.value && typeof domRef.value.clear === 'function') {
      domRef.value.clear()
    }

    request.get('/auth/captcha').then((res: any) => {
      // 兼容 JSON 格式 { code: 0, data: { background: ... } }
      // request 拦截器可能已经脱壳，也可能没脱，这里做个兼容
      const d = res.data || res
      
      if (d) {
        // 优先读取 background，其次读取 image_base64 (兼容不同后端字段)
        cData.image = d.background || d.image_base64 || ""
        cData.thumb = d.block || d.thumb_base64 || ""
        cData.captKey = d.captcha_id || d.captcha_key || ""
        
        // 默认值防止报错
        cData.thumbSize = d.thumb_size || 0 
        
        // Base64 前缀补全 (防抖)
        if (cData.image && !cData.image.startsWith('data:')) {
           cData.image = 'data:image/png;base64,' + cData.image
        }
        if (cData.thumb && !cData.thumb.startsWith('data:')) {
           cData.thumb = 'data:image/png;base64,' + cData.thumb
        }
      }
    }).catch((e) => {
      console.warn('验证码获取失败:', e)
    })
  }

  const refreshEvent = () => {
    requestCaptchaData()
  }

  const confirmEvent = (angle: number, clear: Function) => {
    submitCallback({
      key: cData.captKey,
      angle: String(angle)
    }).then(() => {
        // 成功，由组件外部 (Login/Register) 处理跳转逻辑
    }).catch((err: any) => {
        // 🔥🔥🔥 核心修改：删除了 message.error('验证失败...') 🔥🔥🔥
        // 原因：具体的错误信息（如“邮箱已注册”、“验证码错误”）已由 request.ts 全局拦截器弹出。
        // 这里只负责重置验证码状态，避免误导用户。

        // 1. 重置前端组件状态 (变红/归位)
        if (typeof clear === 'function') clear()
        
        // 2. 失败后必须刷新验证码 (延迟一点体验更好)
        setTimeout(() => {
            requestCaptchaData()
        }, 500)
    })
  }

  const closeEvent = () => { }

  // 这里的 onMounted 留空，由外部控制何时加载，避免页面一进来就请求
  onMounted(() => { })

  return {
    data: cData,
    requestCaptchaData,
    closeEvent,
    refreshEvent,
    confirmEvent,
  }
}