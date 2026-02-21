<script setup lang="ts">
import type { FormInst } from 'naive-ui'
import { useAuthStore } from '@/store'
import { local } from '@/utils'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'

const emit = defineEmits(['update:modelValue'])

const authStore = useAuthStore()
const isGeetestEnabled = computed(() => geetestManager.isEnabled())
const isCaptchaVerified = ref(!geetestManager.isEnabled())
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

const formRef = ref<FormInst | null>(null)

async function handleLogin() {
  if (isGeetestEnabled.value && !isCaptchaVerified.value) {
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

  if (isRemember.value)
    local.set('loginAccount', { account, pwd })
  else local.remove('loginAccount')

  const hadToken = Boolean(local.get('accessToken'))
  await authStore.login(account, pwd)
  const hasTokenNow = Boolean(local.get('accessToken'))

  if (!hadToken && !hasTokenNow) {
    isCaptchaVerified.value = false
    geetestManager.clearCaptchaResult()
    captchaKey.value++ // 登录失败，重新渲染极验
  }
  isLoading.value = false
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
  if (geetestManager.isEnabled()) {
    isCaptchaVerified.value = false
    geetestManager.clearCaptchaResult()
  }
})

watchEffect(() => {
  if (!geetestManager.isEnabled())
    isCaptchaVerified.value = true
})
onMounted(() => {
  checkUserAccount()
})
function checkUserAccount() {
  const loginAccount = local.get('loginAccount')
  if (!loginAccount)
    return

  formValue.value = loginAccount
  isRemember.value = true
}
</script>

<template>
  <div>
    <n-h2 depth="3" class="text-center">
      {{ $t('login.signInTitle') }}
    </n-h2>
    <n-form ref="formRef" :rules="rules" :model="formValue" :show-label="false" size="large">
      <n-form-item path="account">
        <n-input v-model:value="formValue.account" clearable :placeholder="$t('login.accountOrEmailPlaceholder')" :input-props="{ autocomplete: 'username' }" />
      </n-form-item>
      <n-form-item path="pwd">
        <n-input v-model:value="formValue.pwd" type="password" :placeholder="$t('login.passwordPlaceholder')" clearable show-password-on="click" :input-props="{ autocomplete: 'current-password' }">
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
          <n-button type="primary" text @click="toOtherForm('resetPwd')">
            {{ $t('login.forgotPassword') }}
          </n-button>
        </div>
        <GeetestCaptcha v-if="isGeetestEnabled" :key="captchaKey" @success="onGeetestSuccess" @error="onGeetestError" />
        <n-button block type="primary" size="large" :loading="isLoading" :disabled="isLoading" @click="handleLogin">
          {{ $t('login.signIn') }}
        </n-button>
        <n-flex>
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
