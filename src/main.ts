import { createApp } from 'vue'
import App from './App.vue'
import { setupI18n } from './locales'
import { setupAssets, setupScrollbarStyle } from './plugins'
import { setupStore } from './store'
import { setupRouter } from './router'
import { setupSvgIcons } from './plugins/svg-icons'

async function bootstrap() {
  const app = createApp(App)
  setupAssets()
  setupSvgIcons()

  setupScrollbarStyle()

  setupStore(app)

  setupI18n(app)

  await setupRouter(app)
  app.mount('#app')
}

bootstrap()
