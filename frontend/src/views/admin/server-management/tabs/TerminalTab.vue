<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { withSubmitLock } from '@/hooks'
import { adminApi } from '@/service/api/admin'

const { t } = useI18n()
const message = useMessage()

const cmd = ref('echo hello')
const output = ref('')
const loading = ref(false)
const unavailable = ref(false)
const unavailableTip = ref('')
const outputEl = ref<HTMLElement | null>(null)

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

function scrollOutputToBottom() {
  requestAnimationFrame(() => {
    const el = outputEl.value
    if (el)
      el.scrollTop = el.scrollHeight
  })
}

async function runCmd() {
  const line = cmd.value.trim()
  if (!line) {
    message.warning(t('adminServer.terminalEmpty'))
    return
  }
  await withSubmitLock(loading, async () => {
    try {
      const res = await adminApi.terminal.exec(line)
      if (!res.isSuccess) {
        const code = Number((res as any).code || 0)
        if (code === 403 || code === 404) {
          unavailable.value = true
          unavailableTip.value = res.message || t('adminServer.terminalUnavailable')
        }
        output.value = `$ ${line}\n${res.message || t('adminServer.terminalFailed')}`
        message.error(res.message || t('adminServer.terminalFailed'))
        scrollOutputToBottom()
        return
      }
      unavailable.value = false
      const data = res.data
      const parts = [`$ ${line}`, data?.output || '']
      if (data?.error)
        parts.push(`[exit] ${data.error}`)
      output.value = parts.filter(Boolean).join('\n')
      scrollOutputToBottom()
      if (data?.success)
        message.success(t('adminServer.terminalSuccess'))
      else
        message.warning(t('adminServer.terminalFailed'))
    }
    catch (e: any) {
      output.value = `$ ${line}\n${e?.message || t('adminServer.terminalFailed')}`
      message.error(e?.message || t('adminServer.terminalFailed'))
      scrollOutputToBottom()
    }
  })
}

function onInputKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    if (!loading.value && !unavailable.value)
      void runCmd()
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

    <div class="term" :class="{ 'term--disabled': unavailable }">
      <div class="term-bar">
        <span class="term-dot term-dot--r" />
        <span class="term-dot term-dot--y" />
        <span class="term-dot term-dot--g" />
        <span class="term-bar-title">{{ t('adminServer.terminalTab') }}</span>
      </div>

      <div ref="outputEl" class="term-screen">
        <pre class="term-out">{{ output || t('adminServer.terminalEmptyOutput') }}</pre>
      </div>

      <div class="term-prompt-row">
        <span class="term-prompt">$</span>
        <input
          v-model="cmd"
          class="term-input"
          type="text"
          spellcheck="false"
          autocomplete="off"
          :placeholder="t('adminServer.terminalPlaceholder')"
          :disabled="unavailable || loading"
          @keydown="onInputKeydown"
        >
      </div>
    </div>

    <n-space class="mt-12px">
      <n-button type="primary" :loading="loading" :disabled="unavailable" @click="runCmd">
        {{ t('adminServer.terminalRun') }}
      </n-button>
      <n-button quaternary :disabled="!output" @click="output = ''">
        {{ t('common.reset') }}
      </n-button>
    </n-space>
  </n-card>
</template>

<style scoped>
.term {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #2a2a2a;
  border-radius: 8px;
  background: #0d0d0d;
  color: #d0d0d0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.term--disabled {
  opacity: 0.72;
}

.term-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border-bottom: 1px solid #222;
  background: #161616;
  user-select: none;
}

.term-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.term-dot--r {
  background: #ff5f56;
}

.term-dot--y {
  background: #ffbd2e;
}

.term-dot--g {
  background: #27c93f;
}

.term-bar-title {
  margin-left: 8px;
  color: #8a8a8a;
  font-size: 12px;
}

.term-screen {
  min-height: 220px;
  max-height: 420px;
  overflow: auto;
  padding: 12px 14px;
}

.term-out {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  color: #c8c8c8;
}

.term-prompt-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid #222;
  background: #0a0a0a;
}

.term-prompt {
  flex-shrink: 0;
  color: #7dba6f;
  font-weight: 600;
}

.term-input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: #e8e8e8;
  font: inherit;
  caret-color: #7dba6f;
}

.term-input::placeholder {
  color: #555;
}

.term-input:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

/* 滚动条：暗色、不抢眼 */
.term-screen::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.term-screen::-webkit-scrollbar-thumb {
  background: #333;
  border-radius: 4px;
}

.term-screen::-webkit-scrollbar-track {
  background: transparent;
}
</style>
