/**
 * 导入导出中心 MVP：下载用户 CSV / 导入模板
 * 导出用户含 PII：若管理员启用了 TOTP，须二次输入动态码。
 */
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { adminApiUrl } from '@/service/api/admin/base'
import { authStorage } from '@/utils'
import { promptSensitiveTotpCode } from '@/composables/useSensitiveTotp'

const { t } = useI18n()
const message = useMessage()

function authHeaders(extra?: Record<string, string>): HeadersInit {
  const token = authStorage.get('accessToken') || ''
  return {
    Authorization: token ? `Bearer ${token}` : '',
    ...(extra || {}),
  }
}

async function downloadBlob(url: string, method: 'GET' | 'POST', filename: string, headers?: Record<string, string>) {
  try {
    const res = await fetch(url, {
      method,
      headers: authHeaders(headers),
    })
    if (!res.ok) {
      // 尽量解析业务错误
      try {
        const j = await res.json()
        message.error(j?.message || t('adminImportExport.downloadFailed'))
      }
      catch {
        message.error(t('adminImportExport.downloadFailed'))
      }
      return
    }
    const blob = await res.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = filename
    a.click()
    URL.revokeObjectURL(a.href)
    message.success(t('adminImportExport.downloadOk'))
  }
  catch (e) {
    message.error(t('adminImportExport.downloadFailed'))
    if (import.meta.env.DEV)
      console.error(e)
  }
}

async function exportUsers() {
  const totpCode = await promptSensitiveTotpCode()
  if (totpCode === null)
    return
  const headers: Record<string, string> = {}
  if (totpCode)
    headers['X-Totp-Code'] = totpCode
  await downloadBlob(adminApiUrl('/export/users'), 'POST', `users_${Date.now()}.csv`, headers)
}

function downloadTemplate() {
  downloadBlob(adminApiUrl('/export/users/template'), 'GET', 'users_import_template.csv')
}
</script>

<template>
  <n-card :title="t('adminImportExport.title')" :bordered="false">
    <n-alert type="info" class="mb-16px">
      {{ t('adminImportExport.hint') }}
    </n-alert>
    <n-space>
      <n-button type="primary" @click="exportUsers">
        {{ t('adminImportExport.exportUsers') }}
      </n-button>
      <n-button @click="downloadTemplate">
        {{ t('adminImportExport.downloadTemplate') }}
      </n-button>
    </n-space>
  </n-card>
</template>
