<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NGrid,
  NGi,
  NInput,
  NModal,
  NSelect,
  NSpace,
  NStatistic,
  NTag,
  NText,
  useMessage,
} from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import { adminEmailLogApi, type EmailLog, type EmailLogListParams } from '@/service/api/admin/email-log'

const message = useMessage()
const { t } = useI18n()

const loading = ref(false)
const logList = ref<EmailLog[]>([])
const total = ref(0)
const statsData = ref({ total: 0, success: 0, fail: 0 })
const templateNameOptions = ref<{ label: string; value: string }[]>([])

const query = reactive<EmailLogListParams>({
  page: 1,
  page_size: 20,
  to_email: '',
  template_name: undefined,
  status: -1,
  start_time: '',
  end_time: '',
})

const dateRange = ref<[number, number] | null>(null)
const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
})

const showDetail = ref(false)
const detailLoading = ref(false)
const detailData = ref<EmailLog | null>(null)

const showClean = ref(false)
const cleanBefore = ref('')
const cleaning = ref(false)

const statusOptions = [
  { label: t('adminEmailLogs.all'), value: -1 },
  { label: t('adminEmailLogs.success'), value: 1 },
  { label: t('adminEmailLogs.failed'), value: 0 },
]

const formattedContent = computed(() => {
  const raw = detailData.value?.content?.trim()
  if (!raw) return ''
  return raw
})

const columns: DataTableColumns<EmailLog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('adminEmailLogs.toEmail'), key: 'to_email', width: 220, ellipsis: { tooltip: true } },
  { title: t('adminEmailLogs.subject'), key: 'subject', ellipsis: { tooltip: true } },
  { title: t('adminEmailLogs.template'), key: 'template_name', width: 140, ellipsis: { tooltip: true } },
  {
    title: t('adminEmailLogs.status'),
    key: 'status',
    width: 90,
    render(row) {
      return h(NTag, { size: 'small', type: row.status === 1 ? 'success' : 'error' }, () => row.status === 1 ? t('adminEmailLogs.success') : t('adminEmailLogs.failed'))
    },
  },
  {
    title: t('adminEmailLogs.time'),
    key: 'created_at',
    width: 170,
    render(row) {
      return row.created_at ? new Date(row.created_at).toLocaleString() : '-'
    },
  },
  {
    title: t('adminEmailLogs.actions'),
    key: 'actions',
    width: 80,
    render(row) {
      return h(NButton, { text: true, type: 'primary', size: 'small', onClick: () => handleDetail(row.id) }, () => t('adminEmailLogs.detail'))
    },
  },
]

const selectableColumnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'to_email', label: t('adminEmailLogs.toEmail') },
  { key: 'subject', label: t('adminEmailLogs.subject') },
  { key: 'template_name', label: t('adminEmailLogs.template') },
  { key: 'status', label: t('adminEmailLogs.status') },
  { key: 'created_at', label: t('adminEmailLogs.time') },
]

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<EmailLog>({
  storageKey: 'admin-email-logs-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 960,
})

async function fetchList() {
  loading.value = true
  try {
    if (dateRange.value) {
      query.start_time = new Date(dateRange.value[0]).toISOString().slice(0, 19).replace('T', ' ')
      query.end_time = new Date(dateRange.value[1]).toISOString().slice(0, 19).replace('T', ' ')
    }
    else {
      query.start_time = ''
      query.end_time = ''
    }

    const params: Record<string, any> = { ...query }
    if (!params.to_email) delete params.to_email
    if (!params.template_name) delete params.template_name
    if (params.status === -1) delete params.status
    if (!params.start_time) delete params.start_time
    if (!params.end_time) delete params.end_time

    const res = await adminEmailLogApi.list(params)
    logList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = total.value
  }
  catch {
    message.error(t('adminEmailLogs.fetchListFailed'))
  }
  finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await adminEmailLogApi.stats()
    if (res.data) statsData.value = res.data
  }
  catch {}
}

async function fetchTemplateNames() {
  try {
    const res = await adminEmailLogApi.templateNames()
    templateNameOptions.value = [{ label: t('adminEmailLogs.all'), value: '' }, ...(res.data || []).map(item => ({ label: item, value: item }))]
  }
  catch {}
}

async function handleDetail(id: number) {
  showDetail.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res = await adminEmailLogApi.detail(id)
    detailData.value = res.data || null
  }
  catch {
    showDetail.value = false
    message.error(t('adminEmailLogs.loadDetailFailed'))
  }
  finally {
    detailLoading.value = false
  }
}

function handleSearch() {
  query.page = 1
  pagination.page = 1
  fetchList()
}

function handleReset() {
  query.to_email = ''
  query.template_name = undefined
  query.status = -1
  dateRange.value = null
  handleSearch()
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchList()
}

function handleCleanDateChange(val: number | null) {
  cleanBefore.value = val ? new Date(val).toISOString().slice(0, 19).replace('T', ' ') : ''
}

async function handleClean() {
  if (!cleanBefore.value) {
    message.warning(t('adminEmailLogs.selectCleanDate'))
    return
  }
  cleaning.value = true
  try {
    const res = await adminEmailLogApi.clean(cleanBefore.value)
    message.success(t('adminEmailLogs.cleanSuccess', { count: res.data?.affected || 0 }))
    showClean.value = false
    cleanBefore.value = ''
    fetchList()
    fetchStats()
  }
  catch {
    message.error(t('adminEmailLogs.cleanFailed'))
  }
  finally {
    cleaning.value = false
  }
}

onMounted(() => {
  fetchList()
  fetchStats()
  fetchTemplateNames()
})
</script>

<template>
  <div class="email-log-page">
    <NGrid :x-gap="12" :y-gap="12" cols="3" style="margin-bottom: 16px;">
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminEmailLogs.total')" :value="statsData.total" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminEmailLogs.success')">
            <template #default>
              <NText type="success">{{ statsData.success }}</NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic :label="t('adminEmailLogs.failed')">
            <template #default>
              <NText type="error">{{ statsData.fail }}</NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
    </NGrid>

    <NCard :title="t('adminEmailLogs.title')">
      <template #header-extra>
        <NSpace>
          <TableColumnSelector
            v-model="selectedColumnKeys"
            :options="columnOptions"
            :visible-count="visibleColumnCount"
            :total-count="totalColumnCount"
            :button-label="t('common.showFields')"
            :title="t('common.visibleFields')"
            :hint="t('common.columnVisibilityHint')"
            :reset-label="t('common.restoreDefaultFields')"
            @reset="resetSelectedColumns"
          />
          <NButton size="small" type="primary" @click="fetchList">{{ t('adminEmailLogs.refresh') }}</NButton>
          <NButton size="small" type="warning" @click="showClean = true">{{ t('adminEmailLogs.cleanLogs') }}</NButton>
        </NSpace>
      </template>

      <NSpace align="center" style="margin-bottom: 12px;" :wrap="true">
        <NInput v-model:value="query.to_email" :placeholder="t('adminEmailLogs.toEmail')" clearable size="small" style="width: 220px;" @keyup.enter="handleSearch" />
        <NSelect v-model:value="query.template_name" :options="templateNameOptions" :placeholder="t('adminEmailLogs.template')" clearable size="small" style="width: 160px;" />
        <NSelect v-model:value="query.status" :options="statusOptions" :placeholder="t('adminEmailLogs.status')" size="small" style="width: 100px;" />
        <NDatePicker v-model:value="dateRange" type="datetimerange" clearable size="small" style="width: 340px;" />
        <NButton size="small" type="primary" @click="handleSearch">{{ t('adminEmailLogs.search') }}</NButton>
        <NButton size="small" @click="handleReset">{{ t('adminEmailLogs.reset') }}</NButton>
      </NSpace>

      <NDataTable
        remote
        :columns="visibleColumns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="tableScrollX"
        :row-key="(row: EmailLog) => row.id"
        @update:page="handlePageChange"
      />
    </NCard>

    <NModal v-model:show="showDetail" preset="card" :title="t('adminEmailLogs.detailTitle')" style="width: 760px;" :mask-closable="true">
      <NText v-if="detailLoading" depth="3">{{ t('adminEmailLogs.loading') }}</NText>
      <NSpace v-else-if="detailData" vertical :size="16">
        <NGrid cols="2" :x-gap="12" :y-gap="12">
          <NGi>
            <NCard size="small" embedded>
              <NStatistic :label="t('adminEmailLogs.logId')" :value="detailData.id" />
            </NCard>
          </NGi>
          <NGi>
            <NCard size="small" embedded>
              <NStatistic :label="t('adminEmailLogs.sendStatus')">
                <template #default>
                  <NTag :type="detailData.status === 1 ? 'success' : 'error'" size="small">{{ detailData.status === 1 ? t('adminEmailLogs.success') : t('adminEmailLogs.failed') }}</NTag>
                </template>
              </NStatistic>
            </NCard>
          </NGi>
        </NGrid>

        <NCard size="small" embedded :title="t('adminEmailLogs.basicInfo')">
          <NDescriptions bordered :column="2" label-placement="left">
            <NDescriptionsItem :label="t('adminEmailLogs.toEmail')">{{ detailData.to_email }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminEmailLogs.templateName')">{{ detailData.template_name || '-' }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminEmailLogs.subject')" :span="2">{{ detailData.subject || '-' }}</NDescriptionsItem>
            <NDescriptionsItem :label="t('adminEmailLogs.sendTime')" :span="2">{{ detailData.created_at ? new Date(detailData.created_at).toLocaleString() : '-' }}</NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <NCard size="small" embedded :title="t('adminEmailLogs.content')">
          <div class="detail-content-block">{{ formattedContent || '-' }}</div>
        </NCard>

        <NCard v-if="detailData.error_msg" size="small" embedded :title="t('adminEmailLogs.errorMsg')">
          <NText type="error">{{ detailData.error_msg }}</NText>
        </NCard>
      </NSpace>
      <NText v-else depth="3">{{ t('adminEmailLogs.noDetailData') }}</NText>
    </NModal>

    <NModal v-model:show="showClean" preset="card" :title="t('adminEmailLogs.cleanModalTitle')" style="width: 400px;" :mask-closable="false">
      <NSpace vertical>
        <NText>{{ t('adminEmailLogs.cleanWarning') }}</NText>
        <NDivider style="margin: 8px 0;" />
        <NText depth="3">{{ t('adminEmailLogs.cleanBeforeLabel') }}</NText>
        <NDatePicker type="datetime" clearable style="width: 100%;" @update:value="handleCleanDateChange" />
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showClean = false">{{ t('common.cancel') }}</NButton>
          <NButton type="error" :loading="cleaning" :disabled="!cleanBefore" @click="handleClean">{{ t('adminEmailLogs.confirmClean') }}</NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.detail-content-block {
  padding: 12px;
  border-radius: 10px;
  background: rgb(250 250 252);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
