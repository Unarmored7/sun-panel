<script setup lang="ts">
import { NButton, NDropdown, NModal, NSkeleton, NSpin, useDialog, useMessage } from 'naive-ui'
import { computed, defineAsyncComponent, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import AppIcon from './components/AppIcon/index.vue'
import SearchBox from '@/components/deskModule/SearchBox/index.vue'
import SvgIcon from '@/components/common/SvgIcon/index.vue'
import { deletes, getListByGroupId, saveSort } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'

import { setTitle, updateLocalUserInfo } from '@/utils/cmn'
import { useAuthStore, usePanelState } from '@/store'
import { PanelPanelConfigStyleEnum } from '@/enums'
import { VisitMode } from '@/enums/auth'
import { router } from '@/router'
import { t } from '@/locales'
import { resolveGroupCardStyle } from '@/utils/groupCardStyle'
import UsagiFooter from '@/components/usagi/UsagiFooter.vue'
import UsagiHeader from '@/components/usagi/UsagiHeader.vue'
import UsagiCategoryTabs from '@/components/usagi/UsagiCategoryTabs.vue'
import UsagiFloatingIsland from '@/components/usagi/UsagiFloatingIsland.vue'

const AppStarter = defineAsyncComponent(() => import('./components/AppStarter/index.vue'))
const EditItem = defineAsyncComponent(() => import('./components/EditItem/index.vue'))
const SystemMonitor = defineAsyncComponent(() => import('@/components/deskModule/SystemMonitor/index.vue'))
// Drag and drop is only needed while a group is being sorted.
const VueDraggable = defineAsyncComponent(() => import('vue-draggable-plus').then(module => module.VueDraggable))

interface ItemGroup extends Panel.ItemIconGroup {
  sortStatus?: boolean
  items?: Panel.ItemInfo[]
}

const ms = useMessage()
const dialog = useDialog()
const panelState = usePanelState()
const authStore = useAuthStore()


const editItemInfoShow = ref<boolean>(false)
const editItemInfoData = ref<Panel.ItemInfo | null>(null)
const windowShow = ref<boolean>(false)
const windowSrc = ref<string>('')
const windowTitle = ref<string>('')

const windowIframeIsLoad = ref<boolean>(false)

const dropdownMenuX = ref(0)
const dropdownMenuY = ref(0)
const dropdownShow = ref(false)
const currentRightSelectItem = ref<Panel.ItemInfo | null>(null)
const currentAddItenIconGroupId = ref<number | undefined>()

const settingModalShow = ref(false)
const settingsLoaded = ref(false)
const editorLoaded = ref(false)
watch(settingModalShow, (visible) => { if (visible) settingsLoaded.value = true })
watch(editItemInfoShow, (visible) => { if (visible) editorLoaded.value = true })
const pageLoading = ref(true)
const loadError = ref(false)

const items = ref<ItemGroup[]>([])
const filterItems = ref<ItemGroup[]>([])
const activeGroupId = ref<number | 'all'>('all')
const currentSearchKeyword = ref<string>('')
const showBackTop = ref<boolean>(false)

// Long pages paint the first screen of groups, then render the rest in one more pass,
// so the first screen does not wait for every card (one extra layout, not one per chunk).
const renderedGroupCount = ref(0)
let renderFrame = 0
function renderGroupsProgressively() {
  cancelAnimationFrame(renderFrame)
  let end = 0
  let cards = 0
  while (end < filterItems.value.length && (end === 0 || cards < 36))
    cards += filterItems.value[end++].items?.length || 0
  renderedGroupCount.value = end
  if (end < filterItems.value.length) {
    // Two frames: let the first screen paint before the remaining groups render.
    renderFrame = requestAnimationFrame(() => {
      renderFrame = requestAnimationFrame(() => { renderedGroupCount.value = filterItems.value.length })
    })
  }
}
watch(filterItems, renderGroupsProgressively)
const visibleGroups = computed(() => filterItems.value.slice(0, renderedGroupCount.value))

// Each group's card style is parsed once per change, not for every card render.
const groupCardStyles = computed(() => filterItems.value.map(group => resolveGroupCardStyle(group, panelState.panelConfig)))

const totalBookmarkCount = computed(() => {
  return items.value.reduce((acc, g) => acc + (g.items?.length || 0), 0)
})

const categoryGroups = computed(() => {
  return items.value.map(g => ({
    id: g.id,
    title: g.title || '',
    itemCount: g.items?.length || 0,
    publicVisitHide: g.publicVisitHide,
  }))
})

const filteredMatchCount = computed(() => {
  return filterItems.value.reduce((acc, g) => acc + (g.items?.length || 0), 0)
})

async function handleCategorySelect(id: number | 'all') {
  activeGroupId.value = id
  if (renderedGroupCount.value < filterItems.value.length) {
    cancelAnimationFrame(renderFrame)
    renderedGroupCount.value = filterItems.value.length
    await nextTick()
  }
  if (id === 'all') {
    window.scrollTo({ top: 0, behavior: 'smooth' })
    return
  }
  const el = document.getElementById(`usagi-group-${id}`)
  if (el) {
    const y = el.getBoundingClientRect().top + window.pageYOffset - 130
    window.scrollTo({ top: Math.max(0, y), behavior: 'smooth' })
  }
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

let scrollFrame = 0
function handleScroll() {
  if (!scrollFrame)
    scrollFrame = requestAnimationFrame(updateScrollState)
}

function updateScrollState() {
  scrollFrame = 0
  const scrollY = window.pageYOffset || document.documentElement.scrollTop
  showBackTop.value = scrollY > 260

  if (!currentSearchKeyword.value && items.value.length > 0) {
    const scrollPos = scrollY + 150
    let currentId: number | 'all' = 'all'

    for (let i = 0; i < items.value.length; i++) {
      const g = items.value[i]
      if (!g.id)
        continue
      const el = document.getElementById(`usagi-group-${g.id}`)
      if (el) {
        if (scrollPos >= el.offsetTop)
          currentId = g.id
      }
    }
    activeGroupId.value = currentId
  }
}

function openPage(openMethod: number, url: string, title?: string) {
  switch (openMethod) {
    case 1:
      window.location.href = url
      break
    case 2:
      window.open(url)
      break
    case 3:
      windowShow.value = true
      windowSrc.value = url
      windowTitle.value = title || url
      windowIframeIsLoad.value = true
      break

    default:
      break
  }
}

function handleItemClick(itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus) {
    handleEditItem(item)
    return
  }

  openPage(item.openMethod, item.url, item.title)
}

function handWindowIframeIdLoad(payload: Event) {
  windowIframeIsLoad.value = false
}

async function getList() {
  const { code, data } = await getGroupList<Common.ListResponse<ItemGroup[]>>(true)
  if (code !== 0 || !data?.list)
    throw new Error('Navigation could not be loaded')
  items.value = data.list.map(group => ({ ...group, items: group.items || [] }))
  itemFrontEndSearch(currentSearchKeyword.value)
}

async function initializePage() {
  pageLoading.value = true
  loadError.value = false
  try {
    await Promise.all([getList(), panelState.updatePanelConfigByCloud(), updateLocalUserInfo()])
    if (panelState.panelConfig.logoText)
      setTitle(panelState.panelConfig.logoText)
  }
  catch {
    loadError.value = true
  }
  finally {
    pageLoading.value = false
  }
}

// 从后端获取组下面的图标
function updateItemIconGroupByNet(itemIconGroupIndex: number, itemIconGroupId: number) {
  getListByGroupId<Common.ListResponse<Panel.ItemInfo[]>>(itemIconGroupId).then((res) => {
    if (res.code === 0)
      items.value[itemIconGroupIndex].items = res.data.list
  })
}

function handleRightMenuSelect(key: string | number) {
  dropdownShow.value = false
  // console.log(currentRightSelectItem, key)
  switch (key) {
    case 'newWindows':
      window.open(currentRightSelectItem.value?.url)
      break
    case 'edit':
      // 这里有个奇怪的问题，如果不使用{...}的方式 父组件的值会同步修改 标记一下
      handleEditItem({ ...currentRightSelectItem.value } as Panel.ItemInfo)
      break
    case 'delete':
      dialog.warning({
        title: t('common.warning'),
        content: t('common.deleteConfirmByName', { name: currentRightSelectItem.value?.title }),
        positiveText: t('common.confirm'),
        negativeText: t('common.cancel'),
        onPositiveClick: () => {
          deletes([currentRightSelectItem.value?.id as number]).then(({ code, msg }) => {
            if (code === 0) {
              ms.success(t('common.deleteSuccess'))
              getList()
            }
            else {
              ms.error(`${t('common.deleteFail')}:${msg}`)
            }
          })
        },
      })

      break
    default:
      break
  }
}

function handleContextMenu(e: MouseEvent, itemGroupIndex: number, item: Panel.ItemInfo) {
  if (items.value[itemGroupIndex] && items.value[itemGroupIndex].sortStatus)
    return

  e.preventDefault()
  currentRightSelectItem.value = item
  dropdownShow.value = false
  nextTick().then(() => {
    dropdownShow.value = true
    dropdownMenuX.value = e.clientX
    dropdownMenuY.value = e.clientY
  })
}

function onClickoutside() {
  // message.info('clickoutside')
  dropdownShow.value = false
}

function handleEditSuccess(item: Panel.ItemInfo) {
  getList()
}

// 结束拖拽
// function handleEndDrag(event: any, itemIconGroup: Panel.ItemIconGroup) {
//   // console.log(event)
//   // console.log(items.value)
// }

function handleSaveSort(itemGroup: ItemGroup) {
  const saveItems: Common.SortItemRequest[] = []
  if (itemGroup.items) {
    for (let i = 0; i < itemGroup.items.length; i++) {
      const element = itemGroup.items[i]
      saveItems.push({
        id: element.id as number,
        sort: i + 1,
      })
    }

    saveSort({ itemIconGroupId: itemGroup.id as number, sortItems: saveItems }).then(({ code, msg }) => {
      if (code === 0) {
        ms.success(t('common.saveSuccess'))
        itemGroup.sortStatus = false
      }
      else {
        ms.error(`${t('common.saveFail')}:${msg}`)
      }
    })
  }
}

function getDropdownMenuOptions() {
  const dropdownMenuOptions = [
    {
      label: t('iconItem.newWindowOpen'),
      key: 'newWindows',
    },

  ]

  if (authStore.visitMode === VisitMode.VISIT_MODE_LOGIN) {
    dropdownMenuOptions.push({
      label: t('common.edit'),
      key: 'edit',
    }, {
      label: t('common.delete'),
      key: 'delete',
    })
  }

  return dropdownMenuOptions
}

onMounted(() => {
  initializePage()

  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  cancelAnimationFrame(scrollFrame)
  cancelAnimationFrame(renderFrame)
})

// 前端搜索过滤
function itemFrontEndSearch(keyword?: string) {
  const kw = keyword?.trim() || ''
  currentSearchKeyword.value = kw
  if (kw !== '') {
    const filteredData: ItemGroup[] = []
    for (let i = 0; i < items.value.length; i++) {
      const element = items.value[i].items?.filter((item: Panel.ItemInfo) => {
        return (
          item.title.toLowerCase().includes(kw.toLowerCase())
          || item.url.toLowerCase().includes(kw.toLowerCase())
          || (item.description && item.description.toLowerCase().includes(kw.toLowerCase()))
        )
      })
      if (element && element.length > 0)
        filteredData.push({ ...items.value[i], items: element })
    }
    filterItems.value = filteredData
  }
  else {
    filterItems.value = items.value
  }
}

function handleClearSearch() {
  itemFrontEndSearch('')
}

function handleSetSortStatus(groupIndex: number, sortStatus: boolean) {
  if (items.value[groupIndex])
    items.value[groupIndex].sortStatus = sortStatus

  // 并未保存排序重新更新数据
  if (!sortStatus) {
    // 单独更新组
    if (items.value[groupIndex] && items.value[groupIndex].id)
      updateItemIconGroupByNet(groupIndex, items.value[groupIndex].id as number)
  }
}

function handleEditItem(item: Panel.ItemInfo) {
  editItemInfoData.value = item
  editItemInfoShow.value = true
  currentAddItenIconGroupId.value = undefined
}

function handleAddItem(itemIconGroupId?: number) {
  editItemInfoData.value = null
  editItemInfoShow.value = true
  if (itemIconGroupId)
    currentAddItenIconGroupId.value = itemIconGroupId
}
</script>

<template>
  <div class="sun-main usagi-page">
    <!-- 全局统一顶栏 (乌萨奇窝标准顶栏) -->
    <UsagiHeader
      :logo-text="panelState.panelConfig.logoText"
      :clock-show="panelState.panelConfig.clockShow !== false"
      :clock-show-second="panelState.panelConfig.clockShowSecond"
      :visit-mode="authStore.visitMode"
      @open-settings="settingModalShow = !settingModalShow"
      @go-login="router.push('/login')"
    />

    <div
      v-if="panelState.panelConfig.backgroundImageSrc" class="cover wallpaper" :style="{
        filter: `blur(${panelState.panelConfig.backgroundBlur}px)`,
        background: `url(${panelState.panelConfig.backgroundImageSrc}) no-repeat`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }"
    />
    <div class="mask" :style="{ backgroundColor: `rgba(0,0,0,${panelState.panelConfig.backgroundMaskNumber})` }" />
    <div v-if="pageLoading || loadError" class="usagi-content usagi-loading" :aria-busy="pageLoading">
      <div v-if="loadError" class="usagi-load-error" role="alert">
        <p>导航暂时未能加载，请重试。</p>
        <NButton type="primary" @click="initializePage">
          重新加载
        </NButton>
      </div>
      <div v-else role="status" aria-label="正在加载导航">
        <div class="usagi-loading-search" />
        <div class="usagi-loading-tabs" />
        <div class="usagi-loading-grid">
          <div v-for="n in 12" :key="n" class="usagi-loading-card" />
        </div>
      </div>
    </div>
    <div v-else class="usagi-content">
      <main
        class="usagi-main"
        :style="{
          marginTop: `${panelState.panelConfig.marginTop}%`,
          marginBottom: `${panelState.panelConfig.marginBottom}%`,
          maxWidth: (panelState.panelConfig.maxWidth ?? '1200') + panelState.panelConfig.maxWidthUnit,
        }"
      >
        <!-- 站名与时钟位于顶栏；此处只保留搜索框 -->
        <header v-if="panelState.panelConfig.searchBoxShow" class="usagi-header">
          <div class="usagi-search">
            <SearchBox :background="panelState.panelConfig.searchBoxBackground" :text-color="panelState.panelConfig.searchBoxTextColor" :border-color="panelState.panelConfig.searchBoxBorderColor" @itemSearch="itemFrontEndSearch" />
          </div>
        </header>

        <!-- 分类吸顶导航栏 -->
        <UsagiCategoryTabs
          v-if="!currentSearchKeyword && categoryGroups.length > 1"
          :groups="categoryGroups"
          :active-id="activeGroupId"
          :total-count="totalBookmarkCount"
          @select="handleCategorySelect"
        />

        <!-- 搜索匹配反馈条 -->
        <div v-if="currentSearchKeyword" class="usagi-search-feedback">
          <span class="feedback-text">
            搜索 “<strong>{{ currentSearchKeyword }}</strong>”：共匹配到 <strong>{{ filteredMatchCount }}</strong> 个书签
          </span>
          <button type="button" class="feedback-clear-btn" @click="handleClearSearch">
            <SvgIcon icon="line-md:close-small" class="inline text-base" /> 清除搜索
          </button>
        </div>

        <!-- 应用盒子 -->
        <div :style="{ marginLeft: `${panelState.panelConfig.marginX}px`, marginRight: `${panelState.panelConfig.marginX}px` }">
          <!-- 系统监控状态 -->
          <div
            v-if="panelState.panelConfig.systemMonitorShow
              && ((panelState.panelConfig.systemMonitorPublicVisitModeShow && authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
                || authStore.visitMode === VisitMode.VISIT_MODE_LOGIN)"
            class="flex mx-auto"
          >
            <SystemMonitor
              :allow-edit="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
              :show-title="panelState.panelConfig.systemMonitorShowTitle"
            />
          </div>

          <!-- 组纵向排列 -->
          <div
            v-for="(itemGroup, itemGroupIndex) in visibleGroups"
            :id="`usagi-group-${itemGroup.id}`"
            :key="itemGroup.id ?? itemGroupIndex"
            class="item-list"
            :class="itemGroup.sortStatus ? 'usagi-group-sorting' : ''"
          >
            <!-- 分组标题 -->
            <div class="group-heading">
              <span class="group-title text-shadow">
                {{ itemGroup.title }}
              </span>
              <span v-if="itemGroup.items" class="group-count">{{ itemGroup.items.length }}</span>
              <span v-if="itemGroup.publicVisitHide === 1" class="private-badge">{{ $t('apps.itemGroupManage.hidePublic') }}</span>
              <div
                v-if="authStore.visitMode === VisitMode.VISIT_MODE_LOGIN"
                class="group-buttons ml-2 delay-100 transition-opacity flex"
              >
                <span class="mr-2 cursor-pointer" :title="t('common.add')" @click="handleAddItem(itemGroup.id)">
                  <SvgIcon class="usagi-action-icon font-xl" icon="typcn:plus" />
                </span>
                <span class="mr-2 cursor-pointer " :title="t('common.sort')" @click="handleSetSortStatus(itemGroupIndex, !itemGroup.sortStatus)">
                  <SvgIcon class="usagi-action-icon font-xl" icon="ri:drag-drop-line" />
                </span>
              </div>
            </div>

            <!-- 卡片：详情（info）或图标（icon）样式；仅在排序时挂载拖拽 -->
            <component
              :is="itemGroup.sortStatus ? VueDraggable : 'div'"
              v-if="itemGroup.items"
              v-bind="itemGroup.sortStatus ? { 'modelValue': itemGroup.items, 'onUpdate:modelValue': (value: Panel.ItemInfo[]) => { itemGroup.items = value }, 'animation': 300, 'filter': '.not-drag' } : {}"
              :class="groupCardStyles[itemGroupIndex].style === PanelPanelConfigStyleEnum.info ? 'icon-info-box' : 'icon-small-box'"
            >
              <div v-for="(item, index) in itemGroup.items" :key="index" :title="item.description" @contextmenu="(e) => handleContextMenu(e, itemGroupIndex, item)">
                <AppIcon
                  :class="itemGroup.sortStatus ? 'cursor-move' : 'cursor-pointer'"
                  :item-info="item"
                  :eager="itemGroupIndex === 0 && index < 6"
                  :icon-text-color="groupCardStyles[itemGroupIndex].textColor"
                  :icon-text-info-hide-description="groupCardStyles[itemGroupIndex].hideDescription"
                  :icon-text-icon-hide-title="groupCardStyles[itemGroupIndex].hideTitle"
                  :card-style="groupCardStyles[itemGroupIndex].style"
                  @click="handleItemClick(itemGroupIndex, item)"
                />
              </div>

              <div v-if="itemGroup.items.length === 0 && authStore.visitMode === VisitMode.VISIT_MODE_LOGIN" class="not-drag">
                <AppIcon
                  class="cursor-pointer"
                  :item-info="{ icon: { itemType: 1, text: '+' }, title: t('common.add'), url: '', openMethod: 0 }"
                  :icon-text-color="groupCardStyles[itemGroupIndex].textColor"
                  :icon-text-info-hide-description="groupCardStyles[itemGroupIndex].hideDescription"
                  :icon-text-icon-hide-title="groupCardStyles[itemGroupIndex].hideTitle"
                  :card-style="groupCardStyles[itemGroupIndex].style"
                  @click="handleAddItem(itemGroup.id)"
                />
              </div>
            </component>

            <!-- 编辑栏 -->
            <div v-if="itemGroup.sortStatus" class="flex mt-[10px]">
              <div>
                <NButton type="primary" secondary @click="handleSaveSort(itemGroup)">
                  <template #icon>
                    <SvgIcon class="usagi-action-icon font-xl" icon="material-symbols:save" />
                  </template>
                  <div>
                    {{ $t('common.saveSort') }}
                  </div>
                </NButton>
              </div>
            </div>
          </div>
        </div>
        <div v-if="panelState.panelConfig.footerHtml && !panelState.panelConfig.footerHtml.includes('usagi-footer')" class="custom-footer" v-html="panelState.panelConfig.footerHtml" />
      </main>
    </div>
    <UsagiFooter v-if="!pageLoading" />

    <!-- 右键菜单 -->
    <NDropdown
      placement="bottom-start" trigger="manual" :x="dropdownMenuX" :y="dropdownMenuY"
      :options="getDropdownMenuOptions()" :show="dropdownShow" :on-clickoutside="onClickoutside" @select="handleRightMenuSelect"
    />

    <!-- 悬浮控制岛与应用中心 -->
    <UsagiFloatingIsland
      :show-back-top="showBackTop"
      @scroll-to-top="scrollToTop"
    />

    <AppStarter v-if="settingsLoaded" v-model:visible="settingModalShow" />

    <EditItem v-if="editorLoaded" v-model:visible="editItemInfoShow" :item-info="editItemInfoData" :item-group-id="currentAddItenIconGroupId" @done="handleEditSuccess" />

    <!-- 弹窗 -->
    <NModal
      v-model:show="windowShow" :mask-closable="false" preset="card"
      style="max-width: 1000px;height: 600px;border-radius: 1rem;" :bordered="true" size="small" role="dialog"
      aria-modal="true"
    >
      <template #header>
        <div class="flex items-center">
          <span class="mr-[20px]">
            {{ windowTitle }}
          </span>

          <NSpin v-if="windowIframeIsLoad" size="small" />
        </div>
      </template>
      <div class="w-full h-full rounded-2xl overflow-hidden border">
        <div v-if="windowIframeIsLoad" class="flex flex-col p-5">
          <NSkeleton height="50px" width="100%" class="rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
          <NSkeleton height="180px" width="100%" class="mt-[20px] rounded-lg" />
        </div>
        <iframe
          v-show="!windowIframeIsLoad" id="windowIframeId" :src="windowSrc"
          class="w-full h-full" frameborder="0" @load="handWindowIframeIdLoad"
        />
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.mask { width: 100%; height: 100%; }
.usagi-search-feedback {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 680px;
  margin: 14px auto 22px;
  padding: 8px 18px;
  background: var(--usagi-pink-wash, rgba(255, 183, 206, 0.18));
  border: 1px solid var(--usagi-pink-primary, #FFB7CE);
  border-radius: 9999px;
  font-size: 13px;
  color: var(--usagi-text-main, #3A3131);
}
.feedback-text strong {
  color: var(--usagi-pink-dark, #C2597F);
}
.feedback-clear-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px 10px;
  border-radius: 9999px;
  border: none;
  background: rgba(255, 255, 255, 0.85);
  color: var(--usagi-pink-dark, #C2597F);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.feedback-clear-btn:hover {
  background: #FFFFFF;
  color: var(--usagi-pink-deep, #9A284E);
}
</style>
