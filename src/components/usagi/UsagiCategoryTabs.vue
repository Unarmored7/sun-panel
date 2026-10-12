<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

export interface CategoryGroup {
  id?: number
  title: string
  itemCount?: number
  publicVisitHide?: number
}

const props = withDefaults(defineProps<{
  groups: CategoryGroup[]
  activeId?: number | 'all'
  totalCount?: number
}>(), {
  activeId: 'all',
  totalCount: 0,
})

const emit = defineEmits<{
  (e: 'select', id: number | 'all'): void
}>()

const track = ref<HTMLElement | null>(null)
const hasOverflow = ref(false)
const canScrollLeft = ref(false)
const canScrollRight = ref(false)
let resizeObserver: ResizeObserver | undefined

function measureOverflow() {
  const el = track.value
  if (!el)
    return
  const maxScroll = el.scrollWidth - el.clientWidth
  hasOverflow.value = maxScroll > 1
  canScrollLeft.value = el.scrollLeft > 1
  canScrollRight.value = el.scrollLeft < maxScroll - 1
}

function scrollBehavior(): ScrollBehavior {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth'
}

function scrollCategories(direction: -1 | 1) {
  const el = track.value
  if (el)
    el.scrollBy({ left: direction * Math.max(120, el.clientWidth * 0.75), behavior: scrollBehavior() })
}

function revealTab(button: HTMLElement) {
  const el = track.value
  if (!el)
    return
  const bounds = el.getBoundingClientRect()
  const tab = button.getBoundingClientRect()
  const left = bounds.left + (canScrollLeft.value ? 18 : 4)
  const right = bounds.right - (canScrollRight.value ? 18 : 4)
  if (tab.left < left)
    el.scrollBy({ left: tab.left - left, behavior: scrollBehavior() })
  else if (tab.right > right)
    el.scrollBy({ left: tab.right - right, behavior: scrollBehavior() })
}

function revealActiveTab() {
  const button = track.value?.querySelector<HTMLElement>('.usagi-tab-item.is-active')
  if (button)
    revealTab(button)
}

function handleTabFocus(event: FocusEvent) {
  if (event.target instanceof HTMLElement)
    revealTab(event.target)
}

function handleTabClick(id: number | 'all') {
  emit('select', id)
}

watch(() => props.groups, measureOverflow, { deep: true, flush: 'post' })
watch(() => props.activeId, revealActiveTab, { flush: 'post' })

onMounted(() => {
  measureOverflow()
  if (track.value && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(measureOverflow)
    resizeObserver.observe(track.value)
  }
  window.addEventListener('resize', measureOverflow)
  nextTick(revealActiveTab)
})

onUnmounted(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', measureOverflow)
})
</script>

<template>
  <div v-if="groups.length > 0" class="usagi-category-nav-wrapper">
    <nav class="usagi-category-tabs" aria-label="书签分类导航">
      <button
        v-if="hasOverflow"
        type="button"
        class="category-scroll-button scroll-left"
        aria-label="向左滚动分类"
        :disabled="!canScrollLeft"
        @click="scrollCategories(-1)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 6-6 6 6 6" /></svg>
      </button>
      <div
        ref="track"
        class="usagi-category-track"
        :data-scroll-left="canScrollLeft"
        :data-scroll-right="canScrollRight"
        @scroll.passive="measureOverflow"
        @focusin="handleTabFocus"
      >
      <!-- 全部标签 -->
      <button
        type="button"
        class="usagi-tab-item"
        :class="{ 'is-active': activeId === 'all' }"
        @click="handleTabClick('all')"
      >
        <span class="tab-title">全部</span>
        <span v-if="totalCount > 0" class="tab-badge">{{ totalCount }}</span>
      </button>

      <!-- 各分组标签 -->
      <button
        v-for="group in groups"
        :key="group.id ?? group.title"
        type="button"
        class="usagi-tab-item"
        :class="{ 'is-active': activeId === group.id }"
        @click="handleTabClick(group.id ?? 'all')"
      >
        <span class="tab-title">{{ group.title }}</span>
        <span v-if="group.itemCount !== undefined && group.itemCount > 0" class="tab-badge">
          {{ group.itemCount }}
        </span>
      </button>
      </div>
      <button
        v-if="hasOverflow"
        type="button"
        class="category-scroll-button scroll-right"
        aria-label="向右滚动分类"
        :disabled="!canScrollRight"
        @click="scrollCategories(1)"
      >
        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m9 6 6 6-6 6" /></svg>
      </button>
    </nav>
  </div>
</template>

<style scoped>
.usagi-category-nav-wrapper {
  position: sticky;
  top: 66px;
  z-index: 25;
  display: flex;
  justify-content: center;
  margin: 18px auto 26px;
  width: 100%;
  padding: 0 12px;
  pointer-events: none; /* 让外部透明区域穿透点击 */
}

.usagi-category-tabs {
  pointer-events: auto;
  display: flex;
  align-items: center;
  gap: 4px;
  max-width: min(100%, 960px);
  min-width: 0;
  padding: 3px 6px;
  background: rgba(255, 253, 249, 0.88);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid var(--usagi-brown-border, rgba(227, 216, 200, 0.85));
  border-radius: 9999px;
  box-shadow: 0 4px 18px rgba(138, 122, 106, 0.08);
  transition: all 0.2s ease;
}

.usagi-category-track {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1 1 auto;
  min-width: 0;
  padding: 4px 2px;
  overflow-x: auto;
  overflow-y: hidden;
  overscroll-behavior-x: contain;
  scrollbar-width: none;
  scroll-padding-inline: 18px;
}

.usagi-category-track::-webkit-scrollbar {
  display: none;
}

.usagi-category-track[data-scroll-right="true"] {
  mask-image: linear-gradient(to right, #000 calc(100% - 16px), transparent);
}

.usagi-category-track[data-scroll-left="true"] {
  mask-image: linear-gradient(to right, transparent, #000 16px);
}

.usagi-category-track[data-scroll-left="true"][data-scroll-right="true"] {
  mask-image: linear-gradient(to right, transparent, #000 16px, #000 calc(100% - 16px), transparent);
}

.category-scroll-button {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 30px;
  width: 30px;
  height: 30px;
  padding: 0;
  border: 0;
  border-radius: 50%;
  background: var(--usagi-pink-wash);
  color: var(--usagi-pink-dark);
  cursor: pointer;
}

.category-scroll-button:disabled {
  background: transparent;
  color: var(--usagi-text-muted);
  opacity: 0.4;
  cursor: default;
}

.category-scroll-button svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.category-scroll-button:focus-visible,
.usagi-tab-item:focus-visible {
  outline: 2px solid var(--usagi-pink-dark);
  outline-offset: -1px;
}

.usagi-tab-item {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  border-radius: 9999px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--usagi-text-main, #3A3131);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}

.usagi-tab-item:hover {
  color: var(--usagi-pink-dark, #C2597F);
  background: var(--usagi-pink-wash, rgba(255, 183, 206, 0.18));
}

.usagi-tab-item.is-active {
  background: var(--usagi-pink-primary, #FFB7CE);
  color: #FFFFFF;
  font-weight: 600;
  box-shadow: 0 2px 10px rgba(255, 183, 206, 0.45);
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 6px;
  height: 17px;
  min-width: 17px;
  border-radius: 9999px;
  font-size: 10.5px;
  font-weight: 600;
  line-height: 1;
  background: rgba(227, 216, 200, 0.5);
  color: var(--usagi-text-main, #3A3131);
  transition: all 0.2s ease;
}

.usagi-tab-item:hover .tab-badge {
  background: rgba(255, 183, 206, 0.4);
  color: var(--usagi-pink-deep, #9A284E);
}

.usagi-tab-item.is-active .tab-badge {
  background: rgba(255, 255, 255, 0.32);
  color: #FFFFFF;
}

@media (max-width: 640px) {
  .usagi-category-nav-wrapper {
    top: 56px;
    margin: 12px auto 18px;
    padding: 0 8px;
  }
  .usagi-tab-item {
    padding: 4px 10px;
    font-size: 12px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .usagi-tab-item, .usagi-category-tabs { transition: none; }
}
</style>
