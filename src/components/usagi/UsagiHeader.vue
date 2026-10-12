<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
// Direct imports: the component barrels would pull unrelated components into the home chunk.
import SvgIcon from '@/components/common/SvgIcon/index.vue'
import { VisitMode } from '@/enums/auth'
import { t } from '@/locales'

interface Props {
  logoText?: string
  clockShow?: boolean
  clockShowSecond?: boolean
  visitMode: VisitMode
}

const props = withDefaults(defineProps<Props>(), {
  logoText: '乌萨奇导航',
  clockShow: true,
  clockShowSecond: true,
})

const emit = defineEmits<{
  (e: 'openSettings'): void
  (e: 'goLogin'): void
}>()

// A compact clock sized for the bar; the large desk clock does not fit 54px.
const now = ref(new Date())
let clockTimer: ReturnType<typeof setInterval> | undefined
onMounted(() => { clockTimer = setInterval(() => { now.value = new Date() }, 1000) })
onUnmounted(() => clearInterval(clockTimer))
const pad = (value: number) => String(value).padStart(2, '0')
const clockTime = computed(() => {
  const d = now.value
  const time = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  return props.clockShowSecond ? `${time}:${pad(d.getSeconds())}` : time
})
const weekKeys = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat']
const clockDate = computed(() => `${now.value.getMonth() + 1}-${now.value.getDate()} ${t(`deskModule.clock.${weekKeys[now.value.getDay()]}`)}`)

const brandTitle = computed(() => {
  const text = (props.logoText || '').trim()
  if (!text || text === 'Sun-Panel' || text === '乌萨奇导航' || /usagi/i.test(text))
    return { cn: '乌萨奇导航', en: 'Usagi Nav' }
  return { cn: text, en: 'Usagi Nav' }
})
</script>

<template>
  <header class="usagi-global-header" role="banner">
    <div class="header-inner">
      <!-- 左侧：品牌 Logo 与站名 (与网盘/论坛/博客统一格式) -->
      <div class="header-left">
        <a href="/#/" class="header-brand" aria-label="乌萨奇导航首页">
          <span class="brand-cn">{{ brandTitle.cn }}</span>
          <span class="brand-dot">·</span>
          <span class="brand-en">{{ brandTitle.en }}</span>
        </a>
      </div>

      <!-- 右侧：系统功能控件 (微透白胶囊风格) -->
      <div class="header-right">
        <!-- 时钟微胶囊 -->
        <div v-if="clockShow" class="clock-capsule" aria-label="当前时间">
          <span class="clock-time">{{ clockTime }}</span>
          <span class="clock-date">{{ clockDate }}</span>
        </div>

        <!-- 管理中心入口 (已登录) -->
        <button
          v-if="visitMode === VisitMode.VISIT_MODE_LOGIN"
          type="button"
          class="header-btn manage-btn"
          title="应用管理与设置"
          aria-label="应用管理与设置"
          @click="emit('openSettings')"
        >
          <SvgIcon icon="majesticons-applications" class="btn-icon" />
          <span class="btn-text">管理</span>
        </button>

        <!-- 登录入口 (访客) -->
        <button
          v-if="visitMode === VisitMode.VISIT_MODE_PUBLIC"
          type="button"
          class="header-btn login-btn"
          aria-label="登录后台"
          @click="emit('goLogin')"
        >
          <SvgIcon icon="material-symbols:account-circle" class="btn-icon" />
          <span class="btn-text">{{ t('login.loginButton') }}</span>
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.usagi-global-header {
  position: sticky;
  top: 0;
  z-index: 50;
  width: 100%;
  height: 54px;
  background-color: var(--usagi-brown-footer, #8A7A6A);
  color: #FFFFFF;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  transition: background-color 0.25s ease, border-color 0.25s ease;
}

.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1360px;
  height: 100%;
  margin: 0 auto;
  padding: 0 16px;
  box-sizing: border-box;
}

/* 品牌 Logo 标题 (纯白高级质感，与网盘/论坛/博客完全一致) */
.header-left {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.header-brand {
  display: inline-flex;
  align-items: center;
  color: #FFFFFF !important;
  text-decoration: none !important;
  font-size: 18px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.3px;
  transition: opacity 0.2s ease;
  user-select: none;
}

.header-brand:hover {
  opacity: 0.88;
}

.brand-cn {
  color: #FFFFFF !important;
  font-size: 18px;
  font-weight: 700;
}

.brand-dot {
  color: #FFFFFF !important;
  opacity: 0.7;
  margin: 0 6px;
  font-weight: 700;
}

.brand-en {
  color: #FFFFFF !important;
  font-size: 18px;
  font-weight: 700;
}

/* 右侧系统控件 */
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.clock-capsule {
  display: inline-flex;
  align-items: baseline;
  gap: 8px;
  padding: 4px 10px;
  height: 30px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.22);
  color: #FFFFFF;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
  line-height: 20px;
}

.clock-time {
  font-weight: 600;
  color: #FFFFFF;
}

.clock-date {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.82);
}

/* 微透白胶囊按钮 */
.header-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
  height: 30px;
  padding: 0 10px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.22);
  background: rgba(255, 255, 255, 0.16);
  color: #FFFFFF;
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}

.header-btn .btn-icon {
  font-size: 15px;
  color: #FFFFFF;
  transition: transform 0.18s ease;
}

.header-btn:hover {
  background: rgba(255, 255, 255, 0.26);
  border-color: rgba(255, 255, 255, 0.38);
  color: #FFFFFF;
}

.header-btn:hover .btn-icon {
  transform: scale(1.08);
  color: #FFFFFF;
}

.header-btn:active {
  transform: scale(0.96);
}

/* 移动端适配 */
@media (max-width: 640px) {
  .usagi-global-header {
    height: 48px;
  }
  .header-inner {
    padding: 0 10px;
  }
  .header-brand {
    font-size: 15px;
  }
  .brand-cn,
  .brand-en {
    font-size: 15px;
  }
  .brand-dot {
    margin: 0 4px;
  }
  .clock-capsule,
  .manage-btn .btn-text {
    display: none;
  }
  .header-btn {
    height: 28px;
    padding: 0 8px;
    font-size: 12px;
  }
}
</style>
