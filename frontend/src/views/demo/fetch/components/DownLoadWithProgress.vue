<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import {
  downloadFile,
} from '@/service'
import { useRequest } from 'alova/client'
import { normalizeSizeUnits } from '@/utils'

const { t } = useI18n()

const emit = defineEmits<{
  update: [data: any]
}>()

const fileURL = ref('https://images.unsplash.com/photo-1663529628961-80aa6ebcd157?ixlib=rb-1.2.1&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=764&q=80')

const { loading, downloading, abort, send, data } = useRequest(() => downloadFile(fileURL.value), {
  immediate: false,
})

const downloadProcess = computed(() => {
  if (!downloading.value.loaded)
    return 0
  return Math.floor(downloading.value.loaded / downloading.value.total * 100)
})

async function handleDownloadFile() {
  await send()
  emit('update', 'fileOk')
  const urlObj = new URL(fileURL.value)
  const pathname = urlObj.pathname
  const name = pathname.split('/')[1]
  downloadLink(data.value, name)
}

function downloadLink(data: Blob, name: string) {
  const link = URL.createObjectURL(data)
  const eleLink = document.createElement('a')
  eleLink.download = name
  eleLink.style.display = 'none'
  eleLink.href = link
  document.body.appendChild(eleLink)
  eleLink.click()
  document.body.removeChild(eleLink)
}
</script>

<template>
  <n-card :title="t('demo.fetch.downloadWithProgress')" size="small">
    <n-space vertical>
      <n-input v-model:value="fileURL" />
      <div>{{ t('demo.fetch.fileSize') }}：{{ normalizeSizeUnits(downloading.total) }}</div>
      <div>{{ t('demo.fetch.downloaded') }}：{{ normalizeSizeUnits(downloading.loaded) }}</div>
      <n-progress type="line" indicator-placement="inside" :processing="loading" :percentage="downloadProcess" />
      <n-space>
        <n-button strong secondary @click="handleDownloadFile">
          {{ t('demo.fetch.startDownload') }}
        </n-button>
        <n-button strong secondary type="warning" @click="abort">
          {{ t('demo.fetch.abortDownload') }}
        </n-button>
      </n-space>
    </n-space>
  </n-card>
</template>

<style scoped>

</style>
