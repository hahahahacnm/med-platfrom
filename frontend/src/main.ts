import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'

// 引入 Naive UI 的字体（可选，如果你刚才装了的话）
import 'vfonts/Lato.css' 

const app = createApp(App)

app.use(createPinia()) // 启用 Store
app.use(router)        // 启用 路由

app.mount('#app')