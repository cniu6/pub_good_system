<script setup lang="ts">
/**
 * 公告正文预览弹窗（MdPreview）：铃铛 / 工作台 / 登录弹窗共用
 */
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NModal, NButton, NSpace, NTag } from 'naive-ui'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/preview.css'
import { useAppStore } from '@/store'

const props = defineProps<{
  show: boolean
  title?: string
  content?: string
  type?: string
}>()

const emit = defineEmits<{
  (e: 'update:show', v: boolean): void
  (e: 'close'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const editorTheme = computed(() => (appStore.colorMode === 'dark' ? 'dark' : 'light'))

const visible = computed({
  get: () => props.show,
  set: (v: boolean) => emit('update:show', v),
})

const tagType = computed(() => {
  const m: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
    info: 'info',
    success: 'success',
    warning: 'warning',
    error: 'error',
  }
  return m[props.type || 'info'] || 'info'
})

function handleClose() {
  visible.value = false
  emit('close')
}

// 切换内容时重置滚动（由 Modal 自身处理即可）
watch(() => props.content, () => {})
</script>

<template>
  <NModal
    v-model:show="visible"
    preset="card"
    :title="title || t('announcements.previewTitle')"
    style="width: min(720px, 92vw); max-height: 85vh;"
    :bordered="false"
    :segmented="{ content: true, footer: true }"
    @after-leave="emit('close')"
  >
    <template #header-extra>
      <NTag v-if="type" :type="tagType" size="small" :bordered="false">
        {{ t(`announcements.type.${type || 'info'}`) }}
      </NTag>
    </template>
    <div class="announcement-preview-body">
      <MdPreview
        :model-value="content || ''"
        :theme="editorTheme"
        language="zh-CN"
      />
    </div>
    <template #footer>
      <NSpace justify="end">
        <NButton type="primary" @click="handleClose">
          {{ t('announcements.gotIt') }}
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>

<style scoped>
.announcement-preview-body {
  max-height: 60vh;
  overflow: auto;
}
</style>
