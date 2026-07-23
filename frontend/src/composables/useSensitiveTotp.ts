/**
 * 敏感操作二次 TOTP：已启用 2FA 的管理员在补单/调账/导出等写操作前输入动态码，
 * 通过请求头 X-Totp-Code 交给后端 RequireRecentTOTP 校验。
 * 未启用时直接放行（与后端中间件一致）。
 */
import { h, ref } from 'vue'
import { NInput } from 'naive-ui'
import { $t } from '@/utils/i18n'
import { adminProfileApi } from '@/service/api/admin/profile'

let totpEnabledCache: boolean | null = null

/** 刷新本地缓存（启用/禁用 TOTP 后调用） */
export function invalidateSensitiveTotpCache() {
  totpEnabledCache = null
}

async function isTotpEnabled(): Promise<boolean> {
  if (totpEnabledCache !== null)
    return totpEnabledCache
  try {
    const res = await adminProfileApi.me() as Service.ResponseResult<{ totp_enabled?: boolean }>
    totpEnabledCache = !!(res.isSuccess && res.data?.totp_enabled)
  }
  catch {
    totpEnabledCache = false
  }
  return totpEnabledCache
}

/**
 * 若当前管理员启用了 TOTP，弹出输入框拿动态码；否则返回空串。
 * 用户取消时返回 null（调用方应中止操作）。
 */
export async function promptSensitiveTotpCode(): Promise<string | null> {
  const enabled = await isTotpEnabled()
  if (!enabled)
    return ''

  const codeRef = ref('')

  return new Promise((resolve) => {
    const d = window.$dialog?.create({
      title: $t('common.sensitiveTotpTitle'),
      content: () => h('div', { style: 'display:flex;flex-direction:column;gap:12px;' }, [
        h('div', $t('common.sensitiveTotpHint')),
        h(NInput, {
          value: codeRef.value,
          maxlength: 8,
          placeholder: $t('common.sensitiveTotpPlaceholder'),
          autofocus: true,
          'onUpdate:value': (v: string) => { codeRef.value = v },
        }),
      ]),
      positiveText: $t('common.confirm'),
      negativeText: $t('common.cancel'),
      onPositiveClick: () => {
        const code = codeRef.value.trim()
        if (!code) {
          window.$message?.warning($t('common.sensitiveTotpRequired'))
          return false
        }
        resolve(code)
        return true
      },
      onNegativeClick: () => {
        resolve(null)
      },
      onClose: () => {
        resolve(null)
      },
    })
    if (!d) {
      window.$message?.error($t('common.sensitiveTotpRequired'))
      resolve(null)
    }
  })
}

/** 构造带 X-Totp-Code 的请求头（code 为空则不加） */
export function totpHeaders(code?: string | null): Record<string, string> {
  if (!code)
    return {}
  return { 'X-Totp-Code': code }
}
