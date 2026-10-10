<script setup lang="ts">
import { NEllipsis } from 'naive-ui'
import { ItemIcon } from '@/components/common'
import { PanelPanelConfigStyleEnum } from '@/enums'

interface Prop {
  itemInfo?: Panel.ItemInfo
  size?: number // 默认70
  forceBackground?: string // 强制背景色
  iconTextColor?: string
  iconTextInfoHideDescription: boolean
  iconTextIconHideTitle: boolean
  style: PanelPanelConfigStyleEnum
}

withDefaults(defineProps<Prop>(), {
  size: 70,
})

const defaultBackground = 'var(--usagi-bg-card)'

</script>

<template>
  <button type="button" class="app-icon w-full" :aria-label="itemInfo?.title">
    <!-- 详情图标 -->
    <div
      v-if="style === PanelPanelConfigStyleEnum.info"
      class="app-icon-info w-full flex"
      :style="{ background: forceBackground || defaultBackground }"
    >
      <!-- 图标 -->
      <div class="app-icon-info-icon w-[56px] h-[76px]">
        <div class="w-[56px] h-full flex items-center justify-center ">
          <ItemIcon :item-icon="itemInfo?.icon" force-background="transparent" :size="42" :fallback-text="itemInfo?.title?.slice(0, 2)" class="overflow-hidden rounded-xl" />
        </div>
      </div>

      <!-- 文字 -->
      <!-- The upstream white default becomes readable warm text on the Usagi card. -->
      <div class="app-icon-info-content flex items-center" :style="{ color: (!iconTextColor || iconTextColor === '#ffffff') ? 'var(--usagi-text-main)' : iconTextColor }">
        <div class="app-icon-info-text-box w-full">
          <div class="app-icon-info-text-box-title font-semibold w-full">
            <NEllipsis>
              {{ itemInfo?.title }}
            </NEllipsis>
          </div>
          <div v-if="!iconTextInfoHideDescription" class="app-icon-info-text-box-description">
            <NEllipsis :line-clamp="2" class="text-xs">
              {{ itemInfo?.description }}
            </NEllipsis>
          </div>
        </div>
      </div>
    </div>

    <!-- 极简(小)图标（APP） -->
    <div v-if="style === PanelPanelConfigStyleEnum.icon" class="app-icon-small">
      <div
        class="app-icon-small-icon overflow-hidden w-[70px] h-[70px] mx-auto"
        :title="itemInfo?.description"
      >
        <ItemIcon :item-icon="itemInfo?.icon" :fallback-text="itemInfo?.title?.slice(0, 2)" />
      </div>
      <div
        v-if="!iconTextIconHideTitle"
        class="app-icon-small-title text-center app-icon-text-shadow cursor-pointer mt-[2px]"
        :style="{ color: (!iconTextColor || iconTextColor === '#ffffff') ? 'var(--usagi-text-main)' : iconTextColor }"
      >
        <span>{{ itemInfo?.title }}</span>
      </div>
    </div>
  </button>
</template>
