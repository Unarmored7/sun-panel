import { createApp } from 'vue'
import App from './App.vue'
import { setupI18n } from './locales'
import { setupAssets, setupScrollbarStyle } from './plugins'
import { setupStore } from './store'
import { setupRouter } from './router'
import { setupSvgIcons } from './plugins/svg-icons'
import { prefetch } from './utils/request'

async function bootstrap() {
  const app = createApp(App)
  setupAssets()
  setupSvgIcons()

  setupScrollbarStyle()

  setupStore(app)

  setupI18n(app)

  // Fetch the home page's data while its code is still downloading.
  if (location.hash === '' || location.hash === '#/') {
    prefetch('/panel/itemIconGroup/getList', { includeItems: true })
    prefetch('/panel/userConfig/get')
    prefetch('/user/getAuthInfo')
  }

  await setupRouter(app)
  app.mount('#app')
}

bootstrap()
