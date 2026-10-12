<script setup lang="ts">
import { computed, defineEmits, defineProps, onBeforeUnmount, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NForm, NFormItem, NGrid, NGridItem, NInput, NModal, NSelect, useMessage } from 'naive-ui'
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
  lanUrl: '',
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
  title: {
    required: true,
    trigger: 'blur',
    message: t('form.required'),
  },
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

type IconStatus = 'idle' | 'loading' | 'found' | 'missing'
const iconStatus = ref<IconStatus>('idle')
let lookupTimer: ReturnType<typeof setTimeout> | undefined
let generation = 0
let initialURL = ''
let initialIcon: Panel.ItemIcon | null = null
let autoTitle = ''
let autoDescription = ''
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

function hasImageIcon(icon: Panel.ItemIcon | null | undefined) {
  return icon?.itemType === 2 && !!icon.src
}

function fillHostname(url: string) {
  if (!model.value.title || model.value.title === autoTitle) {
    autoTitle = new URL(url).hostname.replace(/^www\./, '').slice(0, 20)
    model.value.title = autoTitle
  }
}

// The icon always comes from the site: a lookup replaces it, and a new URL without one falls back to the title.
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
  iconStatus.value = 'loading'
  const task = async () => {
    let found = false
    try {
      const { code, data } = await getSiteFavicon<{ iconUrl: string; originalIconUrl: string; title: string; description: string }>(url)
      if (stale())
        return
      if (code === 0) {
        if (data.title && (!model.value.title || model.value.title === autoTitle)) {
          autoTitle = [...data.title].slice(0, 20).join('')
          model.value.title = autoTitle
        }
        if (data.description && (!model.value.description || model.value.description === autoDescription)) {
          autoDescription = [...data.description].slice(0, 100).join('')
          model.value.description = autoDescription
        }
        if (data.iconUrl) {
          model.value.icon = { itemType: 2, src: data.originalIconUrl || data.iconUrl, cachedSrc: data.iconUrl }
          found = true
        }
      }
    }
    catch {
      // An unreachable site keeps its URL; the server retries the icon after saving.
    }
    if (stale())
      return
    if (found) {
      iconStatus.value = 'found'
    }
    else if (url === initialURL && (model.value.icon?.src || model.value.icon?.text)) {
      // Keep the saved icon when its own site cannot be reached right now.
      iconStatus.value = 'idle'
      if (force)
        ms.warning(t('iconItem.geticonFail'))
    }
    else {
      model.value.icon = null
      iconStatus.value = 'missing'
    }
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
  if (url !== lookupURL || !model.value.title)
    fillHostname(url)
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
  iconStatus.value = 'idle'
  if (url && url === initialURL && initialIcon && hasImageIcon(initialIcon)) {
    // Back to the saved address: show its saved icon again.
    model.value.icon = { ...initialIcon }
    return
  }
  model.value.icon = null
  if (model.value.description === autoDescription) {
    model.value.description = ''
    autoDescription = ''
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
  autoTitle = autoDescription = ''
  iconStatus.value = 'idle'
  if (visible) {
    model.value = props.itemInfo ? { ...props.itemInfo, icon: props.itemInfo.icon ? { ...props.itemInfo.icon } : null } : { ...restoreDefault }
    initialURL = normalizeURL(model.value.url || '')
    initialIcon = model.value.icon ? { ...model.value.icon } : null
    if (props.itemGroupId)
      model.value.itemIconGroupId = props.itemGroupId
    getGroupListOptions()
    // Bookmarks saved without a site icon look it up again when opened.
    if (initialURL && !hasImageIcon(initialIcon))
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
          <ItemIcon :item-icon="model.icon" :size="44" eager :fallback-text="model.title?.slice(0, 2)" />
          <span class="min-w-0 flex-1 text-sm opacity-70">
            {{ iconStatus === 'loading' ? $t('iconItem.iconDetecting') : iconStatus === 'found' ? $t('iconItem.iconDetected') : iconStatus === 'missing' ? $t('iconItem.iconMissing') : $t('iconItem.iconAuto') }}
          </span>
          <NButton text type="primary" :disabled="!model.url" :loading="iconStatus === 'loading'" @click="detectSite(model.url, true)">
            {{ $t('iconItem.redetect') }}
          </NButton>
        </div>
        <NGrid cols="2" :x-gap="10" item-responsive>
          <NGridItem span="2 500:1">
            <NFormItem path="title" :label="$t('common.title')">
              <NInput v-model:value="model.title" type="text" show-count :maxlength="20" />
            </NFormItem>
          </NGridItem>
          <NGridItem span="2 500:1">
            <NFormItem path="itemIconGroupId" :label="t('iconItem.iconGroup')">
              <NSelect v-model:value="model.itemIconGroupId" :options="itemIconGroupOptions" />
            </NFormItem>
          </NGridItem>
        </NGrid>
        <NFormItem path="description" :label="$t('common.description')">
          <NInput v-model:value="model.description" type="text" show-count :maxlength="100" />
        </NFormItem>
        <NFormItem path="lanUrl" :label="$t('iconItem.lanUrl')">
          <NInput v-model:value="model.lanUrl" type="text" :maxlength="1000" :placeholder="$t('iconItem.lanUrlInputPlaceholder')" />
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
