<script setup lang="ts">
/**
 * 管理端：站内公告 CRUD（MdEditor 编辑正文）
 */
import { computed, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton,
  NCard,
  NDataTable,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { adminAnnouncementApi, type AdminAnnouncement, type AnnouncementUpsertPayload } from '@/service/api/admin/announcement'
import { useAppStore } from '@/store'

const { t } = useI18n()
const message = useMessage()
const appStore = useAppStore()
const editorTheme = computed(() => (appStore.colorMode === 'dark' ? 'dark' : 'light'))

const loading = ref(false)
const list = ref<AdminAnnouncement[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const statusFilter = ref<number | null>(null)

const showModal = ref(false)
const saving = ref(false)
const editingId = ref<number | null>(null)
const form = ref<AnnouncementUpsertPayload>({
  title: '',
  summary: '',
  content: '',
  type: 'info',
  priority: 0,
  popup: 0,
  target_type: 'all',
  target_value: '',
  start_at: 0,
  end_at: 0,
})

const statusOptions = computed(() => [
  { label: t('announcements.statusAll'), value: null },
  { label: t('announcements.statusDraft'), value: 0 },
  { label: t('announcements.statusPublished'), value: 1 },
  { label: t('announcements.statusUnpublished'), value: 2 },
])

const typeOptions = computed(() => [
  { label: t('announcements.type.info'), value: 'info' },
  { label: t('announcements.type.success'), value: 'success' },
  { label: t('announcements.type.warning'), value: 'warning' },
  { label: t('announcements.type.error'), value: 'error' },
])

function formatTime(ts: number) {
  if (!ts)
    return '-'
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function statusTag(status: number) {
  if (status === 1)
    return h(NTag, { type: 'success', size: 'small', bordered: false }, () => t('announcements.statusPublished'))
  if (status === 2)
    return h(NTag, { type: 'warning', size: 'small', bordered: false }, () => t('announcements.statusUnpublished'))
  return h(NTag, { type: 'default', size: 'small', bordered: false }, () => t('announcements.statusDraft'))
}

const columns = computed<DataTableColumns<AdminAnnouncement>>(() => [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('announcements.colTitle'), key: 'title', ellipsis: { tooltip: true } },
  {
    title: t('announcements.colType'),
    key: 'type',
    width: 90,
    render: row => h(NTag, { type: (row.type as any) || 'info', size: 'small', bordered: false }, () => t(`announcements.type.${row.type || 'info'}`)),
  },
  { title: t('announcements.colStatus'), key: 'status', width: 90, render: row => statusTag(row.status) },
  { title: t('announcements.colPriority'), key: 'priority', width: 70 },
  {
    title: t('announcements.colPopup'),
    key: 'popup',
    width: 70,
    render: row => (row.popup ? t('announcements.yes') : t('announcements.no')),
  },
  { title: t('announcements.colPublishedAt'), key: 'published_at', width: 160, render: row => formatTime(row.published_at) },
  {
    title: t('announcements.colAction'),
    key: 'actions',
    width: 280,
    render: (row) => {
      return h(NSpace, { size: 8 }, () => [
        h(NButton, { size: 'tiny', onClick: () => openEdit(row) }, () => t('common.edit')),
        row.status !== 1
          ? h(NButton, { size: 'tiny', type: 'primary', onClick: () => doPublish(row.id) }, () => t('announcements.publish'))
          : h(NButton, { size: 'tiny', onClick: () => doUnpublish(row.id) }, () => t('announcements.unpublish')),
        h(NButton, { size: 'tiny', type: 'error', onClick: () => doDelete(row.id) }, () => t('common.delete')),
      ])
    },
  },
])

async function loadList() {
  loading.value = true
  try {
    const res = await adminAnnouncementApi.list({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value || undefined,
      status: statusFilter.value === null ? undefined : statusFilter.value,
    })
    if (res.isSuccess && res.data) {
      list.value = res.data.list || []
      total.value = res.data.total || 0
    }
    else {
      message.error(res.message || t('announcements.loadFailed'))
    }
  }
  catch (e) {
    if (import.meta.env.DEV)
      console.error(e)
    message.error(t('announcements.loadFailed'))
  }
  finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  form.value = {
    title: '',
    summary: '',
    content: '',
    type: 'info',
    priority: 0,
    popup: 0,
    target_type: 'all',
    target_value: '',
    start_at: 0,
    end_at: 0,
  }
  showModal.value = true
}

function openEdit(row: AdminAnnouncement) {
  editingId.value = row.id
  form.value = {
    title: row.title,
    summary: row.summary || '',
    content: row.content,
    type: row.type || 'info',
    priority: row.priority,
    popup: row.popup,
    target_type: row.target_type === 'role' ? (row.target_value || 'user') : 'all',
    target_value: row.target_value,
    start_at: row.start_at,
    end_at: row.end_at,
  }
  showModal.value = true
}

function buildPayload(): AnnouncementUpsertPayload {
  // 公告面向全体登录用户，不做管理员/用户分层定向
  return {
    title: form.value.title.trim(),
    summary: (form.value.summary || '').trim(),
    content: form.value.content,
    type: form.value.type,
    priority: form.value.priority || 0,
    popup: form.value.popup ? 1 : 0,
    target_type: 'all',
    target_value: '',
    start_at: form.value.start_at || 0,
    end_at: form.value.end_at || 0,
  }
}

async function handleSave() {
  if (!form.value.title.trim() || !form.value.content.trim()) {
    message.warning(t('announcements.titleContentRequired'))
    return
  }
  saving.value = true
  try {
    const payload = buildPayload()
    const res = editingId.value
      ? await adminAnnouncementApi.update(editingId.value, payload)
      : await adminAnnouncementApi.create(payload)
    if (res.isSuccess) {
      message.success(t('announcements.saveSuccess'))
      showModal.value = false
      await loadList()
    }
    else {
      message.error(res.message || t('announcements.saveFailed'))
    }
  }
  catch (e) {
    if (import.meta.env.DEV)
      console.error(e)
    message.error(t('announcements.saveFailed'))
  }
  finally {
    saving.value = false
  }
}

async function doPublish(id: number) {
  const res = await adminAnnouncementApi.publish(id)
  if (res.isSuccess) {
    message.success(t('announcements.publishSuccess'))
    await loadList()
  }
  else {
    message.error(res.message || t('announcements.actionFailed'))
  }
}

async function doUnpublish(id: number) {
  const res = await adminAnnouncementApi.unpublish(id)
  if (res.isSuccess) {
    message.success(t('announcements.unpublishSuccess'))
    await loadList()
  }
  else {
    message.error(res.message || t('announcements.actionFailed'))
  }
}

async function doDelete(id: number) {
  const res = await adminAnnouncementApi.remove(id)
  if (res.isSuccess) {
    message.success(t('announcements.deleteSuccess'))
    await loadList()
  }
  else {
    message.error(res.message || t('announcements.actionFailed'))
  }
}

onMounted(() => loadList())
</script>

<template>
  <NCard :title="t('announcements.pageTitle')">
    <template #header-extra>
      <NSpace>
        <NInput
          v-model:value="keyword"
          clearable
          :placeholder="t('announcements.searchPlaceholder')"
          style="width: 200px"
          @keyup.enter="loadList"
        />
        <NSelect
          v-model:value="statusFilter"
          :options="statusOptions"
          style="width: 140px"
          @update:value="loadList"
        />
        <NButton @click="loadList">
          {{ t('common.refresh') }}
        </NButton>
        <NButton type="primary" @click="openCreate">
          {{ t('announcements.create') }}
        </NButton>
      </NSpace>
    </template>

    <NDataTable
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="{
        page,
        pageSize,
        itemCount: total,
        onUpdatePage: (p: number) => { page = p; loadList() },
        onUpdatePageSize: (s: number) => { pageSize = s; page = 1; loadList() },
      }"
      :bordered="false"
    />

    <NModal
      v-model:show="showModal"
      preset="card"
      :title="editingId ? t('announcements.editTitle') : t('announcements.createTitle')"
      style="width: min(900px, 94vw)"
      :segmented="{ content: true, footer: true }"
    >
      <NForm label-placement="left" label-width="100">
        <NFormItem :label="t('announcements.colTitle')" required>
          <NInput v-model:value="form.title" />
        </NFormItem>
        <NFormItem :label="t('announcements.colType')">
          <NSelect v-model:value="form.type" :options="typeOptions" />
        </NFormItem>
        <NFormItem :label="t('announcements.colPriority')">
          <NInputNumber v-model:value="form.priority" :min="0" />
        </NFormItem>
        <NFormItem :label="t('announcements.colPopup')">
          <NSwitch :value="!!form.popup" @update:value="(v: boolean) => { form.popup = v ? 1 : 0 }" />
        </NFormItem>
        <NFormItem :label="t('announcements.colSummary')">
          <NInput
            v-model:value="form.summary"
            type="textarea"
            :rows="2"
            :maxlength="80"
            show-count
            :placeholder="t('announcements.summaryPlaceholder')"
          />
        </NFormItem>
        <NFormItem :label="t('announcements.colContent')" required>
          <MdEditor
            v-model="form.content"
            :theme="editorTheme"
            language="zh-CN"
            style="height: 360px; width: 100%"
          />
        </NFormItem>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="saving" @click="handleSave">
            {{ t('common.save') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </NCard>
</template>
