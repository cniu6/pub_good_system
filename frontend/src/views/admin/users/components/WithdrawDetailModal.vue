<script setup lang="ts">
/**
 * 提现详情弹窗（用户列表编辑弹窗 / 用户详情页共用）
 */
import { NTag } from 'naive-ui'
import type { WithdrawRecord } from '@/service/api/admin/finance'
import type { UserSimpleInfo } from '@/service/api/admin/user'
import {
  formatTime,
  getAdminDisplayName,
  getWithdrawStatusMeta,
} from '../utils/userDisplay'

const props = defineProps<{
  show: boolean
  detail: WithdrawRecord | null
  adminUserMap: Record<number, UserSimpleInfo>
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

function adminName(id?: number | null) {
  return getAdminDisplayName(id, props.adminUserMap)
}
</script>

<template>
  <n-modal
    :show="show"
    preset="card"
    :title="$t('adminUsers.withdrawDetail')"
    style="width: 620px;"
    :bordered="false"
    @update:show="emit('update:show', $event)"
  >
    <template v-if="detail">
      <n-descriptions :column="1" bordered label-placement="left">
        <n-descriptions-item :label="$t('moneyScore.applicationId')">
          {{ detail.id }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('adminRealname.userId')">
          {{ detail.user_id }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.withdrawAmount')">
          ¥{{ Number(detail.amount).toFixed(2) }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.status')">
          <NTag :type="getWithdrawStatusMeta(detail.status).type">
            {{ getWithdrawStatusMeta(detail.status).label }}
          </NTag>
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.accountType')">
          {{ detail.account_type }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.accountName')">
          {{ detail.account_name }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.accountNo')">
          {{ detail.account_no }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.realName')">
          {{ detail.real_name }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.userRemark')">
          {{ detail.remark || '-' }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.reviewRemark')">
          {{ detail.review_remark || '-' }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.transferRemark')">
          {{ detail.transfer_remark || '-' }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.createdAt')">
          {{ formatTime(detail.create_time) }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.reviewedAt')">
          {{ formatTime(detail.reviewed_at) }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('adminUsers.reviewer')">
          {{ adminName(detail.reviewed_by) }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('moneyScore.paidAt')">
          {{ formatTime(detail.paid_at) }}
        </n-descriptions-item>
        <n-descriptions-item :label="$t('adminUsers.payer')">
          {{ adminName(detail.paid_by) }}
        </n-descriptions-item>
      </n-descriptions>
    </template>
  </n-modal>
</template>
