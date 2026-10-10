<script setup lang="ts">
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

function handleTabClick(id: number | 'all') {
  emit('select', id)
}
</script>

<template>
  <div v-if="groups.length > 0" class="usagi-category-nav-wrapper">
    <nav class="usagi-category-tabs" aria-label="书签分类导航">
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
    </nav>
  </div>
</template>

<style scoped>
.usagi-category-nav-wrapper {
  position: sticky;
  top: 14px;
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
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: min(100%, 960px);
  padding: 5px 8px;
  overflow-x: auto;
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE 10+ */
  background: rgba(255, 253, 249, 0.88);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid var(--usagi-brown-border, rgba(227, 216, 200, 0.85));
  border-radius: 9999px;
  box-shadow: 0 4px 18px rgba(138, 122, 106, 0.08);
  transition: all 0.2s ease;
}

.usagi-category-tabs::-webkit-scrollbar {
  display: none; /* Chrome / Safari */
}

.usagi-tab-item {
  display: inline-flex;
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
    top: 8px;
    margin: 12px auto 18px;
    padding: 0 8px;
  }
  .usagi-tab-item {
    padding: 4px 10px;
    font-size: 12px;
  }
}
</style>
