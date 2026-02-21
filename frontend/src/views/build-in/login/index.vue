<script setup lang="ts">
import { Login, Register, ResetPwd, ResetPwdConfirm } from './components'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()
type IformType = 'login' | 'register' | 'resetPwd' | 'resetPwdConfirm'
const formType = ref<IformType>('login')
const formComponets = {
  login: Login,
  register: Register,
  resetPwd: ResetPwd,
  resetPwdConfirm: ResetPwdConfirm,
}

onMounted(() => {
  if (route.query.email && route.query.token) {
    formType.value = 'resetPwdConfirm'
  }
})
</script>

<template>
  <n-el class="wh-full flex-center" style="background-color: var(--body-color);">
    <div class="fixed top-40px right-40px text-lg">
      <DarkModeSwitch />
      <LangsSwitch />
    </div>
    <div
      class="p-4xl h-full w-full sm:w-450px sm:h-unset"
      style="background: var(--card-color);box-shadow: var(--box-shadow-1);"
    >
      <div class="w-full flex flex-col items-center">
        <SvgIconsLogo class="text-6em" />
        <n-h3>{{ t('app.title') }} </n-h3>
        <transition
          name="fade-slide"
          mode="out-in"
        >
          <component
            :is="formComponets[formType]"
            v-model="formType"
            class="w-85%"
          />
        </transition>
      </div>
    </div>

    <div />
  </n-el>
</template>
