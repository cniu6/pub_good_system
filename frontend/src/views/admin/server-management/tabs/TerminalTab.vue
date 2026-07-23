<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { adminApi } from '@/service/api/admin'

const { t } = useI18n()
const message = useMessage()

const cmd = ref('echo hello')
const output = ref('')
const loading = ref(false)
const unavailable = ref(false)
const unavailableTip = ref('')

async function probeAvailability() {
  // 用一条无害命令探测是否开启 debug ops；403 则提示
  try {
    const res = await adminApi.terminal.exec('echo terminal-probe')
    if (!res.isSuccess) {
      const code = Number((res as any).code || 0)
      if (code === 403 || code === 404) {
        unavailable.value = true
        unavailableTip.value = res.message || t('adminServer.terminalUnavailable')
        return
      }
    }
    unavailable.value = false
  }
  catch (e: any) {
    const msg = String(e?.message || '')
    if (msg.includes('403') || msg.includes('404') || msg.includes('debug')) {
      unavailable.value = true
      unavailableTip.value = t('adminServer.terminalUnavailable')
    }
  }
}

async function runCmd() {
  const line = cmd.value.trim()
  if (!line) {
    message.warning(t('adminServer.terminalEmpty'))
    return
  }
  loading.value = true
  try {
    const res = await adminApi.terminal.exec(line)
    if (!res.isSuccess) {
      const code = Number((res as any).code || 0)
      if (code === 403 || code === 404) {
        unavailable.value = true
        unavailableTip.value = res.message || t('adminServer.terminalUnavailable')
      }
      output.value = res.message || t('adminServer.terminalFailed')
      message.error(res.message || t('adminServer.terminalFailed'))
      return
    }
    unavailable.value = false
    const data = res.data
    const parts = [data?.output || '']
    if (data?.error)
      parts.push(`[exit] ${data.error}`)
    output.value = parts.filter(Boolean).join('\n')
    if (data?.success)
      message.success(t('adminServer.terminalSuccess'))
    else
      message.warning(t('adminServer.terminalFailed'))
  }
  catch (e: any) {
    output.value = e?.message || t('adminServer.terminalFailed')
    message.error(e?.message || t('adminServer.terminalFailed'))
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  void probeAvailability()
})
</script>

<template>
  <n-card :title="t('adminServer.terminalTab')" :bordered="false" size="small">
    <n-alert v-if="unavailable" type="warning" class="mb-12px">
      {{ unavailableTip || t('adminServer.terminalUnavailable') }}
    </n-alert>
    <n-space vertical :size="12">
      <n-input
        v-model:value="cmd"
        type="textarea"
        :rows="3"
        :placeholder="t('adminServer.terminalPlaceholder')"
        :disabled="unavailable"
      />
      <n-space>
        <n-button type="primary" :loading="loading" :disabled="unavailable" @click="runCmd">
          {{ t('adminServer.terminalRun') }}
        </n-button>
        <n-button quaternary :disabled="!output" @click="output = ''">
          {{ t('common.reset') }}
        </n-button>
      </n-space>
      <n-text depth="3">
        {{ t('adminServer.terminalOutput') }}
      </n-text>
      <pre class="terminal-output">{{ output || t('adminServer.terminalEmptyOutput') }}</pre>
    </n-space>
  </n-card>
</template>

<style scoped>
.terminal-output {
  margin: 0;
  padding: 12px;
  min-height: 180px;
  max-height: 480px;
  overflow: auto;
  border-radius: 6px;
  background: rgba(127, 127, 127, 0.12);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
