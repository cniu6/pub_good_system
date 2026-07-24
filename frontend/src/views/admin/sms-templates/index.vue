<script setup lang="ts">
import { computed, h, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  NAlert,
  NButton,
  NCard,
  NCode,
  NDataTable,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NRadioButton,
  NRadioGroup,
  NSpace,
  NTag,
  NText,
} from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility, withSubmitLock } from '@/hooks'
import {
  fetchPreviewSMSTemplate,
  fetchResetSMSTemplate,
  fetchSMSTemplateList,
  fetchUpdateSMSTemplate,

} from '@/service/api/admin/sms-template'
import type { SMSTemplate } from '@/service/api/admin/sms-template'

const { t } = useI18n()
const route = useRoute()

function reportSMSTemplateError(message: string, error?: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

const text = computed(() => ({
  pageTitle: t('smsTemplates.pageTitle'),
  refresh: t('smsTemplates.refresh'),
  infoTip: t('smsTemplates.infoTip'),
  providerWarning: t('smsTemplates.providerWarning'),
  lang: t('smsTemplates.lang'),
  signName: t('smsTemplates.signName'),
  content: t('smsTemplates.content'),
  description: t('smsTemplates.description'),
  status: t('smsTemplates.status'),
  action: t('smsTemplates.action'),
  enabled: t('smsTemplates.enabled'),
  disabled: t('smsTemplates.disabled'),
  unknown: t('smsTemplates.unknown'),
  edit: t('smsTemplates.edit'),
  reset: t('smsTemplates.reset'),
  resetConfirmFirst: t('smsTemplates.resetConfirmFirst'),
  resetConfirmSecond: t('smsTemplates.resetConfirmSecond'),
  loadFailed: t('smsTemplates.loadFailed'),
  resetSuccess: t('smsTemplates.resetSuccess'),
  resetFailed: t('smsTemplates.resetFailed'),
  saveSuccess: t('smsTemplates.saveSuccess'),
  saveFailed: t('smsTemplates.saveFailed'),
  previewFailed: t('smsTemplates.previewFailed'),
  editModalTitle: t('smsTemplates.editModalTitle'),
  contentRequired: t('smsTemplates.contentRequired'),
  signNamePlaceholder: t('smsTemplates.signNamePlaceholder'),
  contentPlaceholder: t('smsTemplates.contentPlaceholder'),
  descriptionPlaceholder: t('smsTemplates.descriptionPlaceholder'),
  cancel: t('smsTemplates.cancel'),
  save: t('smsTemplates.save'),
  registerCode: t('smsTemplates.registerCode'),
  loginCode: t('smsTemplates.loginCode'),
  resetPassword: t('smsTemplates.resetPassword'),
  bindPhone: t('smsTemplates.bindPhone'),
  previewTab: t('smsTemplates.previewTab'),
  variables: t('smsTemplates.variables'),
  varPlaceholder: t('smsTemplates.varPlaceholder'),
  contentLabel: t('smsTemplates.contentLabel'),
  confirmBtn: t('smsTemplates.confirmBtn'),
  cancelBtn: t('smsTemplates.cancelBtn'),
  finalConfirmBtn: t('smsTemplates.finalConfirmBtn'),
  noVarsMsg: t('smsTemplates.noVarsMsg'),
  loadingMsg: t('smsTemplates.loadingMsg'),
}))

const loading = ref(false)
const templates = ref<SMSTemplate[]>([])

const showModal = ref(false)
const currentTemplate = ref<SMSTemplate | null>(null)
const formValue = ref({
  sign_name: '',
  content: '',
  description: '',
  status: 1 as number,
})
const previewText = ref('')
const previewVars = ref<Record<string, string>>({})
const previewLoading = ref(false)
const resetStep = ref(0)
const saving = ref(false)
const resetting = ref(false)

const langMap = computed<Record<string, string>>(() => ({
  'zh-CN': t('adminUsersDetail.chinese'),
  'en-US': t('adminUsersDetail.english'),
}))

const statusMap = computed(() => ({
  0: { label: text.value.disabled, type: 'error' as const },
  1: { label: text.value.enabled, type: 'success' as const },
}))

const groupedTemplates = computed(() => {
  const groups: Record<string, SMSTemplate[]> = {}
  templates.value.forEach((tpl) => {
    if (!groups[tpl.name])
      groups[tpl.name] = []
    groups[tpl.name].push(tpl)
  })
  return groups
})

const templateNameMap = computed<Record<string, string>>(() => ({
  register_code: text.value.registerCode,
  login_code: text.value.loginCode,
  reset_password: text.value.resetPassword,
  bind_phone: text.value.bindPhone,
}))

const columns = computed(() => [
  {
    title: text.value.lang,
    key: 'lang',
    width: 100,
    render: (row: SMSTemplate) => h(NTag, { type: 'info' }, () => langMap.value[row.lang] || row.lang),
  },
  {
    title: text.value.signName,
    key: 'sign_name',
    width: 120,
    ellipsis: { tooltip: true },
    render: (row: SMSTemplate) => row.sign_name || '—',
  },
  {
    title: text.value.content,
    key: 'content',
    ellipsis: { tooltip: true },
  },
  {
    title: text.value.description,
    key: 'description',
    ellipsis: { tooltip: true },
  },
  {
    title: text.value.status,
    key: 'status',
    width: 80,
    render: (row: SMSTemplate) => {
      const status = statusMap.value[row.status as 0 | 1]
      return h(NTag, { type: status?.type || 'default' }, () => status?.label || text.value.unknown)
    },
  },
  {
    title: text.value.action,
    key: 'actions',
    width: 100,
    render: (row: SMSTemplate) => {
      return h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, () => text.value.edit)
    },
  },
])

const selectableColumnOptions = computed(() => [
  { key: 'lang', label: text.value.lang },
  { key: 'sign_name', label: text.value.signName },
  { key: 'content', label: text.value.content },
  { key: 'description', label: text.value.description },
  { key: 'status', label: text.value.status },
])

const {
  columnOptions,
  selectedColumnKeys,
  visibleColumns,
  visibleColumnCount,
  totalColumnCount,
  tableScrollX,
  resetSelectedColumns,
} = useTableColumnVisibility<SMSTemplate>({
  storageKey: 'admin-sms-templates-list',
  columns,
  options: selectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 720,
})

async function loadData() {
  loading.value = true
  try {
    const result = await fetchSMSTemplateList()
    if (result.data)
      templates.value = result.data
  }
  catch (error) {
    reportSMSTemplateError('[smsTemplates] load failed', error)
    window.$message?.error(text.value.loadFailed)
  }
  finally {
    loading.value = false
  }
}

function handleEdit(template: SMSTemplate) {
  currentTemplate.value = template
  formValue.value = {
    sign_name: template.sign_name || '',
    content: template.content,
    description: template.description || '',
    status: template.status,
  }
  resetStep.value = 0
  const vars: Record<string, string> = {}
  if (template.variables) {
    template.variables.split(',').forEach((v) => {
      const key = v.trim()
      if (key)
        vars[key] = ''
    })
  }
  previewVars.value = vars
  previewText.value = ''
  showModal.value = true
  nextTick(() => refreshPreview())
}

async function refreshPreview() {
  if (!currentTemplate.value)
    return
  previewLoading.value = true
  try {
    const result = await fetchPreviewSMSTemplate(currentTemplate.value.id, {
      content: formValue.value.content,
      vars: previewVars.value,
    })
    if (result.data)
      previewText.value = result.data.content
  }
  catch (error) {
    reportSMSTemplateError('[smsTemplates] preview failed', error)
    window.$message?.error(text.value.previewFailed)
  }
  finally {
    previewLoading.value = false
  }
}

let previewTimer: ReturnType<typeof setTimeout> | null = null
function debouncedRefreshPreview() {
  if (previewTimer)
    clearTimeout(previewTimer)
  previewTimer = setTimeout(() => refreshPreview(), 500)
}

watch(() => formValue.value.content, debouncedRefreshPreview)
watch(previewVars, debouncedRefreshPreview, { deep: true })

async function handleSave() {
  if (!currentTemplate.value)
    return
  if (!formValue.value.content.trim()) {
    window.$message?.warning(text.value.contentRequired)
    return
  }
  await withSubmitLock(saving, async () => {
    try {
      const result = await fetchUpdateSMSTemplate(currentTemplate.value!.id, {
        sign_name: formValue.value.sign_name,
        content: formValue.value.content,
        description: formValue.value.description,
        status: formValue.value.status,
      })
      if (result.data) {
        window.$message?.success(text.value.saveSuccess)
        showModal.value = false
        await loadData()
      }
    }
    catch (error) {
      reportSMSTemplateError('[smsTemplates] save failed', error)
      window.$message?.error(text.value.saveFailed)
    }
  })
}

function handleResetClick() {
  resetStep.value = 1
}

function handleResetCancel() {
  resetStep.value = 0
}

async function handleResetFinalConfirm() {
  if (!currentTemplate.value)
    return
  await withSubmitLock(resetting, async () => {
    try {
      const result = await fetchResetSMSTemplate(currentTemplate.value!.id)
      if (result.data) {
        window.$message?.success(text.value.resetSuccess)
        resetStep.value = 0
        await loadData()
        const updated = templates.value.find(t => t.id === currentTemplate.value!.id)
        if (updated) {
          currentTemplate.value = updated
          formValue.value = {
            sign_name: updated.sign_name || '',
            content: updated.content,
            description: updated.description || '',
            status: updated.status,
          }
          await refreshPreview()
        }
      }
    }
    catch (error) {
      reportSMSTemplateError('[smsTemplates] reset failed', error)
      window.$message?.error(text.value.resetFailed)
    }
  })
}

function openByQueryName() {
  const name = typeof route.query.name === 'string' ? route.query.name.trim() : ''
  if (!name || templates.value.length === 0)
    return
  const match = templates.value.find(t => t.name === name)
  if (match)
    handleEdit(match)
}

onMounted(async () => {
  await loadData()
  openByQueryName()
})
</script>

<template>
  <div class="sms-template-page">
    <NCard :title="text.pageTitle">
      <template #header-extra>
        <NSpace>
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
          <NButton type="primary" :loading="loading" @click="loadData">
            {{ text.refresh }}
          </NButton>
        </NSpace>
      </template>

      <NSpace vertical>
        <NAlert type="info">
          {{ text.infoTip }}
        </NAlert>
        <NAlert type="warning">
          {{ text.providerWarning }}
        </NAlert>

        <div v-for="(tpls, name) in groupedTemplates" :key="name">
          <NDivider>
            <span class="font-bold">{{ templateNameMap[name] || name }}</span>
          </NDivider>
          <NDataTable
            :columns="visibleColumns"
            :data="tpls"
            :bordered="false"
            :loading="loading"
            :scroll-x="tableScrollX"
          />
        </div>
      </NSpace>
    </NCard>

    <NModal
      v-model:show="showModal"
      preset="card"
      :title="text.editModalTitle + (currentTemplate ? ` - ${templateNameMap[currentTemplate.name] || currentTemplate.name} (${langMap[currentTemplate.lang] || currentTemplate.lang})` : '')"
      style="width: 92vw; max-width: 1100px;"
      :mask-closable="false"
    >
      <div class="edit-modal-body">
        <div class="edit-panel">
          <div class="edit-panel-header">
            <template v-if="resetStep === 0">
              <NButton size="small" type="warning" @click="handleResetClick">
                {{ text.reset }}
              </NButton>
            </template>
            <template v-else-if="resetStep === 1">
              <NSpace align="center" :size="8">
                <span class="reset-warn-text">{{ text.resetConfirmFirst }}</span>
                <NButton size="small" type="error" @click="resetStep = 2">
                  {{ text.confirmBtn }}
                </NButton>
                <NButton size="small" @click="handleResetCancel">
                  {{ text.cancelBtn }}
                </NButton>
              </NSpace>
            </template>
            <template v-else-if="resetStep === 2">
              <NSpace align="center" :size="8">
                <span class="reset-danger-text">{{ text.resetConfirmSecond }}</span>
                <NButton size="small" type="error" :loading="resetting" @click="handleResetFinalConfirm">
                  {{ text.finalConfirmBtn }}
                </NButton>
                <NButton size="small" @click="handleResetCancel">
                  {{ text.cancelBtn }}
                </NButton>
              </NSpace>
            </template>
          </div>

          <div class="edit-panel-content">
            <NForm label-placement="top">
              <NFormItem :label="text.signName">
                <NInput v-model:value="formValue.sign_name" :placeholder="text.signNamePlaceholder" />
              </NFormItem>
              <NFormItem :label="text.contentLabel">
                <NInput
                  v-model:value="formValue.content"
                  type="textarea"
                  :placeholder="text.contentPlaceholder"
                  :rows="6"
                  style="font-family: Consolas, Monaco, monospace; font-size: 13px;"
                />
              </NFormItem>
              <NFormItem :label="text.description">
                <NInput v-model:value="formValue.description" :placeholder="text.descriptionPlaceholder" />
              </NFormItem>
              <NFormItem :label="text.status" style="margin-bottom: 8px;">
                <NRadioGroup v-model:value="formValue.status">
                  <NRadioButton :value="1">
                    {{ text.enabled }}
                  </NRadioButton>
                  <NRadioButton :value="0">
                    {{ text.disabled }}
                  </NRadioButton>
                </NRadioGroup>
              </NFormItem>

              <NDivider style="margin: 8px 0;">
                <span style="font-size: 13px;">{{ text.variables }}</span>
              </NDivider>
              <div v-if="Object.keys(previewVars).length > 0" class="var-grid">
                <div v-for="key in Object.keys(previewVars)" :key="key" class="var-item">
                  <span class="var-label">{{ key }}</span>
                  <NInput
                    v-model:value="previewVars[key]"
                    :placeholder="text.varPlaceholder"
                    size="small"
                  />
                </div>
              </div>
              <NAlert v-else type="info" style="font-size: 12px;">
                {{ text.noVarsMsg }}
              </NAlert>
            </NForm>
          </div>
        </div>

        <div class="preview-panel">
          <div class="preview-panel-header">
            <span class="preview-title">{{ text.previewTab }}</span>
            <NButton size="tiny" :loading="previewLoading" @click="refreshPreview">
              {{ text.refresh }}
            </NButton>
          </div>
          <div class="preview-box">
            <NCode v-if="previewText" :code="previewText" language="text" word-wrap />
            <NText v-else depth="3">
              {{ text.loadingMsg }}
            </NText>
          </div>
        </div>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">
            {{ text.cancel }}
          </NButton>
          <NButton type="primary" :loading="saving" @click="handleSave">
            {{ text.save }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.edit-modal-body {
  display: flex;
  gap: 16px;
  min-height: 420px;
  max-height: 70vh;
  overflow: hidden;
}

.edit-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.edit-panel-header {
  flex-shrink: 0;
  margin-bottom: 12px;
  min-height: 32px;
  display: flex;
  align-items: center;
}

.edit-panel-content {
  flex: 1;
  overflow-y: auto;
  padding-right: 8px;
}

.reset-warn-text {
  font-size: 13px;
  color: #f0a020;
  font-weight: 500;
}

.reset-danger-text {
  font-size: 13px;
  color: #d03050;
  font-weight: 600;
}

.var-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.var-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.var-label {
  font-size: 12px;
  font-weight: 600;
  color: #666;
  white-space: nowrap;
  min-width: 60px;
}

.preview-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid #e8e8f0;
  padding-left: 16px;
  overflow: hidden;
}

.preview-panel-header {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  min-height: 32px;
}

.preview-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.preview-box {
  flex: 1;
  border: 1px solid #e8e8f0;
  border-radius: 8px;
  padding: 16px;
  background: #fafafa;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}
</style>
