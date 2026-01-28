<script setup lang="ts">
import { NCard, NForm, NFormItem, NInput, NButton, useMessage, NIcon } from 'naive-ui'
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import { useUserStore } from '../stores/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()
const message = useMessage()

// 表单数据
const formModel = ref({
  username: '', 
  password: ''
})

const loading = ref(false)

const handleLogin = async () => {
  // 1. 基础校验
  if (!formModel.value.username || !formModel.value.password) {
    message.warning('请输入账号和密码')
    return
  }

  loading.value = true
  
  // 2. 调用 Store 的登录方法
  // (注意：请确保 userStore.login 内部已经执行了 localStorage.setItem('role', ...))
  const success = await userStore.login(formModel.value)
  
  loading.value = false

  if (success) {
    message.success('登录成功！')
    
    // 🔥🔥🔥 核心修改：根据角色跳转不同页面 🔥🔥🔥
    // 从本地缓存获取角色（由 Store 存入）
    const role = localStorage.getItem('role')

    if (role === 'admin' || role === 'agent') {
      // 如果是管理员或代理，跳到后台
      router.push('/admin')
    } else {
      // 普通用户，跳到前台刷题页
      router.push('/')
    }
  }
}
</script>

<template>
  <div class="login-container">
    <n-card title="医考刷题平台 · 登录" class="login-card" size="huge" :bordered="false">
      <n-form size="large">
        <n-form-item label="账号">
          <n-input v-model:value="formModel.username" placeholder="请输入用户名">
            <template #prefix>
              <n-icon :component="PersonOutline" />
            </template>
          </n-input>
        </n-form-item>
        
        <n-form-item label="密码">
          <n-input
            v-model:value="formModel.password"
            type="password"
            show-password-on="click"
            placeholder="请输入密码"
            @keydown.enter="handleLogin"
          >
            <template #prefix>
              <n-icon :component="LockClosedOutline" />
            </template>
          </n-input>
        </n-form-item>
        
        <div style="margin-top: 10px;">
          <n-button type="primary" block size="large" :loading="loading" @click="handleLogin">
            立即登录
          </n-button>
        </div>

        <div style="margin-top: 20px; text-align: center; font-size: 14px; color: #666;">
          还没有账号？ 
          <a class="register-link" @click="$router.push('/register')">
            去注册
          </a>
        </div>

      </n-form>
    </n-card>
  </div>
</template>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.login-card {
  width: 100%;
  max-width: 400px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.08);
}

.register-link {
  color: #18a058;
  cursor: pointer;
  font-weight: 500;
  transition: color 0.2s;
}

.register-link:hover {
  color: #36ad6a;
  text-decoration: underline;
}
</style>