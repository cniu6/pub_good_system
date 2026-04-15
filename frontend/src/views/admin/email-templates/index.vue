<script setup lang="ts">
import { computed, h, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDivider,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NRadioButton,
  NRadioGroup,
  NSelect,
  NSpace,
  NTag,
} from 'naive-ui'
import {
  adminEmailTemplateApi,
  fetchEmailTemplateList,
  fetchPreviewEmailTemplate,
  fetchResetEmailTemplate,
  fetchUpdateEmailTemplate,
  type EmailTemplate,
} from '@/service/api/admin/email-template'

const { t } = useI18n()

// 开发环境日志辅助函数
function reportEmailTemplateError(message: string, error?: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

const text = computed(() => ({
  pageTitle: t('emailTemplates.pageTitle'),
  refresh: t('emailTemplates.refresh'),
  infoTip: t('emailTemplates.infoTip'),
  lang: t('emailTemplates.lang'),
  subject: t('emailTemplates.subject'),
  description: t('emailTemplates.description'),
  status: t('emailTemplates.status'),
  action: t('emailTemplates.action'),
  enabled: t('emailTemplates.enabled'),
  disabled: t('emailTemplates.disabled'),
  unknown: t('emailTemplates.unknown'),
  edit: t('emailTemplates.edit'),
  reset: t('emailTemplates.reset'),
  resetConfirmFirst: t('emailTemplates.resetConfirmFirst'),
  resetConfirmSecond: t('emailTemplates.resetConfirmSecond'),
  loadFailed: t('emailTemplates.loadFailed'),
  resetSuccess: t('emailTemplates.resetSuccess'),
  resetFailed: t('emailTemplates.resetFailed'),
  saveSuccess: t('emailTemplates.saveSuccess'),
  saveFailed: t('emailTemplates.saveFailed'),
  previewFailed: t('emailTemplates.previewFailed'),
  editModalTitle: t('emailTemplates.editModalTitle'),
  subjectRequired: t('emailTemplates.subjectRequired'),
  contentRequired: t('emailTemplates.contentRequired'),
  subjectPlaceholder: t('emailTemplates.subjectPlaceholder'),
  contentPlaceholder: t('emailTemplates.contentPlaceholder'),
  descriptionPlaceholder: t('emailTemplates.descriptionPlaceholder'),
  cancel: t('emailTemplates.cancel'),
  save: t('emailTemplates.save'),
  registerCode: t('emailTemplates.registerCode'),
  resetPassword: t('emailTemplates.resetPassword'),
  sendTest: t('emailTemplates.sendTest'),
  sendTestDesc: t('emailTemplates.sendTestDesc'),
  testTo: t('emailTemplates.testTo'),
  testToPlaceholder: t('emailTemplates.testToPlaceholder'),
  testSubject: t('emailTemplates.testSubject'),
  testSubjectPlaceholder: t('emailTemplates.testSubjectPlaceholder'),
  sending: t('emailTemplates.sending'),
  send: t('emailTemplates.send'),
  sendSuccess: t('emailTemplates.sendSuccess'),
  sendFailed: t('emailTemplates.sendFailed'),
  fullscreen: t('emailTemplates.fullscreen'),
  exitFullscreen: t('emailTemplates.exitFullscreen'),
  previewTab: t('emailTemplates.previewTab'),
  variables: t('emailTemplates.variables'),
  varPlaceholder: t('emailTemplates.varPlaceholder'),
  contentLabel: t('emailTemplates.contentLabel'),
  confirmBtn: t('emailTemplates.confirmBtn'),
  cancelBtn: t('emailTemplates.cancelBtn'),
  finalConfirmBtn: t('emailTemplates.finalConfirmBtn'),
  noVarsMsg: t('emailTemplates.noVarsMsg'),
  loadingMsg: t('emailTemplates.loadingMsg'),
  inputEmail: t('emailTemplates.inputEmail'),
  selectTemplate: t('emailTemplates.selectTemplate'),
  noTemplate: t('emailTemplates.noTemplate'),
}))

const loading = ref(false)
const templates = ref<EmailTemplate[]>([])

// ---- Send Test ----
const testTo = ref('')
const testSubject = ref('')
const testTemplateId = ref<number | null>(null)
const testSending = ref(false)

const templateOptions = computed(() => {
  const opts: { label: string; value: number }[] = []
  templates.value.forEach((tpl) => {
    const name = templateNameMap.value[tpl.name] || tpl.name
    const lang = langMap.value[tpl.lang] || tpl.lang
    opts.push({ label: `${name} (${lang})`, value: tpl.id })
  })
  return opts
})

// ---- Edit Modal ----
const showModal = ref(false)
const isFullscreen = ref(false)
const currentTemplate = ref<EmailTemplate | null>(null)
const formValue = ref({
  subject: '',
  content: '',
  description: '',
  status: 1 as number,
})
const previewHtml = ref('')
const previewVars = ref<Record<string, string>>({})
const previewLoading = ref(false)
const resetStep = ref(0) // 0=idle, 1=first confirm, 2=second confirm

const langMap = computed<Record<string, string>>(() => ({
  'zh-CN': t('adminUsersDetail.chinese'),
  'en-US': t('adminUsersDetail.english'),
}))

const statusMap = computed(() => ({
  0: { label: text.value.disabled, type: 'error' as const },
  1: { label: text.value.enabled, type: 'success' as const },
}))

const groupedTemplates = computed(() => {
  const groups: Record<string, EmailTemplate[]> = {}
  templates.value.forEach((tpl) => {
    if (!groups[tpl.name])
      groups[tpl.name] = []
    groups[tpl.name].push(tpl)
  })
  return groups
})

const templateNameMap = computed<Record<string, string>>(() => ({
  register_code: text.value.registerCode,
  reset_password: text.value.resetPassword,
}))

const columns = computed(() => [
  {
    title: text.value.lang,
    key: 'lang',
    width: 100,
    render: (row: EmailTemplate) => h(NTag, { type: 'info' }, () => langMap.value[row.lang] || row.lang),
  },
  {
    title: text.value.subject,
    key: 'subject',
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
    render: (row: EmailTemplate) => {
      const status = statusMap.value[row.status as 0 | 1]
      return h(NTag, { type: status?.type || 'default' }, () => status?.label || text.value.unknown)
    },
  },
  {
    title: text.value.action,
    key: 'actions',
    width: 100,
    render: (row: EmailTemplate) => {
      return h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, () => text.value.edit)
    },
  },
])

// ---- Data ----
async function loadData() {
  loading.value = true
  try {
    const result = await fetchEmailTemplateList()
    if (result.data)
      templates.value = result.data
  }
  catch (error) {
    reportEmailTemplateError('[emailTemplates] load failed', error)
    window.$message?.error(text.value.loadFailed)
  }
  finally {
    loading.value = false
  }
}

// ---- Send Test ----
async function handleSendTest() {
  if (!testTo.value.trim()) {
    window.$message?.warning(text.value.inputEmail)
    return
  }
  testSending.value = true
  try {
    const result = await adminEmailTemplateApi.sendTest({
      to: testTo.value.trim(),
      subject: testSubject.value.trim() || undefined,
      template_id: testTemplateId.value || undefined,
    })
    if (result.data)
      window.$message?.success(text.value.sendSuccess)
  }
  catch (error: any) {
    reportEmailTemplateError('[emailTemplates] send test failed', error)
    window.$message?.error(text.value.sendFailed)
  }
  finally {
    testSending.value = false
  }
}

// ---- Edit ----
function handleEdit(template: EmailTemplate) {
  currentTemplate.value = template
  formValue.value = {
    subject: template.subject,
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
  previewHtml.value = ''
  showModal.value = true
  nextTick(() => refreshPreview())
}

async function refreshPreview() {
  if (!currentTemplate.value)
    return
  previewLoading.value = true
  try {
    const result = await fetchPreviewEmailTemplate(currentTemplate.value.id, {
      content: formValue.value.content,
      vars: previewVars.value,
    })
    if (result.data)
      previewHtml.value = result.data.wrapped || result.data.content
  }
  catch (error) {
    reportEmailTemplateError('[emailTemplates] preview failed', error)
  }
  finally {
    previewLoading.value = false
  }
}

let previewTimer: ReturnType<typeof setTimeout> | null = null
function debouncedRefreshPreview() {
  if (previewTimer)
    clearTimeout(previewTimer)
  previewTimer = setTimeout(() => refreshPreview(), 600)
}

watch(() => formValue.value.content, debouncedRefreshPreview)
watch(() => formValue.value.subject, debouncedRefreshPreview)
watch(previewVars, debouncedRefreshPreview, { deep: true })

async function handleSave() {
  if (!currentTemplate.value)
    return
  if (!formValue.value.subject.trim()) {
    window.$message?.warning(text.value.subjectRequired)
    return
  }
  if (!formValue.value.content.trim()) {
    window.$message?.warning(text.value.contentRequired)
    return
  }
  try {
    const result = await fetchUpdateEmailTemplate(currentTemplate.value.id, {
      subject: formValue.value.subject,
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
    reportEmailTemplateError('[emailTemplates] save failed', error)
    window.$message?.error(text.value.saveFailed)
  }
}

// ---- Reset (double confirm) ----
function handleResetClick() {
  resetStep.value = 1
}

function handleResetCancel() {
  resetStep.value = 0
}

async function handleResetFinalConfirm() {
  if (!currentTemplate.value)
    return
  try {
    const result = await fetchResetEmailTemplate(currentTemplate.value.id)
    if (result.data) {
      window.$message?.success(text.value.resetSuccess)
      resetStep.value = 0
      await loadData()
      const updated = templates.value.find(t => t.id === currentTemplate.value!.id)
      if (updated) {
        currentTemplate.value = updated
        formValue.value = {
          subject: updated.subject,
          content: updated.content,
          description: updated.description || '',
          status: updated.status,
        }
      }
    }
  }
  catch (error) {
    reportEmailTemplateError('[emailTemplates] reset failed', error)
    window.$message?.error(text.value.resetFailed)
  }
}

function toggleFullscreen() {
  isFullscreen.value = !isFullscreen.value
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="email-template-page">
    <!-- Send Test Card -->
    <NCard :title="text.sendTest" style="margin-bottom: 16px;">
      <NAlert type="info" style="margin-bottom: 16px;">
        {{ text.sendTestDesc }}
      </NAlert>
      <NForm label-placement="left" label-width="auto" inline>
        <NFormItem :label="text.testTo">
          <NInput
            v-model:value="testTo"
            :placeholder="text.testToPlaceholder"
            style="width: 280px;"
          />
        </NFormItem>
        <NFormItem :label="text.testSubject">
          <NInput
            v-model:value="testSubject"
            :placeholder="text.testSubjectPlaceholder"
            style="width: 280px;"
            :disabled="!!testTemplateId"
          />
        </NFormItem>
        <NFormItem :label="text.selectTemplate">
          <NSelect
            v-model:value="testTemplateId"
            :options="templateOptions"
            clearable
            :placeholder="text.noTemplate"
            style="width: 220px;"
          />
        </NFormItem>
        <NFormItem>
          <NButton
            type="primary"
            :loading="testSending"
            :disabled="!testTo.trim()"
            @click="handleSendTest"
          >
            {{ testSending ? text.sending : text.send }}
          </NButton>
        </NFormItem>
      </NForm>
    </NCard>

    <!-- Template List Card -->
    <NCard :title="text.pageTitle">
      <template #header-extra>
        <NButton type="primary" :loading="loading" @click="loadData">
          {{ text.refresh }}
        </NButton>
      </template>

      <NSpace vertical>
        <NAlert type="info">
          {{ text.infoTip }}
        </NAlert>

        <div v-for="(tpls, name) in groupedTemplates" :key="name">
          <NDivider>
            <span class="font-bold">{{ templateNameMap[name] || name }}</span>
          </NDivider>
          <NDataTable
            :columns="columns"
            :data="tpls"
            :bordered="false"
            :loading="loading"
          />
        </div>
      </NSpace>
    </NCard>

    <!-- Edit Modal (left edit, right preview) -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="text.editModalTitle + (currentTemplate ? ` - ${templateNameMap[currentTemplate.name] || currentTemplate.name} (${langMap[currentTemplate.lang] || currentTemplate.lang})` : '')"
      :style="isFullscreen
        ? 'position:fixed;top:0;left:0;width:100vw;height:100vh;max-width:100vw;border-radius:0;z-index:9999;'
        : 'width:95vw;max-width:1400px;'"
      :mask-closable="false"
    >
      <template #header-extra>
        <NButton quaternary size="small" @click="toggleFullscreen">
          {{ isFullscreen ? text.exitFullscreen : text.fullscreen }}
        </NButton>
      </template>

      <div class="edit-modal-body" :style="{ height: isFullscreen ? 'calc(100vh - 130px)' : '70vh' }">
        <!-- Left: Edit Panel -->
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
                <NButton size="small" type="error" @click="handleResetFinalConfirm">
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
              <NFormItem :label="text.subject">
                <NInput v-model:value="formValue.subject" :placeholder="text.subjectPlaceholder" />
              </NFormItem>
              <NFormItem :label="text.contentLabel">
                <NInput
                  v-model:value="formValue.content"
                  type="textarea"
                  :placeholder="text.contentPlaceholder"
                  :rows="12"
                  style="font-family: 'Consolas', 'Monaco', monospace; font-size: 13px;"
                />
              </NFormItem>
              <NFormItem :label="text.description">
                <NInput v-model:value="formValue.description" :placeholder="text.descriptionPlaceholder" />
              </NFormItem>
              <NFormItem :label="text.status" style="margin-bottom: 8px;">
                <NRadioGroup v-model:value="formValue.status">
                  <NRadioButton :value="1">{{ text.enabled }}</NRadioButton>
                  <NRadioButton :value="0">{{ text.disabled }}</NRadioButton>
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

        <!-- Right: Preview Panel -->
        <div class="preview-panel">
          <div class="preview-panel-header">
            <span class="preview-title">{{ text.previewTab }}</span>
            <NButton size="tiny" :loading="previewLoading" @click="refreshPreview">
              {{ text.refresh }}
            </NButton>
          </div>
          <div class="preview-frame-wrapper">
            <iframe
              v-if="previewHtml"
              class="preview-iframe"
              :srcdoc="previewHtml"
              sandbox="allow-same-origin allow-scripts"
            ></iframe>
            <div v-else class="preview-empty">
              {{ text.loadingMsg }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">
            {{ text.cancel }}
          </NButton>
          <NButton type="primary" @click="handleSave">
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
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
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
  min-width: 50px;
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

.preview-frame-wrapper {
  flex: 1;
  border: 1px solid #e8e8f0;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f5f5;
}

.preview-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: #fff;
}

.preview-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #aaa;
  font-size: 14px;
}
</style>
