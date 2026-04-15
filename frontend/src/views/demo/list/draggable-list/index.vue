<script setup lang="tsx">
import type { DataTableColumns, FormInst, NDataTable } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { Gender } from '@/constants'
import { useBoolean } from '@/hooks'
import { useTableDrag } from '@/hooks/useTableDrag'
import { fetchUserPage } from '@/service'
import { NButton, NPopconfirm, NSpace, NSwitch, NTag } from 'naive-ui'

const { t } = useI18n()

const { bool: loading, setTrue: startLoading, setFalse: endLoading } = useBoolean(false)

const initialModel = {
  condition_1: '',
  condition_2: '',
  condition_3: '',
  condition_4: '',
}
const model = ref({ ...initialModel })

const formRef = ref<FormInst | null>()
void formRef.value
function sendMail(id?: number) {
  window.$message.success(`${t('demo.list.deleteUserId')}${id}`)
}
const columns: DataTableColumns<Entity.User> = [
  {
    title: t('demo.list.name'),
    align: 'center',
    key: 'userName',
  },
  {
    title: t('demo.list.age'),
    align: 'center',
    key: 'age',
  },
  {
    title: t('demo.list.gender'),
    align: 'center',
    key: 'gender',
    render: (row) => {
      const tagType = {
        0: 'primary',
        1: 'success',
        2: 'warning',
      } as const
      if (row.gender !== undefined) {
        return (
          <NTag type={tagType[row.gender]}>
            {Gender[row.gender]}
          </NTag>
        )
      }
    },
  },
  {
    title: t('demo.list.email'),
    align: 'center',
    key: 'email',
  },
  {
    title: t('demo.list.status'),
    align: 'center',
    key: 'status',
    render: (row) => {
      return (
        <NSwitch
          value={row.status}
          checked-value={1}
          unchecked-value={0}
          onUpdateValue={(value: 0 | 1) =>
            handleUpdateDisabled(value, row.id!)}
        >
          {{ checked: () => t('common.enable'), unchecked: () => t('common.disable') }}
        </NSwitch>
      )
    },
  },
  {
    title: t('demo.list.action'),
    align: 'center',
    key: 'actions',
    render: (row) => {
      return (
        <NSpace justify="center">
          <NPopconfirm onPositiveClick={() => sendMail(row.id)}>
            {{
              default: () => t('common.confirmDelete'),
              trigger: () => <NButton size="small">{t('common.delete')}</NButton>,
            }}
          </NPopconfirm>
        </NSpace>
      )
    },
  },
]

const listData = ref<Entity.User[]>([])
function handleUpdateDisabled(value: 0 | 1, id: number) {
  const index = listData.value.findIndex(item => item.id === id)
  if (index > -1)
    listData.value[index].status = value
}

const tableRef = ref<InstanceType<typeof NDataTable>>()
useTableDrag({
  tableRef,
  data: listData,
  onRowDrag(data) {
    const target = data[data.length - 1]
    window.$message.success(`${t('demo.list.dragData')} id: ${target.id} name: ${target.userName}`)
  },
})

onMounted(() => {
  getUserList()
})
async function getUserList() {
  startLoading()
  await fetchUserPage().then((res: any) => {
    listData.value = res.data.list
    endLoading()
  })
}
function changePage(page: number, size: number) {
  window.$message.success(`${t('demo.list.pagination')}:${page},${size}`)
}
function handleResetSearch() {
  model.value = { ...initialModel }
}
</script>

<template>
  <NSpace vertical size="large">
    <n-card>
      <n-form ref="formRef" :model="model" label-placement="left" inline :show-feedback="false">
        <n-flex>
          <n-form-item :label="t('demo.list.name')" path="condition_1">
            <n-input v-model:value="model.condition_1" :placeholder="t('common.inputPlaceholder')" />
          </n-form-item>
          <n-form-item :label="t('demo.list.age')" path="condition_2">
            <n-input v-model:value="model.condition_2" :placeholder="t('common.inputPlaceholder')" />
          </n-form-item>
          <n-form-item :label="t('demo.list.gender')" path="condition_3">
            <n-input v-model:value="model.condition_3" :placeholder="t('common.inputPlaceholder')" />
          </n-form-item>
          <n-form-item :label="t('demo.list.address')" path="condition_4">
            <n-input v-model:value="model.condition_4" :placeholder="t('common.inputPlaceholder')" />
          </n-form-item>
          <n-flex class="ml-auto">
            <NButton type="primary" @click="getUserList">
              <template #icon>
                <icon-park-outline-search />
              </template>
              {{ t('moneyScore.search') }}
            </NButton>
            <NButton strong secondary @click="handleResetSearch">
              <template #icon>
                <icon-park-outline-redo />
              </template>
              {{ t('common.reset') }}
            </NButton>
          </n-flex>
        </n-flex>
      </n-form>
    </n-card>
    <n-card>
      <NSpace vertical size="large">
        <n-data-table ref="tableRef" row-class-name="drag-handle" :columns="columns" :data="listData" :loading="loading" />
        <Pagination :count="100" @change="changePage" />
      </NSpace>
    </n-card>
  </NSpace>
</template>
