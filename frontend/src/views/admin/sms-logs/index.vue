<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
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
import { adminSMSLogApi, type SMSLog, type SMSLogListParams } from '@/service/api/admin/sms-log'

const message = useMessage()

const loading = ref(false)
const logList = ref<SMSLog[]>([])
const total = ref(0)
const statsData = ref<{ total: number; success: number; fail: number }>({ total: 0, success: 0, fail: 0 })
const templateNameOptions = ref<{ label: string; value: string }[]>([])

const query = reactive<SMSLogListParams>({
  page: 1,
  page_size: 20,
  phone: '',
  provider: undefined,
  template_name: undefined,
  lang: undefined,
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
const detailData = ref<SMSLog | null>(null)
const detailLoading = ref(false)

const showClean = ref(false)
const cleanBefore = ref('')
const cleaning = ref(false)

const providerOptions = [
  { label: '全部', value: '' },
  { label: '阿里云', value: 'aliyun' },
  { label: '腾讯云', value: 'tencent' },
  { label: '自定义HTTP', value: 'custom' },
  { label: '控制台', value: 'console' },
]

const statusOptions = [
  { label: '全部', value: -1 },
  { label: '成功', value: 1 },
  { label: '失败', value: 0 },
]

const langOptions = [
  { label: '全部', value: '' },
  { label: '中文', value: 'zh-CN' },
  { label: 'English', value: 'en-US' },
]

const providerMap: Record<string, { label: string; type: 'info' | 'success' | 'warning' | 'default' }> = {
  aliyun: { label: '阿里云', type: 'info' },
  tencent: { label: '腾讯云', type: 'success' },
  custom: { label: '自定义', type: 'warning' },
  console: { label: '控制台', type: 'default' },
}

const detailStatusText = computed(() => detailData.value?.status === 1 ? '成功' : '失败')
const detailStatusType = computed(() => detailData.value?.status === 1 ? 'success' : 'error')
const formattedResponse = computed(() => {
  const raw = detailData.value?.response?.trim()
  if (!raw) return ''
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  }
  catch {
    return raw
  }
})

async function handleCopyResponse() {
  if (!formattedResponse.value) {
    message.warning('暂无可复制的响应内容')
    return
  }
  try {
    await navigator.clipboard.writeText(formattedResponse.value)
    message.success('响应内容已复制')
  }
  catch {
    message.error('复制失败')
  }
}

const columns: DataTableColumns<SMSLog> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: '手机号', key: 'phone', width: 130 },
  {
    title: '服务商',
    key: 'provider',
    width: 100,
    render(row) {
      const p = providerMap[row.provider]
      return h(NTag, { type: p?.type || 'default', size: 'small' }, () => p?.label || row.provider)
    },
  },
  { title: '模板', key: 'template_name', width: 120, ellipsis: { tooltip: true } },
  {
    title: '语言',
    key: 'lang',
    width: 80,
    render(row) {
      return row.lang === 'zh-CN' ? '中文' : row.lang === 'en-US' ? 'EN' : row.lang
    },
  },
  { title: '内容', key: 'content', ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row) {
      return h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () => row.status === 1 ? '成功' : '失败')
    },
  },
  {
    title: '时间',
    key: 'created_at',
    width: 160,
    render(row) {
      if (!row.created_at) return '-'
      return new Date(row.created_at).toLocaleString()
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render(row) {
      return h(NButton, { size: 'small', type: 'primary', text: true, onClick: () => handleDetail(row) }, () => '详情')
    },
  },
]

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
    const params: any = { ...query }
    if (!params.phone) delete params.phone
    if (!params.provider) delete params.provider
    if (!params.template_name) delete params.template_name
    if (!params.lang) delete params.lang
    if (params.status === -1) delete params.status
    if (!params.start_time) delete params.start_time
    if (!params.end_time) delete params.end_time

    const res = await adminSMSLogApi.list(params)
    logList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = total.value
  }
  catch {
    message.error('查询短信日志失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await adminSMSLogApi.stats()
    if (res.data) statsData.value = res.data
  }
  catch { /* ignore */ }
}

async function fetchTemplateNames() {
  try {
    const res = await adminSMSLogApi.templateNames()
    if (res.data) {
      templateNameOptions.value = [
        { label: '全部', value: '' },
        ...res.data.map(n => ({ label: n, value: n })),
      ]
    }
  }
  catch { /* ignore */ }
}

async function handleDetail(row: SMSLog) {
  showDetail.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res = await adminSMSLogApi.detail(row.id)
    detailData.value = res.data || null
    if (!detailData.value) {
      message.warning('未获取到短信日志详情')
    }
  }
  catch {
    showDetail.value = false
    message.error('加载短信日志详情失败')
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
  query.phone = ''
  query.provider = undefined
  query.template_name = undefined
  query.lang = undefined
  query.status = -1
  dateRange.value = null
  handleSearch()
}

function handlePageChange(page: number) {
  query.page = page
  pagination.page = page
  fetchList()
}

async function handleClean() {
  if (!cleanBefore.value) {
    message.warning('请选择清理日期')
    return
  }
  cleaning.value = true
  try {
    const res = await adminSMSLogApi.clean(cleanBefore.value)
    if (res.data) {
      message.success(`已清理 ${res.data.affected} 条记录`)
      showClean.value = false
      cleanBefore.value = ''
      fetchList()
      fetchStats()
    }
  }
  catch {
    message.error('清理失败')
  }
  finally {
    cleaning.value = false
  }
}

function handleCleanDateChange(val: number | null) {
  if (val) {
    cleanBefore.value = new Date(val).toISOString().slice(0, 19).replace('T', ' ')
  }
  else {
    cleanBefore.value = ''
  }
}

onMounted(() => {
  fetchList()
  fetchStats()
  fetchTemplateNames()
})
</script>

<template>
  <div class="sms-log-page">
    <!-- 统计卡片 -->
    <NGrid :x-gap="12" :y-gap="12" cols="3" style="margin-bottom: 16px;">
      <NGi>
        <NCard size="small">
          <NStatistic label="总发送" :value="statsData.total" />
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic label="成功">
            <template #default>
              <NText type="success">{{ statsData.success }}</NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
      <NGi>
        <NCard size="small">
          <NStatistic label="失败">
            <template #default>
              <NText type="error">{{ statsData.fail }}</NText>
            </template>
          </NStatistic>
        </NCard>
      </NGi>
    </NGrid>

    <!-- 列表 -->
    <NCard title="短信日志">
      <template #header-extra>
        <NSpace>
          <NButton size="small" type="primary" :loading="loading" @click="fetchList">刷新</NButton>
          <NButton size="small" type="warning" @click="showClean = true">清理日志</NButton>
        </NSpace>
      </template>

      <!-- 筛选 -->
      <NSpace align="center" style="margin-bottom: 12px;" :wrap="true">
        <NInput v-model:value="query.phone" placeholder="手机号" clearable size="small" style="width: 140px;" @keyup.enter="handleSearch" />
        <NSelect v-model:value="query.provider" :options="providerOptions" placeholder="服务商" clearable size="small" style="width: 120px;" />
        <NSelect v-model:value="query.template_name" :options="templateNameOptions" placeholder="模板" clearable size="small" style="width: 140px;" />
        <NSelect v-model:value="query.lang" :options="langOptions" placeholder="语言" clearable size="small" style="width: 100px;" />
        <NSelect v-model:value="query.status" :options="statusOptions" placeholder="状态" size="small" style="width: 90px;" />
        <NDatePicker v-model:value="dateRange" type="datetimerange" clearable size="small" style="width: 340px;" />
        <NButton size="small" type="primary" @click="handleSearch">搜索</NButton>
        <NButton size="small" @click="handleReset">重置</NButton>
      </NSpace>

      <NDataTable
        remote
        :columns="columns"
        :data="logList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: SMSLog) => row.id"
        @update:page="handlePageChange"
      />
    </NCard>

    <!-- 详情弹窗 -->
    <NModal v-model:show="showDetail" preset="card" title="短信日志详情" style="width: 760px;" :mask-closable="true">
      <NText v-if="detailLoading" depth="3">加载中...</NText>
      <NSpace v-else-if="detailData" vertical :size="16">
        <NGrid cols="2" :x-gap="12" :y-gap="12">
          <NGi>
            <NCard size="small" embedded>
              <NStatistic label="日志ID" :value="detailData.id" />
            </NCard>
          </NGi>
          <NGi>
            <NCard size="small" embedded>
              <NStatistic label="发送状态">
                <template #default>
                  <NTag :type="detailStatusType" size="small">{{ detailStatusText }}</NTag>
                </template>
              </NStatistic>
            </NCard>
          </NGi>
        </NGrid>

        <NCard size="small" embedded title="基础信息">
          <NDescriptions bordered :column="2" label-placement="left">
            <NDescriptionsItem label="手机号">{{ detailData.phone }}</NDescriptionsItem>
            <NDescriptionsItem label="服务商">
              <NTag :type="providerMap[detailData.provider]?.type || 'default'" size="small">
                {{ providerMap[detailData.provider]?.label || detailData.provider }}
              </NTag>
            </NDescriptionsItem>
            <NDescriptionsItem label="模板ID">{{ detailData.template_code || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="模板名称">{{ detailData.template_name || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="语言">{{ detailData.lang || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="请求ID">{{ detailData.request_id || '-' }}</NDescriptionsItem>
            <NDescriptionsItem label="发送时间" :span="2">
              {{ detailData.created_at ? new Date(detailData.created_at).toLocaleString() : '-' }}
            </NDescriptionsItem>
          </NDescriptions>
        </NCard>

        <NCard size="small" embedded title="发送内容">
          <div class="detail-content-block">
            {{ detailData.content || '-' }}
          </div>
        </NCard>

        <NCard v-if="detailData.error_msg" size="small" embedded title="错误信息">
          <NText type="error">{{ detailData.error_msg }}</NText>
        </NCard>

        <NCard v-if="formattedResponse" size="small" embedded title="完整响应">
          <template #header-extra>
            <NButton size="small" quaternary @click="handleCopyResponse">复制内容</NButton>
          </template>
          <div class="detail-response-block">{{ formattedResponse }}</div>
        </NCard>
      </NSpace>
      <NText v-else depth="3">暂无详情数据</NText>
    </NModal>

    <!-- 清理弹窗 -->
    <NModal v-model:show="showClean" preset="card" title="清理短信日志" style="width: 400px;" :mask-closable="false">
      <NSpace vertical>
        <NText>删除指定日期之前的所有短信日志，此操作不可撤销。</NText>
        <NDivider style="margin: 8px 0;" />
        <NText depth="3">清理此日期之前的记录：</NText>
        <NDatePicker type="datetime" clearable style="width: 100%;" @update:value="handleCleanDateChange" />
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showClean = false">取消</NButton>
          <NButton type="error" :loading="cleaning" :disabled="!cleanBefore" @click="handleClean">确认清理</NButton>
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

.detail-response-block {
  max-height: 320px;
  overflow-y: auto;
  padding: 12px;
  border-radius: 10px;
  background: rgb(17 24 39);
  color: rgb(229 231 235);
  font-family: Consolas, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
