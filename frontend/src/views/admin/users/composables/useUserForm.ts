/**
 * 用户新建/编辑表单 CRUD（不含余额积分提现，那些在 useUserFinance）
 */
import { computed, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage } from 'naive-ui'
import type { FormRules } from 'naive-ui'
import {
  createUser,
  updateAdminUserProfile,
} from '@/service/api/admin/user'
import type { AdminUser } from '@/service/api/admin/user'
import { useSettingsStore } from '@/store'
import { withSubmitLock } from '@/hooks'
import { normalizeAndValidateMobile } from '@/utils/phone'

function reportAdminUsersError(message: string, error?: unknown) {
  if (import.meta.env.DEV)
    console.error(message, error)
}

export function useUserForm(options?: {
  /** 保存成功后回调（通常刷新列表） */
  onSuccess?: () => void
}) {
  const message = useMessage()
  const { t } = useI18n()

  const showUserModal = ref(false)
  const isEdit = ref(false)
  const submitting = ref(false)
  const formRef = ref()
  const selectedUser = ref<AdminUser | null>(null)
  const activeTab = ref('details')
  const isFullscreen = ref(false)

  const userForm = reactive({
    username: '',
    nickname: '',
    email: '',
    mobile: '',
    language: 'zh-CN',
    country: '',
    admin_remark: '',
    password: '',
    role: 'user',
    level: 1,
    status: 1,
    avatar: '',
    gender: 0,
    birthday: null as number | null,
    motto: '',
  })

  const roleOptions = computed(() => [
    { label: t('adminUsers.admin'), value: 'admin' },
    { label: t('adminUsers.normalUser'), value: 'user' },
  ])

  const userStatusOptions = computed(() => [
    { label: t('adminUsers.enabled'), value: 1 },
    { label: t('adminUsers.disabled'), value: 0 },
  ])

  const genderOptions = computed(() => [
    { label: t('adminUsers.unknown'), value: 0 },
    { label: t('adminUsers.male'), value: 1 },
    { label: t('adminUsers.female'), value: 2 },
  ])

  const languageOptions = computed(() => [
    { label: t('adminUsersDetail.chinese'), value: 'zh-CN' },
    { label: t('adminUsersDetail.english'), value: 'en-US' },
  ])

  const rules: FormRules = {
    username: [
      { required: true, message: t('adminUsers.enterUsername'), trigger: 'blur' },
      { min: 3, max: 20, message: t('adminUsers.usernameLength'), trigger: 'blur' },
    ],
    email: [
      { required: true, message: t('adminUsers.enterEmail'), trigger: 'blur' },
      { type: 'email' as any, message: t('adminUsers.invalidEmail'), trigger: 'blur' },
    ],
    role: [
      { required: true, message: t('adminUsers.selectRole'), trigger: 'change' },
    ],
    level: [
      { type: 'number' as any, message: t('adminUsers.enterValidLevel'), trigger: 'change' },
    ],
    status: [
      {
        required: true,
        validator: (_rule: any, value: any) => {
          if (value === null || value === undefined || value === '')
            return new Error(t('adminUsers.selectStatus'))
          return true
        },
        trigger: 'change',
      },
    ],
  }

  /** 仅新建时校验密码；编辑请走详情页重置密码 */
  const passwordRule = computed(() => {
    if (isEdit.value)
      return []
    return [
      { required: true, message: t('adminUsers.enterPassword'), trigger: 'blur' },
      { min: 8, message: t('adminUsers.passwordLength'), trigger: 'blur' },
    ]
  })

  function resetUserForm() {
    Object.assign(userForm, {
      username: '',
      nickname: '',
      email: '',
      mobile: '',
      language: 'zh-CN',
      country: '',
      admin_remark: '',
      password: '',
      role: 'user',
      level: 1,
      status: 1,
      avatar: '',
      gender: 0,
      birthday: null,
      motto: '',
    })
  }

  function handleAdd() {
    isEdit.value = false
    selectedUser.value = null
    resetUserForm()
    activeTab.value = 'details'
    showUserModal.value = true
  }

  function handleEdit(user: AdminUser) {
    isEdit.value = true
    selectedUser.value = user
    activeTab.value = 'details'

    Object.assign(userForm, {
      username: user.username,
      nickname: user.nickname || '',
      email: user.email || '',
      mobile: user.mobile || '',
      language: user.language || 'zh-CN',
      country: user.country || '',
      admin_remark: user.admin_remark || '',
      password: '',
      role: user.role,
      level: user.level,
      status: Number(user.status) || 0,
      avatar: user.avatar || '',
      gender: user.gender || 0,
      birthday: user.birthday ? Number(user.birthday) * 1000 : null,
      motto: user.motto || '',
    })
    showUserModal.value = true
  }

  function toggleFullscreen() {
    isFullscreen.value = !isFullscreen.value
  }

  function handleAvatarError() {
    reportAdminUsersError('[adminUsers] avatar load failed')
  }

  /** 按后台 mobile_cn_only 规范化手机号；空号允许；非法则提示并返回 null */
  function resolveMobileOrWarn(raw: string): string | null {
    const trimmed = String(raw || '').trim()
    if (!trimmed)
      return ''
    const settingsStore = useSettingsStore()
    const cnOnly = settingsStore.mobileCnOnly
    const normalized = normalizeAndValidateMobile(trimmed, cnOnly)
    if (!normalized) {
      message.error(cnOnly ? t('adminUsers.invalidMobileCN') : t('adminUsers.invalidMobile'))
      return null
    }
    return normalized
  }

  async function handleSubmit() {
    if (submitting.value)
      return
    try {
      await formRef.value?.validate()
    }
    catch (error) {
      reportAdminUsersError('[adminUsers] form validation failed', error)
      return
    }

    const resolvedMobile = resolveMobileOrWarn(userForm.mobile)
    if (resolvedMobile === null)
      return
    // 回写规范化结果，便于后续对比与展示
    userForm.mobile = resolvedMobile

    await withSubmitLock(submitting, async () => {
      if (isEdit.value) {
        const originalUser = selectedUser.value
        const changedData: Record<string, any> = {}

        if (userForm.nickname !== (originalUser?.nickname || ''))
          changedData.nickname = userForm.nickname
        if (userForm.email !== (originalUser?.email || ''))
          changedData.email = userForm.email
        if (userForm.mobile !== (originalUser?.mobile || ''))
          changedData.mobile = userForm.mobile
        if (userForm.language !== (originalUser?.language || 'zh-CN'))
          changedData.language = userForm.language
        if (userForm.country !== (originalUser?.country || ''))
          changedData.country = userForm.country
        if (userForm.admin_remark !== (originalUser?.admin_remark || ''))
          changedData.admin_remark = userForm.admin_remark
        if (userForm.role !== originalUser?.role)
          changedData.role = userForm.role
        if (userForm.level !== originalUser?.level)
          changedData.level = userForm.level

        const originalStatus = Number(originalUser?.status) || 0
        if (userForm.status !== originalStatus)
          changedData.status = userForm.status

        if (userForm.avatar !== (originalUser?.avatar || ''))
          changedData.avatar = userForm.avatar
        if (userForm.gender !== (originalUser?.gender || 0))
          changedData.gender = userForm.gender

        const originalBirthday = originalUser?.birthday ? Number(originalUser.birthday) * 1000 : null
        const formBirthday = userForm.birthday ? Number(userForm.birthday) : null
        if (originalBirthday !== formBirthday)
          changedData.birthday = formBirthday ? Math.floor(formBirthday / 1000) : null

        if (userForm.motto !== (originalUser?.motto || ''))
          changedData.motto = userForm.motto

        if (Object.keys(changedData).length === 0) {
          message.warning(t('adminUsers.noChangesDetected'))
          return
        }

        const response: any = await updateAdminUserProfile(selectedUser.value?.id as number, changedData)
        if (response.isSuccess) {
          message.success(t('adminUsers.updateSuccess'))
          showUserModal.value = false
          options?.onSuccess?.()
        }
        else {
          message.error(response.message || t('adminUsers.updateFailed'))
        }
        return
      }

      const userPayload = {
        username: userForm.username,
        password: userForm.password,
        email: userForm.email,
        nickname: userForm.nickname,
        mobile: userForm.mobile,
        language: userForm.language,
        country: userForm.country,
        admin_remark: userForm.admin_remark,
        level: userForm.level,
        role: userForm.role,
        status: userForm.status,
      }
      const response: any = await createUser(userPayload as any)
      if (response.isSuccess) {
        message.success(t('adminUsers.createSuccess'))
        showUserModal.value = false
        options?.onSuccess?.()
      }
      else {
        message.error(response.message || t('adminUsers.createFailed'))
      }
    })
  }

  return {
    showUserModal,
    isEdit,
    submitting,
    formRef,
    selectedUser,
    activeTab,
    isFullscreen,
    userForm,
    rules,
    passwordRule,
    roleOptions,
    userStatusOptions,
    genderOptions,
    languageOptions,
    handleAdd,
    handleEdit,
    handleSubmit,
    toggleFullscreen,
    handleAvatarError,
  }
}
