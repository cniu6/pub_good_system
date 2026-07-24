import { reactive, ref } from 'vue'
import { useRequestGuard, withSubmitLock } from '@/hooks'

// money-logs / score-logs 两个页面的「搜索 + 分页 + 拉取列表 + 删除确认」逻辑几乎完全一样
// （只是金额 vs 积分的字段/格式化不同），抽成通用 composable。
// 列定义、变更弹窗（handleAdd/handleSubmit）因为字段和接口不同，仍留在各自页面里。

export interface LedgerLogListParams {
  page: number
  page_size: number
  keyword?: string
  user_id?: number
}

export interface LedgerLogListResponse<T> {
  isSuccess: boolean
  message?: string
  data?: { list?: T[], total?: number }
}

export interface LedgerLogActionResponse {
  isSuccess: boolean
  message?: string
}

export interface UseLedgerLogPageOptions<T> {
  /** 拉取列表 */
  fetchList: (params: LedgerLogListParams) => Promise<LedgerLogListResponse<T>>
  /** 删除单条记录 */
  deleteItem: (id: number) => Promise<LedgerLogActionResponse>
  /** 拉取失败时的提示文案 */
  fetchErrorMessage: string
  /** 删除成功/失败提示文案 */
  deleteSuccessMessage: string
  deleteFailedMessage: string
  /** 删除二次确认弹窗文案 */
  deleteConfirmTitle: string
  deleteConfirmContent: string
}

export function useLedgerLogPage<T>(options: UseLedgerLogPageOptions<T>) {
  const message = useMessage()
  const dialog = useDialog()
  const { t } = useI18n()
  const listFetchGuard = useRequestGuard()

  const loading = ref(false)
  /** 删除写操作防连点 */
  const deleting = ref(false)
  const logList = ref<T[]>([])

  const searchForm = reactive({
    keyword: '',
    user_id: null as number | null,
  })

  const pagination = reactive({
    page: 1,
    pageSize: 20,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
  })

  async function fetchData() {
    const token = listFetchGuard.begin()
    loading.value = true
    try {
      const res = await options.fetchList({
        page: pagination.page,
        page_size: pagination.pageSize,
        keyword: searchForm.keyword || undefined,
        user_id: searchForm.user_id || undefined,
      })
      if (!listFetchGuard.isLatest(token))
        return
      if (res.isSuccess) {
        logList.value = res.data?.list || []
        pagination.itemCount = res.data?.total || 0
      }
      else {
        message.error(res.message || options.fetchErrorMessage)
      }
    }
    catch {
      if (listFetchGuard.isLatest(token))
        message.error(options.fetchErrorMessage)
    }
    finally {
      if (listFetchGuard.isLatest(token))
        loading.value = false
    }
  }

  function handleSearch() {
    pagination.page = 1
    fetchData()
  }

  function handleReset() {
    searchForm.keyword = ''
    searchForm.user_id = null
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

  function handleDelete(id: number) {
    if (deleting.value)
      return
    dialog.warning({
      title: options.deleteConfirmTitle,
      content: options.deleteConfirmContent,
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: () => withSubmitLock(deleting, async () => {
        try {
          const res = await options.deleteItem(id)
          if (res.isSuccess) {
            message.success(res.message || options.deleteSuccessMessage)
            fetchData()
            return
          }
          message.error(res.message || options.deleteFailedMessage)
          return false
        }
        catch {
          message.error(options.deleteFailedMessage)
          return false
        }
      }),
    })
  }

  return {
    loading,
    logList,
    searchForm,
    pagination,
    fetchData,
    handleSearch,
    handleReset,
    handlePageChange,
    handlePageSizeChange,
    handleDelete,
  }
}
