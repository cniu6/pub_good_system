<template>
  <n-card title="实名认证管理">
    <n-space vertical>
      <n-space align="center">
        <n-text depth="3">共 {{ total }} 条记录</n-text>
        <n-divider vertical />
        <n-text depth="3">状态筛选</n-text>
        <n-select
          v-model:value="queryStatus"
          :options="realnameStatusOptions"
          placeholder="全部状态"
          clearable
          size="small"
          style="width: 120px"
          @update:value="handleStatusChange"
        />
        <n-text depth="3">关键词</n-text>
        <n-input
          v-model:value="searchKeyword"
          placeholder="姓名/证件号"
          clearable
          size="small"
          style="width: 160px"
          @update:value="handleKeywordChange"
        />
      </n-space>

      <n-data-table
        remote
        :columns="columns"
        :data="verificationList"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
      />
    </n-space>

    <!-- 审核弹窗 -->
    <n-modal v-model:show="showReviewModal" preset="card" title="审核实名认证" style="width: 500px;">
      <n-form label-placement="left" label-width="100">
        <n-form-item label="申请人">
          <n-text>{{ currentVerification?.real_name || '-' }}</n-text>
        </n-form-item>
        <n-form-item label="证件类型">
          <n-text>{{ getCertificateTypeText(currentVerification?.certificate_type) }}</n-text>
        </n-form-item>
        <n-form-item label="证件号码">
          <n-text>{{ currentVerification?.certificate_no || '-' }}</n-text>
        </n-form-item>
        <n-form-item label="证件照片">
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
        <n-form-item label="审核操作">
          <n-radio-group v-model:value="reviewStatus">
            <n-space>
              <n-radio :value="1">通过</n-radio>
              <n-radio :value="2">拒绝</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>
        <n-form-item v-if="reviewStatus === 2" label="拒绝原因">
          <n-input
            v-model:value="rejectReason"
            type="textarea"
            placeholder="请填写拒绝原因"
            :rows="3"
          />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showReviewModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleReview">确认</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 详情弹窗 -->
    <n-modal v-model:show="showDetailModal" preset="card" title="实名认证详情" style="width: 600px;">
      <n-descriptions v-if="currentVerification" :column="1" label-placement="left" bordered>
        <n-descriptions-item label="ID">{{ currentVerification.id }}</n-descriptions-item>
        <n-descriptions-item label="用户ID">{{ currentVerification.user_id }}</n-descriptions-item>
        <n-descriptions-item label="真实姓名">{{ currentVerification.real_name }}</n-descriptions-item>
        <n-descriptions-item label="证件类型">
          {{ getCertificateTypeText(currentVerification.certificate_type) }}
        </n-descriptions-item>
        <n-descriptions-item label="证件号码">{{ currentVerification.certificate_no }}</n-descriptions-item>
        <n-descriptions-item label="证件正面">
          <n-image
            v-if="currentVerification.certificate_front"
            :src="currentVerification.certificate_front"
            width="200"
            height="130"
            object-fit="cover"
          />
          <span v-else>-</span>
        </n-descriptions-item>
        <n-descriptions-item label="证件背面">
          <n-image
            v-if="currentVerification.certificate_back"
            :src="currentVerification.certificate_back"
            width="200"
            height="130"
            object-fit="cover"
          />
          <span v-else>-</span>
        </n-descriptions-item>
        <n-descriptions-item label="状态">
          <n-tag :type="getStatusType(currentVerification.status)">
            {{ getStatusText(currentVerification.status) }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="拒绝原因">{{ currentVerification.reject_reason || '-' }}</n-descriptions-item>
        <n-descriptions-item label="提交时间">
          {{ currentVerification.submitted_at ? new Date(currentVerification.submitted_at * 1000).toLocaleString() : '-' }}
        </n-descriptions-item>
        <n-descriptions-item label="审核时间">
          {{ currentVerification.reviewed_at ? new Date(currentVerification.reviewed_at * 1000).toLocaleString() : '-' }}
        </n-descriptions-item>
      </n-descriptions>
    </n-modal>
  </n-card>
</template>

<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import { NButton, NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRouter } from 'vue-router'
import {
  realnameStatusOptions,
  type RealnameVerification,
  type RealnameStatus,
} from '@/service/api/admin/realname'
import { adminApi } from '@/service/api/admin'
import type { UserSimpleInfo } from '@/service/api/admin/user'

const router = useRouter()
const message = useMessage()
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
  if (!user) return `用户#${userId}`
  return user.nickname || user.username || `用户#${userId}`
}

// 获取证件类型文本
function getCertificateTypeText(type_: number | undefined): string {
  const map: Record<number, string> = { 1: '身份证', 2: '护照', 3: '军官证' }
  return type_ ? map[type_] || '未知' : '-'
}

// 获取状态文本
function getStatusText(status: number | undefined): string {
  const map: Record<number, string> = { 0: '待审核', 1: '已通过', 2: '已拒绝' }
  return status !== undefined ? map[status] || '未知' : '-'
}

// 获取状态颜色
function getStatusType(status: number | undefined): 'warning' | 'success' | 'error' {
  const map: Record<number, 'warning' | 'success' | 'error'> = { 0: 'warning', 1: 'success', 2: 'error' }
  return status !== undefined ? map[status] || 'warning' : 'warning'
}

const columns: DataTableColumns<RealnameVerification> = [
  { title: 'ID', key: 'id', width: 80 },
  {
    title: '用户',
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
  { title: '真实姓名', key: 'real_name', width: 100 },
  {
    title: '证件类型',
    key: 'certificate_type',
    width: 80,
    render(row) {
      return getCertificateTypeText(row.certificate_type)
    },
  },
  { title: '证件号码', key: 'certificate_no', width: 180, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render(row) {
      return h(NTag, { type: getStatusType(row.status), size: 'small' }, () => getStatusText(row.status))
    },
  },
  {
    title: '提交时间',
    key: 'submitted_at',
    width: 160,
    render(row) {
      return row.submitted_at ? new Date(row.submitted_at * 1000).toLocaleString() : '-'
    },
  },
  {
    title: '操作',
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
          () => '详情',
        ),
        row.status === 0
          ? h(
              NButton,
              {
                size: 'small',
                type: 'primary',
                onClick: () => openReviewModal(row),
              },
              () => '审核',
            )
          : null,
      ])
    },
  },
]

// 批量获取用户信息
async function fetchUserInfos(verifications: RealnameVerification[]) {
  const userIds = [...new Set(verifications.map((v) => v.user_id).filter(Boolean))]
  if (userIds.length === 0) return

  try {
    userMap.value = await adminApi.user.batchSimpleInfo(userIds as number[])
  } catch {
    console.error('Failed to fetch user infos')
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
      message.error(res.message || '加载失败')
      verificationList.value = []
      total.value = 0
      pagination.itemCount = 0
      return
    }
    verificationList.value = res.data?.list || []
    total.value = res.data?.total || 0
    pagination.itemCount = res.data?.total || 0

    await fetchUserInfos(verificationList.value)
  } catch (e) {
    console.error(e)
    message.error('加载失败')
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
      message.error(res.message || '加载详情失败')
    }
  } catch {
    message.error('加载详情失败')
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
    message.warning('请填写拒绝原因')
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
      message.error(res.message || '审核失败')
      return
    }
    message.success('审核成功')
    showReviewModal.value = false
    loadData()
  } catch (e: any) {
    message.error(e?.message || '审核失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
