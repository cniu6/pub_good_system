/**
 * 实名认证相关选项常量（支持 i18n）
 */
import { i18n } from '@/modules/i18n'

/** 获取证件类型选项 */
export function getCertificateTypeOptions() {
  return [
    { label: i18n.global.t('realname.idCard'), value: 1 },
    { label: i18n.global.t('realname.passport'), value: 2 },
    { label: i18n.global.t('realname.officer'), value: 3 },
  ]
}

/** 获取实名认证状态选项 */
export function getRealnameStatusOptions() {
  return [
    { label: i18n.global.t('realname.pending'), value: 0 },
    { label: i18n.global.t('realname.approved'), value: 1 },
    { label: i18n.global.t('realname.rejected'), value: 2 },
  ]
}

/** 证件类型选项（向后兼容） */
export const certificateTypeOptions = getCertificateTypeOptions()

/** 实名认证状态选项（向后兼容） */
export const realnameStatusOptions = getRealnameStatusOptions()
