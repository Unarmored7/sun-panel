<script setup lang="ts">
import { computed, defineEmits, defineProps, onBeforeUnmount, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NForm, NFormItem, NInput, NModal, NSelect, useMessage } from 'naive-ui'
import ItemIcon from '@/components/common/ItemIcon/index.vue'
import { edit, getSiteFavicon } from '@/api/panel/itemIcon'
import { getList as getGroupList } from '@/api/panel/itemIconGroup'
import { t } from '@/locales'

interface Props {
  visible: boolean
  itemInfo: Panel.Info | null
  itemGroupId?: number
}

const props = defineProps<Props>()
const emit = defineEmits<Emit>()
const ms = useMessage()
const submitLoading = ref(false)
const itemIconGroupOptions = ref<{
  label: string
  value: number
}[]>([])

const restoreDefault: Panel.Info = {
  icon: null,
  title: '',
  url: '',
  description: '',
  openMethod: 2,
}

interface Emit {
  (e: 'update:visible', visible: boolean): void
  (e: 'done', item: Panel.Info): void// 创建完成
}

const model = ref<Panel.Info>(props.itemInfo ? { ...props.itemInfo } : { ...restoreDefault })
const formRef = ref<FormInst | null>(null)

const rules: FormRules = {
  url: {
    required: true,
    trigger: 'blur',
    type: 'string',
    message: t('form.required'),
  },
  // itemIconGroupId: {
  //   required: true,
  //   trigger: ['blur', 'change'],
  //   message: t('form.required'),
  // },
}

const options = [
  {
    default: true,
    label: t('iconItem.currentPageOpen'),
    value: 1,
  },
  {
    label: t('iconItem.newWindowOpen'),
    value: 2,
  },
  {
    label: t('iconItem.currentPageLayerOpen'),
    value: 3,
  },
]

// 更新值父组件传来的值
const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => {
    emit('update:visible', visible)
  },
})

async function editApi() {
  submitLoading.value = true
  try {
    const { code, data, msg } = await edit<Panel.ItemInfo>(model.value)
    if (code === 0) {
      show.value = false
      model.value = { ...restoreDefault }

      emit('done', data)
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  }
  catch {
    ms.error(t('common.saveFail'))
  }
  submitLoading.value = false
}

interface SiteInfo { iconUrl: string; originalIconUrl: string; title: string; description: string }
type DetectStatus = 'idle' | 'loading' | 'found' | 'missing'
const detectStatus = ref<DetectStatus>('idle')
let lookupTimer: ReturnType<typeof setTimeout> | undefined
let generation = 0
let initialURL = ''
let initial: Pick<Panel.Info, 'title' | 'description' | 'icon'> = { title: '', description: '', icon: null }
let lookupSeq = 0
let lookupURL = ''
let lookupPromise: Promise<void> | undefined

function normalizeURL(raw: string) {
  const value = raw.trim()
  if (!value)
    return ''
  try {
    return new URL(value.includes('://') ? value : `https://${value}`).href
  }
  catch {
    return ''
  }
}

function hostname(url: string) {
  return new URL(url).hostname.replace(/^www\./, '').slice(0, 20)
}

function hasImageIcon(icon: Panel.ItemIcon | null | undefined) {
  return icon?.itemType === 2 && !!icon.src
}

// Name, description and icon all come from the site. A new address replaces
// them; the saved address only fills gaps unless detection is requested again.
function detectSite(raw: string, force = false) {
  const url = normalizeURL(raw)
  if (!url)
    return
  if (!force && lookupURL === url && lookupPromise)
    return lookupPromise
  const currentGeneration = generation
  const seq = ++lookupSeq
  // Any later URL edit or lookup supersedes this one.
  const stale = () => !show.value || currentGeneration !== generation || seq !== lookupSeq
  const replace = url !== initialURL
  detectStatus.value = 'loading'
  const task = async () => {
    let info: SiteInfo | undefined
    try {
      const { code, data } = await getSiteFavicon<SiteInfo>(url)
      if (code === 0)
        info = data
    }
    catch {
      // An unreachable site keeps its address; the server retries after saving.
    }
    if (stale())
      return
    const fallback = hostname(url)
    const title = info?.title && info.title !== fallback ? [...info.title].slice(0, 20).join('') : ''
    const description = [...(info?.description || '')].slice(0, 100).join('')
    const icon: Panel.ItemIcon | null = info?.iconUrl ? { itemType: 2, src: info.originalIconUrl || info.iconUrl, cachedSrc: info.iconUrl } : null
    if (replace) {
      model.value.title = title || fallback
      model.value.description = description
      model.value.icon = icon
    }
    else {
      // Never trade saved details for empty results from a briefly unreachable site.
      if (title && (force || !model.value.title || model.value.title === fallback))
        model.value.title = title
      if (description && (force || !model.value.description))
        model.value.description = description
      if (icon && (force || !hasImageIcon(model.value.icon)))
        model.value.icon = icon
    }
    if (!model.value.title)
      model.value.title = fallback
    if (icon)
      detectStatus.value = 'found'
    else if (!replace && (model.value.icon?.src || model.value.icon?.text))
      detectStatus.value = 'idle'
    else
      detectStatus.value = 'missing'
    if (force && !replace && !icon && !title && !description)
      ms.warning(t('iconItem.geticonFail'))
  }
  // Keep a completed lookup reusable across debounce, blur and save.
  lookupURL = url
  lookupPromise = task()
  return lookupPromise
}

function handleUrlBlur() {
  clearTimeout(lookupTimer)
  const url = normalizeURL(model.value.url || '')
  if (!url)
    return
  model.value.url = url
  if (!model.value.title)
    model.value.title = hostname(url)
  if (url !== initialURL || !hasImageIcon(model.value.icon))
    return detectSite(url)
}

watch(() => model.value.url, (value, old) => {
  const url = normalizeURL(value || '')
  if (!show.value || url === normalizeURL(old || '') || url === lookupURL)
    return
  clearTimeout(lookupTimer)
  lookupSeq++
  lookupURL = ''
  lookupPromise = undefined
  detectStatus.value = 'idle'
  if (url && url === initialURL) {
    // Back to the saved address: show its saved details again.
    model.value.title = initial.title
    model.value.description = initial.description
    model.value.icon = initial.icon ? { ...initial.icon } : null
    if (hasImageIcon(initial.icon))
      return
  }
  else {
    model.value.title = ''
    model.value.description = ''
    model.value.icon = null
  }
  lookupTimer = setTimeout(() => { handleUrlBlur() }, 450)
})

async function handleValidateButtonClick(e: MouseEvent) {
  e.preventDefault()
  if (submitLoading.value)
    return
  submitLoading.value = true
  try {
    let pendingURL: string
    do {
      pendingURL = normalizeURL(model.value.url || '')
      await handleUrlBlur()
    } while (pendingURL !== normalizeURL(model.value.url || ''))
    await formRef.value?.validate()
    await editApi()
  }
  catch {
    // Form fields explain validation errors inline.
  }
  finally {
    submitLoading.value = false
  }
}

watch(() => props.visible, (visible) => {
  generation++
  clearTimeout(lookupTimer)
  lookupURL = ''
  lookupPromise = undefined
  detectStatus.value = 'idle'
  if (visible) {
    model.value = props.itemInfo ? { ...props.itemInfo, icon: props.itemInfo.icon ? { ...props.itemInfo.icon } : null } : { ...restoreDefault }
    initialURL = normalizeURL(model.value.url || '')
    initial = { title: model.value.title, description: model.value.description, icon: model.value.icon ? { ...model.value.icon } : null }
    if (props.itemGroupId)
      model.value.itemIconGroupId = props.itemGroupId
    getGroupListOptions()
    // Bookmarks saved without a site icon look the site up again when opened.
    if (initialURL && !hasImageIcon(initial.icon))
      detectSite(initialURL)
  }
}, { immediate: true })

onBeforeUnmount(() => { generation++; clearTimeout(lookupTimer) })

function getGroupListOptions() {
  getGroupList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ data, code, msg }) => {
    if (code === 0) {
      itemIconGroupOptions.value = []

      for (let i = 0; i < data.list.length; i++) {
        const element = data.list[i]
        if (i === 0 && !model.value.itemIconGroupId) {
          model.value.itemIconGroupId = element.id
          restoreDefault.itemIconGroupId = element.id
        }

        itemIconGroupOptions.value.push({
          value: element.id as number,
          label: element.title as string,
        })
      }
    }
    else {
      ms.error(`${t('iconItem.getGroupFail')}:${msg}`)
    }
  })
}
</script>

<template>
  <NModal v-model:show="show" preset="card" size="small" style="width: 600px;border-radius: 1rem;" :title="itemInfo ? t('iconItem.edit') : t('iconItem.add')">
    <div class="max-h-[65vh] overflow-auto p-[5px]">
      <NForm ref="formRef" :model="model" :rules="rules">
        <NFormItem path="url" :label="$t('iconItem.url')">
          <NInput v-model:value="model.url" type="text" :maxlength="1000" :placeholder="$t('iconItem.urlPlaceholder')" @blur="handleUrlBlur" />
        </NFormItem>
        <div class="mb-5 flex items-center gap-3">
          <ItemIcon :item-icon="model.icon" :size="48" eager :fallback-text="model.title?.slice(0, 2)" />
          <div class="min-w-0 flex-1">
            <div class="truncate font-medium">
              {{ model.title || $t('iconItem.titlePending') }}
            </div>
            <div v-if="model.description" class="truncate text-sm opacity-70" :title="model.description">
              {{ model.description }}
            </div>
            <div class="text-xs opacity-60">
              {{ detectStatus === 'loading' ? $t('iconItem.iconDetecting') : detectStatus === 'found' ? $t('iconItem.iconDetected') : detectStatus === 'missing' ? $t('iconItem.iconMissing') : $t('iconItem.iconAuto') }}
            </div>
          </div>
          <NButton text type="primary" :disabled="!model.url" :loading="detectStatus === 'loading'" @click="detectSite(model.url, true)">
            {{ $t('iconItem.redetect') }}
          </NButton>
        </div>
        <NFormItem path="itemIconGroupId" :label="t('iconItem.iconGroup')">
          <NSelect v-model:value="model.itemIconGroupId" :options="itemIconGroupOptions" />
        </NFormItem>
        <NFormItem path="openMethod" :label="$t('iconItem.openMethod')">
          <NSelect v-model:value="model.openMethod" :options="options" />
        </NFormItem>
      </NForm>
    </div>

    <template #footer>
      <NButton type="primary" :loading="submitLoading" style="float: right;" @click="handleValidateButtonClick">
        {{ $t('common.save') }}
      </NButton>
    </template>
  </NModal>
</template>
