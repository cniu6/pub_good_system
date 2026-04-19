<template>
  <n-card :title="t('adminRealname.title')">
    <n-space vertical>
      <n-space align="center">
        <n-text depth="3">{{ t('adminRealname.totalRecords', { total }) }}</n-text>
        <n-divider vertical />
        <n-text depth="3">{{ t('adminRealname.statusFilter') }}</n-text>
        <n-select
          v-model:value="queryStatus"
          :options="realnameStatusOptions"
          :placeholder="t('adminRealname.allStatus')"
          clearable
          size="small"
          style="width: 120px"
          @update:value="handleStatusChange"
        />
        <n-text depth="3">{{ t('adminRealname.keyword') }}</n-text>
        <n-input
          v-model:value="searchKeyword"
          :placeholder="t('adminRealname.keywordPlaceholder')"
          clearable
          size="small"
          style="width: 160px"
          @update:value="handleKeywordChange"
        />
      </n-space>

      <n-space justify="end">
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
      </n-space>

      <n-data-table
        remote
        :columns="visibleColumns"
        :data="verificationList"
        :loading="loading"
        :pagination="pagination"
        :scroll-x="tableScrollX"
        @update:page="handlePageChange"
      />
    </n-space>

    <!-- 审核弹窗 -->
    <n-modal v-model:show="showReviewModal" preset="card" :title="t('adminRealname.reviewTitle')" style="width: 500px;">
      <n-form label-placement="left" label-width="100">
        <n-form-item :label="t('adminRealname.applicant')">
          <n-text>{{ currentVerification?.real_name || '-' }}</n-text>
        </n-form-item>
        <n-form-item :label="t('realname.certificateType')">
          <n-text>{{ getCertificateTypeText(currentVerification?.certificate_type) }}</n-text>
        </n-form-item>
        <n-form-item :label="t('realname.certificateNo')">
          <n-text>{{ currentVerification?.certificate_no || '-' }}</n-text>
        </n-form-item>
        <n-form-item :label="t('adminRealname.certificatePhotos')">
          <n-space>
            <n-image
              v-if="currentVerification?.certificate_front"
              :src="currentVerification.certificate_front"
              width="120"
              height="80"
              object-fit="cover"
            />
            <n-image
              v-if="currentVerification?.certificate_back"
              :src="currentVerification.certificate_back"
              width="120"
              height="80"
              object-fit="cover"
            />
          </n-space>
        </n-form-item>
        <n-form-item :label="t('adminRealname.reviewAction')">
          <n-radio-group v-model:value="reviewStatus">
            <n-space>
              <n-radio :value="1">{{ t('realname.approved') }}</n-radio>
              <n-radio :value="2">{{ t('realname.rejected') }}</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>
        <n-form-item v-if="reviewStatus === 2" :label="t('realname.rejectReason')">
          <n-input
            v-model:value="rejectReason"
            type="textarea"
            :placeholder="t('adminRealname.enterRejectReason')"
            :rows="3"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showReviewModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="submitting" @click="handleReview">{{ t('common.confirm') }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 详情弹窗 -->
    <n-modal v-model:show="showDetailModal" preset="card" :title="t('realname.detailTitle')" style="width: 600px;">
      <n-descriptions v-if="currentVerification" :column="1" label-placement="left" bordered>
        <n-descriptions-item label="ID">{{ currentVerification.id }}</n-descriptions-item>
        <n-descriptions-item :label="t('adminRealname.userId')">{{ currentVerification.user_id }}</n-descriptions-item>
        <n-descriptions-item :label="t('realname.realName')">{{ currentVerification.real_name }}</n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateType')">
          {{ getCertificateTypeText(currentVerification.certificate_type) }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateNo')">{{ currentVerification.certificate_no }}</n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateFront')">
          <n-image
            v-if="currentVerification.certificate_front"
            :src="currentVerification.certificate_front"
            width="200"
            height="130"
            object-fit="cover"
          />
          <span v-else>-</span>
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateBack')">
          <n-image
            v-if="currentVerification.certificate_back"
            :src="currentVerification.certificate_back"
            width="200"
            height="130"
            object-fit="cover"
          />
          <span v-else>-</span>
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.status')">
          <n-tag :type="getStatusType(currentVerification.status)">
            {{ getStatusText(currentVerification.status) }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.rejectReason')">{{ currentVerification.reject_reason || '-' }}</n-descriptions-item>
        <n-descriptions-item :label="t('realname.submittedAt')">
          {{ currentVerification.submitted_at ? new Date(currentVerification.submitted_at * 1000).toLocaleString() : '-' }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.reviewedAt')">
          {{ currentVerification.reviewed_at ? new Date(currentVerification.reviewed_at * 1000).toLocaleString() : '-' }}
        </n-descriptions-item>
      </n-descriptions>
    </n-modal>
  </n-card>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import {
  realnameStatusOptions,
  type RealnameVerification,
  type RealnameStatus,
} from '@/service/api/admin/realname'
import { adminApi } from '@/service/api/admin'
import type { UserSimpleInfo } from '@/service/api/admin/user'

const router = useRouter()
const message = useMessage()
const { t } = useI18n()
const loading = ref(false)
const verificationList = ref<RealnameVerification[]>([])
const userMap = ref<Record<number, UserSimpleInfo>>({})
const total = ref(0)
const searchKeyword = ref('')
const queryStatus = ref<RealnameStatus | null>(null)

const query = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  status: undefined as RealnameStatus | undefined,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
})

// 审核相关
const showReviewModal = ref(false)
const showDetailModal = ref(false)
const currentVerification = ref<RealnameVerification | null>(null)
const reviewStatus = ref<1 | 2>(1)
const rejectReason = ref('')
const submitting = ref(false)

// 跳转到用户详情页
function goToUserDetail(userId: number) {
  const adminPath = import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr'
  if (userId) {
    router.push(`${adminPath}/users/${userId}`)
  }
}

// 获取用户显示名称
function getUserDisplayName(userId: number): string {
  const user = userMap.value[userId]
  if (!user) return t('adminRealname.userFallback', { id: userId })
  return user.nickname || user.username || t('adminRealname.userFallback', { id: userId })
}

// 获取证件类型文本
function getCertificateTypeText(type_: number | undefined): string {
  const map: Record<number, string> = { 1: t('realname.idCard'), 2: t('realname.passport'), 3: t('realname.officer') }
  return type_ ? map[type_] || t('realname.unknown') : '-'
}

// 获取状态文本
function getStatusText(status: number | undefined): string {
  const map: Record<number, string> = { 0: t('realname.pending'), 1: t('realname.approved'), 2: t('realname.rejected') }
  return status !== undefined ? map[status] || t('realname.unknown') : '-'
}

// 获取状态颜色
function getStatusType(status: number | undefined): 'warning' | 'success' | 'error' {
  const map: Record<number, 'warning' | 'success' | 'error'> = { 0: 'warning', 1: 'success', 2: 'error' }
  return status !== undefined ? map[status] || 'warning' : 'warning'
}

const columns: DataTableColumns<RealnameVerification> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: t('adminRealname.user'),
    key: 'user_id',
    width: 120,
    render(row) {
      const userId = row.user_id
      if (!userId) return '-'
      const displayName = getUserDisplayName(userId)
      return h(
        NButton,
        {
          text: true,
          type: 'primary',
          onClick: () => goToUserDetail(userId),
        },
        { default: () => displayName },
      )
    },
  },
  { title: t('realname.realName'), key: 'real_name', width: 100 },
  {
    title: t('realname.certificateType'),
    key: 'certificate_type',
    width: 80,
    render(row) {
      return getCertificateTypeText(row.certificate_type)
    },
  },
  { title: t('realname.certificateNo'), key: 'certificate_no', width: 180, ellipsis: { tooltip: true } },
  {
    title: t('realname.status'),
    key: 'status',
    width: 80,
    render(row) {
      return h(NTag, { type: getStatusType(row.status), size: 'small' }, () => getStatusText(row.status))
    },
  },
  {
    title: t('realname.submittedAt'),
    key: 'submitted_at',
    width: 160,
    render(row) {
      return row.submitted_at ? new Date(row.submitted_at * 1000).toLocaleString() : '-'
    },
  },
  {
    title: t('adminRealname.actions'),
    key: 'actions',
    width: 150,
    render(row) {
      return h('div', { style: { display: 'flex', gap: '8px' } }, [
        h(
          NButton,
          {
            size: 'small',
            onClick: () => showDetail(row),
          },
          () => t('adminRealname.detail'),
        ),
        row.status === 0
          ? h(
              NButton,
              {
                size: 'small',
                type: 'primary',
                onClick: () => openReviewModal(row),
              },
              () => t('adminRealname.review'),
            )
          : null,
      ])
    },
  },
]

 const selectableColumnOptions = [
   { key: 'id', label: 'ID' },
   { key: 'user_id', label: t('adminRealname.user') },
   { key: 'real_name', label: t('realname.realName') },
   { key: 'certificate_type', label: t('realname.certificateType') },
   { key: 'certificate_no', label: t('realname.certificateNo') },
   { key: 'status', label: t('realname.status') },
   { key: 'submitted_at', label: t('realname.submittedAt') },
 ]

 const {
   columnOptions,
   selectedColumnKeys,
   visibleColumns,
   visibleColumnCount,
   totalColumnCount,
   tableScrollX,
   resetSelectedColumns,
 } = useTableColumnVisibility<RealnameVerification>({
   storageKey: 'admin-realname-list',
   columns,
   options: selectableColumnOptions,
   minVisibleCount: 1,
   minScrollX: 900,
 })

// 批量获取用户信息
async function fetchUserInfos(verifications: RealnameVerification[]) {
  const userIds = [...new Set(verifications.map((v) => v.user_id).filter(Boolean))]
  if (userIds.length === 0) return

  try {
    userMap.value = await adminApi.user.batchSimpleInfo(userIds as number[])
  } catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminRealname] fetch user infos failed', error)
  }
}

// 加载数据
async function loadData() {
  loading.value = true
  try {
    query.keyword = searchKeyword.value
    query.status = queryStatus.value ?? undefined
    query.page = pagination.page

    const res = await adminApi.realname.list(query)
    if (!res.isSuccess) {
      message.error(res.message || t('adminRealname.loadFailed'))
      verificationList.value = []
      total.value = 0
      pagination.itemCount = 0
      return
    }
    verificationList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = res.data?.total || 0

    await fetchUserInfos(verificationList.value)
  } catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminRealname] load data failed', error)
    message.error(t('adminRealname.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 分页变化
function handlePageChange(page: number) {
  pagination.page = page
  loadData()
}

// 状态筛选变化
function handleStatusChange() {
  pagination.page = 1
  loadData()
}

// 关键词搜索变化
let searchTimer: number
function handleKeywordChange() {
  clearTimeout(searchTimer)
  searchTimer = window.setTimeout(() => {
    pagination.page = 1
    loadData()
  }, 300)
}

// 显示详情
async function showDetail(row: RealnameVerification) {
  try {
    const res = await adminApi.realname.detail(row.id)
    if (res.isSuccess && res.data?.verification) {
      currentVerification.value = res.data.verification
      showDetailModal.value = true
    } else {
      message.error(res.message || t('adminRealname.loadDetailFailed'))
    }
  } catch {
    message.error(t('adminRealname.loadDetailFailed'))
  }
}

// 打开审核弹窗
function openReviewModal(row: RealnameVerification) {
  currentVerification.value = row
  reviewStatus.value = 1
  rejectReason.value = ''
  showReviewModal.value = true
}

// 提交审核
async function handleReview() {
  if (reviewStatus.value === 2 && !rejectReason.value.trim()) {
    message.warning(t('adminRealname.enterRejectReason'))
    return
  }

  submitting.value = true
  try {
    const res = await adminApi.realname.review({
      id: currentVerification.value!.id,
      status: reviewStatus.value,
      reject_reason: rejectReason.value,
    })
    if (!res.isSuccess) {
      message.error(res.message || t('adminRealname.reviewFailed'))
      return
    }
    message.success(t('adminRealname.reviewSuccess'))
    showReviewModal.value = false
    loadData()
  } catch (e: any) {
    message.error(e?.message || t('adminRealname.reviewFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
