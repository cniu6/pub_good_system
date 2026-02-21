<script setup lang="ts">
import { Regex } from '@/constants/Regex'
import GeetestCaptcha from '@/components/common/GeetestCaptcha.vue'
import { geetestManager } from '@/utils/geetest'
import { fetchRegister, fetchSendRegisterCode } from '@/service'
import { i18n } from '@/modules/i18n'

const emit = defineEmits(['update:modelValue'])
function toLogin() {
  emit('update:modelValue', 'login')
}
const { t } = useI18n()

const isGeetestEnabled = computed(() => geetestManager.isEnabled())
const isCaptchaVerified = ref(!geetestManager.isEnabled())
const captchaKey = ref(0)

const formValue = ref({
  username: '',
  account: '',
  code: '',
  pwd: '',
  rePwd: '',
})

const rules = computed(() => {
  return {
    username: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.usernameRuleTip'),
      },
      {
        pattern: Regex.Username,
        trigger: 'blur',
        message: t('login.usernameInvalid'),
      },
    ],
    account: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.accountRuleTip'),
      },
      {
        pattern: Regex.Email,
        trigger: 'blur',
        message: t('login.emailInvalid'),
      },
    ],
    code: {
      required: true,
      trigger: 'blur',
      message: t('login.codeRequired'),
    },
    pwd: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.passwordRuleTip'),
      },
      {
        pattern: Regex.Password,
        trigger: 'blur',
        message: t('login.passwordInvalid'),
      },
    ],
    rePwd: [
      {
        required: true,
        trigger: 'blur',
        message: t('login.checkPasswordRuleTip'),
      },
      {
        validator: (rule: any, value: string) => {
          if (value !== formValue.value.pwd) {
            return new Error(t('login.passwordNotMatch'))
          }
          return true
        },
        trigger: 'blur',
      },
    ],
  }
})

const formRef = ref<any>(null)
const isLoading = ref(false)
const isSending = ref(false)
const count = ref(0)
const timer = ref<any>(null)

function startCountDown() {
  count.value = 60
  timer.value = setInterval(() => {
    count.value--
    if (count.value <= 0) {
      clearInterval(timer.value)
      timer.value = null
    }
  }, 1000)
}

onUnmounted(() => {
  if (timer.value)
    clearInterval(timer.value)
})

const geetestRef = ref<any>(null)
const captchaPurpose = ref<'sendCode' | 'register'>('sendCode')

async function handleSendCode() {
  if (!formValue.value.account || !new RegExp(Regex.Email).test(formValue.value.account)) {
    window.$message.warning(t('login.emailInvalid'))
    return
  }

  captchaPurpose.value = 'sendCode'
  if (isGeetestEnabled.value) {
    geetestRef.value?.showCaptcha()
  }
  else {
    sendCode()
  }
}

async function sendCode() {
  isSending.value = true
  try {
    const { isSuccess } = await fetchSendRegisterCode({
      email: formValue.value.account,
      lang: i18n.global.locale.value,
    })
    if (isSuccess) {
      window.$message.success(t('login.codeSent'))
      startCountDown()
    }
  }
  finally {
    isSending.value = false
  }
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

async function onGeetestSuccess() {
  isCaptchaVerified.value = true
  if (captchaPurpose.value === 'sendCode') {
    await sendCode()
  }
  else if (captchaPurpose.value === 'register') {
    await doRegister()
  }
}

function onGeetestError() {
  isCaptchaVerified.value = false
  geetestManager.clearCaptchaResult()
  captchaKey.value++ // 验证错误，重新渲染极验
}

const isRead = ref(false)

function openAgreement() {
  const url = import.meta.env.VITE_USER_AGREEMENT_URL
  if (url) {
    window.open(url, '_blank')
  }
}

async function handleRegister() {
  if (isGeetestEnabled.value && !isCaptchaVerified.value) {
    captchaPurpose.value = 'register'
    geetestRef.value?.showCaptcha()
    return
  }

  if (!isRead.value) {
    window.$message.warning(t('login.readAndAgreeTip'))
    return
  }

  await doRegister()
}

async function doRegister() {
  const hasErrors = await new Promise<boolean>((resolve) => {
    formRef.value?.validate((errors: any) => {
      resolve(Boolean(errors))
    })
  })

  if (hasErrors)
    return

  isLoading.value = true
  try {
    const result = await fetchRegister({
      username: formValue.value.username,
      email: formValue.value.account,
      password: formValue.value.pwd,
      code: formValue.value.code,
    })
    if (result?.isSuccess) {
      window.$message.success(t('login.registerSuccess'))
      toLogin()
    }
  }
  catch {
    // ignore
  }
  finally {
    isLoading.value = false
  }
}

watch(() => [formValue.value.account, formValue.value.pwd, formValue.value.rePwd], () => {
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
      {{ $t('login.registerTitle') }}
    </n-h2>
    <n-form
      ref="formRef"
      :rules="rules"
      :model="formValue"
      :show-label="false"
      size="large"
    >
      <n-form-item path="username">
        <n-input
          v-model:value="formValue.username"
          clearable
          :placeholder="$t('login.usernamePlaceholder')"
        />
      </n-form-item>
      <n-form-item path="account">
        <n-input
          v-model:value="formValue.account"
          clearable
          :placeholder="$t('login.emailPlaceholder')"
          :input-props="{ autocomplete: 'username' }"
        />
      </n-form-item>
      <n-form-item path="code">
        <n-input-group>
          <n-input
            v-model:value="formValue.code"
            :placeholder="$t('login.codePlaceholder')"
          />
          <n-button
            type="primary"
            ghost
            :disabled="count > 0 || isSending"
            :loading="isSending"
            @click="handleSendCode"
          >
            {{ count > 0 ? `${count}s` : $t('login.sendCode') }}
          </n-button>
        </n-input-group>
      </n-form-item>
      <n-form-item path="pwd">
        <n-input
          v-model:value="formValue.pwd"
          type="password"
          :placeholder="$t('login.passwordPlaceholder')"
          clearable
          show-password-on="click"
          :input-props="{ autocomplete: 'new-password' }"
        >
          <template #password-invisible-icon>
            <icon-park-outline-preview-close-one />
          </template>
          <template #password-visible-icon>
            <icon-park-outline-preview-open />
          </template>
        </n-input>
      </n-form-item>
      <n-form-item path="rePwd">
        <n-input
          v-model:value="formValue.rePwd"
          type="password"
          :placeholder="$t('login.checkPasswordPlaceholder')"
          clearable
          show-password-on="click"
          :input-props="{ autocomplete: 'new-password' }"
        >
          <template #password-invisible-icon>
            <icon-park-outline-preview-close-one />
          </template>
          <template #password-visible-icon>
            <icon-park-outline-preview-open />
          </template>
        </n-input>
      </n-form-item>
      <n-form-item>
        <n-space
          vertical
          :size="20"
          class="w-full"
        >
          <n-checkbox v-model:checked="isRead">
            {{ $t('login.readAndAgree') }} <n-button
              type="primary"
              text
              @click="openAgreement"
            >
              {{ $t('login.userAgreement') }}
            </n-button>
          </n-checkbox>
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
            @click="handleRegister"
          >
            {{ $t('login.signUp') }}
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
