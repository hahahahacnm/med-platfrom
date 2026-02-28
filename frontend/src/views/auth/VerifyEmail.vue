<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NResult, NButton, NSpin, NIcon } from 'naive-ui'
import { MailUnreadOutline } from '@vicons/ionicons5'
import request from '../../utils/request'

const route = useRoute()
const router = useRouter()

// 页面状态：loading | success | error
const status = ref<'loading' | 'success' | 'error'>('loading')
const errorMessage = ref('')
const verifyType = ref('')

onMounted(async () => {
  const token = route.query.token as string
  const type = route.query.type as string
  verifyType.value = type

  if (!token || !type) {
    status.value = 'error'
    errorMessage.value = '链接格式不正确，缺少关键参数'
    return
  }

  try {
    // 发起验证请求
    await request.get('/auth/verify-email', {
      params: { token, type }
    })
    
    // 延迟 500ms 让用户看清加载动画（优化体验）
    setTimeout(() => {
      status.value = 'success'
    }, 500)

  } catch (error: any) {
    status.value = 'error'
    // 捕获后端的具体报错信息（如：链接已过期、已被使用等）
    errorMessage.value = error.response?.data?.error || '网络异常或验证链接已失效'
  }
})

const handleFinish = () => {
  if (verifyType.value === 'register') {
    router.replace('/login') // 注册成功去登录
  } else {
    router.replace('/profile') // 换绑成功回个人中心
  }
}
</script>

<template>
  <div class="verify-container">
    <div class="verify-card">
      
      <div v-if="status === 'loading'" class="state-wrapper">
        <n-spin size="large" />
        <h2 class="title">正在安全核验...</h2>
        <p class="desc">请勿关闭页面，系统正在确认您的邮箱所有权</p>
      </div>

      <div v-else-if="status === 'success'" class="state-wrapper animate-in">
        <n-result
          status="success"
          title="验证成功"
          :description="verifyType === 'register' ? '您的账号已成功激活，欢迎加入题酷！' : '新邮箱换绑成功！'"
        >
          <template #footer>
            <n-button type="primary" size="large" @click="handleFinish" class="action-btn">
              {{ verifyType === 'register' ? '立即登录' : '返回个人中心' }}
            </n-button>
          </template>
        </n-result>
      </div>

      <div v-else class="state-wrapper animate-in">
        <n-result
          status="error"
          title="验证失败"
          :description="errorMessage"
        >
          <template #footer>
            <n-button secondary type="error" size="large" @click="router.replace('/')" class="action-btn">
              返回首页
            </n-button>
          </template>
        </n-result>
      </div>

    </div>

    <div class="bg-decoration">
      <n-icon size="400" color="rgba(59, 130, 246, 0.03)"><MailUnreadOutline /></n-icon>
    </div>
  </div>
</template>

<style scoped>
.verify-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f8fafc;
  position: relative;
  overflow: hidden;
}

.verify-card {
  width: 100%;
  max-width: 460px;
  background: white;
  padding: 40px;
  border-radius: 20px;
  box-shadow: 0 10px 40px -10px rgba(0, 0, 0, 0.08);
  position: relative;
  z-index: 10;
  text-align: center;
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.state-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
}

.title {
  margin-top: 24px;
  font-size: 20px;
  color: #1e293b;
  font-weight: 600;
}

.desc {
  margin-top: 8px;
  color: #64748b;
  font-size: 14px;
}

.action-btn {
  margin-top: 12px;
  width: 200px;
  font-weight: bold;
  border-radius: 8px;
}

.animate-in {
  animation: scaleUp 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes scaleUp {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.bg-decoration {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 1;
  pointer-events: none;
}
</style>