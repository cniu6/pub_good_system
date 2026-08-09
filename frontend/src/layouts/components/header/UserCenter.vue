<script setup lang="ts">
import { ref } from 'vue'
import { getAdminBasePath } from '@/router/constants'
import { useAuthStore } from '@/store'
import UserCenterCard from './UserCenterCard.vue'

const { t } = useI18n()

const { userInfo, logout } = useAuthStore()
const router = useRouter()

const showPopover = ref(false)

/** 是否处于管理端入口（pathname 匹配 VITE_ADMIN_BASE_PATH） */
function isAdminEntry() {
  const adminBase = getAdminBasePath()
  const pathname = window.location.pathname.replace(/\/+$/, '') || '/'
  return pathname === adminBase || pathname.startsWith(`${adminBase}/`)
}

function handleUserCenter() {
  showPopover.value = false
  // 管理端没有用户中心页；跳系统配置（hash 路由内部路径），避免跳到用户端 /user/...
  if (isAdminEntry()) {
    router.push({ name: 'admin-settings-config' }).catch(() => {
      router.push('/settings/config')
    })
    return
  }
  router.push('/user/account/user-center')
}

function handleSettings() {
  showPopover.value = false
  if (isAdminEntry()) {
    router.push({ name: 'admin-settings-config' }).catch(() => {
      router.push('/settings/config')
    })
    return
  }
  router.push('/settings/config')
}

function handleDashboard() {
  showPopover.value = false
  if (isAdminEntry()) {
    router.push({ name: 'admin-dashboard' }).catch(() => {
      router.push('/dashboard')
    })
    return
  }
  router.push('/dashboard')
}

function handleLogout() {
  showPopover.value = false
  window.$dialog?.info({
    title: t('app.loginOutTitle'),
    content: t('app.loginOutContent'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => {
      logout()
    },
  })
}
</script>

<template>
  <n-popover
    v-model:show="showPopover"
    placement="bottom-end"
    trigger="click"
    arrow-point-to-center
    class="!p-0"
    style="width: auto"
  >
    <template #trigger>
      <n-avatar
        round
        class="cursor-pointer"
        :src="userInfo?.avatar"
      >
        <template #fallback>
          <div class="wh-full flex-center">
            <icon-park-outline-user />
          </div>
        </template>
      </n-avatar>
    </template>
    <UserCenterCard
      :user-info="userInfo"
      @user-center="handleUserCenter"
      @settings="handleSettings"
      @dashboard="handleDashboard"
      @logout="handleLogout"
    />
  </n-popover>
</template>

<style scoped></style>
