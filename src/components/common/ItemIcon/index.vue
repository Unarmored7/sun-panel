<script setup lang="ts">
import { NAvatar } from 'naive-ui'
import { computed, ref, watch, withDefaults } from 'vue'
import { SvgIconOnline } from '@/components/common'

interface Prop {
  fallbackText?: string
  itemIcon?: Panel.ItemIcon | null
  size?: number // 默认70
  forceBackground?: string // 强制背景色
}

const props = withDefaults(defineProps<Prop>(), { size: 70 })
const defaultBackground = 'var(--usagi-bg-alt)'
const defaultStyle = computed(() => ({ width: `${props.size}px`, height: `${props.size}px` }))
const imageFailed = ref(false)
const imageReady = ref(false)
const background = computed(() => {
  const value = props.forceBackground ?? props.itemIcon?.backgroundColor
  return !value || value.toLowerCase() === '#2a2a2a6b' ? defaultBackground : value
})
watch(() => props.itemIcon?.src, () => { imageFailed.value = false; imageReady.value = false })
</script>

<template>
  <div class="item-icon" :style="defaultStyle">
    <slot>
      <template v-if="itemIcon">
        <template v-if="itemIcon?.itemType === 1">
          <NAvatar :size="props.size" :style="{ backgroundColor: background, color: 'var(--usagi-pink-deep)' }">
            {{ itemIcon.text }}
          </NAvatar>
        </template>

        <template v-else-if="itemIcon?.itemType === 2">
          <div class="item-icon-image" :style="{ backgroundColor: background, ...defaultStyle }">
            <img v-if="!imageFailed && itemIcon.src" :src="itemIcon.src" alt="" loading="lazy" decoding="async" :style="{ opacity: imageReady ? 1 : 0 }" @load="imageReady = true" @error="imageFailed = true">
            <span v-if="!imageReady" class="item-icon-fallback">{{ fallbackText || itemIcon.text || '↗' }}</span>
          </div>
        </template>

        <template v-else-if="itemIcon?.itemType === 3">
          <NAvatar :size="props.size" :style="{ backgroundColor: background, color: 'var(--usagi-pink-deep)' }">
            <SvgIconOnline style="font-size: 35px;" :icon="itemIcon.text" />
          </NAvatar>
        </template>
      </template>

      <template v-else>
        <NAvatar :size="props.size" />
      </template>
    </slot>
  </div>
</template>

<style scoped>
.item-icon-image { display: grid; place-items: center; }
.item-icon-image img { grid-area: 1 / 1; width: 78%; height: 78%; object-fit: contain; }
.item-icon-fallback { grid-area: 1 / 1; color: var(--usagi-pink-deep); font-size: 16px; font-weight: 600; }
</style>
