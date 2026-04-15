<template>
  <n-card>
    <n-tabs v-model:value="activeTab" type="line" animated>
      <n-tab-pane name="money" :tab="t('moneyScore.moneyRecords')">
        <n-space vertical>
          <n-space>
            <n-input v-model:value="moneyKeyword" :placeholder="t('moneyScore.searchRemark')" clearable style="width: 200px" @keyup.enter="fetchMoneyLogs" />
            <n-button type="primary" @click="fetchMoneyLogs">{{ t('moneyScore.search') }}</n-button>
            <n-button @click="moneyKeyword = ''; moneyPagination.page = 1; fetchMoneyLogs()">{{ t('moneyScore.reset') }}</n-button>
          </n-space>
          <n-data-table
            :columns="moneyColumns"
            :data="moneyLogs"
            :loading="moneyLoading"
            :pagination="moneyPagination"
            striped
            size="small"
            @update:page="(p: number) => { moneyPagination.page = p; fetchMoneyLogs() }"
            @update:page-size="(s: number) => { moneyPagination.pageSize = s; moneyPagination.page = 1; fetchMoneyLogs() }"
          />
        </n-space>
      </n-tab-pane>

      <n-tab-pane name="score" :tab="t('moneyScore.scoreRecords')">
        <n-space vertical>
          <n-space>
            <n-input v-model:value="scoreKeyword" :placeholder="t('moneyScore.searchRemark')" clearable style="width: 200px" @keyup.enter="fetchScoreLogs" />
            <n-button type="primary" @click="fetchScoreLogs">{{ t('moneyScore.search') }}</n-button>
            <n-button @click="scoreKeyword = ''; scorePagination.page = 1; fetchScoreLogs()">{{ t('moneyScore.reset') }}</n-button>
          </n-space>
          <n-data-table
            :columns="scoreColumns"
            :data="scoreLogs"
            :loading="scoreLoading"
            :pagination="scorePagination"
            striped
            size="small"
            @update:page="(p: number) => { scorePagination.page = p; fetchScoreLogs() }"
            @update:page-size="(s: number) => { scorePagination.pageSize = s; scorePagination.page = 1; fetchScoreLogs() }"
          />
        </n-space>
      </n-tab-pane>
    </n-tabs>
  </n-card>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { fetchMyMoneyLogs, fetchMyScoreLogs } from '@/service/api/user/user-center'
import { parseMemo } from '@/utils/memo'

const message = useMessage()
const { t } = useI18n()

const activeTab = ref('money')

const moneyLoading = ref(false)
const moneyKeyword = ref('')
const moneyLogs = ref<Entity.UserMoneyLog[]>([])
const moneyPagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const moneyColumns: DataTableColumns<Entity.UserMoneyLog> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('moneyScore.moneyChange'),
    key: 'money',
    width: 120,
    render: row => {
      const money = Number(row.money) || 0
      const isPositive = money > 0
      return h('span', {
        style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: '500' },
      }, `${isPositive ? '+' : ''}¥${money.toFixed(2)}`)
    },
  },
  {
    title: t('moneyScore.beforeChange'),
    key: 'before',
    width: 110,
    render: row => `¥${(Number(row.before) || 0).toFixed(2)}`,
  },
  {
    title: t('moneyScore.afterChange'),
    key: 'after',
    width: 110,
    render: row => `¥${(Number(row.after) || 0).toFixed(2)}`,
  },
  {
    title: t('moneyScore.remark'),
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: t('moneyScore.time'),
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

async function fetchMoneyLogs() {
  moneyLoading.value = true
  try {
    const res = await fetchMyMoneyLogs({
      page: moneyPagination.page,
      page_size: moneyPagination.pageSize,
      keyword: moneyKeyword.value || undefined,
    })
    if (res.isSuccess) {
      moneyLogs.value = res.data?.list || []
      moneyPagination.itemCount = res.data?.total || 0
    } else {
      message.error(res.message || t('moneyScore.fetchMoneyFailed'))
    }
  } catch {
    message.error(t('moneyScore.fetchMoneyFailed'))
  } finally {
    moneyLoading.value = false
  }
}

const scoreLoading = ref(false)
const scoreKeyword = ref('')
const scoreLogs = ref<Entity.UserScoreLog[]>([])
const scorePagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
})

const scoreColumns: DataTableColumns<Entity.UserScoreLog> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('moneyScore.scoreChange'),
    key: 'score',
    width: 120,
    render: row => {
      const score = Number(row.score) || 0
      const isPositive = score > 0
      return h('span', {
        style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: '500' },
      }, `${isPositive ? '+' : ''}${score}`)
    },
  },
  {
    title: t('moneyScore.beforeChange'),
    key: 'before',
    width: 100,
    render: row => `${Number(row.before) || 0}`,
  },
  {
    title: t('moneyScore.afterChange'),
    key: 'after',
    width: 100,
    render: row => `${Number(row.after) || 0}`,
  },
  {
    title: t('moneyScore.remark'),
    key: 'memo',
    ellipsis: { tooltip: true },
    render: row => parseMemo(row.memo),
  },
  {
    title: t('moneyScore.time'),
    key: 'create_time',
    width: 170,
    render: row => row.create_time ? new Date(row.create_time * 1000).toLocaleString() : '-',
  },
]

async function fetchScoreLogs() {
  scoreLoading.value = true
  try {
    const res = await fetchMyScoreLogs({
      page: scorePagination.page,
      page_size: scorePagination.pageSize,
      keyword: scoreKeyword.value || undefined,
    })
    if (res.isSuccess) {
      scoreLogs.value = res.data?.list || []
      scorePagination.itemCount = res.data?.total || 0
    } else {
      message.error(res.message || t('moneyScore.fetchScoreFailed'))
    }
  } catch {
    message.error(t('moneyScore.fetchScoreFailed'))
  } finally {
    scoreLoading.value = false
  }
}

watch(activeTab, (val) => {
  if (val === 'money' && moneyLogs.value.length === 0) fetchMoneyLogs()
  if (val === 'score' && scoreLogs.value.length === 0) fetchScoreLogs()
})

onMounted(() => fetchMoneyLogs())
</script>
