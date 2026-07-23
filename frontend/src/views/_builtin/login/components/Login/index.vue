<script setup lang="ts">
import type { FormInst } from 'naive-ui'
import { useAuthStore, useSettingsStore } from '@/store'
import { authStorage, local } from '@/utils'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'
import { getRuntimeRouteMode } from '@/router/runtime-mode'

const emit = defineEmits(['update:modelValue'])

const authStore = useAuthStore()
const settingsStore = useSettingsStore()

// 极验是否启用：以后端配置为准
const isGeetestEnabled = computed(() => settingsStore.geetestEnabled)
// 检查是否配置了 captchaId（从后端获取）
const hasCaptchaId = computed(() => Boolean(settingsStore.geetestCaptchaId))
// 综合判断：后端启用 且 有配置 captchaId
const shouldShowCaptcha = computed(() => isGeetestEnabled.value && hasCaptchaId.value)

// 「禁止网页端登录」开关仅拦截普通用户（管理端路由模式不受影响）；
// 命中时直接禁用登录表单，避免用户提交后才收到后端 403 提示。
const isWebLoginDisabled = computed(() => getRuntimeRouteMode() !== 'admin' && settingsStore.webLoginDisabled)

const isCaptchaVerified = ref(false)
const captchaKey = ref(0)

function toOtherForm(type: any) {
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
// 管理端 TOTP 第二步
const needTotp = ref(false)
const totpTempToken = ref('')
const totpCode = ref('')

const formRef = ref<FormInst | null>(null)

async function handleLogin() {
  if (isWebLoginDisabled.value) {
    window.$message.warning(t('login.webLoginDisabledTip'))
    return
  }

  // 只有当需要显示验证码且未验证时才提示
  if (shouldShowCaptcha.value && !isCaptchaVerified.value) {
    window.$message.warning(t('login.captchaRequired'))
    return
  }

  const hasErrors = await new Promise<boolean>((resolve) => {
    formRef.value?.validate((errors) => {
      resolve(Boolean(errors))
    })
  })
  if (hasErrors)
    return

  isLoading.value = true
  const { account, pwd } = formValue.value

  // 「记住我」= 只记住账号到 localStorage；密码交给浏览器密码管理器保存，禁止写本地明文
  if (isRemember.value)
    local.set('loginAccount', { account })
  else local.remove('loginAccount')

  const hadToken = Boolean(authStorage.get('accessToken'))
  const loginResult = await authStore.login(account, pwd)
  if (loginResult.status === 'need_totp' && loginResult.tempToken) {
    needTotp.value = true
    totpTempToken.value = loginResult.tempToken
    totpCode.value = ''
    isLoading.value = false
    return
  }

  const hasTokenNow = Boolean(authStorage.get('accessToken'))

  if (!hadToken && !hasTokenNow) {
    isCaptchaVerified.value = false
    geetestManager.clearCaptchaResult()
    captchaKey.value++ // 登录失败，重新渲染极验
  }
  isLoading.value = false
}

async function handleTotpLogin() {
  if (!totpCode.value || !totpTempToken.value) {
    window.$message.warning(t('login.totpRequired'))
    return
  }
  isLoading.value = true
  const ok = await authStore.loginWithTotp(totpTempToken.value, totpCode.value.trim())
  if (!ok) {
    isLoading.value = false
    return
  }
  needTotp.value = false
  totpTempToken.value = ''
  totpCode.value = ''
  isLoading.value = false
}

function cancelTotp() {
  needTotp.value = false
  totpTempToken.value = ''
  totpCode.value = ''
}

async function onGeetestSuccess() {
  isCaptchaVerified.value = true
  await handleLogin()
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
      {{ needTotp ? $t('login.totpTitle') : $t('login.signInTitle') }}
    </n-h2>
    <n-alert v-if="isWebLoginDisabled" type="warning" :show-icon="true" class="mb-16" :title="$t('login.webLoginDisabledTip')" />

    <!-- 管理端 TOTP 第二步：输入动态码 -->
    <n-space v-if="needTotp" vertical :size="20">
      <n-alert type="info" :show-icon="true" :title="$t('login.totpHint')" />
      <n-input
        v-model:value="totpCode"
        size="large"
        maxlength="8"
        :placeholder="$t('login.totpPlaceholder')"
        :input-props="{ autocomplete: 'one-time-code', inputmode: 'numeric' }"
        @keyup.enter="handleTotpLogin"
      />
      <n-button block type="primary" size="large" :loading="isLoading" :disabled="isLoading" @click="handleTotpLogin">
        {{ $t('login.totpConfirm') }}
      </n-button>
      <n-button block quaternary :disabled="isLoading" @click="cancelTotp">
        {{ $t('common.cancel') }}
      </n-button>
    </n-space>

    <n-form v-else ref="formRef" :rules="rules" :model="formValue" :show-label="false" size="large" :disabled="isWebLoginDisabled">
      <!-- 账号 username / 密码 current-password：配合浏览器密码管理器，不写 localStorage 明文密码 -->
      <n-form-item path="account">
        <n-input v-model:value="formValue.account" clearable :placeholder="$t('login.accountOrEmailPlaceholder')" name="username" :input-props="{ autocomplete: 'username', name: 'username' }" />
      </n-form-item>
      <n-form-item path="pwd">
        <n-input v-model:value="formValue.pwd" type="password" :placeholder="$t('login.passwordPlaceholder')" clearable show-password-on="click" name="password" :input-props="{ autocomplete: 'current-password', name: 'password' }">
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
          <n-button type="primary" text :disabled="isWebLoginDisabled" @click="toOtherForm('resetPwd')">
            {{ $t('login.forgotPassword') }}
          </n-button>
        </div>
        <GeetestCaptcha v-if="shouldShowCaptcha" :key="captchaKey" @success="onGeetestSuccess" @error="onGeetestError" />
        <n-button block type="primary" size="large" :loading="isLoading" :disabled="isLoading || isWebLoginDisabled" @click="handleLogin">
          {{ $t('login.signIn') }}
        </n-button>
        <n-flex v-if="!isWebLoginDisabled">
          <n-text>{{ $t('login.noAccountText') }}</n-text>
          <n-button type="primary" text @click="toOtherForm('register')">
            {{ $t('login.signUp') }}
          </n-button>
        </n-flex>
      </n-space>
    </n-form>
    <template v-if="!needTotp">
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
    </template>
  </div>
</template>

<style scoped></style>
