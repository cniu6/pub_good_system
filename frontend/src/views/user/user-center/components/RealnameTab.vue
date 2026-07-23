<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import type { FormItemRule, FormRules } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '@/store'
import {

  certificateTypeOptions,
  fetchMyRealnameStatus,

  submitRealname,
} from '@/service/api/user/realname'
import type { CertificateType, RealnameStatusResponse } from '@/service/api/user/realname'

const message = useMessage()
const { t } = useI18n()
const settingsStore = useSettingsStore()

// 状态
const loading = ref(false)
const submitting = ref(false)
const showSubmitModal = ref(false)
const showDetailModal = ref(false)
const realnameInfo = ref<RealnameStatusResponse | null>(null)

// 表单
const formRef = ref()
const form = ref({
  real_name: '',
  certificate_type: null as CertificateType | null,
  certificate_no: '',
  certificate_front: '',
  certificate_back: '',
})

// 计算属性
const hasVerification = computed(() => realnameInfo.value?.hasVerification)
const verificationStatus = computed(() => realnameInfo.value?.status as 0 | 1 | 2 | undefined)
const realnameEnabled = computed(() => settingsStore.realnameEnabled)
const realnameNotifyText = computed(() => settingsStore.realnameNotifyText || t('realname.notVerifiedDesc'))
const submittedTime = computed(() => {
  const ts = realnameInfo.value?.submittedAt
  return ts ? new Date(ts * 1000).toLocaleString() : '-'
})
const rejectReason = computed(() => realnameInfo.value?.rejectReason || '')
const certificateNoPlaceholder = computed(() => {
  switch (form.value.certificate_type) {
    case 1: return t('realname.inputIdCardNo')
    case 2: return t('realname.inputPassportNo')
    case 3: return t('realname.inputOfficerNo')
    default: return t('realname.chooseCertificateTypeFirst')
  }
})

// 表单验证
function validateCertificateNo(_rule: FormItemRule, value: string) {
  if (!value)
    return new Error(t('realname.certificateNoRequired'))
  if (form.value.certificate_type === 1) {
    if (!/^\d{17}[\dX]$/i.test(value))
      return new Error(t('realname.idCardInvalid'))
  }
  return true
}

const rules: FormRules = {
  real_name: [
    { required: true, message: t('realname.realNameRequired'), trigger: 'blur' },
    { max: 50, message: t('realname.realNameTooLong'), trigger: 'blur' },
  ],
  certificate_type: [
    { required: true, type: 'number', message: t('realname.certificateTypeRequired'), trigger: 'change' },
  ],
  certificate_no: [
    { required: true, message: t('realname.certificateNoRequired'), trigger: 'blur' },
    { validator: validateCertificateNo, trigger: 'blur' },
  ],
  certificate_front: [
    { required: true, message: t('realname.certificateFrontRequired'), trigger: 'blur' },
  ],
  certificate_back: [
    { required: true, message: t('realname.certificateBackRequired'), trigger: 'blur' },
  ],
}

// 获取证件类型文本
function getCertificateTypeText(type_: number | undefined): string {
  const map: Record<number, string> = {
    1: t('realname.idCard'),
    2: t('realname.passport'),
    3: t('realname.officer'),
  }
  return type_ ? map[type_] || t('realname.unknown') : '-'
}

// 获取状态文本
function getStatusText(status: number | undefined): string {
  const map: Record<number, string> = {
    0: t('realname.pending'),
    1: t('realname.approved'),
    2: t('realname.rejected'),
  }
  return status !== undefined ? map[status] || t('realname.unknown') : '-'
}

// 获取状态颜色
function getStatusType(status: number | undefined): 'warning' | 'success' | 'error' {
  const map: Record<number, 'warning' | 'success' | 'error'> = { 0: 'warning', 1: 'success', 2: 'error' }
  return status !== undefined ? map[status] || 'warning' : 'warning'
}

// 证件号码脱敏
function maskCertificateNo(no: string | undefined): string {
  if (!no || no.length < 8)
    return no || '-'
  return `${no.slice(0, 4)}****${no.slice(-4)}`
}

// 加载实名状态
async function loadRealnameStatus() {
  loading.value = true
  try {
    const res = await fetchMyRealnameStatus()
    if (res.isSuccess && res.data) {
      realnameInfo.value = res.data
    }
  }
  catch (e) {
    if (import.meta.env.DEV)
      console.error('[realnameTab] load status failed', e)
    message.error(t('realname.loadFailed'))
  }
  finally {
    loading.value = false
  }
}

// 提交认证
async function handleSubmit() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  submitting.value = true
  try {
    const res = await submitRealname({
      real_name: form.value.real_name.trim(),
      certificate_type: form.value.certificate_type!,
      certificate_no: form.value.certificate_no.trim(),
      certificate_front: form.value.certificate_front.trim(),
      certificate_back: form.value.certificate_back.trim(),
    })
    if (!res.isSuccess) {
      message.error(res.message || t('realname.submitFailed'))
      return
    }
    message.success(t('realname.submitSuccess'))
    showSubmitModal.value = false
    loadRealnameStatus()
    // 重置表单
    form.value = {
      real_name: '',
      certificate_type: null,
      certificate_no: '',
      certificate_front: '',
      certificate_back: '',
    }
  }
  catch (e: unknown) {
    if (import.meta.env.DEV)
      console.error('[realnameTab] submit failed', e)
    message.error(t('realname.submitFailed'))
  }
  finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadRealnameStatus()
})
</script>

<template>
  <div class="realname-container">
    <n-card v-if="loading" :title="t('realname.title')" class="realname-card">
      <n-spin size="small">
        <div style="padding: 24px 0;" />
      </n-spin>
    </n-card>

    <!-- 未认证状态 -->
    <n-card v-else-if="!realnameEnabled" :title="t('realname.title')" class="realname-card">
      <n-result status="warning" :title="t('realname.disabledTitle')" :description="t('realname.disabledDesc')" />
    </n-card>

    <n-card v-else-if="!hasVerification" :title="t('realname.title')" class="realname-card">
      <n-result status="info" :title="t('realname.notVerifiedTitle')" :description="realnameNotifyText">
        <template #footer>
          <n-button type="primary" @click="showSubmitModal = true">
            {{ t('realname.verifyNow') }}
          </n-button>
        </template>
      </n-result>
    </n-card>

    <!-- 待审核状态 -->
    <n-card v-else-if="verificationStatus === 0" :title="t('realname.title')" class="realname-card">
      <n-result status="warning" :title="t('realname.pendingTitle')" :description="`${t('realname.pendingDescPrefix')}${submittedTime}${t('realname.pendingDescSuffix')}`">
        <template #footer>
          <n-space>
            <n-button @click="showDetailModal = true">
              {{ t('realname.viewDetail') }}
            </n-button>
          </n-space>
        </template>
      </n-result>
    </n-card>

    <!-- 已拒绝状态 -->
    <n-card v-else-if="verificationStatus === 2" :title="t('realname.title')" class="realname-card">
      <n-result status="error" :title="t('realname.rejectedTitle')" :description="`${t('realname.rejectedReasonPrefix')}${rejectReason || t('realname.rejectedFallbackReason')}`">
        <template #footer>
          <n-space>
            <n-button type="primary" @click="showSubmitModal = true">
              {{ t('realname.retry') }}
            </n-button>
          </n-space>
        </template>
      </n-result>
    </n-card>

    <!-- 已通过状态 -->
    <n-card v-else-if="verificationStatus === 1" :title="t('realname.title')" class="realname-card">
      <n-result status="success" :title="t('realname.approvedTitle')" :description="t('realname.approvedDesc')">
        <template #footer>
          <n-space>
            <n-button @click="showDetailModal = true">
              {{ t('realname.viewDetail') }}
            </n-button>
          </n-space>
        </template>
      </n-result>
      <n-descriptions :column="2" label-placement="left" bordered size="small">
        <n-descriptions-item :label="t('realname.realName')">
          {{ realnameInfo?.realName }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateType')">
          {{ getCertificateTypeText(realnameInfo?.certificateType) }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateNo')">
          {{ maskCertificateNo(realnameInfo?.certificateNo) }}
        </n-descriptions-item>
      </n-descriptions>
    </n-card>

    <!-- 提交认证弹窗 -->
    <n-modal v-model:show="showSubmitModal" preset="card" :title="t('realname.submitTitle')" style="width: 520px;">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="100">
        <n-form-item :label="t('realname.realName')" path="real_name">
          <n-input v-model:value="form.real_name" :placeholder="t('realname.inputRealName')" />
        </n-form-item>
        <n-form-item :label="t('realname.certificateType')" path="certificate_type">
          <n-select v-model:value="form.certificate_type" :options="certificateTypeOptions" :placeholder="t('realname.selectCertificateType')" />
        </n-form-item>
        <n-form-item :label="t('realname.certificateNo')" path="certificate_no">
          <n-input v-model:value="form.certificate_no" :placeholder="certificateNoPlaceholder" />
        </n-form-item>
        <n-form-item :label="t('realname.certificateFront')" path="certificate_front">
          <n-input v-model:value="form.certificate_front" :placeholder="t('realname.inputCertificateFrontUrl')" />
          <n-image v-if="form.certificate_front" :src="form.certificate_front" width="200" height="130" object-fit="cover" style="margin-top: 8px; border-radius: 4px;" />
        </n-form-item>
        <n-form-item :label="t('realname.certificateBack')" path="certificate_back">
          <n-input v-model:value="form.certificate_back" :placeholder="t('realname.inputCertificateBackUrl')" />
          <n-image v-if="form.certificate_back" :src="form.certificate_back" width="200" height="130" object-fit="cover" style="margin-top: 8px; border-radius: 4px;" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showSubmitModal = false">
            {{ t('realname.cancel') }}
          </n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ t('realname.submit') }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 详情弹窗 -->
    <n-modal v-model:show="showDetailModal" preset="card" :title="t('realname.detailTitle')" style="width: 500px;">
      <n-descriptions v-if="realnameInfo" :column="1" label-placement="left" bordered>
        <n-descriptions-item :label="t('realname.realName')">
          {{ realnameInfo.realName }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateType')">
          {{ getCertificateTypeText(realnameInfo.certificateType) }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateNo')">
          {{ maskCertificateNo(realnameInfo.certificateNo) }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateFront')">
          <n-image v-if="realnameInfo.certificateFront" :src="realnameInfo.certificateFront" width="200" height="130" object-fit="cover" />
          <span v-else>-</span>
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.certificateBack')">
          <n-image v-if="realnameInfo.certificateBack" :src="realnameInfo.certificateBack" width="200" height="130" object-fit="cover" />
          <span v-else>-</span>
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.status')">
          <n-tag :type="getStatusType(realnameInfo.status)">
            {{ getStatusText(realnameInfo.status) }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item v-if="realnameInfo.status === 2" :label="t('realname.rejectReason')">
          {{ realnameInfo.rejectReason || '-' }}
        </n-descriptions-item>
        <n-descriptions-item :label="t('realname.submittedAt')">
          {{ realnameInfo.submittedAt ? new Date(realnameInfo.submittedAt * 1000).toLocaleString() : '-' }}
        </n-descriptions-item>
        <n-descriptions-item v-if="realnameInfo.reviewedAt" :label="t('realname.reviewedAt')">
          {{ realnameInfo.reviewedAt ? new Date(realnameInfo.reviewedAt * 1000).toLocaleString() : '-' }}
        </n-descriptions-item>
      </n-descriptions>
    </n-modal>
  </div>
</template>

<style scoped>
.realname-card {
  max-width: 600px;
}
</style>
