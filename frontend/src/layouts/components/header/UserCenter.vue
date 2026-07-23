<script setup lang="ts">
import { getAdminBasePath } from '@/router/constants'
import { useAuthStore } from '@/store'
import IconLogout from '~icons/icon-park-outline/logout'
import IconUser from '~icons/icon-park-outline/user'

const { t } = useI18n()

const { userInfo, logout } = useAuthStore()
const router = useRouter()

/** 是否处于管理端入口（pathname 匹配 VITE_ADMIN_BASE_PATH） */
function isAdminEntry() {
  const adminBase = getAdminBasePath()
  const pathname = window.location.pathname.replace(/\/+$/, '') || '/'
  return pathname === adminBase || pathname.startsWith(`${adminBase}/`)
}

// 头像下拉只保留本站入口（用户中心 / 退出），不再外链上游 Nova/chansee97 仓库与文档
const options = computed(() => {
  return [
    {
      label: t('app.userCenter'),
      key: 'userCenter',
      icon: () => h(IconUser),
    },
    {
      type: 'divider',
      key: 'd1',
    },
    {
      label: t('app.loginOut'),
      key: 'loginOut',
      icon: () => h(IconLogout),
    },
  ]
})
function handleSelect(key: string | number) {
  if (key === 'loginOut') {
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
  if (key === 'userCenter') {
    // 管理端没有用户中心页；跳系统配置（hash 路由内部路径），避免跳到用户端 /user/...
    if (isAdminEntry()) {
      router.push({ name: 'admin-settings-config' }).catch(() => {
        router.push('/settings/config')
      })
      return
    }
    router.push('/user/account/user-center')
  }
}
</script>

<template>
  <n-dropdown
    trigger="click"
    :options="options"
    @select="handleSelect"
  >
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
  </n-dropdown>
</template>

<style scoped></style>
