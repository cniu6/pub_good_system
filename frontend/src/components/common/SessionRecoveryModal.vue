<script setup lang="ts">
import { Login, Register, ResetPwd } from '@/views/_builtin/login/components'
import { useAuthStore } from '@/store'

type FormType = 'login' | 'register' | 'resetPwd'

const authStore = useAuthStore()
const { t } = useI18n()
const formType = ref<FormType>('login')

watch(() => authStore.needsReauthentication, (visible) => {
  if (visible)
    formType.value = 'login'
})
</script>

<template>
  <n-modal
    :show="authStore.needsReauthentication"
    preset="card"
    class="session-recovery-modal"
    style="width: min(460px, calc(100vw - 32px));"
    :mask-closable="false"
    :close-on-esc="false"
    :closable="false"
    :auto-focus="true"
    :trap-focus="true"
    :bordered="false"
  >
    <div class="mb-20 flex items-center gap-12">
      <SvgIconsLogo class="text-2.5em" />
      <div>
        <n-h3 class="m-0">
          {{ t('login.sessionExpiredTitle') }}
        </n-h3>
        <n-text depth="3">
          {{ t('login.sessionExpiredDescription') }}
        </n-text>
      </div>
    </div>

    <transition name="fade-slide" mode="out-in">
      <Login
        v-if="formType === 'login'"
        v-model="formType"
        :preserve-current-page="true"
      />
      <Register
        v-else-if="formType === 'register'"
        v-model="formType"
      />
      <ResetPwd
        v-else
        v-model="formType"
      />
    </transition>
  </n-modal>
</template>

<style scoped>
.session-recovery-modal :deep(.n-card__content) {
  padding: 28px;
}

@media (max-width: 640px) {
  .session-recovery-modal :deep(.n-card__content) {
    padding: 20px;
  }
}
</style>
