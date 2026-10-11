<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, onUnmounted, ref, watch, withDefaults } from 'vue'
import { loadNearViewport } from '@/utils/imageVisibility'

interface Prop {
  fallbackText?: string
  itemIcon?: Panel.ItemIcon | null
  size?: number // 默认70
  eager?: boolean
  forceBackground?: string // 强制背景色
}

const props = withDefaults(defineProps<Prop>(), { size: 70 })
const SvgIconOnline = defineAsyncComponent(() => import('../SvgIconOnline/index.vue'))
const imageFailed = ref(false)
const imageReady = ref(false)
const root = ref<HTMLElement>()
const shouldLoad = ref(props.eager || false)
const useOriginal = ref(false)
const imageSource = computed(() => !useOriginal.value && props.itemIcon?.cachedSrc ? props.itemIcon.cachedSrc : props.itemIcon?.src)
let stopObserving = () => {}
onMounted(() => {
  if (root.value && !shouldLoad.value)
    stopObserving = loadNearViewport(root.value, () => { shouldLoad.value = true })
})
onUnmounted(() => stopObserving())
function handleImageError() {
  imageReady.value = false
  if (!useOriginal.value && props.itemIcon?.cachedSrc && props.itemIcon.cachedSrc !== props.itemIcon.src)
    useOriginal.value = true
  else
    imageFailed.value = true
}
const configuredBackground = computed(() => {
  const value = props.forceBackground ?? props.itemIcon?.backgroundColor
  return value && value.toLowerCase() !== '#2a2a2a6b' ? value : undefined
})
const background = computed(() => {
  if (configuredBackground.value)
    return configuredBackground.value
  return props.itemIcon?.itemType === 2 && imageReady.value ? 'transparent' : 'var(--usagi-pink-primary)'
})
const defaultStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`,
  fontSize: `${props.size * 0.32}px`,
  backgroundColor: background.value,
  color: configuredBackground.value ? 'var(--usagi-pink-deep)' : '#9A284E',
}))
const fallback = computed(() => props.fallbackText || props.itemIcon?.text || '↗')
watch([() => props.itemIcon?.src, () => props.itemIcon?.cachedSrc], () => { imageFailed.value = false; imageReady.value = false; useOriginal.value = false })
</script>

<template>
  <div ref="root" class="item-icon" :style="defaultStyle">
    <slot>
      <span v-if="itemIcon?.itemType === 1" class="item-icon-text">{{ itemIcon.text || fallback }}</span>
      <template v-else-if="itemIcon?.itemType === 2">
        <img v-if="shouldLoad && !imageFailed && imageSource" :src="imageSource" alt="" :width="size" :height="size" :loading="eager ? 'eager' : 'lazy'" :fetchpriority="eager ? 'high' : 'auto'" decoding="async" :style="{ opacity: imageReady ? 1 : 0 }" @load="imageReady = true" @error="handleImageError">
        <span v-if="!imageReady" class="item-icon-text">{{ fallback }}</span>
      </template>
      <SvgIconOnline v-else-if="itemIcon?.itemType === 3" class="item-icon-symbol" :icon="itemIcon.text" />
      <span v-else class="item-icon-text">{{ fallback }}</span>
    </slot>
  </div>
</template>

<style scoped>
.item-icon {
  /* One continuous rounded-square mask for images, symbols, and text at every size. */
  --app-icon-mask: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'%3E%3Cpath fill='white' d='M35 0H65C77 0 83 0 88.5 2.5C92.5 4.3 95.7 7.5 97.5 11.5C100 17 100 23 100 35V65C100 77 100 83 97.5 88.5C95.7 92.5 92.5 95.7 88.5 97.5C83 100 77 100 65 100H35C23 100 17 100 11.5 97.5C7.5 95.7 4.3 92.5 2.5 88.5C0 83 0 77 0 65V35C0 23 0 17 2.5 11.5C4.3 7.5 7.5 4.3 11.5 2.5C17 0 23 0 35 0Z'/%3E%3C/svg%3E");
  display: grid;
  place-items: center;
  flex-shrink: 0;
  padding: 0;
  border: 0;
  border-radius: 22.4%;
  overflow: hidden;
  color: var(--usagi-pink-deep);
  -webkit-mask: var(--app-icon-mask) center / 100% 100% no-repeat;
  mask: var(--app-icon-mask) center / 100% 100% no-repeat;
}
@supports (mask-image: linear-gradient(white, white)) {
  .item-icon { border-radius: 0; }
}
.item-icon img { grid-area: 1 / 1; display: block; width: 100%; height: 100%; min-width: 0; min-height: 0; object-fit: cover; }
.item-icon-text { grid-area: 1 / 1; max-width: 86%; overflow: hidden; white-space: nowrap; line-height: 1; font-weight: 600; }
.item-icon-symbol { grid-area: 1 / 1; width: 55%; height: 55%; }
</style>
