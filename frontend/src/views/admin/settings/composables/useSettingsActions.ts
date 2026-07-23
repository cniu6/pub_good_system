/**
 * 管理端设置：load / save / toggle 动作
 */
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import type { SettingDTO, SettingType } from '@/service/api/admin/settings'
import { useSettingsStore } from '@/store/settings'
import { parseBooleanSetting } from '@/utils'
import {
  addForm,
  addFormRef,
  adding,
  basicForm,
  customSettings,
  editForm,
  emailForm,
  loading,
  paymentForm,
  realnameApiForm,
  restartingBackend,
  savingBasic,
  savingEdit,
  savingEmail,
  savingPayment,
  savingRealnameApi,
  savingSecurity,
  savingSms,
  securityForm,
  showAddModal,
  showEditModal,
  smsForm,
  switchLoading,
  testEmailTo,
  testingEmail,
  testingSms,
  testSmsPhone,
} from './settingsState'

export function useSettingsActions() {
  const message = useMessage()
  const settingsStore = useSettingsStore()
  const { t } = useI18n()

  async function loadSettings() {
    loading.value = true
    try {
      const response = await adminApi.settings.list()
      // 必须先判断业务成功，避免失败响应误写入表单（如密钥被空串覆盖）
      if (!response.isSuccess) {
        message.error(response.message || t('adminSettings.loadSettingsFailed'))
        return
      }
      if (response.data?.categories) {
        for (const category of response.data.categories) {
          for (const item of category.items) {
            if (item.key === 'site_name')
              basicForm.site_name = String(item.value || '')
            if (item.key === 'site_desc')
              basicForm.site_desc = String(item.value || '')
            if (item.key === 'site_logo')
              basicForm.site_logo = String(item.value || '')
            if (item.key === 'copyright')
              basicForm.copyright = String(item.value || '')
            if (item.key === 'icp')
              basicForm.icp = String(item.value || '')
            if (item.key === 'version')
              basicForm.version = String(item.value || '')
            if (item.key === 'default_lang')
              basicForm.default_lang = String(item.value || 'zhCN')
            if (item.key === 'allow_register')
              basicForm.allow_register = parseBooleanSetting(item.value)
            if (item.key === 'announcement_enabled')
              basicForm.announcement_enabled = parseBooleanSetting(item.value)
            if (item.key === 'allow_delete_account')
              securityForm.allow_delete_account = parseBooleanSetting(item.value)
            if (item.key === 'frontend_url')
              basicForm.frontend_url = String(item.value || '')
            if (item.key === 'backend_api_url')
              basicForm.backend_api_url = String(item.value || '')

            if (item.key === 'email_verify_enabled')
              emailForm.email_verify_enabled = parseBooleanSetting(item.value)
            if (item.key === 'smtp_host')
              emailForm.smtp_host = String(item.value || '')
            if (item.key === 'smtp_port')
              emailForm.smtp_port = Number(item.value) || 587
            if (item.key === 'smtp_username')
              emailForm.smtp_username = String(item.value || '')
            if (item.key === 'smtp_password')
              emailForm.smtp_password = String(item.value || '')
            if (item.key === 'smtp_ssl')
              emailForm.smtp_ssl = parseBooleanSetting(item.value)
            if (item.key === 'smtp_proxy_enabled')
              emailForm.smtp_proxy_enabled = parseBooleanSetting(item.value)
            if (item.key === 'smtp_proxy_type')
              emailForm.smtp_proxy_type = String(item.value || 'socks5')
            if (item.key === 'smtp_proxy_host')
              emailForm.smtp_proxy_host = String(item.value || '')
            if (item.key === 'smtp_proxy_port')
              emailForm.smtp_proxy_port = Number(item.value) || 1080
            if (item.key === 'smtp_proxy_username')
              emailForm.smtp_proxy_username = String(item.value || '')
            if (item.key === 'smtp_proxy_password')
              emailForm.smtp_proxy_password = String(item.value || '')
            if (item.key === 'system_email_address')
              emailForm.system_email_address = String(item.value || '')
            if (item.key === 'system_email_name')
              emailForm.system_email_name = String(item.value || '')

            if (item.key === 'sms_verify_enabled')
              smsForm.sms_verify_enabled = parseBooleanSetting(item.value)
            if (item.key === 'mobile_cn_only')
              smsForm.mobile_cn_only = parseBooleanSetting(item.value)
            if (item.key === 'mobile_ip_country_detect')
              smsForm.mobile_ip_country_detect = parseBooleanSetting(item.value)
            if (item.key === 'sms_provider')
              smsForm.sms_provider = String(item.value || 'console')
            if (item.key === 'sms_access_key')
              smsForm.sms_access_key = String(item.value || '')
            if (item.key === 'sms_secret_key')
              smsForm.sms_secret_key = String(item.value || '')
            if (item.key === 'sms_sign_name')
              smsForm.sms_sign_name = String(item.value || '')
            if (item.key === 'sms_template_code')
              smsForm.sms_template_code = String(item.value || '')
            if (item.key === 'sms_template_code_en')
              smsForm.sms_template_code_en = String(item.value || '')
            if (item.key === 'sms_region')
              smsForm.sms_region = String(item.value || '')
            if (item.key === 'sms_sdk_app_id')
              smsForm.sms_sdk_app_id = String(item.value || '')
            if (item.key === 'sms_endpoint')
              smsForm.sms_endpoint = String(item.value || '')
            if (item.key === 'sms_body_format')
              smsForm.sms_body_format = String(item.value || 'json')

            if (item.key === 'geetest_enabled')
              securityForm.geetest_enabled = parseBooleanSetting(item.value)
            if (item.key === 'geetest_captcha_id')
              securityForm.geetest_captcha_id = String(item.value || '')
            if (item.key === 'geetest_captcha_key')
              securityForm.geetest_captcha_key = String(item.value || '')
            if (item.key === 'jwt_access_expire')
              securityForm.jwt_access_expire = Number(item.value) || 7200
            if (item.key === 'jwt_refresh_expire')
              securityForm.jwt_refresh_expire = Number(item.value) || 604800
            if (item.key === 'login_max_failure')
              securityForm.login_max_failure = Number(item.value) || 5
            if (item.key === 'login_lock_duration')
              securityForm.login_lock_duration = Number(item.value) || 10
            if (item.key === 'disable_web_login')
              securityForm.disable_web_login = parseBooleanSetting(item.value)
            if (item.key === 'realname_enabled')
              securityForm.realname_enabled = parseBooleanSetting(item.value)
            if (item.key === 'realname_review_required')
              securityForm.realname_review_required = parseBooleanSetting(item.value)
            if (item.key === 'realname_notify_text')
              securityForm.realname_notify_text = String(item.value || t('adminSettings.realnameNotifyTextDefault'))

            if (item.key === 'realname_api_enabled')
              realnameApiForm.realname_api_enabled = parseBooleanSetting(item.value)
            if (item.key === 'realname_api_provider')
              realnameApiForm.realname_api_provider = String(item.value || 'aliyun')
            if (item.key === 'realname_api_app_key')
              realnameApiForm.realname_api_app_key = String(item.value || '')
            if (item.key === 'realname_api_app_secret')
              realnameApiForm.realname_api_app_secret = String(item.value || '')
            if (item.key === 'realname_api_endpoint')
              realnameApiForm.realname_api_endpoint = String(item.value || '')

            if (item.key === 'payment_enabled')
              paymentForm.payment_enabled = parseBooleanSetting(item.value)
            if (item.key === 'payment_order_expire_minutes')
              paymentForm.payment_order_expire_minutes = Number(item.value) || 30
            if (item.key === 'withdraw_enabled')
              paymentForm.withdraw_enabled = parseBooleanSetting(item.value)
            if (item.key === 'withdraw_min_amount')
              paymentForm.withdraw_min_amount = Number(item.value) || 10
            if (item.key === 'withdraw_notify_text')
              paymentForm.withdraw_notify_text = String(item.value || t('adminSettings.withdrawNotifyTextDefault'))
            if (item.key === 'withdraw_account_types')
              paymentForm.withdraw_account_types_text = typeof item.value === 'string' ? item.value : JSON.stringify(item.value || ['bank', 'alipay', 'wechat', 'usdt'])
            if (item.key === 'withdraw_require_realname')
              paymentForm.withdraw_require_realname = parseBooleanSetting(item.value)
            if (item.key === 'finance_dual_approval')
              paymentForm.finance_dual_approval = parseBooleanSetting(item.value)
          }
          if (category.category === 'custom')
            customSettings.value = category.items
        }
      }
    }
    catch {
      message.error(t('adminSettings.loadSettingsFailed'))
    }
    finally {
      loading.value = false
    }
  }

  /** 开关类：乐观更新 + 失败回滚 */
  async function toggleSetting(
    key: keyof typeof switchLoading,
    getPrev: () => boolean,
    setVal: (v: boolean) => void,
    nextValue: boolean,
    onSuccess?: () => void,
    successKey?: string,
    failKey?: string,
  ) {
    const prev = getPrev()
    setVal(nextValue)
    switchLoading[key] = true
    try {
      const res = await adminApi.settings.update(key, String(nextValue))
      if (res.isSuccess) {
        onSuccess?.()
        message.success(res.message || (successKey ? t(successKey) : 'OK'))
      }
      else {
        setVal(prev)
        message.error(res.message || (failKey ? t(failKey) : t('adminSettings.updateFailed')))
      }
    }
    catch {
      setVal(prev)
      message.error(t('adminSettings.updateFailed'))
    }
    finally {
      switchLoading[key] = false
    }
  }

  async function handleUpdateAllowRegister(nextValue: boolean) {
    await toggleSetting(
      'allow_register',
      () => basicForm.allow_register,
      (v) => { basicForm.allow_register = v },
      nextValue,
      () => settingsStore.updateConfig({ allow_register: nextValue }),
      'adminSettings.registerSwitchUpdated',
      'adminSettings.registerSwitchUpdateFailed',
    )
  }

  async function handleUpdateAnnouncementEnabled(nextValue: boolean) {
    await toggleSetting(
      'announcement_enabled',
      () => basicForm.announcement_enabled,
      (v) => { basicForm.announcement_enabled = v },
      nextValue,
      () => settingsStore.updateConfig({ announcement_enabled: nextValue }),
      'adminSettings.announcementSwitchUpdated',
      'adminSettings.announcementSwitchUpdateFailed',
    )
  }

  async function handleUpdateAllowDeleteAccount(nextValue: boolean) {
    await toggleSetting(
      'allow_delete_account',
      () => securityForm.allow_delete_account,
      (v) => { securityForm.allow_delete_account = v },
      nextValue,
      () => settingsStore.updateConfig({ allow_delete_account: nextValue }),
      'adminSettings.deleteAccountSwitchUpdated',
      'adminSettings.deleteAccountSwitchUpdateFailed',
    )
  }

  async function handleUpdateDisableWebLogin(nextValue: boolean) {
    await toggleSetting(
      'disable_web_login',
      () => securityForm.disable_web_login,
      (v) => { securityForm.disable_web_login = v },
      nextValue,
      () => settingsStore.updateConfig({ web_login_disabled: nextValue }),
      'adminSettings.disableWebLoginSwitchUpdated',
      'adminSettings.disableWebLoginSwitchUpdateFailed',
    )
  }

  async function handleUpdateSmtpSSL(nextValue: boolean) {
    await toggleSetting(
      'smtp_ssl',
      () => emailForm.smtp_ssl,
      (v) => { emailForm.smtp_ssl = v },
      nextValue,
      undefined,
      'adminSettings.smtpSslUpdated',
      'adminSettings.smtpSslUpdateFailed',
    )
  }

  async function handleUpdateSmtpProxyEnabled(nextValue: boolean) {
    await toggleSetting(
      'smtp_proxy_enabled',
      () => emailForm.smtp_proxy_enabled,
      (v) => { emailForm.smtp_proxy_enabled = v },
      nextValue,
      undefined,
      'adminSettings.smtpProxyUpdated',
      'adminSettings.smtpProxyUpdateFailed',
    )
  }

  async function handleUpdateEmailVerifyEnabled(nextValue: boolean) {
    await toggleSetting(
      'email_verify_enabled',
      () => emailForm.email_verify_enabled,
      (v) => { emailForm.email_verify_enabled = v },
      nextValue,
      () => settingsStore.updateConfig({ email_verify_enabled: nextValue }),
      'adminSettings.emailVerifySwitchUpdated',
      'adminSettings.emailVerifySwitchUpdateFailed',
    )
  }

  async function handleUpdateSmsVerifyEnabled(nextValue: boolean) {
    await toggleSetting(
      'sms_verify_enabled',
      () => smsForm.sms_verify_enabled,
      (v) => { smsForm.sms_verify_enabled = v },
      nextValue,
      () => settingsStore.updateConfig({ sms_verify_enabled: nextValue }),
      'adminSettings.smsVerifySwitchUpdated',
      'adminSettings.smsVerifySwitchUpdateFailed',
    )
  }

  async function handleUpdateMobileCnOnly(nextValue: boolean) {
    await toggleSetting(
      'mobile_cn_only',
      () => smsForm.mobile_cn_only,
      (v) => { smsForm.mobile_cn_only = v },
      nextValue,
      () => settingsStore.updateConfig({ mobile_cn_only: nextValue }),
      'adminSettings.mobileCnOnlySwitchUpdated',
      'adminSettings.mobileCnOnlySwitchUpdateFailed',
    )
    // 切回仅大陆时，顺手关掉 IP 探测（避免无效配置）
    if (nextValue && smsForm.mobile_ip_country_detect) {
      await handleUpdateMobileIpCountryDetect(false)
    }
  }

  async function handleUpdateMobileIpCountryDetect(nextValue: boolean) {
    // 仅大陆模式下不允许开 IP 探测（后端也会忽略）
    if (smsForm.mobile_cn_only && nextValue) {
      message.warning(t('adminSettings.mobileIpDetectNeedIntl'))
      return
    }
    await toggleSetting(
      'mobile_ip_country_detect',
      () => smsForm.mobile_ip_country_detect,
      (v) => { smsForm.mobile_ip_country_detect = v },
      nextValue,
      () => settingsStore.updateConfig({ mobile_ip_country_detect: nextValue }),
      'adminSettings.mobileIpDetectSwitchUpdated',
      'adminSettings.mobileIpDetectSwitchUpdateFailed',
    )
  }

  async function handleUpdateGeetestEnabled(nextValue: boolean) {
    await toggleSetting(
      'geetest_enabled',
      () => securityForm.geetest_enabled,
      (v) => { securityForm.geetest_enabled = v },
      nextValue,
      () => settingsStore.updateConfig({ geetest_enabled: nextValue }),
      'adminSettings.geetestSwitchUpdated',
      'adminSettings.geetestSwitchUpdateFailed',
    )
  }

  async function handleUpdateRealnameEnabled(nextValue: boolean) {
    await toggleSetting(
      'realname_enabled',
      () => securityForm.realname_enabled,
      (v) => { securityForm.realname_enabled = v },
      nextValue,
      undefined,
      'adminSettings.realnameSwitchUpdated',
      'adminSettings.realnameSwitchUpdateFailed',
    )
  }

  async function handleUpdateRealnameReviewRequired(nextValue: boolean) {
    await toggleSetting(
      'realname_review_required',
      () => securityForm.realname_review_required,
      (v) => { securityForm.realname_review_required = v },
      nextValue,
      undefined,
      'adminSettings.realnameReviewSwitchUpdated',
      'adminSettings.realnameReviewSwitchUpdateFailed',
    )
  }

  async function handleUpdateRealnameApiEnabled(nextValue: boolean) {
    await toggleSetting(
      'realname_api_enabled',
      () => realnameApiForm.realname_api_enabled,
      (v) => { realnameApiForm.realname_api_enabled = v },
      nextValue,
      undefined,
      'adminSettings.realnameApiSwitchUpdated',
      'adminSettings.realnameApiSwitchUpdateFailed',
    )
  }

  async function handleUpdatePaymentEnabled(nextValue: boolean) {
    await toggleSetting(
      'payment_enabled',
      () => paymentForm.payment_enabled,
      (v) => { paymentForm.payment_enabled = v },
      nextValue,
      undefined,
      'adminSettings.paymentSwitchUpdated',
      'adminSettings.paymentSwitchUpdateFailed',
    )
  }

  async function handleUpdateWithdrawEnabled(nextValue: boolean) {
    await toggleSetting(
      'withdraw_enabled',
      () => paymentForm.withdraw_enabled,
      (v) => { paymentForm.withdraw_enabled = v },
      nextValue,
      () => settingsStore.updateConfig({ withdraw_enabled: nextValue }),
      'adminSettings.withdrawSwitchUpdated',
      'adminSettings.withdrawSwitchUpdateFailed',
    )
  }

  async function handleUpdateWithdrawRequireRealname(nextValue: boolean) {
    await toggleSetting(
      'withdraw_require_realname',
      () => paymentForm.withdraw_require_realname,
      (v) => { paymentForm.withdraw_require_realname = v },
      nextValue,
      () => settingsStore.updateConfig({ withdraw_require_realname: nextValue }),
      'adminSettings.withdrawRequireRealnameSwitchUpdated',
      'adminSettings.withdrawRequireRealnameSwitchUpdateFailed',
    )
  }

  async function handleUpdateFinanceDualApproval(nextValue: boolean) {
    await toggleSetting(
      'finance_dual_approval',
      () => paymentForm.finance_dual_approval,
      (v) => { paymentForm.finance_dual_approval = v },
      nextValue,
      undefined,
      'adminSettings.financeDualApprovalSwitchUpdated',
      'adminSettings.financeDualApprovalSwitchUpdateFailed',
    )
  }

  async function handleSaveSms(options?: { silent?: boolean }): Promise<boolean> {
    savingSms.value = true
    try {
      const res = await adminApi.settings.batchUpdate({
        sms_provider: smsForm.sms_provider,
        sms_access_key: smsForm.sms_access_key,
        sms_secret_key: smsForm.sms_secret_key,
        sms_sign_name: smsForm.sms_sign_name,
        sms_template_code: smsForm.sms_template_code,
        sms_template_code_en: smsForm.sms_template_code_en,
        sms_region: smsForm.sms_region,
        sms_sdk_app_id: smsForm.sms_sdk_app_id,
        sms_endpoint: smsForm.sms_endpoint,
        sms_body_format: smsForm.sms_body_format,
      })
      if (res.isSuccess) {
        if (!options?.silent)
          message.success(res.message || t('adminSettings.smsSettingsSaved'))
        return true
      }
      message.error(res.message || t('adminSettings.smsSettingsSaveFailed'))
      return false
    }
    catch {
      message.error(t('adminSettings.saveFailed'))
      return false
    }
    finally {
      savingSms.value = false
    }
  }

  async function handleTestSms() {
    const phone = String(testSmsPhone.value || '').trim()
    if (!phone) {
      message.warning(t('adminSettings.testSmsPhoneRequired'))
      return
    }

    testingSms.value = true
    try {
      // 先静默保存当前表单，避免「改了配置却没保存」测的还是旧通道
      const saved = await handleSaveSms({ silent: true })
      if (!saved) {
        message.error(t('adminSettings.testSmsNeedSaveFirst'))
        return
      }

      const res = await adminApi.smsTemplate.sendTest({ phone })
      if (res.isSuccess)
        message.success(res.data?.message || t('adminSettings.testSmsSent'))
      else
        message.error(res.message || t('adminSettings.testSmsFailed'))
    }
    catch (error: any) {
      message.error(error?.message || t('adminSettings.testSmsFailed'))
    }
    finally {
      testingSms.value = false
    }
  }

  async function handleSaveRealnameApi() {
    savingRealnameApi.value = true
    try {
      const res = await adminApi.settings.batchUpdate({
        realname_api_provider: realnameApiForm.realname_api_provider,
        realname_api_app_key: realnameApiForm.realname_api_app_key,
        realname_api_app_secret: realnameApiForm.realname_api_app_secret,
        realname_api_endpoint: realnameApiForm.realname_api_endpoint,
      })
      if (res.isSuccess)
        message.success(res.message || t('adminSettings.realnameApiSaveSuccess'))
      else message.error(res.message || t('adminSettings.realnameApiSaveFailed'))
    }
    catch {
      message.error(t('adminSettings.saveFailed'))
    }
    finally {
      savingRealnameApi.value = false
    }
  }

  async function handleSavePayment() {
    savingPayment.value = true
    try {
      const parsedAccountTypes = JSON.parse(paymentForm.withdraw_account_types_text || '[]')
      if (!Array.isArray(parsedAccountTypes) || parsedAccountTypes.length === 0 || parsedAccountTypes.some(item => typeof item !== 'string' || !item.trim()))
        throw new Error(t('adminSettings.invalidAccountTypes'))
      const res = await adminApi.settings.batchUpdate({
        payment_order_expire_minutes: String(paymentForm.payment_order_expire_minutes),
        withdraw_min_amount: String(paymentForm.withdraw_min_amount),
        withdraw_notify_text: paymentForm.withdraw_notify_text,
        withdraw_account_types: paymentForm.withdraw_account_types_text,
      })
      if (res.isSuccess) {
        settingsStore.updateConfig({
          withdraw_min_amount: paymentForm.withdraw_min_amount,
          withdraw_notify_text: paymentForm.withdraw_notify_text,
          withdraw_account_types: parsedAccountTypes,
        })
        message.success(res.message || t('adminSettings.paymentSettingsSaved'))
      }
      else {
        message.error(res.message || t('adminSettings.paymentSettingsSaveFailed'))
      }
    }
    catch {
      message.error(t('adminSettings.saveFailed'))
    }
    finally {
      savingPayment.value = false
    }
  }

  async function handleSaveBasic() {
    savingBasic.value = true
    try {
      const frontendUrl = basicForm.frontend_url.trim().replace(/\/+$/, '')
      const backendApiUrl = basicForm.backend_api_url.trim().replace(/\/+$/, '')
      basicForm.frontend_url = frontendUrl
      basicForm.backend_api_url = backendApiUrl
      const res = await adminApi.settings.batchUpdate({
        site_name: basicForm.site_name,
        site_desc: basicForm.site_desc,
        site_logo: basicForm.site_logo,
        copyright: basicForm.copyright,
        icp: basicForm.icp,
        version: basicForm.version,
        default_lang: basicForm.default_lang,
        frontend_url: frontendUrl,
        backend_api_url: backendApiUrl,
      })
      if (res.isSuccess) {
        settingsStore.updateConfig({
          site_name: basicForm.site_name,
          site_desc: basicForm.site_desc,
          site_logo: basicForm.site_logo,
          copyright: basicForm.copyright,
          icp: basicForm.icp,
          version: basicForm.version,
          default_lang: basicForm.default_lang,
        })
        message.success(res.message || t('adminSettings.basicSettingsSaved'))
      }
      else {
        message.error(res.message || t('adminSettings.basicSettingsSaveFailed'))
      }
    }
    catch {
      message.error(t('adminSettings.saveFailed'))
    }
    finally {
      savingBasic.value = false
    }
  }

  async function handleSaveEmail(opts?: { silent?: boolean }) {
    savingEmail.value = true
    try {
      const res = await adminApi.settings.batchUpdate({
        smtp_host: emailForm.smtp_host,
        smtp_port: String(emailForm.smtp_port),
        smtp_username: emailForm.smtp_username,
        smtp_password: emailForm.smtp_password,
        smtp_ssl: String(emailForm.smtp_ssl),
        smtp_proxy_enabled: String(emailForm.smtp_proxy_enabled),
        smtp_proxy_type: emailForm.smtp_proxy_type,
        smtp_proxy_host: emailForm.smtp_proxy_host,
        smtp_proxy_port: String(emailForm.smtp_proxy_port),
        smtp_proxy_username: emailForm.smtp_proxy_username,
        smtp_proxy_password: emailForm.smtp_proxy_password,
        system_email_address: emailForm.system_email_address,
        system_email_name: emailForm.system_email_name,
      })
      if (res.isSuccess) {
        if (!opts?.silent)
          message.success(res.message || t('adminSettings.emailSettingsSaved'))
      }
      else {
        message.error(res.message || t('adminSettings.emailSettingsSaveFailed'))
      }
      return res.isSuccess
    }
    catch {
      message.error(t('adminSettings.saveFailed'))
      return false
    }
    finally {
      savingEmail.value = false
    }
  }

  async function handleTestEmail() {
    const to = String(testEmailTo.value || '').trim()
    if (!to) {
      message.warning(t('adminSettings.testEmailToRequired'))
      return
    }
    // 简单校验邮箱格式，避免打到后端才报错
    const emailOk = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\])|(([a-z\-0-9]+\.)+[a-z]{2,}))$/i.test(to)
    if (!emailOk) {
      message.warning(t('adminSettings.testEmailToInvalid'))
      return
    }

    testingEmail.value = true
    try {
      // 先静默保存当前表单，避免「改了配置却没保存」测的还是旧 SMTP
      const saved = await handleSaveEmail({ silent: true })
      if (!saved) {
        message.error(t('adminSettings.testEmailNeedSaveFirst'))
        return
      }

      const res = await adminApi.emailTemplate.sendTest({ to })
      if (res.isSuccess)
        message.success(res.data?.message || t('adminSettings.testEmailSent'))
      else message.error(res.message || t('adminSettings.testEmailFailed'))
    }
    catch (error: any) {
      message.error(error?.message || t('adminSettings.testEmailFailed'))
    }
    finally {
      testingEmail.value = false
    }
  }

  async function handleSaveSecurity() {
    savingSecurity.value = true
    try {
      const res = await adminApi.settings.batchUpdate({
        geetest_captcha_id: securityForm.geetest_captcha_id,
        geetest_captcha_key: securityForm.geetest_captcha_key,
        jwt_access_expire: String(securityForm.jwt_access_expire),
        jwt_refresh_expire: String(securityForm.jwt_refresh_expire),
        login_max_failure: String(securityForm.login_max_failure),
        login_lock_duration: String(securityForm.login_lock_duration),
        realname_notify_text: securityForm.realname_notify_text,
      })
      if (res.isSuccess) {
        settingsStore.updateConfig({ geetest_captcha_id: securityForm.geetest_captcha_id })
        message.success(res.message || t('adminSettings.securitySettingsSaved'))
      }
      else {
        message.error(res.message || t('adminSettings.securitySettingsSaveFailed'))
      }
    }
    catch {
      message.error(t('adminSettings.saveFailed'))
    }
    finally {
      savingSecurity.value = false
    }
  }

  async function handleRestartBackend() {
    restartingBackend.value = true
    try {
      const res = await adminApi.settings.restartBackend()
      if (res.isSuccess)
        message.success(res.message || t('adminSettings.restartBackendRequested'))
      else message.error(res.message || t('adminSettings.restartBackendFailed'))
    }
    catch {
      message.error(t('adminSettings.restartFailed'))
    }
    finally {
      restartingBackend.value = false
    }
  }

  async function handleAddSetting() {
    try {
      await addFormRef.value?.validate()
    }
    catch {
      return
    }
    adding.value = true
    try {
      const res = await adminApi.settings.create({
        key: addForm.key,
        value: addForm.value,
        type: addForm.type as SettingType,
        category: 'custom',
        label: addForm.label,
        description: addForm.description,
        is_public: addForm.is_public,
        is_editable: true,
      })
      if (res.isSuccess) {
        message.success(res.message || t('adminSettings.configItemAdded'))
        showAddModal.value = false
        addForm.key = ''
        addForm.value = ''
        addForm.label = ''
        addForm.type = 'string'
        addForm.description = ''
        addForm.is_public = false
        await loadSettings()
      }
      else {
        message.error(res.message || t('adminSettings.configItemAddFailed'))
      }
    }
    catch {
      message.error(t('adminSettings.addFailed'))
    }
    finally {
      adding.value = false
    }
  }

  async function handleDeleteSetting(key: string) {
    try {
      const res = await adminApi.settings.delete(key)
      if (res.isSuccess) {
        message.success(res.message || t('adminSettings.configItemDeleted'))
        await loadSettings()
      }
      else {
        message.error(res.message || t('adminSettings.configItemDeleteFailed'))
      }
    }
    catch (error: any) {
      message.error(t('adminSettings.deleteFailed') + (error.message || ''))
    }
  }

  function handleEditSetting(row: SettingDTO) {
    editForm.key = row.key
    editForm.value = row.value == null ? '' : String(row.value)
    editForm.label = row.label || ''
    editForm.type = row.type
    editForm.description = row.description || ''
    editForm.is_public = Boolean(row.is_public)
    showEditModal.value = true
  }

  async function handleSaveSettingEdit() {
    if (!editForm.key)
      return
    savingEdit.value = true
    try {
      const res = await adminApi.settings.updateMeta(editForm.key, {
        value: editForm.value,
        type: editForm.type,
        category: 'custom',
        label: editForm.label,
        description: editForm.description,
        is_public: editForm.is_public,
        is_editable: true,
      })
      if (res.isSuccess) {
        message.success(res.message || t('adminSettings.configItemUpdated'))
        showEditModal.value = false
        await loadSettings()
      }
      else {
        message.error(res.message || t('adminSettings.configItemUpdateFailed'))
      }
    }
    catch {
      message.error(t('adminSettings.editFailed'))
    }
    finally {
      savingEdit.value = false
    }
  }

  return {
    loadSettings,
    handleUpdateAllowRegister,
    handleUpdateAnnouncementEnabled,
    handleUpdateAllowDeleteAccount,
    handleUpdateDisableWebLogin,
    handleUpdateSmtpSSL,
    handleUpdateSmtpProxyEnabled,
    handleUpdateEmailVerifyEnabled,
    handleUpdateSmsVerifyEnabled,
    handleUpdateMobileCnOnly,
    handleUpdateMobileIpCountryDetect,
    handleUpdateGeetestEnabled,
    handleUpdateRealnameEnabled,
    handleUpdateRealnameReviewRequired,
    handleUpdateRealnameApiEnabled,
    handleUpdatePaymentEnabled,
    handleUpdateWithdrawEnabled,
    handleUpdateWithdrawRequireRealname,
    handleUpdateFinanceDualApproval,
    handleSaveBasic,
    handleSaveEmail,
    handleTestEmail,
    handleSaveSms,
    handleTestSms,
    handleSaveSecurity,
    handleRestartBackend,
    handleSaveRealnameApi,
    handleSavePayment,
    handleAddSetting,
    handleDeleteSetting,
    handleEditSetting,
    handleSaveSettingEdit,
  }
}
