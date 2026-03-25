import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import { router } from './router'
import './styles/main.css'
import { ElMessage } from 'element-plus'

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.config.errorHandler = (err, instance, info) => {
  console.error('Global error:', err)
  console.error('Component:', instance)
  console.error('Error info:', info)
  
  const errorMessage = err instanceof Error ? err.message : '未知错误'
  ElMessage.error({
    message: `系统错误: ${errorMessage}`,
    duration: 5000,
  })
}

app.config.warnHandler = (msg, instance, trace) => {
  console.warn('Warning:', msg)
  console.warn('Trace:', trace)
}

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason)
  ElMessage.error({
    message: '异步操作错误',
    duration: 5000,
  })
})

app.mount('#app')
