<script setup lang="ts">
import type { FormInst } from 'naive-ui'
import { useAuthStore, useSettingsStore } from '@/store'
import { authStorage, local } from '@/utils'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'
import { getRuntimeRouteMode } from '@/router/runtime-mode'

const props = withDefaults(defineProps<{
  preserveCurrentPage?: boolean
}>(), {
  preserveCurrentPage: false,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: 'login' | 'register' | 'resetPwd'): void
  (e: 'success'): void
}>()

const authStore = useAuthStore()
const settingsStore = useSettingsStore()

// 极验是否启用：以后端配置为准
const isGeetestEnabled = computed(() => settingsStore.geetestEnabled)
// 检查是否配置了 captchaId（从后端获取）
const hasCaptchaId = computed(() => Boolean(settingsStore.geetestCaptchaId))
// 综合判断：后端启用 且 有配置 captchaId
const shouldShowCaptcha = computed(() => isGeetestEnabled.value && hasCaptchaId.value)

// 关闭「允许用户登录」后禁用用户端登录表单（管理端不受影响）；注册入口单独看 allowRegister
const isUserLoginDisabled = computed(() => getRuntimeRouteMode() !== 'admin' && !settingsStore.allowUserLogin)
const showRegisterEntry = computed(() => settingsStore.allowRegister)

const isCaptchaVerified = ref(false)
const captchaKey = ref(0)
const geetestRef = ref<{ showCaptcha: () => void } | null>(null)

function toOtherForm(type: 'login' | 'register' | 'resetPwd') {
  emit('update:modelValue', type)
}

const { t } = useI18n()
const rules = computed(() => {
  return {
    account: {
      required: true,
      trigger: 'blur',
      message: t('login.accountRuleTip'),
    },
    pwd: {
      required: true,
      trigger: 'blur',
      message: t('login.passwordRuleTip'),
    },
  }
})
const formValue = ref({
  account: '',
  pwd: '',
})
const isRemember = ref(false)
const isLoading = ref(false)

const formRef = ref<FormInst | null>(null)

/** 校验表单；有错返回 true */
async function validateLoginForm() {
  return await new Promise<boolean>((resolve) => {
    formRef.value?.validate((errors) => {
      resolve(Boolean(errors))
    })
  })
}

/**
 * 登录入口：按钮点击 / 回车共用。
 * 有极验且未通过时，先弹出极验；通过后再真正登录。
 */
async function handleLogin() {
  if (isLoading.value)
    return
  if (isUserLoginDisabled.value) {
    window.$message.warning(t('login.userLoginDisabledTip'))
    return
  }

  // 先校验表单，避免空字段就弹极验
  if (await validateLoginForm())
    return

  if (shouldShowCaptcha.value && !isCaptchaVerified.value) {
    geetestRef.value?.showCaptcha()
    return
  }

  await doLogin()
}

/** 极验通过（或未启用）后执行实际登录 */
async function doLogin() {
  if (isLoading.value)
    return

  isLoading.value = true
  const { account, pwd } = formValue.value

  // 「记住我」= 只记住账号到 localStorage；密码交给浏览器密码管理器保存，禁止写本地明文
  if (isRemember.value)
    local.set('loginAccount', { account })
  else local.remove('loginAccount')

  const hadToken = Boolean(authStorage.get('accessToken'))
  const result = await authStore.login(account, pwd, {
    preserveCurrentPage: props.preserveCurrentPage,
  })

  const hasTokenNow = Boolean(authStorage.get('accessToken'))

  if (result.status === 'ok')
    emit('success')

  if (!hadToken && !hasTokenNow) {
    isCaptchaVerified.value = false
    geetestManager.clearCaptchaResult()
    captchaKey.value++ // 登录失败，重新渲染极验
  }
  isLoading.value = false
}

async function onGeetestSuccess() {
  isCaptchaVerified.value = true
  await doLogin()
}

function onGeetestError() {
  isCaptchaVerified.value = false
  geetestManager.clearCaptchaResult()
  captchaKey.value++ // 验证错误，重新渲染极验
}

watch(() => formValue.value.account, (val) => {
  if (!val)
    return
  // 使用正则表达式提取 @ 及其后面的部分，并将该部分转为小写
  const formatted = val.replace(/@.*$/, match => match.toLowerCase())
  if (formatted !== val) {
    formValue.value.account = formatted
  }
})

watch(() => [formValue.value.account, formValue.value.pwd], () => {
  if (shouldShowCaptcha.value) {
    isCaptchaVerified.value = false
    geetestManager.clearCaptchaResult()
  }
})

watchEffect(() => {
  if (!shouldShowCaptcha.value)
    isCaptchaVerified.value = true
})

// 监听后端配置变化，初始化验证状态
watch(shouldShowCaptcha, (show) => {
  if (!show) {
    isCaptchaVerified.value = true
  }
}, { immediate: true })
onMounted(() => {
  checkUserAccount()
})
function checkUserAccount() {
  // 只恢复账号；密码依赖浏览器密码管理器（input autocomplete=current-password）
  const loginAccount = local.get('loginAccount') as { account?: string, pwd?: string } | null
  if (!loginAccount)
    return

  // 兼容旧数据：若本地曾存过 pwd 字段，读入后立即清掉并重写为仅账号
  formValue.value = {
    account: loginAccount.account || '',
    pwd: '',
  }
  isRemember.value = true
  if (loginAccount.pwd !== undefined) {
    local.set('loginAccount', { account: loginAccount.account || '' })
  }
}
</script>

<template>
  <div>
    <n-h2 depth="3" class="text-center">
      {{ $t('login.signInTitle') }}
    </n-h2>
    <n-alert v-if="isUserLoginDisabled" type="warning" :show-icon="true" class="mb-16" :title="$t('login.userLoginDisabledTip')" />

    <n-form ref="formRef" :rules="rules" :model="formValue" :show-label="false" size="large" :disabled="isUserLoginDisabled">
      <!-- 账号 username / 密码 current-password：配合浏览器密码管理器，不写 localStorage 明文密码 -->
      <n-form-item path="account">
        <n-input
          v-model:value="formValue.account"
          clearable
          :placeholder="$t('login.accountOrEmailPlaceholder')"
          name="username"
          :input-props="{ autocomplete: 'username', name: 'username' }"
          @keyup.enter="handleLogin"
        />
      </n-form-item>
      <n-form-item path="pwd">
        <n-input
          v-model:value="formValue.pwd"
          type="password"
          :placeholder="$t('login.passwordPlaceholder')"
          clearable
          show-password-on="click"
          name="password"
          :input-props="{ autocomplete: 'current-password', name: 'password' }"
          @keyup.enter="handleLogin"
        >
          <template #password-invisible-icon>
            <icon-park-outline-preview-close-one />
          </template>
          <template #password-visible-icon>
            <icon-park-outline-preview-open />
          </template>
        </n-input>
      </n-form-item>
      <n-space vertical :size="20">
        <div class="flex-y-center justify-between">
          <n-checkbox v-model:checked="isRemember">
            {{ $t('login.rememberMe') }}
          </n-checkbox>
          <n-button type="primary" text :disabled="isUserLoginDisabled" @click="toOtherForm('resetPwd')">
            {{ $t('login.forgotPassword') }}
          </n-button>
        </div>
        <!-- bind：不占位按钮，由登录/回车主动 showCaptcha -->
        <GeetestCaptcha
          v-if="shouldShowCaptcha"
          ref="geetestRef"
          :key="captchaKey"
          :config="{ product: 'bind' }"
          @success="onGeetestSuccess"
          @error="onGeetestError"
        />
        <n-button block type="primary" size="large" :loading="isLoading" :disabled="isLoading || isUserLoginDisabled" @click="handleLogin">
          {{ $t('login.signIn') }}
        </n-button>
        <n-flex v-if="showRegisterEntry">
          <n-text>{{ $t('login.noAccountText') }}</n-text>
          <n-button type="primary" text @click="toOtherForm('register')">
            {{ $t('login.signUp') }}
          </n-button>
        </n-flex>
      </n-space>
    </n-form>
    <n-divider>
      <span op-80>{{ $t('login.or') }}</span>
    </n-divider>
    <n-space justify="center">
      <n-button circle>
        <template #icon>
          <n-icon><icon-park-outline-wechat /></n-icon>
        </template>
      </n-button>
      <n-button circle>
        <template #icon>
          <n-icon><icon-park-outline-tencent-qq /></n-icon>
        </template>
      </n-button>
      <n-button circle>
        <template #icon>
          <n-icon><icon-park-outline-github-one /></n-icon>
        </template>
      </n-button>
    </n-space>
  </div>
</template>

<style scoped></style>
