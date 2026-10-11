<script setup lang="ts">
import { computed, defineEmits, defineProps, onBeforeUnmount, ref, watch } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NForm, NFormItem, NGrid, NGridItem, NInput, NInputGroup, NModal, NSelect, useMessage } from 'naive-ui'
import IconEditor from './IconEditor.vue'
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
const getIconLoading = ref([false, false])
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

let lookupTimer: ReturnType<typeof setTimeout> | undefined
let generation = 0
let initialURL = ''
let autoTitle = ''
let autoDescription = ''
let autoIcon = ''
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

function fillHostname(url: string) {
  if (!model.value.title || model.value.title === autoTitle) {
    autoTitle = new URL(url).hostname.replace(/^www\./, '').slice(0, 20)
    model.value.title = autoTitle
  }
}

async function getIconByUrl(raw: string, loadingIndex: number, force = true) {
  const url = normalizeURL(raw)
  if (!url)
    return
  if (!force && lookupURL === url && lookupPromise)
    return lookupPromise
  const currentGeneration = generation
  const task = async () => {
    getIconLoading.value[loadingIndex] = true
    try {
      const { code, data } = await getSiteFavicon<{ iconUrl: string; originalIconUrl: string; title: string; description: string }>(url)
      const currentURL = loadingIndex === 0 ? model.value.url : model.value.lanUrl || ''
      if (!show.value || currentGeneration !== generation || normalizeURL(currentURL) !== url)
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
        if (data.iconUrl && (force || !model.value.icon || model.value.icon.src === autoIcon)) {
          autoIcon = data.originalIconUrl || data.iconUrl
          model.value.icon = { itemType: 2, src: autoIcon, cachedSrc: data.iconUrl }
        }
        else if (force && !data.iconUrl) {
          ms.warning(t('iconItem.geticonFail'))
        }
      }
    }
    finally {
      if (currentGeneration === generation && normalizeURL(loadingIndex === 0 ? model.value.url || '' : model.value.lanUrl || '') === url)
        getIconLoading.value[loadingIndex] = false
    }
  }
  // Keep a completed lookup reusable across debounce, blur and save.
  const promise = task().catch(() => {
    if (force && currentGeneration === generation && show.value)
      ms.warning(t('iconItem.geticonFail'))
  })
  if (loadingIndex === 0) {
    lookupURL = url
    lookupPromise = promise
  }
  return promise
}

function handleUrlBlur() {
  clearTimeout(lookupTimer)
  const url = normalizeURL(model.value.url || '')
  if (!url)
    return
  model.value.url = url
  if (url !== lookupURL || !model.value.title)
    fillHostname(url)
  if (url !== initialURL)
    return getIconByUrl(url, 0, false)
}

watch(() => model.value.url, (value) => {
  clearTimeout(lookupTimer)
  if (!show.value || normalizeURL(value || '') === initialURL || normalizeURL(value || '') === lookupURL)
    return
  if (model.value.icon?.src === autoIcon) {
    model.value.icon = null
    autoIcon = ''
  }
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
  autoTitle = autoDescription = autoIcon = ''
  getIconLoading.value = [false, false]
  if (visible) {
    model.value = props.itemInfo ? { ...props.itemInfo, icon: props.itemInfo.icon ? { ...props.itemInfo.icon } : null } : { ...restoreDefault }
    initialURL = normalizeURL(model.value.url || '')
    if (props.itemGroupId)
      model.value.itemIconGroupId = props.itemGroupId
    getGroupListOptions()
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
          <NInputGroup>
            <NInput v-model:value="model.url" type="text" :maxlength="1000" placeholder="输入或粘贴网站网址" @blur="handleUrlBlur" />
            <NButton :disabled="!model.url" :loading="getIconLoading[0]" @click="getIconByUrl(model.url, 0)">
              {{ $t('iconItem.getIcon') }}
            </NButton>
          </NInputGroup>
        </NFormItem>
        <p class="mb-5 text-sm opacity-70">
          只需输入网址，自动识别名称和图标。重复网址会更新原书签。
        </p>
        <NGrid cols="2" :x-gap="10" item-responsive>
          <NGridItem span="2 500:1">
            <NFormItem path="itemIconGroupId" :label="t('iconItem.iconGroup')">
              <NSelect v-model:value="model.itemIconGroupId" :options="itemIconGroupOptions" />
            </NFormItem>
          </NGridItem>
          <NGridItem span="2 500:1">
            <NFormItem path="title" :label="$t('common.title')">
              <NInput v-model:value="model.title" type="text" show-count :maxlength="20" />
            </NFormItem>
          </NGridItem>
        </NGrid>

        <NFormItem path="icon" :label="$t('common.icon')">
          <IconEditor v-model:item-icon="model.icon" />
        </NFormItem>
        <NFormItem path="lanUrl" :label="$t('iconItem.lanUrl')">
          <NInputGroup>
            <NInput v-model:value="model.lanUrl" type="text" :maxlength="1000" :placeholder="$t('iconItem.lanUrlInputPlaceholder')" />
            <NButton :disabled="!model.lanUrl" :loading="getIconLoading[1]" @click="getIconByUrl(model.lanUrl || '', 1)">
              {{ $t('iconItem.getIcon') }}
            </NButton>
          </NInputGroup>
        </NFormItem>
        <NFormItem path="description" :label="$t('common.description')">
          <NInput v-model:value="model.description" type="text" show-count :maxlength="100" />
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
