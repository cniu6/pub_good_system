/**
 * 用户管理页公共展示工具（index / detail 共用）
 * 纯展示逻辑，无副作用
 */
import { i18n } from '@/modules/i18n'
import type { AdminUser, UserSimpleInfo } from '@/service/api/admin/user'

const t = (key: string, params?: Record<string, unknown>) => i18n.global.t(key, params as any) as string

/** 金额保留两位小数 */
export function formatCurrency(value?: number | null) {
  return Number(value || 0).toFixed(2)
}

/** 语言代码 → 中文/英文展示名 */
export function formatLanguage(language?: string) {
  if (!language)
    return '-'
  if (language === 'zh-CN')
    return t('adminUsersDetail.chinese')
  if (language === 'en-US')
    return t('adminUsersDetail.english')
  return language
}

/** 充值留存比例（无累计充值时显示 -） */
export function formatRechargeRetentionRatio(row?: Pick<AdminUser, 'total_paid_amount' | 'balance_paid_ratio'> | null) {
  const totalPaid = Number(row?.total_paid_amount || 0)
  if (totalPaid <= 0)
    return '-'
  return `${(Number(row?.balance_paid_ratio || 0) * 100).toFixed(2)}%`
}

/** Unix 秒时间戳 → 本地时间字符串 */
export function formatTime(ts?: number | null) {
  return ts ? new Date(ts * 1000).toLocaleString() : '-'
}

/**
 * 实名状态文案（复用 realname.* i18n，与 constants/realname 选项一致）
 * status 未传视为未认证
 */
export function getRealnameStatusText(status?: number) {
  const map: Record<number, string> = {
    0: t('realname.pending'),
    1: t('realname.approved'),
    2: t('realname.rejected'),
  }
  return status !== undefined ? (map[status] || t('adminUsers.unknown')) : t('adminUsers.unverified')
}

/** 实名状态对应 NTag type */
export function getRealnameStatusType(status?: number): 'default' | 'warning' | 'success' | 'error' {
  const map: Record<number, 'warning' | 'success' | 'error'> = {
    0: 'warning',
    1: 'success',
    2: 'error',
  }
  return status !== undefined ? (map[status] || 'default') : 'default'
}

/** 证件号脱敏：保留前后 4 位 */
export function maskCertificateNo(no?: string) {
  if (!no)
    return '-'
  if (no.length < 8)
    return no
  return `${no.slice(0, 4)}****${no.slice(-4)}`
}

/** 账户号脱敏：保留前后 4 位 */
export function maskAccountNo(accountNo?: string) {
  if (!accountNo)
    return '-'
  if (accountNo.length <= 8)
    return accountNo
  return `${accountNo.slice(0, 4)}****${accountNo.slice(-4)}`
}

/** 提现状态 → NTag type + 文案 */
export function getWithdrawStatusMeta(status?: number): { type: 'warning' | 'info' | 'error' | 'success', label: string } {
  const map: Record<number, { type: 'warning' | 'info' | 'error' | 'success', label: string }> = {
    0: { type: 'warning', label: t('moneyScore.statusPending') },
    1: { type: 'info', label: t('moneyScore.statusApproved') },
    2: { type: 'error', label: t('moneyScore.statusRejected') },
    3: { type: 'success', label: t('moneyScore.statusPaid') },
  }
  return status !== undefined
    ? (map[status] || { type: 'info', label: t('adminUsers.unknown') })
    : { type: 'info', label: t('adminUsers.unknown') }
}

/**
 * 管理员 ID → 展示名
 * @param adminId 管理员用户 ID
 * @param adminUserMap batchSimpleInfo 得到的映射表
 */
export function getAdminDisplayName(
  adminId: number | null | undefined,
  adminUserMap: Record<number, UserSimpleInfo>,
) {
  if (!adminId)
    return '-'
  const admin = adminUserMap[adminId]
  return admin?.nickname || admin?.username || t('adminUsers.adminFallback', { id: adminId })
}
