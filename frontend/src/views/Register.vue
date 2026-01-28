<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, useMessage, NIcon } from 'naive-ui'
import { PersonOutline, LockClosedOutline, HappyOutline } from '@vicons/ionicons5'
import request from '../utils/request'

const router = useRouter()
const message = useMessage()
const formRef = ref(null)
const loading = ref(false)

// 表单数据
const model = reactive({
  username: '',
  nickname: '', // 昵称
  password: '',
  confirmPassword: ''
})

// 校验规则
const rules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' },
    { min: 3, message: '账号长度不能少于3位', trigger: 'blur' }
  ],
  nickname: [
    { required: true, message: '请输入昵称/姓名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入密码', trigger: ['input', 'blur'] },
    {
      validator: (rule: any, value: string) => {
        return value === model.password
      },
      message: '两次输入的密码不一致',
      trigger: ['input', 'blur']
    }
  ]
}

const handleRegister = (e: MouseEvent) => {
  e.preventDefault()
  formRef.value?.validate(async (errors: any) => {
    if (!errors) {
      loading.value = true
      try {
        // 调用后端注册接口
        await request.post('/auth/register', {
          username: model.username,
          password: model.password,
          nickname: model.nickname
        })
        
        message.success('注册成功！请登录')
        // 注册成功后跳转到登录页
        router.push('/login')
      } catch (error) {
        // 错误处理交给了 request 拦截器，这里只需关闭 loading
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<template>
  <div class="register-container">
    <div class="register-box">
      <div class="header">
        <div class="title">🏥 医考刷题平台</div>
        <div class="subtitle">创建新账号</div>
      </div>

      <n-card :bordered="false" size="large" style="box-shadow: 0 4px 16px rgba(0,0,0,0.08); border-radius: 12px;">
        <n-form ref="formRef" :model="model" :rules="rules" size="large">
          
          <n-form-item path="username" label="账号">
            <n-input v-model:value="model.username" placeholder="请输入用户名 (唯一标识)">
              <template #prefix><n-icon :component="PersonOutline" /></template>
            </n-input>
          </n-form-item>

          <n-form-item path="nickname" label="昵称">
            <n-input v-model:value="model.nickname" placeholder="怎么称呼您？">
              <template #prefix><n-icon :component="HappyOutline" /></template>
            </n-input>
          </n-form-item>

          <n-form-item path="password" label="密码">
            <n-input
              v-model:value="model.password"
              type="password"
              show-password-on="click"
              placeholder="请输入密码"
            >
              <template #prefix><n-icon :component="LockClosedOutline" /></template>
            </n-input>
          </n-form-item>

          <n-form-item path="confirmPassword" label="确认密码">
            <n-input
              v-model:value="model.confirmPassword"
              type="password"
              show-password-on="click"
              placeholder="请再次输入密码"
              @keydown.enter="handleRegister"
            >
              <template #prefix><n-icon :component="LockClosedOutline" /></template>
            </n-input>
          </n-form-item>

          <div style="margin-top: 10px;">
            <n-button type="primary" block size="large" :loading="loading" @click="handleRegister">
              立即注册
            </n-button>
          </div>

          <div class="footer-links">
            <span>已有账号？</span>
            <a class="login-link" @click="$router.push('/login')">去登录</a>
          </div>

        </n-form>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  /* 或者用这种医用绿色渐变背景 */
  /* background: linear-gradient(135deg, #e0f2f1 0%, #a7ffeb 100%); */
}

.register-box {
  width: 100%;
  max-width: 420px;
  padding: 20px;
}

.header {
  text-align: center;
  margin-bottom: 30px;
}

.title {
  font-size: 28px;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 8px;
}

.subtitle {
  font-size: 16px;
  color: #7f8c8d;
}

.footer-links {
  margin-top: 20px;
  text-align: center;
  font-size: 14px;
  color: #666;
}

.login-link {
  color: #18a058;
  cursor: pointer;
  font-weight: 500;
  margin-left: 5px;
}
.login-link:hover {
  text-decoration: underline;
}
</style>