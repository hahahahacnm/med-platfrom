import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'

// 引入 Naive UI 的字体
import 'vfonts/Lato.css' 

// 🔥🔥🔥【必须】引入验证码样式，否则弹窗会乱！🔥🔥🔥
import 'go-captcha-vue/dist/style.css' 
import './router/guard'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')