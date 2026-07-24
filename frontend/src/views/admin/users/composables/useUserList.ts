/**
 * 用户列表：搜索 / 分页 / 表格列 / 拉取 / 删除 / 登录为用户
 */
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import { useRequestGuard, useTableColumnVisibility, withSubmitLock } from '@/hooks'
import {
  deleteUser,
  fetchAdminUserPage,
  loginAsUser,
  openLoginAsUserWindow,
} from '@/service/api/admin/user'
import type { AdminUser, LoginAsAuthGuard } from '@/service/api/admin/user'
import { buildAdminEntryUrl } from '@/router/constants'
import {
  formatCurrency,
  formatLanguage,
  formatRechargeRetentionRatio,
  getRealnameStatusText,
  getRealnameStatusType,
} from '../utils/userDisplay'

export function useUserList(options?: {
  /** 编辑按钮回调（打开编辑弹窗） */
  onEdit?: (user: AdminUser) => void
}) {
  const route = useRoute()
  const router = useRouter()
  const message = useMessage()
  const dialog = useDialog()
  const { t } = useI18n()

  const searchForm = reactive({
    keyword: '',
    realnameStatus: null as 0 | 1 | 2 | null,
  })

  const realnameFilterOptions = computed(() => [
    { label: t('realname.pending'), value: 0 },
    { label: t('realname.approved'), value: 1 },
    { label: t('realname.rejected'), value: 2 },
  ])

  const pagination = reactive({
    page: 1,
    pageSize: 10,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
  })

  const userData = ref<AdminUser[]>([])
  const loading = ref(false)
  /** 删除等危险写操作防连点 */
  const actionLock = ref(false)

  /** 跳转详情页（唯一入口，不再弹窗） */
  function handleViewUserDetail(user: AdminUser) {
    router.push({ name: 'admin-user-detail', params: { id: user.id } })
  }

  function isAdminRole(role?: string) {
    return String(role || '').trim().toLowerCase() === 'admin'
  }

  async function doLoginAsUser(user: AdminUser, authGuard: LoginAsAuthGuard) {
    await withSubmitLock(actionLock, async () => {
      const res: any = await loginAsUser(user.id, { auth_guard: authGuard })
      if (!(res.isSuccess && res.data?.user && res.data?.token))
        return

      const targetUrl = authGuard === 'admin'
        ? `${buildAdminEntryUrl('/dashboard')}?_t=${Date.now()}`
        : `/user/dashboard?_t=${Date.now()}`

      openLoginAsUserWindow(res.data.user, res.data.token, res.data.refreshToken, res.data.expiresAt, targetUrl, authGuard)
      message.success(
        authGuard === 'admin'
          ? t('adminUsers.openedAdminConsole')
          : t('adminUsers.openedUserConsole'),
      )
    })
  }

  function confirmAdminLoginTarget(user: AdminUser) {
    const d = dialog.create({
      type: 'warning',
      title: t('adminUsers.confirmAdminLoginTitle'),
      content: t('adminUsers.confirmAdminLoginContent', { username: user.username }),
      closable: true,
      maskClosable: true,
      action: () => h(NSpace, { justify: 'end' }, () => [
        h(NButton, { size: 'small', onClick: () => d.destroy() }, () => t('common.cancel')),
        h(NButton, {
          size: 'small',
          type: 'info',
          disabled: actionLock.value,
          onClick: () => {
            d.destroy()
            void doLoginAsUser(user, 'user')
          },
        }, () => t('adminUsers.loginAsUserFrontend')),
        h(NButton, {
          size: 'small',
          type: 'warning',
          disabled: actionLock.value,
          onClick: () => {
            d.destroy()
            void doLoginAsUser(user, 'admin')
          },
        }, () => t('adminUsers.loginAsAdminConsole')),
      ]),
    })
  }

  function handleLoginAsUser(user: AdminUser) {
    if (actionLock.value)
      return
    dialog.warning({
      title: t('adminUsers.confirmLoginTitle'),
      content: t('adminUsers.confirmLoginContent', { username: user.username }),
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: () => {
        if (isAdminRole(user.role)) {
          confirmAdminLoginTarget(user)
          return
        }
        void doLoginAsUser(user, 'user')
      },
    })
  }

  function handleDelete(userId: number) {
    if (actionLock.value)
      return
    dialog.warning({
      title: t('adminUsers.confirmDeleteTitle'),
      content: t('adminUsers.confirmDeleteContent'),
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: () => withSubmitLock(actionLock, async () => {
        try {
          const response: any = await deleteUser(userId)
          if (response.isSuccess) {
            message.success(t('adminUsers.deleteSuccess'))
            fetchData()
            return
          }
          message.error(response.message || t('adminUsers.deleteFailed'))
          return false
        }
        catch {
          message.error(t('adminUsers.deleteFailed'))
          return false
        }
      }),
    })
  }

  const columns: DataTableColumns<AdminUser> = [
    { title: 'ID', key: 'id', width: 80 },
    { title: t('adminUsers.username'), key: 'username', width: 120 },
    { title: t('adminUsers.nickname'), key: 'nickname', width: 120 },
    {
      title: t('adminUsers.userGroup'),
      key: 'group_id',
      width: 80,
      render: (row: AdminUser) => row.group_id?.toString() || '-',
    },
    { title: t('adminUsers.email'), key: 'email', width: 200, ellipsis: true },
    { title: t('adminUsers.mobile'), key: 'mobile', width: 120 },
    {
      title: t('adminUsers.language'),
      key: 'language',
      width: 100,
      render: (row: AdminUser) => formatLanguage(row.language),
    },
    {
      title: t('adminUsers.role'),
      key: 'role',
      width: 100,
      render: (row: AdminUser) => {
        const roleMap: Record<string, { type: string, label: string }> = {
          admin: { type: 'error', label: t('adminUsers.admin') },
          user: { type: 'success', label: t('adminUsers.user') },
        }
        const role = roleMap[row.role] || { type: 'default', label: row.role }
        return h(NTag, { type: role.type as any }, () => role.label)
      },
    },
    { title: t('adminUsers.level'), key: 'level', width: 80 },
    {
      title: t('adminUsers.balance'),
      key: 'money',
      width: 100,
      render: (row: AdminUser) => `¥${formatCurrency(row.money)}`,
    },
    {
      title: t('adminUsers.rechargeRetentionRatio'),
      key: 'balance_paid_ratio',
      width: 130,
      render: (row: AdminUser) => formatRechargeRetentionRatio(row),
    },
    {
      title: t('adminUsers.score'),
      key: 'score',
      width: 80,
      render: (row: AdminUser) => (Number(row.score) || 0).toString(),
    },
    {
      title: t('adminUsers.status'),
      key: 'status',
      width: 100,
      render: (row: AdminUser) => {
        const statusMap: Record<string, { type: string, text: string }> = {
          1: { type: 'success', text: t('adminUsers.enabled') },
          0: { type: 'error', text: t('adminUsers.disabled') },
        }
        const status = statusMap[String(row.status)] || { type: 'default', text: t('adminUsers.unknown') }
        return h(NTag, { type: status.type as any }, { default: () => status.text })
      },
    },
    {
      title: t('adminUsers.realnameStatus'),
      key: 'realname_status',
      width: 110,
      render: (row: AdminUser) => {
        const status = row.realname_status
        return h(
          NTag,
          { type: getRealnameStatusType(status === null ? undefined : status), size: 'small' },
          { default: () => getRealnameStatusText(status === null ? undefined : status) },
        )
      },
    },
    {
      title: t('adminUsers.adminRemark'),
      key: 'admin_remark',
      width: 220,
      ellipsis: { tooltip: true },
      render: (row: AdminUser) => row.admin_remark || '-',
    },
    {
      // 「最近登录」用于快速判断用户上次活跃时间，无需进详情页；在线设备详情见详情页「安全」Tab
      title: t('adminUsers.lastLoginTime'),
      key: 'last_login_time',
      width: 180,
      render: (row: AdminUser) => {
        if (!row.last_login_time)
          return '-'
        try {
          return new Date(row.last_login_time * 1000).toLocaleString()
        }
        catch {
          return row.last_login_time as any
        }
      },
    },
    {
      // 「上次在线」来自会话心跳（user_sessions.last_seen_at），比「最近登录」更贴近真实活跃时间
      title: t('adminUsers.lastSeenAt'),
      key: 'last_seen_at',
      width: 190,
      render: (row: AdminUser) => {
        if (!row.last_seen_at)
          return '-'
        let text = ''
        try {
          text = new Date(row.last_seen_at * 1000).toLocaleString()
        }
        catch {
          text = String(row.last_seen_at)
        }
        return h(NSpace, { size: 4, align: 'center' }, () => [
          h(NTag, { size: 'small', type: row.is_online ? 'success' : 'default' }, () => row.is_online ? t('adminUsers.online') : t('adminUsers.offline')),
          h('span', text),
        ])
      },
    },
    {
      title: t('adminUsers.registerTime'),
      key: 'create_time',
      width: 180,
      render: (row: AdminUser) => {
        if (!row.create_time)
          return '-'
        try {
          return new Date(row.create_time * 1000).toLocaleString()
        }
        catch {
          return row.create_time as any
        }
      },
    },
    {
      title: t('adminUsers.updateTime'),
      key: 'update_time',
      width: 180,
      render: (row: AdminUser) => {
        if (!row.update_time)
          return '-'
        try {
          return new Date(row.update_time * 1000).toLocaleString()
        }
        catch {
          return row.update_time as any
        }
      },
    },
    {
      title: t('adminUsers.actions'),
      key: 'actions',
      width: 250,
      render: (row: AdminUser) => h(NSpace, { size: 'small' }, () => [
        h(NButton, {
          size: 'small',
          type: 'info',
          onClick: () => handleViewUserDetail(row),
        }, () => t('adminUsers.detail')),
        h(NButton, {
          size: 'small',
          type: 'primary',
          onClick: () => options?.onEdit?.(row),
        }, () => t('adminUsers.edit')),
        h(NButton, {
          size: 'small',
          type: 'success',
          onClick: () => handleLoginAsUser(row),
        }, () => t('adminUsers.loginAs')),
        h(NButton, {
          size: 'small',
          type: 'error',
          onClick: () => handleDelete(row.id),
        }, () => t('adminUsers.delete')),
      ]),
    },
  ]

  const selectableColumnOptions = [
    { key: 'id', label: 'ID' },
    { key: 'username', label: t('adminUsers.username') },
    { key: 'nickname', label: t('adminUsers.nickname') },
    { key: 'group_id', label: t('adminUsers.userGroup') },
    { key: 'email', label: t('adminUsers.email') },
    { key: 'mobile', label: t('adminUsers.mobile') },
    { key: 'language', label: t('adminUsers.language') },
    { key: 'role', label: t('adminUsers.role') },
    { key: 'level', label: t('adminUsers.level') },
    { key: 'money', label: t('adminUsers.balance') },
    { key: 'balance_paid_ratio', label: t('adminUsers.rechargeRetentionRatio') },
    { key: 'score', label: t('adminUsers.score') },
    { key: 'status', label: t('adminUsers.status') },
    { key: 'realname_status', label: t('adminUsers.realnameStatus') },
    { key: 'admin_remark', label: t('adminUsers.adminRemark') },
    { key: 'last_login_time', label: t('adminUsers.lastLoginTime') },
    { key: 'last_seen_at', label: t('adminUsers.lastSeenAt') },
    { key: 'create_time', label: t('adminUsers.registerTime') },
    { key: 'update_time', label: t('adminUsers.updateTime') },
  ]

  const {
    columnOptions,
    selectedColumnKeys,
    visibleColumns,
    visibleColumnCount,
    totalColumnCount,
    tableScrollX,
    resetSelectedColumns,
  } = useTableColumnVisibility<AdminUser>({
    storageKey: 'admin-users-list',
    columns,
    options: selectableColumnOptions,
    minVisibleCount: 1,
    minScrollX: 1280,
  })

  const { begin: beginFetch, isLatest: isLatestFetch } = useRequestGuard()

  async function fetchData() {
    const token = beginFetch()
    loading.value = true
    try {
      const response: any = await fetchAdminUserPage({
        page: pagination.page,
        page_size: pagination.pageSize,
        keyword: searchForm.keyword || undefined,
        realname_status: searchForm.realnameStatus ?? undefined,
      })
      if (!isLatestFetch(token))
        return

      if (response.isSuccess) {
        const data = response.data
        if (data && data.list) {
          userData.value = data.list || []
          pagination.itemCount = data.total || response.total || 0
        }
        else {
          userData.value = []
          pagination.itemCount = response.total || 0
        }
      }
      else {
        message.error(response.message || t('adminUsers.fetchListFailed'))
      }
    }
    catch {
      if (isLatestFetch(token))
        message.error(t('adminUsers.fetchListFailed'))
    }
    finally {
      if (isLatestFetch(token))
        loading.value = false
    }
  }

  function handleRefresh() {
    fetchData()
  }

  function handleSearch() {
    pagination.page = 1
    fetchData()
  }

  function handleReset() {
    Object.assign(searchForm, { keyword: '', realnameStatus: null })
    pagination.page = 1
    fetchData()
  }

  function handlePageChange(page: number) {
    pagination.page = page
    fetchData()
  }

  function handlePageSizeChange(pageSize: number) {
    pagination.pageSize = pageSize
    pagination.page = 1
    fetchData()
  }

  function handleRealnameStatusChange() {
    pagination.page = 1
    fetchData()
  }

  onMounted(() => {
    const searchKeyword = route.query.search as string
    if (searchKeyword)
      searchForm.keyword = searchKeyword
    fetchData()
  })

  return {
    searchForm,
    realnameFilterOptions,
    pagination,
    userData,
    loading,
    columnOptions,
    selectedColumnKeys,
    visibleColumns,
    visibleColumnCount,
    totalColumnCount,
    tableScrollX,
    resetSelectedColumns,
    fetchData,
    handleRefresh,
    handleSearch,
    handleReset,
    handlePageChange,
    handlePageSizeChange,
    handleRealnameStatusChange,
  }
}
