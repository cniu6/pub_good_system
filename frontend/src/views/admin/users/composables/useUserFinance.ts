/**
 * 编辑弹窗内：余额 / 积分 / 提现 tabs
 */
import { computed, h, reactive, ref, type Ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTag, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { adminUserApi } from '@/service/api/admin/user'
import type { AdminUser, UserSimpleInfo } from '@/service/api/admin/user'
import { adminApi } from '@/service/api/admin'
import type { MoneyOperationPayload, WithdrawRecord } from '@/service/api/admin/finance'
import {
  formatTime,
  getAdminDisplayName,
  getWithdrawStatusMeta,
  maskAccountNo,
} from '../utils/userDisplay'

function reportAdminUsersError(message: string, error?: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

export function useUserFinance(options: {
  selectedUser: Ref<AdminUser | null>
  submitting: Ref<boolean>
  /** 操作成功后：刷新列表并关弹窗 */
  onSuccess?: () => void
}) {
  const message = useMessage()
  const { t } = useI18n()

  const balanceForm = reactive({
    amount: 0,
    operation: 'balance_only',
    memo: '',
    orderNo: '',
    tradeNo: '',
    orderStatus: 1,
  })

  const scoreForm = reactive({
    amount: 0,
    operation: 'modify',
    memo: '',
  })

  const orderStatusOptions = computed(() => [
    { label: `${t('adminUsersDetail.pendingPayment')}(0)`, value: 0 },
    { label: `${t('adminUsersDetail.paid')}(1)`, value: 1 },
    { label: `${t('adminUsersDetail.cancelled')}(2)`, value: 2 },
    { label: `${t('adminUsersDetail.refunded')}(3)`, value: 3 },
    { label: `${t('adminUsersDetail.paymentFailed')}(4)`, value: 4 },
  ])

  const balanceAmountLabel = computed(() => {
    if (['log_only', 'log_order'].includes(balanceForm.operation))
      return t('adminUsers.logAmount')
    return t('adminUsers.amount')
  })

  const balanceAmountPlaceholder = computed(() => {
    if (['log_only', 'log_order'].includes(balanceForm.operation))
      return t('adminUsers.logAmountPlaceholder')
    return t('adminUsers.amountPlaceholder')
  })

  const withdrawLoading = ref(false)
  const withdrawData = ref<WithdrawRecord[]>([])
  const withdrawPagination = reactive({
    page: 1,
    pageSize: 10,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
  })
  const showWithdrawDetailModal = ref(false)
  const withdrawDetail = ref<WithdrawRecord | null>(null)
  const adminUserMap = ref<Record<number, UserSimpleInfo>>({})

  function resetForms() {
    balanceForm.amount = 0
    balanceForm.operation = 'balance_only'
    balanceForm.memo = ''
    balanceForm.orderNo = ''
    balanceForm.tradeNo = ''
    balanceForm.orderStatus = 1

    scoreForm.amount = 0
    scoreForm.operation = 'modify'
    scoreForm.memo = ''

    withdrawData.value = []
    withdrawPagination.page = 1
    withdrawPagination.pageSize = 10
    withdrawPagination.itemCount = 0
  }

  async function fetchWithdrawData() {
    if (!options.selectedUser.value)
      return

    withdrawLoading.value = true
    try {
      const response: any = await adminApi.finance.fetchWithdrawRecords({
        page: withdrawPagination.page,
        page_size: withdrawPagination.pageSize,
        user_id: options.selectedUser.value.id,
      })
      if (response.isSuccess) {
        withdrawData.value = response.data?.list || []
        withdrawPagination.itemCount = response.data?.total || 0
        const adminIds = Array.from(
          new Set(
            withdrawData.value.flatMap(item => [item.reviewed_by, item.paid_by]).filter(Boolean) as number[],
          ),
        )
        adminUserMap.value = await adminUserApi.batchSimpleInfo(adminIds)
      }
      else {
        message.error(response.message || t('moneyScore.fetchWithdrawFailed'))
      }
    }
    catch {
      message.error(t('moneyScore.fetchWithdrawFailed'))
    }
    finally {
      withdrawLoading.value = false
    }
  }

  function handleWithdrawPageChange(page: number) {
    withdrawPagination.page = page
    fetchWithdrawData()
  }

  function handleWithdrawPageSizeChange(pageSize: number) {
    withdrawPagination.pageSize = pageSize
    withdrawPagination.page = 1
    fetchWithdrawData()
  }

  function openWithdrawDetail(row: WithdrawRecord) {
    withdrawDetail.value = row
    showWithdrawDetailModal.value = true
  }

  const withdrawColumns: DataTableColumns<WithdrawRecord> = [
    { title: 'ID', key: 'id', width: 80 },
    {
      title: t('adminUsers.amount'),
      key: 'amount',
      width: 100,
      render: row => `¥${(Number(row.amount) || 0).toFixed(2)}`,
    },
    {
      title: t('adminUsers.status'),
      key: 'status',
      width: 100,
      render: (row) => {
        const status = getWithdrawStatusMeta(row.status)
        return h(NTag, { type: status.type }, () => status.label)
      },
    },
    { title: t('adminUsers.method'), key: 'account_type', width: 90 },
    { title: t('moneyScore.accountName'), key: 'account_name', width: 120, ellipsis: true },
    {
      title: t('moneyScore.accountNo'),
      key: 'account_no',
      width: 160,
      ellipsis: true,
      render: row => maskAccountNo(row.account_no),
    },
    {
      title: t('adminUsers.reviewer'),
      key: 'reviewed_by',
      width: 110,
      render: row => getAdminDisplayName(row.reviewed_by, adminUserMap.value),
    },
    {
      title: t('adminUsers.payer'),
      key: 'paid_by',
      width: 110,
      render: row => getAdminDisplayName(row.paid_by, adminUserMap.value),
    },
    {
      title: t('moneyScore.createdAt'),
      key: 'create_time',
      width: 170,
      render: row => formatTime(row.create_time),
    },
    {
      title: t('adminUsers.actions'),
      key: 'actions',
      width: 90,
      render: row => h(
        'a',
        {
          style: 'color: var(--n-primary-color); cursor: pointer;',
          onClick: () => openWithdrawDetail(row),
        },
        t('adminUsers.detail'),
      ),
    },
  ]

  async function handleBalanceOperation() {
    if (!options.selectedUser.value)
      return
    // 防重复提交
    if (options.submitting.value)
      return

    const isOrder = ['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)
    const needsAmount = balanceForm.operation !== 'order_only'

    // 校验放在加锁前，避免校验失败也占用 submitting
    if (needsAmount && balanceForm.amount === null) {
      message.warning(t('adminUsers.amountRequired'))
      return
    }
    if (needsAmount && Number(balanceForm.amount) === 0) {
      message.warning(t('adminUsers.amountCannotBeZero'))
      return
    }
    if (isOrder && !balanceForm.orderNo) {
      message.warning(t('adminUsers.orderNoRequired'))
      return
    }

    try {
      options.submitting.value = true

      const response: any = await adminApi.finance.operateUserMoney(options.selectedUser.value.id, {
        money: Number(balanceForm.amount || 0),
        memo: balanceForm.memo,
        operation: balanceForm.operation as MoneyOperationPayload['operation'],
        order_no: balanceForm.orderNo || undefined,
        trade_no: balanceForm.tradeNo || undefined,
        order_status: isOrder ? balanceForm.orderStatus : undefined,
      })

      if (response.isSuccess) {
        message.success(t('adminUsers.balanceOperationSuccess'))
        options.onSuccess?.()
      }
      else {
        message.error(response.message || t('adminUsers.balanceOperationFailed'))
      }
    }
    catch (error) {
      reportAdminUsersError('[adminUsers] balance operation failed', error)
      message.error(t('adminUsers.operationFailed'))
    }
    finally {
      options.submitting.value = false
    }
  }

  /**
   * 自动填充订单号或交易号（合并原 handleAutoFillOrderNo / handleAutoFillTradeNo）
   * @param field 'order' | 'trade'
   */
  async function handleAutoFillNo(field: 'order' | 'trade') {
    try {
      const res: any = await adminApi.finance.generateNos()
      if (!res.isSuccess) {
        message.error(res.message || (
          field === 'order' ? t('adminUsers.generateOrderNoFailed') : t('adminUsers.generateTradeNoFailed')
        ))
        return
      }
      if (field === 'order') {
        if (res.data?.order_no) {
          balanceForm.orderNo = res.data.order_no
          message.success(t('adminUsers.orderNoGenerated'))
        }
        else {
          message.error(t('adminUsers.generateOrderNoFailed'))
        }
      }
      else if (res.data?.trade_no) {
        balanceForm.tradeNo = res.data.trade_no
        message.success(t('adminUsers.tradeNoGenerated'))
      }
      else {
        message.error(t('adminUsers.generateTradeNoFailed'))
      }
    }
    catch {
      message.error(field === 'order' ? t('adminUsers.generateOrderNoFailed') : t('adminUsers.generateTradeNoFailed'))
    }
  }

  async function handleScoreOperation() {
    if (!options.selectedUser.value)
      return
    if (options.submitting.value)
      return

    if (Number(scoreForm.amount) === 0) {
      message.warning(t('adminUsers.scoreCannotBeZero'))
      return
    }

    try {
      options.submitting.value = true

      if (scoreForm.operation === 'modify') {
        const response: any = await adminUserApi.changeScore(options.selectedUser.value.id, {
          score: scoreForm.amount,
          memo: scoreForm.memo,
        })
        if (response.isSuccess) {
          message.success(t('adminUsers.scoreChangedSuccess'))
          options.onSuccess?.()
        }
        else {
          message.error(response.message || t('adminUsers.scoreChangedFailed'))
        }
      }
      else if (scoreForm.operation === 'log') {
        const response: any = await adminApi.finance.addScoreLog(options.selectedUser.value.id, {
          score: scoreForm.amount,
          memo: scoreForm.memo,
        })
        if (response.isSuccess) {
          message.success(t('adminUsers.scoreLogAddedSuccess'))
          options.onSuccess?.()
        }
        else {
          message.error(response.message || t('adminUsers.scoreLogAddedFailed'))
        }
      }
    }
    catch (error) {
      reportAdminUsersError('[adminUsers] score operation failed', error)
      message.error(t('adminUsers.operationFailed'))
    }
    finally {
      options.submitting.value = false
    }
  }

  return {
    balanceForm,
    scoreForm,
    orderStatusOptions,
    balanceAmountLabel,
    balanceAmountPlaceholder,
    withdrawLoading,
    withdrawData,
    withdrawPagination,
    withdrawColumns,
    showWithdrawDetailModal,
    withdrawDetail,
    adminUserMap,
    resetForms,
    fetchWithdrawData,
    handleWithdrawPageChange,
    handleWithdrawPageSizeChange,
    handleBalanceOperation,
    handleScoreOperation,
    handleAutoFillNo,
  }
}
