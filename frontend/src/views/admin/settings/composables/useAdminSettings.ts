/**
 * 管理端系统设置：聚合入口（状态 + 选项 + 动作）
 * Tab 仍从此处 import，对外入口不变。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import * as state from './settingsState'
import { useSettingsActions } from './useSettingsActions'

export function useAdminSettings() {
  const { t } = useI18n()
  const actions = useSettingsActions()

  // 首次进入时补默认文案（模块级表单只初始化一次）
  if (!state.securityForm.realname_notify_text)
    state.securityForm.realname_notify_text = t('adminSettings.realnameNotifyTextDefault')
  if (!state.paymentForm.withdraw_notify_text)
    state.paymentForm.withdraw_notify_text = t('adminSettings.withdrawNotifyTextDefault')

  const langOptions = [
    { label: t('adminSettings.langZhCN'), value: 'zhCN' },
    { label: t('adminSettings.langEnUS'), value: 'enUS' },
  ]
  const smsProviderOptions = [
    { label: t('adminSettings.smsProviderConsole'), value: 'console' },
    { label: t('adminSettings.smsProviderAliyun'), value: 'aliyun' },
    { label: t('adminSettings.smsProviderTencent'), value: 'tencent' },
    { label: t('adminSettings.smsProviderCustom'), value: 'custom' },
  ]
  const smsBodyFormatOptions = [
    { label: t('adminSettings.formatJSON'), value: 'json' },
    { label: t('adminSettings.formatForm'), value: 'form' },
  ]
  const realnameApiProviderOptions = [
    { label: t('adminSettings.providerAliyun'), value: 'aliyun' },
    { label: t('adminSettings.providerTencent'), value: 'tencent' },
    { label: t('adminSettings.providerBaidu'), value: 'baidu' },
    { label: t('adminSettings.providerCustom'), value: 'custom' },
  ]
  const typeOptions = [
    { label: t('adminSettings.typeString'), value: 'string' },
    { label: t('adminSettings.typeNumber'), value: 'number' },
    { label: t('adminSettings.typeBoolean'), value: 'boolean' },
    { label: t('adminSettings.typeJSON'), value: 'json' },
    { label: t('adminSettings.typePassword'), value: 'password' },
  ]

  const smsProviderNeedsSignName = computed(() => state.smsForm.sms_provider !== 'console')
  const smsProviderNeedsTemplateCode = computed(() => ['aliyun', 'tencent'].includes(state.smsForm.sms_provider))
  const smsAccessKeyPlaceholder = computed(() => {
    if (state.smsForm.sms_provider === 'tencent') return t('adminSettings.smsTencentSecretId')
    if (state.smsForm.sms_provider === 'custom') return t('adminSettings.smsCustomApiKeyOptional')
    return t('adminSettings.smsProviderAccessKey')
  })
  const smsSecretKeyPlaceholder = computed(() => {
    if (state.smsForm.sms_provider === 'tencent') return t('adminSettings.smsTencentSecretKey')
    if (state.smsForm.sms_provider === 'custom') return t('adminSettings.smsCustomApiSecretOptional')
    return t('adminSettings.smsProviderSecretKey')
  })
  const smsTemplateLabel = computed(() => state.smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCode') : t('adminSettings.smsTemplateId'))
  const smsTemplatePlaceholder = computed(() => state.smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeExample') : t('adminSettings.smsTemplateIdPlaceholder'))
  const smsTemplateEnLabel = computed(() => state.smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeEnglish') : t('adminSettings.smsTemplateIdEnglish'))
  const smsTemplateEnPlaceholder = computed(() => state.smsForm.sms_provider === 'aliyun' ? t('adminSettings.smsTemplateCodeEnglishOptional') : t('adminSettings.smsTemplateIdEnglishOptional'))

  const addFormRules = {
    key: [
      { required: true, message: () => t('adminSettings.keyRequired'), trigger: 'blur' },
      { pattern: /^[a-z][a-z0-9_]*$/, message: () => t('adminSettings.keyPattern'), trigger: 'blur' },
    ],
    label: [{ required: true, message: () => t('adminSettings.labelRequired'), trigger: 'blur' }],
  }

  return {
    loading: state.loading,
    topTab: state.topTab,
    systemSubTab: state.systemSubTab,
    switchLoading: state.switchLoading,
    basicForm: state.basicForm,
    emailForm: state.emailForm,
    testEmailTo: state.testEmailTo,
    smsForm: state.smsForm,
    securityForm: state.securityForm,
    realnameApiForm: state.realnameApiForm,
    paymentForm: state.paymentForm,
    customSettings: state.customSettings,
    addFormRef: state.addFormRef,
    addForm: state.addForm,
    editForm: state.editForm,
    addFormRules,
    langOptions,
    smsProviderOptions,
    smsBodyFormatOptions,
    realnameApiProviderOptions,
    typeOptions,
    smsProviderNeedsSignName,
    smsProviderNeedsTemplateCode,
    smsAccessKeyPlaceholder,
    smsSecretKeyPlaceholder,
    smsTemplateLabel,
    smsTemplatePlaceholder,
    smsTemplateEnLabel,
    smsTemplateEnPlaceholder,
    savingBasic: state.savingBasic,
    savingEmail: state.savingEmail,
    savingSms: state.savingSms,
    savingSecurity: state.savingSecurity,
    savingRealnameApi: state.savingRealnameApi,
    savingPayment: state.savingPayment,
    testingEmail: state.testingEmail,
    restartingBackend: state.restartingBackend,
    adding: state.adding,
    savingEdit: state.savingEdit,
    showAddModal: state.showAddModal,
    showEditModal: state.showEditModal,
    ...actions,
  }
}
