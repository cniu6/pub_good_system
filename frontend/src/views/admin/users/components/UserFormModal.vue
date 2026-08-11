<script setup lang="ts">
/**
 * 新建/编辑用户弹窗（含详情 / 余额 / 积分 / 提现 tabs）
 * 状态与逻辑由父级 composable 传入，本组件只负责 UI
 */
import type { DataTableColumns, FormInst, FormRules } from 'naive-ui'
import type { AdminUser } from '@/service/api/admin/user'
import type { WithdrawRecord } from '@/service/api/admin/finance'
import NovaIcon from '@/components/common/NovaIcon.vue'
import PhoneInput from '@/components/common/PhoneInput.vue'
import { useSettingsStore } from '@/store'

/* eslint-disable vue/no-mutating-props -- 父组件传入共享表单对象，约定由本弹窗直接写入字段 */
defineProps<{
  show: boolean
  isEdit: boolean
  isFullscreen: boolean
  submitting: boolean
  activeTab: string
  userForm: Record<string, any>
  rules: FormRules
  passwordRule: any[]
  selectedUser: AdminUser | null
  roleOptions: { label: string, value: string }[]
  userStatusOptions: { label: string, value: number }[]
  genderOptions: { label: string, value: number }[]
  languageOptions: { label: string, value: string }[]
  balanceForm: Record<string, any>
  scoreForm: Record<string, any>
  orderStatusOptions: { label: string, value: number }[]
  balanceAmountLabel: string
  balanceAmountPlaceholder: string
  withdrawColumns: DataTableColumns<WithdrawRecord>
  withdrawData: WithdrawRecord[]
  withdrawLoading: boolean
  withdrawPagination: Record<string, any>
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'update:activeTab': [value: string]
  'setFormRef': [el: FormInst | null]
  'toggleFullscreen': []
  'submit': []
  'avatarError': []
  'balanceOperation': []
  'scoreOperation': []
  'autoFillNo': [field: 'order' | 'trade']
  'withdrawPageChange': [page: number]
  'withdrawPageSizeChange': [pageSize: number]
  'tabChange': [tab: string]
}>()

const settingsStore = useSettingsStore()
const mobileCnOnly = computed(() => settingsStore.mobileCnOnly)
function onTabChange(tab: string) {
  emit('update:activeTab', tab)
  emit('tabChange', tab)
}

function bindFormRef(el: any) {
  emit('setFormRef', el as FormInst | null)
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    :title="isEdit ? $t('adminUsers.editUser') : $t('adminUsers.addUser')"
    :style="isFullscreen ? 'width: 100vw; height: 100vh; max-width: none; max-height: none;' : 'width: 800px;'"
    :bordered="false"
    :closable="!isFullscreen"
    :mask-closable="!isFullscreen"
    @update:show="emit('update:show', $event)"
  >
    <template #header-extra>
      <NButton quaternary circle @click="emit('toggleFullscreen')">
        <template #icon>
          <NovaIcon :icon="isFullscreen ? 'icon-park-outline:off-screen' : 'icon-park-outline:full-screen'" />
        </template>
      </NButton>
    </template>

    <n-tabs :value="activeTab" type="line" animated @update:value="onTabChange">
      <n-tab-pane name="details" :tab="$t('adminUsers.detail')">
        <n-form
          :ref="bindFormRef"
          :model="userForm"
          :rules="rules"
          label-placement="left"
          :label-width="100"
        >
          <n-grid :cols="2" :x-gap="16">
            <n-form-item-gi :label="$t('adminUsers.username')" path="username">
              <n-input v-model:value="userForm.username" :placeholder="$t('adminUsers.enterUsername')" :disabled="isEdit" />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.nickname')" path="nickname">
              <n-input v-model:value="userForm.nickname" :placeholder="$t('adminUsers.enterNickname')" />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.email')" path="email">
              <n-input
                v-model:value="userForm.email"
                :placeholder="$t('adminUsers.enterEmail')"
                @blur="userForm.email = userForm.email.includes('@') ? `${userForm.email.split('@')[0]}@${userForm.email.split('@')[1].toLowerCase()}` : userForm.email"
              />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.mobile')" path="mobile">
              <PhoneInput
                v-model="userForm.mobile"
                :cn-only="mobileCnOnly"
                :auto-detect="!isEdit"
              />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.language')" path="language">
              <n-select v-model:value="userForm.language" :options="languageOptions" :placeholder="$t('adminUsers.selectLanguage')" />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.country')" path="country">
              <n-input v-model:value="userForm.country" :placeholder="$t('adminUsers.enterCountry')" />
            </n-form-item-gi>
            <n-form-item-gi span="2" :label="$t('adminUsers.adminRemark')" path="admin_remark">
              <n-input v-model:value="userForm.admin_remark" type="textarea" :placeholder="$t('adminUsers.enterAdminRemark')" :rows="3" />
            </n-form-item-gi>
            <n-form-item-gi v-if="!isEdit" span="2" :label="$t('adminUsers.password')" path="password" :rule="passwordRule">
              <n-input v-model:value="userForm.password" type="password" :placeholder="$t('adminUsers.enterNewPassword')" show-password-on="click" />
              <template #feedback>
                <span class="password-tip">{{ $t('adminUsers.setPasswordTip') }}</span>
              </template>
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.role')" path="role">
              <n-select v-model:value="userForm.role" :options="roleOptions" :placeholder="$t('adminUsers.selectRole')" />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.level')" path="level">
              <n-input-number v-model:value="userForm.level" :placeholder="$t('adminUsers.enterUserLevel')" :min="0" :max="100" />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.status')" path="status">
              <n-select v-model:value="userForm.status" :options="userStatusOptions" :placeholder="$t('adminUsers.selectStatus')" />
            </n-form-item-gi>
            <n-form-item-gi :label="$t('adminUsers.gender')" path="gender">
              <n-select v-model:value="userForm.gender" :options="genderOptions" :placeholder="$t('adminUsers.selectGender')" />
            </n-form-item-gi>
            <n-form-item-gi span="2" :label="$t('adminUsers.birthday')" path="birthday">
              <n-date-picker v-model:value="userForm.birthday" type="date" :placeholder="$t('adminUsers.selectBirthday')" clearable />
            </n-form-item-gi>
            <n-form-item-gi span="2" :label="$t('adminUsers.avatar')" path="avatar">
              <NSpace vertical>
                <n-input v-model:value="userForm.avatar" :placeholder="$t('adminUsers.enterAvatarUrl')" />
                <div v-if="userForm.avatar" class="avatar-preview">
                  <n-text depth="3" style="font-size: 12px;">
                    {{ $t('adminUsers.preview') }}:
                  </n-text>
                  <n-avatar
                    :src="userForm.avatar"
                    size="large"
                    fallback-src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjQiIGhlaWdodD0iNjQiIHZpZXdCb3g9IjAgMCA2NCA2NCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPGNpcmNsZSBjeD0iMzIiIGN5PSIzMiIgcj0iMzIiIGZpbGw9IiNGNUY1RjUiLz4KPHN2ZyB3aWR0aD0iMzIiIGhlaWdodD0iMzIiIHZpZXdCb3g9IjAgMCAzMiAzMiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4PSIxNiIgeT0iMTYiPgo8cGF0aCBkPSJNMTYgMTZDMTguMjA5MSAxNiAyMCAxNC4yMDkxIDIwIDEyQzIwIDkuNzkwODYgMTguMjA5MSA4IDE2IDhDMTMuNzkwOSA4IDEyIDkuNzkwODYgMTIgMTJDMTIgMTQuMjA5MSAxMy43OTA5IDE2IDE2IDE2WiIgZmlsbD0iIzk5OTk5OSIvPgo8cGF0aCBkPSJNMjQgMjRWMjJDMjQgMTkuNzkwOSAyMi4yMDkxIDE4IDIwIDE4SDEyQzkuNzkwODYgMTggOCAxOS43OTA5IDggMjJWMMjQiIGZpbGw9IiM5OTk5OTkiLz4KPC9zdmc+Cjwvc3ZnPgo="
                    @error="emit('avatarError')"
                  />
                </div>
              </NSpace>
            </n-form-item-gi>
            <n-form-item-gi span="2" :label="$t('adminUsers.motto')" path="motto">
              <n-input v-model:value="userForm.motto" type="textarea" :placeholder="$t('adminUsers.enterMotto')" :rows="3" />
            </n-form-item-gi>
          </n-grid>
        </n-form>
      </n-tab-pane>

      <n-tab-pane v-if="isEdit" name="balance" :tab="$t('adminUsers.balance')">
        <div class="balance-management">
          <NSpace vertical size="large">
            <n-card :title="$t('adminUsers.currentBalance')" size="small">
              <n-statistic :label="$t('adminUsers.userBalance')" :value="selectedUser?.money || 0" :precision="2">
                <template #prefix>
                  {{ settingsStore.currencySymbol }}
                </template>
              </n-statistic>
            </n-card>
            <n-form label-placement="left" :label-width="100">
              <n-form-item v-if="balanceForm.operation !== 'order_only'" :label="balanceAmountLabel">
                <n-input-number v-model:value="balanceForm.amount" :placeholder="balanceAmountPlaceholder" :precision="2" :step="0.01" />
              </n-form-item>
              <n-form-item :label="$t('adminUsers.operationType')">
                <n-radio-group v-model:value="balanceForm.operation">
                  <NSpace wrap>
                    <n-radio value="balance_only">
                      {{ $t('adminUsers.balanceOnly') }}
                    </n-radio>
                    <n-radio value="log_only">
                      {{ $t('adminUsers.logOnly') }}
                    </n-radio>
                    <n-radio value="order_only">
                      {{ $t('adminUsers.orderOnly') }}
                    </n-radio>
                    <n-radio value="balance_log">
                      {{ $t('adminUsers.balanceLog') }}
                    </n-radio>
                    <n-radio value="balance_order">
                      {{ $t('adminUsers.balanceOrder') }}
                    </n-radio>
                    <n-radio value="log_order">
                      {{ $t('adminUsers.logOrder') }}
                    </n-radio>
                    <n-radio value="both">
                      {{ $t('adminUsers.allInOne') }}
                    </n-radio>
                  </NSpace>
                </n-radio-group>
              </n-form-item>
              <n-form-item v-if="['log_only', 'balance_log', 'log_order', 'both'].includes(balanceForm.operation)" :label="$t('moneyScore.remark')">
                <n-input v-model:value="balanceForm.memo" type="textarea" :placeholder="$t('adminUsers.enterOperationRemark')" :rows="3" />
              </n-form-item>
              <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" :label="$t('recharge.orderNo')">
                <n-input-group>
                  <n-input v-model:value="balanceForm.orderNo" :placeholder="$t('adminUsers.enterOrderNoRequired')" style="flex: 1" />
                  <NButton type="primary" ghost @click="emit('autoFillNo', 'order')">
                    {{ $t('adminUsers.autoGenerate') }}
                  </NButton>
                </n-input-group>
              </n-form-item>
              <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" :label="$t('recharge.tradeNo')">
                <n-input-group>
                  <n-input v-model:value="balanceForm.tradeNo" :placeholder="$t('adminUsers.enterTradeNoOptional')" style="flex: 1" />
                  <NButton type="primary" ghost @click="emit('autoFillNo', 'trade')">
                    {{ $t('adminUsers.autoGenerate') }}
                  </NButton>
                </n-input-group>
              </n-form-item>
              <n-form-item v-if="['order_only', 'balance_order', 'log_order', 'both'].includes(balanceForm.operation)" :label="$t('recharge.orderStatus')">
                <n-select v-model:value="balanceForm.orderStatus" :options="orderStatusOptions" :placeholder="$t('adminUsers.selectOrderStatus')" />
              </n-form-item>
              <n-alert v-if="balanceForm.operation === 'order_only'" type="info" :show-icon="false" style="margin-bottom: 16px;">
                {{ $t('adminUsers.orderOnlyHint') }}
              </n-alert>
              <n-alert v-if="balanceForm.operation === 'balance_order'" type="info" :show-icon="false" style="margin-bottom: 16px;">
                {{ $t('adminUsers.balanceOrderHint') }}
              </n-alert>
              <n-alert v-if="balanceForm.operation === 'log_order'" type="info" :show-icon="false" style="margin-bottom: 16px;">
                {{ $t('adminUsers.logOrderHint') }}
              </n-alert>
              <n-alert v-if="balanceForm.operation === 'both'" type="info" :show-icon="false" style="margin-bottom: 16px;">
                {{ $t('adminUsers.allInOneHint') }}
              </n-alert>
              <n-form-item>
                <NButton type="primary" :loading="submitting" @click="emit('balanceOperation')">
                  {{ $t('adminUsers.confirmOperation') }}
                </NButton>
              </n-form-item>
            </n-form>
          </NSpace>
        </div>
      </n-tab-pane>

      <n-tab-pane v-if="isEdit" name="score" :tab="$t('adminUsers.score')">
        <div class="score-management">
          <NSpace vertical size="large">
            <n-card :title="$t('adminUsers.currentScore')" size="small">
              <n-statistic :label="$t('adminUsers.userScore')" :value="selectedUser?.score || 0" />
            </n-card>
            <n-form label-placement="left" :label-width="100">
              <n-form-item :label="$t('adminUsers.score')">
                <n-input-number v-model:value="scoreForm.amount" :placeholder="$t('adminUsers.scorePlaceholder')" :step="1" />
              </n-form-item>
              <n-form-item :label="$t('adminUsers.operationType')">
                <n-radio-group v-model:value="scoreForm.operation">
                  <NSpace>
                    <n-radio value="modify">
                      {{ $t('adminUsers.modifyScore') }}
                    </n-radio>
                    <n-radio value="log">
                      {{ $t('adminUsers.logOnly') }}
                    </n-radio>
                  </NSpace>
                </n-radio-group>
              </n-form-item>
              <n-form-item :label="$t('moneyScore.remark')">
                <n-input v-model:value="scoreForm.memo" type="textarea" :placeholder="$t('adminUsers.enterOperationRemark')" :rows="3" />
              </n-form-item>
              <n-form-item>
                <NButton type="primary" :loading="submitting" @click="emit('scoreOperation')">
                  {{ $t('adminUsers.confirmOperation') }}
                </NButton>
              </n-form-item>
            </n-form>
          </NSpace>
        </div>
      </n-tab-pane>

      <n-tab-pane v-if="isEdit" name="withdraw" :tab="$t('moneyScore.withdrawRecords')">
        <n-data-table
          :columns="withdrawColumns"
          :data="withdrawData"
          :loading="withdrawLoading"
          :pagination="withdrawPagination"
          size="small"
          :row-key="(row: WithdrawRecord) => row.id"
          @update:page="emit('withdrawPageChange', $event)"
          @update:page-size="emit('withdrawPageSizeChange', $event)"
        />
      </n-tab-pane>
    </n-tabs>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="emit('update:show', false)">
          {{ $t('common.cancel') }}
        </NButton>
        <NButton v-if="activeTab === 'details'" type="primary" :loading="submitting" @click="emit('submit')">
          {{ isEdit ? $t('adminUsers.update') : $t('adminUsers.create') }}
        </NButton>
      </NSpace>
    </template>
  </n-modal>
</template>

<style scoped>
.password-tip {
  font-size: 12px;
  color: #909399;
  font-style: italic;
}

.avatar-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.balance-management,
.score-management {
  padding: 16px 0;
}

.balance-management .n-card,
.score-management .n-card {
  margin-bottom: 16px;
}

.balance-management .n-statistic,
.score-management .n-statistic {
  text-align: center;
}
</style>
