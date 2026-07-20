/**
 * 管理端设置：共享表单与加载态（模块级单例）
 */
import { reactive, ref } from 'vue'
import type { SettingDTO, SettingType } from '@/service/api/admin/settings'

export const loading = ref(true)
export const adding = ref(false)
export const savingEdit = ref(false)
export const showAddModal = ref(false)
export const showEditModal = ref(false)
export const savingBasic = ref(false)
export const savingEmail = ref(false)
export const savingSms = ref(false)
export const savingSecurity = ref(false)
export const savingRealnameApi = ref(false)
export const savingPayment = ref(false)
export const testingEmail = ref(false)
export const testingSms = ref(false)
/** 邮件设置页「发送测试邮件」的收件人（不入库，仅本页临时） */
export const testEmailTo = ref('')
/** 短信设置页「发送测试短信」的手机号（不入库，仅本页临时） */
export const testSmsPhone = ref('')
export const restartingBackend = ref(false)
export const topTab = ref('system-config')
export const systemSubTab = ref('basic')

export const switchLoading = reactive({
  allow_register: false,
  announcement_enabled: false,
  allow_delete_account: false,
  smtp_ssl: false,
  smtp_proxy_enabled: false,
  geetest_enabled: false,
  realname_enabled: false,
  realname_review_required: false,
  realname_api_enabled: false,
  email_verify_enabled: false,
  sms_verify_enabled: false,
  payment_enabled: false,
  withdraw_enabled: false,
  mobile_cn_only: false,
  mobile_ip_country_detect: false,
})

export const basicForm = reactive({
  site_name: '',
  site_desc: '',
  site_logo: '',
  copyright: '',
  icp: '',
  version: '',
  default_lang: 'zhCN',
  allow_register: true,
  announcement_enabled: true,
  frontend_url: '',
  backend_api_url: '',
})

export const emailForm = reactive({
  email_verify_enabled: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  // 默认关：587 走 STARTTLS；465 再开隐式 SSL
  smtp_ssl: false,
  // 默认关：国内直连国外 SMTP 不通时再开代理
  smtp_proxy_enabled: false,
  smtp_proxy_type: 'socks5',
  smtp_proxy_host: '',
  smtp_proxy_port: 1080,
  smtp_proxy_username: '',
  smtp_proxy_password: '',
  system_email_address: '',
  system_email_name: '',
})

export const smsForm = reactive({
  sms_verify_enabled: false,
  mobile_cn_only: true,
  mobile_ip_country_detect: false,
  sms_provider: 'console',
  sms_access_key: '',
  sms_secret_key: '',
  sms_sign_name: '',
  sms_template_code: '',
  sms_template_code_en: '',
  sms_region: '',
  sms_sdk_app_id: '',
  sms_endpoint: '',
  sms_body_format: 'json',
})

export const securityForm = reactive({
  geetest_enabled: false,
  geetest_captcha_id: '',
  geetest_captcha_key: '',
  jwt_access_expire: 7200,
  jwt_refresh_expire: 604800,
  login_max_failure: 5,
  login_lock_duration: 10,
  allow_delete_account: false,
  realname_enabled: true,
  realname_review_required: true,
  realname_notify_text: '',
})

export const realnameApiForm = reactive({
  realname_api_enabled: false,
  realname_api_provider: 'aliyun',
  realname_api_app_key: '',
  realname_api_app_secret: '',
  realname_api_endpoint: '',
})

export const paymentForm = reactive({
  payment_enabled: false,
  payment_order_expire_minutes: 30,
  withdraw_enabled: true,
  withdraw_min_amount: 10,
  withdraw_notify_text: '',
  withdraw_account_types_text: '["bank","alipay","wechat","usdt"]',
})

export const customSettings = ref<SettingDTO[]>([])
export const addFormRef = ref()
export const addForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as string,
  description: '',
  is_public: false,
})
export const editForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as SettingType,
  description: '',
  is_public: false,
})
