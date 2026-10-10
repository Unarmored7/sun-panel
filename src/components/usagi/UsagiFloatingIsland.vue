<script setup lang="ts">
import { NTooltip } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { PanelStateNetworkModeEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { t } from '@/locales'

withDefaults(defineProps<{
  networkMode: PanelStateNetworkModeEnum
  netModeChangeButtonShow?: boolean
  visitMode: VisitMode
  showBackTop: boolean
}>(), {
  netModeChangeButtonShow: true,
})

const emit = defineEmits<{
  (e: 'changeNetwork', mode: PanelStateNetworkModeEnum): void
  (e: 'openSettings'): void
  (e: 'goLogin'): void
  (e: 'scrollToTop'): void
}>()
</script>

<template>
  <aside class="usagi-floating-island" aria-label="快捷工具栏">
    <!-- 回到顶部 -->
    <Transition name="island-fade">
      <NTooltip v-if="showBackTop" placement="left" trigger="hover">
        <template #trigger>
          <button
            type="button"
            class="island-btn back-top-btn"
            :title="t('common.backToTop') || '回到顶部'"
            @click="emit('scrollToTop')"
          >
            <SvgIcon class="island-icon" icon="icon-park-outline:to-top" />
          </button>
        </template>
        <span>回到顶部</span>
      </NTooltip>
    </Transition>

    <!-- 网络模式切换：内网 -> 外网 -->
    <NTooltip
      v-if="networkMode === PanelStateNetworkModeEnum.lan && netModeChangeButtonShow"
      placement="left"
      trigger="hover"
    >
      <template #trigger>
        <button
          type="button"
          class="island-btn"
          @click="emit('changeNetwork', PanelStateNetworkModeEnum.wan)"
        >
          <SvgIcon class="island-icon" icon="material-symbols:lan-outline-rounded" />
        </button>
      </template>
      <span>{{ t('panelHome.changeToWanModel') }}</span>
    </NTooltip>

    <!-- 网络模式切换：外网 -> 内网 -->
    <NTooltip
      v-if="networkMode === PanelStateNetworkModeEnum.wan && netModeChangeButtonShow"
      placement="left"
      trigger="hover"
    >
      <template #trigger>
        <button
          type="button"
          class="island-btn"
          @click="emit('changeNetwork', PanelStateNetworkModeEnum.lan)"
        >
          <SvgIcon class="island-icon" icon="mdi:wan" />
        </button>
      </template>
      <span>{{ t('panelHome.changeToLanModel') }}</span>
    </NTooltip>

    <!-- 登录后：管理应用中心 -->
    <NTooltip v-if="visitMode === VisitMode.VISIT_MODE_LOGIN" placement="left" trigger="hover">
      <template #trigger>
        <button
          type="button"
          class="island-btn manage-btn"
          @click="emit('openSettings')"
        >
          <SvgIcon class="island-icon" icon="majesticons-applications" />
        </button>
      </template>
      <span>应用管理与设置</span>
    </NTooltip>

    <!-- 访客模式：登录入口 -->
    <NTooltip v-if="visitMode === VisitMode.VISIT_MODE_PUBLIC" placement="left" trigger="hover">
      <template #trigger>
        <button
          type="button"
          class="island-btn login-btn"
          @click="emit('goLogin')"
        >
          <SvgIcon class="island-icon" icon="material-symbols:account-circle" />
        </button>
      </template>
      <span>{{ $t('panelHome.goToLogin') }}</span>
    </NTooltip>
  </aside>
</template>

<style scoped>
.usagi-floating-island {
  position: fixed;
  right: 18px;
  bottom: 84px;
  z-index: 40;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 6px;
  background: rgba(255, 253, 249, 0.88);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--usagi-brown-border, rgba(227, 216, 200, 0.85));
  border-radius: 9999px;
  box-shadow: 0 8px 30px rgba(138, 122, 106, 0.16);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.island-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: none;
  background: transparent;
  color: var(--usagi-text-main, #3A3131);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.island-icon {
  font-size: 19px;
  color: var(--usagi-pink-dark, #C2597F);
  transition: transform 0.2s ease, color 0.2s ease;
}

.island-btn:hover {
  background: var(--usagi-pink-wash, rgba(255, 183, 206, 0.22));
}

.island-btn:hover .island-icon {
  transform: scale(1.12);
  color: var(--usagi-pink-deep, #9A284E);
}

.island-btn:active {
  transform: scale(0.92);
}

.manage-btn:hover {
  background: rgba(255, 183, 206, 0.28);
}

/* 动效过渡 */
.island-fade-enter-active,
.island-fade-leave-active {
  transition: all 0.25s ease;
}

.island-fade-enter-from,
.island-fade-leave-to {
  opacity: 0;
  transform: translateY(8px) scale(0.85);
}

@media (max-width: 640px) {
  .usagi-floating-island {
    right: 12px;
    bottom: 78px;
    padding: 5px;
    gap: 4px;
  }
  .island-btn {
    width: 34px;
    height: 34px;
  }
  .island-icon {
    font-size: 17px;
  }
}
</style>
