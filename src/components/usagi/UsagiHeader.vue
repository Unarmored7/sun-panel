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

const crossLinks = [
  { title: '导航', href: 'https://nav.usaginest.com', icon: 'lucide:compass', active: true },
  { title: '博客', href: 'https://blog.usaginest.com', icon: 'lucide:pen' },
  { title: '网盘', href: 'https://files.usaginest.com', icon: 'lucide:folder-open' },
  { title: '论坛', href: 'https://bbs.usaginest.com', icon: 'lucide:messages' },
]

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
  if (!props.logoText || props.logoText === 'Sun-Panel' || props.logoText === '乌萨奇导航')
    return { cn: '乌萨奇导航', en: 'Usagi Nav' }
  return { cn: props.logoText, en: 'Usagi Nav' }
})
</script>

<template>
  <header class="usagi-global-header" role="banner">
    <div class="header-inner">
      <!-- 左侧：品牌 Logo 与站名 -->
      <div class="header-left">
        <a href="/" class="header-brand" aria-label="乌萨奇导航首页">
          <span class="brand-cn">{{ brandTitle.cn }}</span>
          <span class="brand-dot">·</span>
          <span class="brand-en">{{ brandTitle.en }}</span>
        </a>
      </div>

      <!-- 中间：跨站核心导航 (与网盘/论坛/博客统一互通) -->
      <nav class="header-center" aria-label="全站导航">
        <a
          v-for="link in crossLinks"
          :key="link.title"
          :href="link.href"
          class="header-nav-link"
          :class="{ 'is-active': link.active }"
          :target="link.active ? undefined : '_blank'"
          rel="noopener noreferrer"
        >
          <SvgIcon :icon="link.icon" class="link-icon" />
          <span>{{ link.title }}</span>
        </a>
      </nav>

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
          <span class="btn-text">{{ t('panelHome.goToLogin') }}</span>
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
  color: #FFF2D5;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
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

/* 品牌 Logo 标题 */
.header-left {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.header-brand {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #FFFFFF !important;
  text-decoration: none !important;
  font-size: 16px;
  font-weight: 700;
  line-height: 1;
  letter-spacing: -0.01em;
  transition: opacity 0.2s ease;
  user-select: none;
}

.header-brand:hover {
  opacity: 0.92;
}

.brand-dot {
  color: var(--usagi-pink-primary, #FFB7CE);
  font-weight: 700;
  padding: 0 1px;
}

.brand-en {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 242, 213, 0.9);
}

/* 中间导航链接 */
.header-center {
  display: none;
  align-items: center;
  gap: 6px;
}

@media (min-width: 768px) {
  .header-center { display: flex; }
}

.header-nav-link {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 6px;
  color: #FFF2D5 !important;
  text-decoration: none !important;
  font-size: 13px;
  font-weight: 500;
  line-height: 1.4;
  transition: all 0.18s ease;
}

.header-nav-link .link-icon {
  font-size: 13px;
  color: var(--usagi-pink-primary, #FFB7CE);
  transition: transform 0.18s ease;
}

.header-nav-link:hover {
  color: #FFFFFF !important;
  background-color: rgba(255, 181, 200, 0.18);
}

.header-nav-link:hover .link-icon {
  transform: scale(1.1);
  color: var(--usagi-pink-hover, #F28BB6);
}

.header-nav-link.is-active {
  background-color: rgba(255, 255, 255, 0.20);
  color: #FFFFFF !important;
  font-weight: 600;
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
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  color: #FFF2D5;
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
  color: var(--usagi-brown-mute, rgba(255, 242, 213, 0.82));
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
  color: #FFF2D5;
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}

.header-btn .btn-icon {
  font-size: 15px;
  color: var(--usagi-pink-primary, #FFB7CE);
  transition: transform 0.18s ease;
}

.header-btn:hover {
  background: rgba(255, 255, 255, 0.26);
  border-color: rgba(255, 255, 255, 0.38);
  color: #FFFFFF;
}

.header-btn:hover .btn-icon {
  transform: scale(1.1);
  color: var(--usagi-pink-hover, #F28BB6);
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
  .brand-en,
  .brand-dot,
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
