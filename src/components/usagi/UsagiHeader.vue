<script setup lang="ts">
import { computed } from 'vue'
import { NTooltip } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { Clock } from '@/components/deskModule'
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
      <nav class="header-center hidden md:flex" aria-label="全站导航">
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
        <div v-if="clockShow" class="clock-capsule hidden sm:inline-flex">
          <Clock :hide-second="!clockShowSecond" />
        </div>

        <!-- 管理中心入口 (已登录) -->
        <NTooltip v-if="visitMode === VisitMode.VISIT_MODE_LOGIN" placement="bottom" trigger="hover">
          <template #trigger>
            <button
              type="button"
              class="header-btn manage-btn"
              aria-label="应用管理中心"
              @click="emit('openSettings')"
            >
              <SvgIcon icon="majesticons-applications" class="btn-icon" />
              <span class="btn-text hidden sm:inline">管理</span>
            </button>
          </template>
          <span>应用管理与设置</span>
        </NTooltip>

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
  align-items: center;
  gap: 6px;
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
  align-items: center;
  padding: 4px 10px;
  height: 30px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.18);
  color: #FFF2D5;
  font-size: 13px;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
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
  .brand-en {
    display: none;
  }
  .header-btn {
    height: 28px;
    padding: 0 8px;
    font-size: 12px;
  }
}

/* 暖夜暗黑模式适配 */
:global(.dark) .usagi-global-header {
  background-color: var(--usagi-brown-footer, #181413);
  border-bottom-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.4);
}

:global(.dark) .clock-capsule {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}

:global(.dark) .header-btn {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.14);
}

:global(.dark) .header-btn:hover {
  background: rgba(255, 255, 255, 0.16);
  border-color: rgba(255, 183, 206, 0.4);
}
</style>
