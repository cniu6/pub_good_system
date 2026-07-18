<script setup lang="ts">
/**
 * 用户管理列表页：搜索 + 表格 + 挂载编辑/提现弹窗
 * 详情统一跳转 detail.vue，不再弹窗
 */
import { onActivated, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import NovaIcon from '@/components/common/NovaIcon.vue'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useUserList } from './composables/useUserList'
import { useUserForm } from './composables/useUserForm'
import { useUserFinance } from './composables/useUserFinance'
import UserFormModal from './components/UserFormModal.vue'
import WithdrawDetailModal from './components/WithdrawDetailModal.vue'
import { adminUserApi } from '@/service/api/admin/user'
import type { AdminUser } from '@/service/api/admin/user'
import { takePendingUserEditId } from './utils/pendingEdit'

const message = useMessage()
const { t } = useI18n()

/** 延后绑定：form/finance 回调里需要 list.fetchData */
const listApi = { fetchData: () => {} }

const form = useUserForm({
  onSuccess: () => listApi.fetchData(),
})

const finance = useUserFinance({
  selectedUser: form.selectedUser,
  submitting: form.submitting,
  onSuccess: () => {
    listApi.fetchData()
    form.showUserModal.value = false
  },
})

// 打开编辑时：重置财务表单并拉提现
const originalHandleEdit = form.handleEdit
function handleEdit(user: AdminUser) {
  finance.resetForms()
  form.activeTab.value = 'details'
  originalHandleEdit(user)
  finance.fetchWithdrawData()
}

const {
  showUserModal,
  isEdit,
  submitting,
  formRef,
  selectedUser,
  activeTab,
  isFullscreen,
  userForm,
  rules,
  passwordRule,
  roleOptions,
  userStatusOptions,
  genderOptions,
  languageOptions,
  handleAdd,
  handleSubmit,
  toggleFullscreen,
  handleAvatarError,
} = form

const {
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
  fetchWithdrawData,
  handleWithdrawPageChange,
  handleWithdrawPageSizeChange,
  handleBalanceOperation,
  handleScoreOperation,
  handleAutoFillNo,
} = finance

const {
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
} = useUserList({
  onEdit: handleEdit,
})

listApi.fetchData = fetchData

/** 防止重复打开 */
let openingPendingEdit = false

/**
 * 详情页「编辑用户」通过 sessionStorage 传递意图。
 * 不能用 ?edit=：布局里 component :key="route.fullPath"，query 一变就会销毁重建，弹窗必丢。
 */
async function tryOpenPendingEdit() {
  if (openingPendingEdit || showUserModal.value)
    return

  const editId = takePendingUserEditId()
  if (!editId)
    return

  openingPendingEdit = true
  try {
    let user = userData.value.find(item => Number(item.id) === editId) || null
    if (!user) {
      const res: any = await adminUserApi.detail(editId)
      if (res?.isSuccess && res.data?.user)
        user = res.data.user as AdminUser
    }

    if (!user) {
      message.error(t('adminUsersDetail.fetchUserFailed'))
      return
    }

    handleEdit(user)
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[adminUsers] open pending edit failed', error)
    message.error(t('adminUsersDetail.fetchUserFailed'))
  }
  finally {
    openingPendingEdit = false
  }
}

onMounted(() => {
  void tryOpenPendingEdit()
})

// keep-alive 场景：从详情返回可能走 activated 而不是重新 mount
onActivated(() => {
  void tryOpenPendingEdit()
})

function handleUserModalTabChange(tab: string) {
  if (tab === 'withdraw' && isEdit.value)
    fetchWithdrawData()
}

function setFormRef(el: any) {
  formRef.value = el
}
</script>

<template>
  <div>
    <n-card class="header-card" :bordered="false">
      <div class="header-content">
        <div class="header-title">
          <NovaIcon :size="24" class="title-icon" icon="icon-park-outline:user" />
          <span>{{ t('adminUsers.userManagement') }}</span>
        </div>
        <NSpace :wrap="false" :size="12" class="header-actions">
          <NButton @click="handleRefresh">
            <template #icon>
              <NovaIcon icon="icon-park-outline:refresh" />
            </template>
            {{ t('common.reload') }}
          </NButton>
          <NButton type="primary" @click="handleAdd">
            <template #icon>
              <NovaIcon icon="icon-park-outline:plus" />
            </template>
            {{ t('adminUsers.addUser') }}
          </NButton>
        </NSpace>
      </div>
    </n-card>

    <n-card class="search-card" :bordered="false">
      <n-form :model="searchForm" label-placement="left" :label-width="80">
        <n-grid :cols="24" :x-gap="16" responsive="screen">
          <n-form-item-gi span="24 600:10 800:10" :label="t('adminRealname.keyword')">
            <n-input
              v-model:value="searchForm.keyword"
              :placeholder="t('adminUsers.searchPlaceholder')"
              clearable
              @keyup.enter="handleSearch"
            />
          </n-form-item-gi>
          <n-form-item-gi span="24 600:6 800:6" :label="t('adminUsers.realnameStatus')">
            <n-select
              v-model:value="searchForm.realnameStatus"
              :options="realnameFilterOptions"
              clearable
              :placeholder="t('common.all')"
              @update:value="handleRealnameStatusChange"
            />
          </n-form-item-gi>
          <n-form-item-gi span="24 600:8 800:8" class="search-actions">
            <NSpace justify="center">
              <NButton type="primary" class="search-btn" @click="handleSearch">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:search" />
                </template>
                {{ t('moneyScore.search') }}
              </NButton>
              <NButton class="reset-btn" @click="handleReset">
                <template #icon>
                  <NovaIcon icon="icon-park-outline:refresh" />
                </template>
                {{ t('common.reset') }}
              </NButton>
            </NSpace>
          </n-form-item-gi>
        </n-grid>
      </n-form>
    </n-card>

    <n-card class="table-card" :bordered="false">
      <NSpace justify="end" style="margin-bottom: 12px;">
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
      </NSpace>
      <n-data-table
        :columns="visibleColumns"
        :data="userData"
        :loading="loading"
        :pagination="false"
        :row-key="(row) => row.id"
        :scrollbar-props="{ trigger: 'hover' }"
        :scroll-x="tableScrollX"
      />
      <div class="pagination-container">
        <div class="pagination-info">
          <n-text depth="3">
            {{ t('adminUsers.paginationInfo', { total: pagination.itemCount, page: pagination.page, pageSize: pagination.pageSize }) }}
          </n-text>
        </div>
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :item-count="pagination.itemCount"
          :page-sizes="pagination.pageSizes"
          :show-size-picker="pagination.showSizePicker"
          show-quick-jumper
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </div>
    </n-card>

    <UserFormModal
      v-model:show="showUserModal"
      v-model:active-tab="activeTab"
      :is-edit="isEdit"
      :is-fullscreen="isFullscreen"
      :submitting="submitting"
      :user-form="userForm"
      :rules="rules"
      :password-rule="passwordRule"
      :selected-user="selectedUser"
      :role-options="roleOptions"
      :user-status-options="userStatusOptions"
      :gender-options="genderOptions"
      :language-options="languageOptions"
      :balance-form="balanceForm"
      :score-form="scoreForm"
      :order-status-options="orderStatusOptions"
      :balance-amount-label="balanceAmountLabel"
      :balance-amount-placeholder="balanceAmountPlaceholder"
      :withdraw-columns="withdrawColumns"
      :withdraw-data="withdrawData"
      :withdraw-loading="withdrawLoading"
      :withdraw-pagination="withdrawPagination"
      @set-form-ref="setFormRef"
      @toggle-fullscreen="toggleFullscreen"
      @submit="handleSubmit"
      @avatar-error="handleAvatarError"
      @balance-operation="handleBalanceOperation"
      @score-operation="handleScoreOperation"
      @auto-fill-no="handleAutoFillNo"
      @withdraw-page-change="handleWithdrawPageChange"
      @withdraw-page-size-change="handleWithdrawPageSizeChange"
      @tab-change="handleUserModalTabChange"
    />

    <WithdrawDetailModal
      v-model:show="showWithdrawDetailModal"
      :detail="withdrawDetail"
      :admin-user-map="adminUserMap"
    />
  </div>
</template>

<style scoped>
.header-card {
  margin-bottom: 16px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #ffffff;
}

.title-icon {
  color: #ffffff;
}

.search-card {
  margin-bottom: 16px;
}

.search-actions {
  display: flex;
  align-items: flex-end;
}

.table-card {
  min-height: 400px;
}

.pagination-container {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 12px 0;
  border-top: 1px solid var(--n-border-color);
}

.pagination-info {
  display: flex;
  align-items: center;
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
    justify-content: space-between;
  }

  .search-card :deep(.n-grid) {
    grid-template-columns: repeat(12, 1fr) !important;
  }

  .search-card :deep(.n-form-item-gi) {
    grid-column: span 12 !important;
  }

  .search-actions {
    margin-top: 8px;
  }

  .search-btn,
  .reset-btn {
    flex: 1;
  }

  .table-card :deep(.n-data-table .n-data-table__td--last) .n-space {
    flex-wrap: wrap;
    gap: 8px;
  }
}

@media (max-width: 480px) {
  .header-title {
    font-size: 16px;
  }

  .table-card :deep(.n-data-table .n-data-table__td--last) .n-space {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    width: 100%;
  }
}
</style>
