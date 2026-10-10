<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import { NButton, NCard, NForm, NFormItem, NInput, NSelect, NSwitch, useDialog, useMessage } from 'naive-ui'
import { VueDraggable } from 'vue-draggable-plus'
import { deletes, edit, getList, saveSort } from '@/api/panel/itemIconGroup'
import { RoundCardModal, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { parseGroupCardStyle } from '@/utils/groupCardStyle'

interface EditModalArg {
  show: boolean
  editStatus: number // 1.添加 2.编辑
  model: Panel.ItemIconGroup
  rules: FormRules
}

const formRef = ref<FormInst | null>(null)
const ms = useMessage()
const dialog = useDialog()
const sortStatus = ref(false)

const defaultMNodal = {
  title: '',
  icon: 'material-symbols:folder-outline',
  sort: 9999,
  publicVisitHide: 0,
  cardStyle: '',
}

const editModalArg = ref<EditModalArg>({
  show: false,
  editStatus: 1,
  model: defaultMNodal,
  rules: {
    title: [
      {
        required: true,
        trigger: 'blur',
        message: t('form.required'),
      },
    ],
  },
})

const groups = ref<Panel.ItemIconGroup[]>([])
const cardStyleForm = ref({ style: -1, textColor: '', textInfoHideDescription: false, textIconHideTitle: false })
const styleOptions = [
  { label: t('apps.itemGroupManage.styleGlobal'), value: -1 },
  { label: t('apps.itemGroupManage.styleInfo'), value: 0 },
  { label: t('apps.itemGroupManage.styleIcon'), value: 1 },
]

function loadCardStyle(group: Panel.ItemIconGroup) {
  cardStyleForm.value = { style: -1, textColor: '', textInfoHideDescription: false, textIconHideTitle: false, ...parseGroupCardStyle(group.cardStyle) }
}

function setPublicVisibility(value: boolean) {
  editModalArg.value.model.publicVisitHide = value ? 1 : 0
}

function handleAddGroup() {
  editModalArg.value.model = { ...defaultMNodal }
  editModalArg.value.editStatus = 1
  loadCardStyle(editModalArg.value.model)
  editModalArg.value.show = true
}

function handleEditGroup(groupInfo: Panel.ItemIconGroup) {
  editModalArg.value.show = true
  editModalArg.value.model = { ...groupInfo }
  loadCardStyle(groupInfo)
  editModalArg.value.editStatus = 2
}

function handleDragSort() {
  sortStatus.value = true
}

function handleSaveSort() {
  const saveItems: Common.SortItemRequest[] = []
  for (let i = 0; i < groups.value.length; i++) {
    const element = groups.value[i]
    saveItems.push({
      id: element.id as number,
      sort: i + 1,
    })
  }
  saveSort(saveItems).then(({ code, msg }) => {
    if (code === 0) {
      ms.success(t('common.saveSuccess'))
      sortStatus.value = false
    }
    else {
      ms.error(`${t('common.saveFail')}:${msg}`)
    }
  })
}

function handleDelete(groupInfo: Panel.ItemIconGroup) {
  dialog.warning({
    title: t('common.warning'),
    content: t('apps.itemGroupManage.deleteWarnText', { name: groupInfo.title }),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      if (groupInfo.id) {
        deletes([groupInfo.id]).then(({ code, msg }) => {
          if (code !== 0)
            ms.error(t('common.deleteFail'))
          else
            refreshList()
        })
      }
    },

  })
}

function handleSaveGroup() {
  formRef.value?.validate((errors) => {
    if (!errors) {
      edit({ ...editModalArg.value.model, cardStyle: cardStyleForm.value.style === -1 ? '' : JSON.stringify(cardStyleForm.value) }).then(({ code, msg }) => {
        if (code !== 0) {
          ms.error(msg)
          return
        }
        refreshList()
        editModalArg.value.show = false
        editModalArg.value.model = { ...defaultMNodal }
      })
    }
    else { console.log(errors) }
  })
}

function refreshList() {
  getList<Common.ListResponse<Panel.ItemIconGroup[]>>().then(({ code, data }) => {
    groups.value = data.list
  })
}

onMounted(() => {
  refreshList()
})
</script>

<template>
  <div class="h-full">
    <div class="p-2">
      <NButton type="primary" size="small" style="margin-right: 10px;" @click="handleAddGroup">
        {{ $t('common.add') }}
      </NButton>

      <NButton v-if="!sortStatus" size="small" @click="handleDragSort">
        {{ $t('common.sort') }}
      </NButton>

      <NButton v-else type="warning" size="small" @click="handleSaveSort">
        {{ $t('common.saveSort') }}
      </NButton>
    </div>

    <div class=" overflow-auto w-full mt-[20px] bg-stone-100 dark:bg-zinc-900/60 rounded-xl" style="height:calc(100% - 65px)">
      <VueDraggable
        v-model="groups"
        item-key="sort" :animation="300"
        :style="{ padding: sortStatus ? '20px' : '10px' }"
        :disabled="!sortStatus"
      >
        <div v-for="(item, index) in groups" :key="index" class="w-full">
          <NCard size="small" style="border-radius:10px;margin-bottom: 10px;">
            <div class="flex" :class="sortStatus ? 'cursor-move' : ''">
              <div class="flex items-center">
                <span class="mr-[10px]">
                  <SvgIcon class="text-[20px] usagi-action-icon" icon="material-symbols:ad-group-outline-rounded" />
                  <!-- <SvgIcon class="text-[20px]" :icon="item.icon" /> -->
                </span>
                <span>
                  {{ item.title }}
                </span>
              </div>
              <div class="ml-auto">
                <span>
                  <NButton strong secondary type="primary" size="small" @click="handleEditGroup(item)">
                    <template #icon>
                      <SvgIcon icon="basil:edit-solid" />
                    </template>
                  </NButton>
                </span>
                <span class="ml-[10px]">
                  <NButton strong secondary type="error" size="small" class="ml-[10px]" @click="handleDelete(item)">
                    <template #icon>
                      <SvgIcon icon="material-symbols:delete" />
                    </template>
                  </NButton>
                </span>
              </div>
            </div>
          </NCard>
        </div>
      </VueDraggable>
    </div>

    <RoundCardModal v-model:show="editModalArg.show" size="small" type="small" :title="editModalArg.editStatus === 1 ? '添加' : '编辑'" style="width: 400px;">
      <NForm ref="formRef" :model="editModalArg.model" :rules="editModalArg.rules">
        <NFormItem path="title" :label="$t('apps.itemGroupManage.groupName')">
          <NInput v-model:value="editModalArg.model.title" type="text" :maxlength="20" show-count />
        </NFormItem>

        <NFormItem :label="$t('apps.itemGroupManage.hidePublic')">
          <NSwitch :value="editModalArg.model.publicVisitHide === 1" @update:value="setPublicVisibility" />
        </NFormItem>
        <NFormItem :label="$t('apps.itemGroupManage.cardStyle')">
          <NSelect v-model:value="cardStyleForm.style" :options="styleOptions" />
        </NFormItem>
        <template v-if="cardStyleForm.style !== -1">
          <NFormItem :label="$t('apps.itemGroupManage.textColor')">
            <NInput v-model:value="cardStyleForm.textColor" placeholder="#ffffff" />
          </NFormItem>
          <NFormItem :label="$t('apps.itemGroupManage.hideDescription')">
            <NSwitch v-model:value="cardStyleForm.textInfoHideDescription" />
          </NFormItem>
          <NFormItem :label="$t('apps.itemGroupManage.hideTitle')">
            <NSwitch v-model:value="cardStyleForm.textIconHideTitle" />
          </NFormItem>
        </template>
      </NForm>
      <template #footer>
        <NButton type="primary" size="small" class="float-right" @click="handleSaveGroup">
          {{ $t('common.confirm') }}
        </NButton>
      </template>
    </RoundCardModal>
  </div>
</template>
