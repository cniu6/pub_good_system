<script setup lang="ts">
import { computed, h, nextTick, onMounted, ref, watch } from 'vue'
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

const text = {
  pageTitle: '\u90AE\u4EF6\u6A21\u677F\u7BA1\u7406',
  refresh: '\u5237\u65B0',
  infoTip: '\u90AE\u4EF6\u6A21\u677F\u652F\u6301\u53D8\u91CF\u66FF\u6362\uFF0C\u4F7F\u7528 {变量名} \u683C\u5F0F\uFF0C\u4F8B\u5982 {code} \u4F1A\u88AB\u66FF\u6362\u4E3A\u9A8C\u8BC1\u7801\u3002',
  lang: '\u8BED\u8A00',
  subject: '\u4E3B\u9898',
  description: '\u63CF\u8FF0',
  status: '\u72B6\u6001',
  action: '\u64CD\u4F5C',
  enabled: '\u542F\u7528',
  disabled: '\u7981\u7528',
  unknown: '\u672A\u77E5',
  edit: '\u4FEE\u6539',
  reset: '\u91CD\u7F6E\u4E3A\u9ED8\u8BA4',
  resetConfirmFirst: '\u786E\u8BA4\u8981\u91CD\u7F6E\u6B64\u6A21\u677F\u4E3A\u9ED8\u8BA4\u5185\u5BB9\u5417\uFF1F',
  resetConfirmSecond: '\u6B64\u64CD\u4F5C\u4E0D\u53EF\u64A4\u9500\uFF0C\u786E\u8BA4\u91CD\u7F6E\uFF1F',
  loadFailed: '\u52A0\u8F7D\u90AE\u4EF6\u6A21\u677F\u5931\u8D25',
  resetSuccess: '\u91CD\u7F6E\u6210\u529F',
  resetFailed: '\u91CD\u7F6E\u5931\u8D25',
  saveSuccess: '\u4FDD\u5B58\u6210\u529F',
  saveFailed: '\u4FDD\u5B58\u5931\u8D25',
  previewFailed: '\u9884\u89C8\u5931\u8D25',
  editModalTitle: '\u7F16\u8F91\u90AE\u4EF6\u6A21\u677F',
  subjectRequired: '\u4E3B\u9898\u4E0D\u80FD\u4E3A\u7A7A',
  contentRequired: '\u5185\u5BB9\u4E0D\u80FD\u4E3A\u7A7A',
  subjectPlaceholder: '\u8BF7\u8F93\u5165\u90AE\u4EF6\u4E3B\u9898',
  contentPlaceholder: '\u8BF7\u8F93\u5165\u90AE\u4EF6\u5185\u5BB9\uFF08\u652F\u6301 HTML\uFF09',
  descriptionPlaceholder: '\u8BF7\u8F93\u5165\u6A21\u677F\u63CF\u8FF0',
  cancel: '\u53D6\u6D88',
  save: '\u4FDD\u5B58',
  registerCode: '\u6CE8\u518C\u9A8C\u8BC1\u7801',
  resetPassword: '\u91CD\u7F6E\u5BC6\u7801',
  sendTest: '\u53D1\u4EF6\u6D4B\u8BD5',
  sendTestDesc: '\u9A8C\u8BC1 SMTP \u914D\u7F6E\u662F\u5426\u6B63\u5E38\uFF0C\u53D1\u9001\u4E00\u5C01\u6D4B\u8BD5\u90AE\u4EF6\u5230\u6307\u5B9A\u90AE\u7BB1\u3002',
  testTo: '\u6536\u4EF6\u90AE\u7BB1',
  testToPlaceholder: '\u8BF7\u8F93\u5165\u6D4B\u8BD5\u6536\u4EF6\u90AE\u7BB1',
  testSubject: '\u90AE\u4EF6\u4E3B\u9898',
  testSubjectPlaceholder: '\u7559\u7A7A\u5219\u4F7F\u7528\u9ED8\u8BA4\u4E3B\u9898',
  sending: '\u53D1\u9001\u4E2D...',
  send: '\u53D1\u9001\u6D4B\u8BD5',
  sendSuccess: '\u6D4B\u8BD5\u90AE\u4EF6\u5DF2\u53D1\u9001',
  sendFailed: '\u53D1\u9001\u5931\u8D25',
  fullscreen: '\u5168\u5C4F',
  exitFullscreen: '\u9000\u51FA\u5168\u5C4F',
  previewTab: '\u9884\u89C8',
  variables: '\u53D8\u91CF',
  varPlaceholder: '\u8BF7\u8F93\u5165\u53D8\u91CF\u503C',
  contentLabel: '\u5185\u5BB9 (HTML)',
  confirmBtn: '\u786E\u8BA4',
  cancelBtn: '\u53D6\u6D88',
  finalConfirmBtn: '\u786E\u5B9A\u91CD\u7F6E',
  noVarsMsg: '\u5F53\u524D\u6A21\u677F\u672A\u5B9A\u4E49\u53D8\u91CF',
  loadingMsg: '\u52A0\u8F7D\u4E2D...',
  inputEmail: '\u8BF7\u8F93\u5165\u6536\u4EF6\u90AE\u7BB1',
  selectTemplate: '\u9009\u62E9\u6A21\u677F',
  noTemplate: '\u4E0D\u4F7F\u7528\u6A21\u677F',
} as const

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
    const name = templateNameMap[tpl.name] || tpl.name
    const lang = langMap[tpl.lang] || tpl.lang
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

const langMap: Record<string, string> = {
  'zh-CN': '\u4E2D\u6587',
  'en-US': 'English',
}

const statusMap: Record<number, { label: string; type: 'success' | 'error' }> = {
  0: { label: text.disabled, type: 'error' },
  1: { label: text.enabled, type: 'success' },
}

const groupedTemplates = computed(() => {
  const groups: Record<string, EmailTemplate[]> = {}
  templates.value.forEach((tpl) => {
    if (!groups[tpl.name])
      groups[tpl.name] = []
    groups[tpl.name].push(tpl)
  })
  return groups
})

const templateNameMap: Record<string, string> = {
  register_code: text.registerCode,
  reset_password: text.resetPassword,
}

const columns = [
  {
    title: text.lang,
    key: 'lang',
    width: 100,
    render: (row: EmailTemplate) => h(NTag, { type: 'info' }, () => langMap[row.lang] || row.lang),
  },
  {
    title: text.subject,
    key: 'subject',
    ellipsis: { tooltip: true },
  },
  {
    title: text.description,
    key: 'description',
    ellipsis: { tooltip: true },
  },
  {
    title: text.status,
    key: 'status',
    width: 80,
    render: (row: EmailTemplate) => {
      const status = statusMap[row.status]
      return h(NTag, { type: status?.type || 'default' }, () => status?.label || text.unknown)
    },
  },
  {
    title: text.action,
    key: 'actions',
    width: 100,
    render: (row: EmailTemplate) => {
      return h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, () => text.edit)
    },
  },
]

// ---- Data ----
async function loadData() {
  loading.value = true
  try {
    const result = await fetchEmailTemplateList()
    if (result.data)
      templates.value = result.data
  }
  catch (error) {
    console.error(text.loadFailed, error)
    window.$message?.error(text.loadFailed)
  }
  finally {
    loading.value = false
  }
}

// ---- Send Test ----
async function handleSendTest() {
  if (!testTo.value.trim()) {
    window.$message?.warning(text.inputEmail)
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
      window.$message?.success(text.sendSuccess)
  }
  catch (error: any) {
    console.error(text.sendFailed, error)
    window.$message?.error(error?.message || text.sendFailed)
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
    console.error(text.previewFailed, error)
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
    window.$message?.warning(text.subjectRequired)
    return
  }
  if (!formValue.value.content.trim()) {
    window.$message?.warning(text.contentRequired)
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
      window.$message?.success(text.saveSuccess)
      showModal.value = false
      await loadData()
    }
  }
  catch (error) {
    console.error(text.saveFailed, error)
    window.$message?.error(text.saveFailed)
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
      window.$message?.success(text.resetSuccess)
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
    console.error(text.resetFailed, error)
    window.$message?.error(text.resetFailed)
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
