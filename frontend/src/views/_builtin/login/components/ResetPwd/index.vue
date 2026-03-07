<script setup lang="ts">
import type { FormInst } from 'naive-ui'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'
import { fetchSendResetEmail } from '@/service'
import { i18n } from '@/modules/i18n'

const emit = defineEmits(['update:modelValue'])
function toLogin() {
  emit('update:modelValue', 'login')
}
const { t } = useI18n()

const isGeetestEnabled = computed(() => geetestManager.isEnabled())
const isCaptchaVerified = ref(!geetestManager.isEnabled())
const captchaKey = ref(0)

const rules = computed(() => {
  return {
    account: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.resetPasswordRuleTip'),
      },
    ],
  }
})
const formValue = ref({
  account: '',
})
const formRef = ref<FormInst | null>(null)
const isLoading = ref(false)

const geetestRef = ref<any>(null)

async function handleReset() {
  if (!formValue.value.account) {
    window.$message.warning(t('login.resetPasswordRuleTip'))
    return
  }

  if (isGeetestEnabled.value) {
    geetestRef.value?.showCaptcha()
  }
  else {
    sendResetEmail()
  }
}

async function sendResetEmail() {
  const hasErrors = await new Promise<boolean>((resolve) => {
    formRef.value?.validate((errors) => {
      resolve(Boolean(errors))
    })
  })

  if (hasErrors)
    return

  isLoading.value = true
  try {
    const { isSuccess } = await fetchSendResetEmail({
      email: formValue.value.account,
      lang: i18n.global.locale.value,
    })
    if (isSuccess) {
      window.$message.success(t('login.resetEmailSent'))
    }
  }
  finally {
    isLoading.value = false
  }
}

async function onGeetestSuccess() {
  isCaptchaVerified.value = true
  await sendResetEmail()
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

  if (geetestManager.isEnabled()) {
    isCaptchaVerified.value = false
    geetestManager.clearCaptchaResult()
  }
})

watchEffect(() => {
  if (!geetestManager.isEnabled())
    isCaptchaVerified.value = true
})
</script>

<template>
  <div>
    <n-h2 depth="3" class="text-center">
      {{ $t('login.resetPasswordTitle') }}
    </n-h2>
    <n-form
      ref="formRef"
      :rules="rules"
      :model="formValue"
      :show-label="false"
      size="large"
    >
      <n-form-item path="account">
        <n-input
          v-model:value="formValue.account"
          clearable
          :placeholder="$t('login.accountOrEmailPlaceholder')"
        />
      </n-form-item>
      <n-form-item>
        <n-space
          vertical
          :size="20"
          class="w-full"
        >
          <GeetestCaptcha
            v-if="isGeetestEnabled"
            ref="geetestRef"
            :key="captchaKey"
            :config="{ product: 'bind' }"
            @success="onGeetestSuccess"
            @error="onGeetestError"
          />
          <n-button
            block
            type="primary"
            :loading="isLoading"
            :disabled="isLoading"
            @click="handleReset"
          >
            {{ $t('login.resetPassword') }}
          </n-button>
          <n-flex justify="center">
            <n-text>{{ $t('login.haveAccountText') }}</n-text>
            <n-button
              text
              type="primary"
              @click="toLogin"
            >
              {{ $t('login.signIn') }}
            </n-button>
          </n-flex>
        </n-space>
      </n-form-item>
    </n-form>
  </div>
</template>

<style scoped></style>
