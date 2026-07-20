<script setup lang="ts">
/**
 * 顶栏铃铛：三 Tab 壳保留；「通知」= 站内公告；消息/待办空态。
 * 点击条目弹出可关闭 MdPreview 预览窗并标记已读。
 */
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import NoticeList from '../common/NoticeList.vue'
import AnnouncementPreviewModal from '@/components/common/AnnouncementPreviewModal.vue'
import { userAnnouncementApi, type UserAnnouncementItem } from '@/service/api/user/announcement'
import { useAuthStore } from '@/store'

const { t } = useI18n()
const message = useMessage()
const authStore = useAuthStore()

const currentTab = ref(0)
const loading = ref(false)
const enabled = ref(false)
const announcements = ref<UserAnnouncementItem[]>([])
const unreadCount = ref(0)

const previewShow = ref(false)
const previewItem = ref<UserAnnouncementItem | null>(null)

const typeIcon: Record<string, string> = {
  info: 'icon-park-outline:tips-one',
  success: 'icon-park-outline:check-one',
  warning: 'icon-park-outline:attention',
  error: 'icon-park-outline:close-one',
}

function formatTime(ts?: number) {
  if (!ts)
    return ''
  const d = new Date(ts * 1000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const noticeList = computed<Entity.Message[]>(() => {
  return announcements.value.map(a => ({
    id: a.id,
    type: 0,
    title: a.title,
    icon: typeIcon[a.type] || typeIcon.info,
    tagTitle: t(`announcements.type.${a.type || 'info'}`),
    tagType: (a.type as any) || 'info',
    description: a.summary || '',
    date: formatTime(a.published_at),
    isRead: a.is_read,
  }))
})

const badgeTotal = computed(() => unreadCount.value)

async function refresh() {
  if (!authStore.token)
    return
  loading.value = true
  try {
    const [listRes, countRes] = await Promise.all([
      userAnnouncementApi.list({ limit: 50 }),
      userAnnouncementApi.unreadCount(),
    ])
    if (listRes.isSuccess && listRes.data) {
      enabled.value = !!listRes.data.enabled
      announcements.value = listRes.data.list || []
    }
    if (countRes.isSuccess && countRes.data) {
      unreadCount.value = countRes.data.count || 0
      enabled.value = countRes.data.enabled ?? enabled.value
    }
  }
  catch (e) {
    if (import.meta.env.DEV)
      console.error('[Notices] load failed', e)
  }
  finally {
    loading.value = false
  }
}

async function openPreview(id: number) {
  const item = announcements.value.find(a => a.id === id)
  if (!item)
    return
  // 若列表无全文，再拉详情
  let content = item.content
  if (!content || content === item.summary) {
    try {
      const res = await userAnnouncementApi.detail(id)
      if (res.isSuccess && res.data)
        content = res.data.content
    }
    catch { /* ignore */ }
  }
  previewItem.value = { ...item, content: content || item.content }
  previewShow.value = true
  try {
    await userAnnouncementApi.markRead(id)
    const wasUnread = !item.is_read
    item.is_read = true
    if (wasUnread)
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    const c = await userAnnouncementApi.unreadCount()
    if (c.isSuccess && c.data)
      unreadCount.value = c.data.count || 0
  }
  catch { /* ignore */ }
}

async function markAllRead() {
  try {
    const res = await userAnnouncementApi.markAllRead()
    if (res.isSuccess) {
      announcements.value.forEach(a => { a.is_read = true })
      unreadCount.value = 0
      message.success(t('announcements.markAllReadSuccess'))
    }
  }
  catch {
    message.error(t('announcements.actionFailed'))
  }
}

function onPresenceAnnouncement() {
  refresh()
}

onMounted(() => {
  refresh()
  window.addEventListener('fst:announcement', onPresenceAnnouncement)
})
onUnmounted(() => {
  window.removeEventListener('fst:announcement', onPresenceAnnouncement)
})

defineExpose({ refresh })
</script>

<template>
  <n-popover placement="bottom" trigger="click" arrow-point-to-center class="!p-0">
    <template #trigger>
      <n-tooltip placement="bottom" trigger="hover">
        <template #trigger>
          <CommonWrapper>
            <n-badge :value="badgeTotal" :max="99" style="color: unset">
              <icon-park-outline-remind />
            </n-badge>
          </CommonWrapper>
        </template>
        <span>{{ $t('app.notificationsTips') }}</span>
      </n-tooltip>
    </template>
    <n-tabs v-model:value="currentTab" type="line" animated justify-content="space-evenly" class="w-390px">
      <n-tab-pane :name="0">
        <template #tab>
          <n-space class="w-130px" justify="center">
            {{ $t('app.notifications') }}
            <n-badge type="info" :value="badgeTotal" :max="99" />
          </n-space>
        </template>
        <!-- 开关关或列表空：统一「暂无」，不向用户暴露「未启用」管理态 -->
        <div class="flex justify-end px-12px py-6px">
          <n-button
            text
            type="primary"
            size="tiny"
            :disabled="!enabled || badgeTotal === 0"
            @click="markAllRead"
          >
            {{ t('announcements.markAllRead') }}
          </n-button>
        </div>
        <NoticeList v-if="enabled && noticeList.length" :list="noticeList" @read="openPreview" />
        <div v-else class="px-16px py-32px text-center opacity-60 text-13px">
          {{ t('announcements.empty') }}
        </div>
      </n-tab-pane>
      <n-tab-pane :name="1">
        <template #tab>
          <n-space class="w-130px" justify="center">
            {{ $t('app.messages') }}
            <n-badge type="warning" :value="0" :max="99" />
          </n-space>
        </template>
        <div class="px-16px py-32px text-center opacity-60 text-13px">
          {{ t('announcements.messagesEmpty') }}
        </div>
      </n-tab-pane>
      <n-tab-pane :name="2">
        <template #tab>
          <n-space class="w-130px" justify="center">
            {{ $t('app.todos') }}
            <n-badge type="error" :value="0" :max="99" />
          </n-space>
        </template>
        <div class="px-16px py-32px text-center opacity-60 text-13px">
          {{ t('announcements.todosEmpty') }}
        </div>
      </n-tab-pane>
    </n-tabs>
  </n-popover>

  <AnnouncementPreviewModal
    v-model:show="previewShow"
    :title="previewItem?.title"
    :content="previewItem?.content"
    :type="previewItem?.type"
  />
</template>

<style scoped></style>
