<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  paymentForm,
  switchLoading,
  savingPayment,
  handleUpdatePaymentEnabled,
  handleUpdateWithdrawEnabled,
  handleUpdateWithdrawRequireRealname,
  handleUpdateFinanceDualApproval,
  handleSavePayment,
} = useAdminSettings()
</script>

<template>
  <n-space vertical>
    <n-form :model="paymentForm" label-placement="left" label-width="140px" style="max-width: 640px;">
      <n-form-item :label="t('adminSettings.paymentEnabled')">
        <n-space align="center">
          <n-switch
            :value="paymentForm.payment_enabled"
            :loading="switchLoading.payment_enabled"
            @update:value="handleUpdatePaymentEnabled"
          />
          <n-text depth="3">
            {{ paymentForm.payment_enabled ? t('adminSettings.enabled') : t('adminSettings.disabled') }}
          </n-text>
        </n-space>
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.orderExpireMinutes')">
        <n-input-number v-model:value="paymentForm.payment_order_expire_minutes" :min="1" :max="1440" style="width: 100%;" />
      </n-form-item>
      <n-divider />
      <n-form-item :label="t('adminSettings.withdrawEnabled')">
        <n-space align="center">
          <n-switch
            :value="paymentForm.withdraw_enabled"
            :loading="switchLoading.withdraw_enabled"
            @update:value="handleUpdateWithdrawEnabled"
          />
          <n-text depth="3">
            {{ paymentForm.withdraw_enabled ? t('adminSettings.withdrawEnabledText') : t('adminSettings.withdrawDisabledText') }}
          </n-text>
        </n-space>
      </n-form-item>
      <n-form-item :label="t('adminSettings.withdrawMinAmount')">
        <n-input-number v-model:value="paymentForm.withdraw_min_amount" :min="0.01" :precision="2" :step="1" style="width: 100%;" />
      </n-form-item>
      <n-form-item :label="t('adminSettings.withdrawNotifyText')">
        <n-input
          v-model:value="paymentForm.withdraw_notify_text"
          type="textarea"
          :placeholder="t('adminSettings.withdrawNotifyTextPlaceholder')"
          :rows="3"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.withdrawAccountTypes')">
        <n-input
          v-model:value="paymentForm.withdraw_account_types_text"
          type="textarea"
          :placeholder="t('adminSettings.withdrawAccountTypesPlaceholder')"
          :rows="3"
        />
      </n-form-item>
      <n-form-item :label="t('adminSettings.withdrawRequireRealname')">
        <n-space align="center">
          <n-switch
            :value="paymentForm.withdraw_require_realname"
            :loading="switchLoading.withdraw_require_realname"
            @update:value="handleUpdateWithdrawRequireRealname"
          />
          <n-text depth="3">
            {{ t('adminSettings.withdrawRequireRealnameText') }}
          </n-text>
        </n-space>
      </n-form-item>
      <n-divider />
      <!-- 双人复核：仅约束管理员人工强制补单，不影响网关回调自动入账/自动发货 -->
      <n-form-item :label="t('adminSettings.financeDualApproval')">
        <n-space vertical :size="6">
          <n-space align="center">
            <n-switch
              :value="paymentForm.finance_dual_approval"
              :loading="switchLoading.finance_dual_approval"
              @update:value="handleUpdateFinanceDualApproval"
            />
            <n-text depth="3">
              {{ paymentForm.finance_dual_approval ? t('adminSettings.enabled') : t('adminSettings.disabled') }}
            </n-text>
          </n-space>
          <n-text depth="3" style="font-size: 12px; line-height: 1.5;">
            {{ t('adminSettings.financeDualApprovalHint') }}
          </n-text>
        </n-space>
      </n-form-item>
      <n-form-item>
        <n-button type="primary" :loading="savingPayment" @click="handleSavePayment">
          {{ t('adminSettings.saveSettings') }}
        </n-button>
      </n-form-item>
    </n-form>
    <n-alert type="info" :title="t('adminSettings.configDesc')" :bordered="false">
      <ul style="margin: 0; padding-left: 18px;">
        <li>{{ t('adminSettings.paymentAlert1') }}</li>
        <li>{{ t('adminSettings.paymentAlert2') }}</li>
        <li>{{ t('adminSettings.paymentAlert3') }}</li>
      </ul>
    </n-alert>
  </n-space>
</template>
