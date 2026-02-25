<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
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
  NPopconfirm,
  NRadioButton,
  NRadioGroup,
  NSpace,
  NTabPane,
  NTabs,
  NTag,
  NText,
} from 'naive-ui'
import {
  fetchEmailTemplateList,
  fetchPreviewEmailTemplate,
  fetchResetEmailTemplate,
  fetchUpdateEmailTemplate,
  type EmailTemplate,
} from '@/service/api/admin/email-template'

const text = {
  pageTitle: '\u90ae\u4ef6\u6a21\u677f\u7ba1\u7406',
  refresh: '\u5237\u65b0',
  infoTip: '\u90ae\u4ef6\u6a21\u677f\u652f\u6301\u53d8\u91cf\u66ff\u6362\uff0c\u4f7f\u7528 `{\u53d8\u91cf\u540d}` \u683c\u5f0f\uff0c\u4f8b\u5982 `{code}` \u4f1a\u88ab\u66ff\u6362\u4e3a\u9a8c\u8bc1\u7801\u3002',
  lang: '\u8bed\u8a00',
  subject: '\u4e3b\u9898',
  content: '\u5185\u5bb9',
  description: '\u63cf\u8ff0',
  status: '\u72b6\u6001',
  action: '\u64cd\u4f5c',
  enabled: '\u542f\u7528',
  disabled: '\u7981\u7528',
  unknown: '\u672a\u77e5',
  edit: '\u7f16\u8f91',
  preview: '\u9884\u89c8',
  reset: '\u91cd\u7f6e',
  resetConfirm: '\u786e\u8ba4\u8981\u91cd\u7f6e\u4e3a\u9ed8\u8ba4\u6a21\u677f\u5417\uff1f',
  loadFailed: '\u52a0\u8f7d\u90ae\u4ef6\u6a21\u677f\u5931\u8d25',
  resetSuccess: '\u91cd\u7f6e\u6210\u529f',
  resetFailed: '\u91cd\u7f6e\u5931\u8d25',
  saveSuccess: '\u4fdd\u5b58\u6210\u529f',
  saveFailed: '\u4fdd\u5b58\u5931\u8d25',
  previewUpdated: '\u9884\u89c8\u5df2\u66f4\u65b0',
  previewFailed: '\u9884\u89c8\u5931\u8d25',
  editModalTitle: '\u7f16\u8f91\u90ae\u4ef6\u6a21\u677f',
  previewModalTitle: '\u9884\u89c8\u90ae\u4ef6\u6a21\u677f',
  subjectRequired: '\u4e3b\u9898\u4e0d\u80fd\u4e3a\u7a7a',
  contentRequired: '\u5185\u5bb9\u4e0d\u80fd\u4e3a\u7a7a',
  subjectPlaceholder: '\u8bf7\u8f93\u5165\u90ae\u4ef6\u4e3b\u9898',
  contentPlaceholder: '\u8bf7\u8f93\u5165\u90ae\u4ef6\u5185\u5bb9\uff08\u652f\u6301 HTML\uff09',
  descriptionPlaceholder: '\u8bf7\u8f93\u5165\u6a21\u677f\u63cf\u8ff0',
  cancel: '\u53d6\u6d88',
  save: '\u4fdd\u5b58',
  varsTab: '\u53d8\u91cf\u8bbe\u7f6e',
  resultTab: '\u9884\u89c8\u7ed3\u679c',
  renderPreview: '\u6e32\u67d3\u9884\u89c8',
  noVars: '\u5f53\u524d\u6a21\u677f\u672a\u5b9a\u4e49\u53d8\u91cf\u3002',
  registerCode: '\u6ce8\u518c\u9a8c\u8bc1\u7801',
  resetPassword: '\u91cd\u7f6e\u5bc6\u7801',
} as const

const loading = ref(false)
const templates = ref<EmailTemplate[]>([])
const showModal = ref(false)
const showPreviewModal = ref(false)
const currentTemplate = ref<EmailTemplate | null>(null)

const previewTemplateContent = ref('')
const previewRenderedContent = ref('')
const previewVars = ref<Record<string, string>>({})

const formValue = ref({
  subject: '',
  content: '',
  description: '',
  status: 1 as number,
})

const langMap: Record<string, string> = {
  'zh-CN': '\u4e2d\u6587',
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
    width: 180,
    render: (row: EmailTemplate) => {
      return h(NSpace, {}, () => [
        h(NButton, { size: 'small', onClick: () => handleEdit(row) }, () => text.edit),
        h(NButton, { size: 'small', onClick: () => handlePreview(row) }, () => text.preview),
        h(
          NPopconfirm,
          {
            onPositiveClick: () => handleReset(row),
          },
          {
            trigger: () => h(NButton, { size: 'small', type: 'warning' }, () => text.reset),
            default: () => text.resetConfirm,
          },
        ),
      ])
    },
  },
]

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

function handleEdit(template: EmailTemplate) {
  currentTemplate.value = template
  formValue.value = {
    subject: template.subject,
    content: template.content,
    description: template.description || '',
    status: template.status,
  }
  showModal.value = true
}

function handlePreview(template: EmailTemplate) {
  currentTemplate.value = template
  previewTemplateContent.value = template.content
  previewRenderedContent.value = template.content

  const vars: Record<string, string> = {}
  if (template.variables) {
    template.variables.split(',').forEach((v) => {
      const key = v.trim()
      if (key)
        vars[key] = ''
    })
  }
  previewVars.value = vars
  showPreviewModal.value = true
}

async function handleReset(template: EmailTemplate) {
  try {
    const result = await fetchResetEmailTemplate(template.id)
    if (result.data) {
      window.$message?.success(text.resetSuccess)
      await loadData()
    }
  }
  catch (error) {
    console.error(text.resetFailed, error)
    window.$message?.error(text.resetFailed)
  }
}

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

async function handlePreviewRender() {
  if (!currentTemplate.value)
    return

  try {
    const result = await fetchPreviewEmailTemplate(currentTemplate.value.id, {
      content: previewTemplateContent.value,
      vars: previewVars.value,
    })
    if (result.data) {
      previewRenderedContent.value = result.data.content
      window.$message?.success(text.previewUpdated)
    }
  }
  catch (error) {
    console.error(text.previewFailed, error)
    window.$message?.error(text.previewFailed)
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="email-template-page">
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

    <NModal
      v-model:show="showModal"
      preset="card"
      :title="text.editModalTitle"
      style="width: 800px; max-width: 90vw;"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="80">
        <NFormItem :label="text.subject">
          <NInput v-model:value="formValue.subject" :placeholder="text.subjectPlaceholder" />
        </NFormItem>
        <NFormItem :label="text.content">
          <NInput
            v-model:value="formValue.content"
            type="textarea"
            :placeholder="text.contentPlaceholder"
            :rows="15"
          />
        </NFormItem>
        <NFormItem :label="text.description">
          <NInput v-model:value="formValue.description" :placeholder="text.descriptionPlaceholder" />
        </NFormItem>
        <NFormItem :label="text.status">
          <NRadioGroup v-model:value="formValue.status">
            <NRadioButton :value="1">
              {{ text.enabled }}
            </NRadioButton>
            <NRadioButton :value="0">
              {{ text.disabled }}
            </NRadioButton>
          </NRadioGroup>
        </NFormItem>
      </NForm>
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

    <NModal
      v-model:show="showPreviewModal"
      preset="card"
      :title="text.previewModalTitle"
      style="width: 900px; max-width: 90vw;"
    >
      <NTabs type="line">
        <NTabPane name="vars" :tab="text.varsTab">
          <NSpace vertical>
            <NSpace v-if="Object.keys(previewVars).length > 0" vertical>
              <NSpace v-for="key in Object.keys(previewVars)" :key="key" align="center">
                <NText style="width: 96px; text-align: right;">
                  {{ key }}:
                </NText>
                <NInput
                  v-model:value="previewVars[key]"
                  :placeholder="`\u8bf7\u8f93\u5165 ${key} \u7684\u503c`"
                  style="flex: 1;"
                />
              </NSpace>
              <NButton type="primary" @click="handlePreviewRender">
                {{ text.renderPreview }}
              </NButton>
            </NSpace>
            <NAlert v-else type="warning">
              {{ text.noVars }}
            </NAlert>
          </NSpace>
        </NTabPane>
        <NTabPane name="preview" :tab="text.resultTab">
          <NSpace vertical>
            <div class="email-preview" v-html="previewRenderedContent"></div>
          </NSpace>
        </NTabPane>
      </NTabs>
    </NModal>
  </div>
</template>

<style scoped>
.email-preview {
  max-height: 400px;
  overflow-y: auto;
}
</style>
