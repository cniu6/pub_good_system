<script setup lang="ts">
const props = defineProps<{
  ddl: string
  loading: boolean
}>()

const emit = defineEmits<{ refresh: [] }>()
const { t } = useI18n()
const message = useMessage()

async function copyDdl() {
  if (!props.ddl)
    return
  try {
    await navigator.clipboard.writeText(props.ddl)
    message.success(t('adminServer.dbCopySuccess'))
  }
  catch {
    message.error(t('adminServer.dbCopyFailed'))
  }
}
</script>

<template>
  <NCard size="small" :title="t('adminServer.dbDdl')" :segmented="{ content: true }">
    <template #header-extra>
      <NSpace>
        <NButton size="small" :loading="loading" @click="emit('refresh')">
          {{ t('common.refresh') }}
        </NButton>
        <NButton size="small" type="primary" secondary :disabled="!ddl" @click="copyDdl">
          {{ t('adminServer.dbCopy') }}
        </NButton>
      </NSpace>
    </template>
    <NCode :code="ddl || t('adminServer.dbNoDdl')" language="sql" show-line-numbers word-wrap />
  </NCard>
</template>
