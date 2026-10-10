<script setup lang="ts">
import { defineEmits, onMounted, ref } from 'vue'
import { NAvatar, NCheckbox } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useModuleConfig } from '@/store/modules'
import { useAuthStore } from '@/store'
import { VisitMode } from '@/enums/auth'

import SvgSrcBaidu from '@/assets/search_engine_svg/baidu.svg'
import SvgSrcBing from '@/assets/search_engine_svg/bing.svg'
import SvgSrcGoogle from '@/assets/search_engine_svg/google.svg'

withDefaults(defineProps<{
  borderColor?: string
  background?: string
  textColor?: string
}>(), {
  background: 'var(--usagi-bg-card)',
  textColor: 'var(--usagi-text-main)',
})

const emits = defineEmits(['itemSearch'])

interface State {
  currentSearchEngine: DeskModule.SearchBox.SearchEngine
  searchEngineList: DeskModule.SearchBox.SearchEngine[]
  newWindowOpen: boolean
}

const moduleConfigName = 'deskModuleSearchBox'
const moduleConfig = useModuleConfig()
const authStore = useAuthStore()
const searchTerm = ref('')
const isFocused = ref(false)
const searchSelectListShow = ref(false)
const defaultSearchEngineList = ref<DeskModule.SearchBox.SearchEngine[]>([
  {
    iconSrc: SvgSrcGoogle,
    title: 'Google',
    url: 'https://www.google.com/search?q=%s',
  },
  {
    iconSrc: SvgSrcBaidu,
    title: 'Baidu',
    url: 'https://www.baidu.com/s?wd=%s',
  },
  {
    iconSrc: SvgSrcBing,
    title: 'Bing',
    url: 'https://www.bing.com/search?q=%s',
  },
])

const defaultState: State = {
  currentSearchEngine: defaultSearchEngineList.value[0],
  searchEngineList: [...defaultSearchEngineList.value],
  newWindowOpen: false,
}

const state = ref<State>({ ...defaultState })

const onFocus = (): void => {
  isFocused.value = true
}

const onBlur = (): void => {
  isFocused.value = false
}

function handleEngineClick() {
  // 访客模式不允许修改
  if (authStore.visitMode === VisitMode.VISIT_MODE_PUBLIC)
    return
  searchSelectListShow.value = !searchSelectListShow.value
}

function handleEngineUpdate(engine: DeskModule.SearchBox.SearchEngine) {
  state.value.currentSearchEngine = engine
  moduleConfig.saveToCloud(moduleConfigName, state.value)
  searchSelectListShow.value = false
}

function handleSearchClick() {
  const url = state.value.currentSearchEngine.url
  const keyword = searchTerm
  // 如果网址中存在 %s，则直接替换为关键字
  const fullUrl = replaceOrAppendKeywordToUrl(url, keyword.value)
  handleClearSearchTerm()
  if (state.value.newWindowOpen)
    window.open(fullUrl)
  else
    window.location.href = fullUrl
}

function replaceOrAppendKeywordToUrl(url: string, keyword: string) {
  // 如果网址中存在 %s，则直接替换为关键字
  if (url.includes('%s'))
    return url.replace('%s', encodeURIComponent(keyword))

  // 如果网址中不存在 %s，则将关键字追加到末尾
  return url + (keyword ? `${encodeURIComponent(keyword)}` : '')
}

const handleItemSearch = () => {
  emits('itemSearch', searchTerm.value)
}

function handleClearSearchTerm() {
  searchTerm.value = ''
  emits('itemSearch', searchTerm.value)
}

onMounted(() => {
  moduleConfig.getValueByNameFromCloud<State>('deskModuleSearchBox').then(({ code, data }) => {
    if (code === 0)
      state.value = data || defaultState
    else
      state.value = defaultState
  })
})
</script>

<template>
  <div class="search-box w-full" @keydown.enter="handleSearchClick" @keydown.esc="handleClearSearchTerm">
    <!-- 快捷搜索引擎胶囊栏 -->
    <div class="search-engine-pills flex items-center justify-center gap-2 mb-2">
      <button
        v-for="(item, index) in state.searchEngineList"
        :key="index"
        type="button"
        class="engine-pill-btn"
        :class="{ 'is-active': state.currentSearchEngine.title === item.title }"
        @click="handleEngineUpdate(item)"
      >
        <NAvatar :src="item.iconSrc" style="background-color: transparent;" :size="15" />
        <span class="engine-title text-xs font-medium">{{ item.title }}</span>
      </button>
    </div>

    <div class="search-container flex rounded-2xl items-center justify-center text-white w-full" :style="{ background, color: textColor, borderColor }" :class="{ focused: isFocused }">
      <div class="search-box-btn-engine w-[38px] flex justify-center cursor-pointer" :title="'当前引擎: ' + state.currentSearchEngine.title" @click="handleEngineClick">
        <NAvatar :src="state.currentSearchEngine.iconSrc" style="background-color: transparent;" :size="20" />
      </div>

      <input
        v-model="searchTerm"
        :placeholder="`搜索书签，或按 Enter 在 ${state.currentSearchEngine.title} 中全网搜索...`"
        @focus="onFocus"
        @blur="onBlur"
        @input="handleItemSearch"
      >

      <div v-if="searchTerm !== ''" class="search-box-btn-clear w-[25px] mr-[6px] flex justify-center cursor-pointer" title="清空 (Esc)" @click="handleClearSearchTerm">
        <SvgIcon style="width: 18px;height: 18px;" icon="line-md:close-small" />
      </div>
      <div class="search-box-btn-search w-[32px] flex justify-center cursor-pointer" :title="'在 ' + state.currentSearchEngine.title + ' 中搜索'" @click="handleSearchClick">
        <SvgIcon style="width: 18px;height: 18px;" icon="iconamoon:search-fill" />
      </div>
    </div>

    <!-- 搜索引擎详细弹层（更多选项） -->
    <div v-if="searchSelectListShow" class="w-full mt-[10px] rounded-xl p-[10px] shadow-sm" :style="{ background }">
      <div class="flex items-center">
        <div class="flex items-center">
          <div
            v-for="item, index in defaultSearchEngineList"
            :key="index"
            :title="item.title"
            class="w-[40px] h-[40px] mr-[10px] cursor-pointer bg-[#ffffff] flex items-center justify-center rounded-xl border border-stone-200 hover:border-pink-300 transition-all"
            @click="handleEngineUpdate(item)"
          >
            <NAvatar :src="item.iconSrc" style="background-color: transparent;" :size="20" />
          </div>
        </div>
      </div>

      <div class="mt-[10px]">
        <NCheckbox v-model:checked="state.newWindowOpen" @update-checked="moduleConfig.saveToCloud(moduleConfigName, state)">
          <span :style="{ color: textColor }">
            {{ $t('deskModule.searchBox.openWithNewOpen') }}
          </span>
        </NCheckbox>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-container {
  border: 1px solid var(--usagi-brown-border, #E3D8C8);
  transition: box-shadow 0.3s ease, border-color 0.3s ease;
  padding: 4px 12px;
}

.engine-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 9px;
  border-radius: 9999px;
  border: 1px solid transparent;
  background: rgba(255, 255, 255, 0.6);
  color: var(--usagi-text-muted, #8A7A6A);
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.engine-pill-btn:hover {
  background: var(--usagi-pink-wash, rgba(255, 183, 206, 0.2));
  color: var(--usagi-pink-dark, #C2597F);
}

.engine-pill-btn.is-active {
  background: rgba(255, 255, 255, 0.95);
  border-color: var(--usagi-pink-primary, #FFB7CE);
  color: var(--usagi-pink-dark, #C2597F);
  box-shadow: 0 2px 8px rgba(255, 183, 206, 0.3);
}

.focused, .search-container:hover {
  border-color: var(--usagi-pink-primary, #FFB7CE) !important;
  box-shadow: 0 0 0 3px var(--usagi-pink-wash, rgba(255, 183, 206, 0.25)) !important;
}

input {
  background-color: transparent;
  box-sizing: border-box;
  width: 100%;
  height: 40px;
  padding: 8px 10px;
  border: none;
  outline: none;
  font-size: 15px;
  color: inherit;
}

input::placeholder {
  color: var(--usagi-text-muted, rgba(58, 49, 49, 0.45));
}
</style>
