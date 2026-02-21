<script setup lang="ts">
import type { FormInst } from 'naive-ui'
import { useRoute } from 'vue-router'
import { fetchResetPasswordConfirm } from '@/service'

const emit = defineEmits(['update:modelValue'])
function toLogin() {
  emit('update:modelValue', 'login')
}
const { t } = useI18n()
const route = useRoute()

const formValue = ref({
  pwd: '',
  rePwd: '',
})

const rules = computed(() => {
  return {
    pwd: {
      required: true,
      trigger: 'blur',
      message: t('login.passwordRuleTip'),
    },
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

const email = ref('')
const token = ref('')

onMounted(() => {
  email.value = (route.query.email as string) || ''
  token.value = (route.query.token as string) || ''

  if (!email.value || !token.value) {
    window.$message.error(t('login.invalidResetLink'))
    toLogin()
  }
})

const formRef = ref<FormInst | null>(null)
const isLoading = ref(false)

async function handleConfirm() {
  const hasErrors = await new Promise<boolean>((resolve) => {
    formRef.value?.validate((errors) => {
      resolve(Boolean(errors))
    })
  })

  if (hasErrors)
    return

  isLoading.value = true
  try {
    const { isSuccess } = await fetchResetPasswordConfirm({
      email: email.value,
      code: token.value,
      new_password: formValue.value.pwd,
    })
    if (isSuccess) {
      window.$message.success(t('login.resetSuccess'))
      toLogin()
    }
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div>
    <n-h2 depth="3" class="text-center">
      {{ $t('login.resetPasswordConfirmTitle') }}
    </n-h2>
    <n-form
      ref="formRef"
      :rules="rules"
      :model="formValue"
      :show-label="false"
      size="large"
    >
      <n-form-item path="pwd">
        <n-input
          v-model:value="formValue.pwd"
          type="password"
          :placeholder="$t('login.newPasswordPlaceholder')"
          clearable
          show-password-on="click"
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
          <n-button
            block
            type="primary"
            :loading="isLoading"
            :disabled="isLoading"
            @click="handleConfirm"
          >
            {{ $t('login.confirmReset') }}
          </n-button>
          <n-flex justify="center">
            <n-button
              text
              type="primary"
              @click="toLogin"
            >
              {{ $t('login.backToLogin') }}
            </n-button>
          </n-flex>
        </n-space>
      </n-form-item>
    </n-form>
  </div>
</template>

<style scoped></style>
