<script setup lang="ts">
import IconLogout from '~icons/icon-park-outline/logout'
import IconUser from '~icons/icon-park-outline/user'
import IconSetting from '~icons/icon-park-outline/setting'
import IconDashboard from '~icons/icon-park-outline/dashboard'
import IconWallet from '~icons/icon-park-outline/wallet'
import IconDiamond from '~icons/icon-park-outline/diamond'
import IconLevel from '~icons/icon-park-outline/level'
import IconCrown from '~icons/icon-park-outline/crown'
import IconPhone from '~icons/icon-park-outline/phone'
import IconMail from '~icons/icon-park-outline/mail'
import IconIdCard from '~icons/icon-park-outline/id-card'
import IconTeam from '~icons/icon-park-outline/people'

type QuickAction = 'userCenter' | 'settings' | 'dashboard' | 'logout'

// userInfo 可能是 Api.Login.Info（role 为数组）或 Entity.User（role 为字符串），这里做兼容
interface UserInfoLike extends Omit<Entity.User, 'role'> {
  role?: string | string[]
}

const props = defineProps<{
  userInfo?: UserInfoLike | null
}>()

const emit = defineEmits<{
  (e: 'userCenter'): void
  (e: 'settings'): void
  (e: 'dashboard'): void
  (e: 'logout'): void
}>()

const { t } = useI18n()

const displayName = computed(() => props.userInfo?.nickname || props.userInfo?.userName || t('userCenter.user'))
const displayAccount = computed(() => props.userInfo?.userName || '')

const quickActions = computed(() => [
  { key: 'userCenter' as QuickAction, icon: IconUser, label: t('app.userCenter') },
  { key: 'settings' as QuickAction, icon: IconSetting, label: t('app.setting') },
  { key: 'dashboard' as QuickAction, icon: IconDashboard, label: t('userCenter.dashboard') },
  { key: 'logout' as QuickAction, icon: IconLogout, label: t('app.loginOut') },
])

const roleLabel = computed(() => {
  const role = props.userInfo?.role
  if (Array.isArray(role) && role.length > 0) {
    return role.includes('admin') ? t('userCenter.admin') : t('userCenter.normalUser')
  }
  if (role === 'admin')
    return t('userCenter.admin')
  return t('userCenter.normalUser')
})

const statusLabel = computed(() =>
  props.userInfo?.status === 1 ? t('userCenter.normal') : t('userCenter.disabled'),
)

const statusType = computed<'success' | 'error'>(() =>
  props.userInfo?.status === 1 ? 'success' : 'error',
)

function handleQuickClick(key: QuickAction) {
  switch (key) {
    case 'userCenter':
      emit('userCenter')
      break
    case 'settings':
      emit('settings')
      break
    case 'dashboard':
      emit('dashboard')
      break
    case 'logout':
      emit('logout')
      break
  }
}
</script>

<template>
  <div class="user-center-card w-340px max-w-90vw overflow-hidden rounded-2xl">
    <!-- 头部：头像 + 名称 -->
    <div class="relative px-20px pt-24px pb-16px text-center">
      <div class="avatar-ring mx-auto mb-12px inline-block rounded-full p-2px">
        <n-avatar
          round
          :size="72"
          :src="userInfo?.avatar"
          class="user-center-avatar"
        >
          <template #fallback>
            <div class="wh-full flex-center">
              <icon-park-outline-user />
            </div>
          </template>
        </n-avatar>
      </div>
      <div class="text-16px font-bold leading-tight">
        {{ displayName }}
      </div>
      <div v-if="displayAccount" class="mt-4px text-13px opacity-60">
        @{{ displayAccount }}
      </div>
      <n-space class="mt-12px" justify="center" size="small" :wrap="false">
        <n-tag v-if="userInfo?.id" type="info" size="small" round>
          ID: {{ userInfo.id }}
        </n-tag>
        <n-tag type="primary" size="small" round>
          {{ roleLabel }}
        </n-tag>
        <n-tag :type="statusType" size="small" round>
          {{ statusLabel }}
        </n-tag>
      </n-space>
    </div>

    <!-- 联系信息 -->
    <div class="px-16px pb-16px">
      <div class="info-grid grid grid-cols-1 gap-8px rounded-xl p-12px text-13px">
        <div v-if="userInfo?.mobile" class="flex items-center gap-8px opacity-80">
          <IconPhone class="text-14px" />
          <span>{{ userInfo.mobile }}</span>
        </div>
        <div v-if="userInfo?.email" class="flex items-center gap-8px opacity-80">
          <IconMail class="text-14px" />
          <span class="truncate">{{ userInfo.email }}</span>
        </div>
        <div v-if="userInfo?.id" class="flex items-center gap-8px opacity-80">
          <IconIdCard class="text-14px" />
          <span>ID: {{ userInfo.id }}</span>
        </div>
      </div>
    </div>

    <!-- 团队中心 -->
    <div class="px-16px pb-16px">
      <div class="team-card rounded-xl p-14px">
        <div class="mb-10px flex items-center justify-between">
          <div class="flex items-center gap-8px text-14px font-semibold">
            <IconTeam />
            {{ t('userCenter.teamCenter') }}
          </div>
          <n-tag v-if="userInfo?.groupId" type="success" size="tiny" round>
            TEAM
          </n-tag>
          <n-tag v-else type="default" size="tiny" round>
            TEAM
          </n-tag>
        </div>
        <div v-if="userInfo?.groupId" class="text-13px opacity-90">
          <span class="font-medium">{{ t('userCenter.teamId') }}:</span> {{ userInfo.groupId }}
        </div>
        <div v-else class="text-13px opacity-70">
          {{ t('userCenter.teamNotOpen') }}
        </div>
      </div>
    </div>

    <!-- 余额 / 积分 / 等级 -->
    <div class="px-16px pb-16px">
      <n-grid :x-gap="8" :y-gap="8" :cols="2" responsive="screen">
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-xl p-12px text-center">
            <div class="mb-4px flex items-center gap-4px text-12px opacity-70">
              <IconWallet class="text-14px" />
              {{ t('userCenter.balance') }}
            </div>
            <div class="text-16px font-bold text-[var(--primary-color)]">
              ¥{{ Number(userInfo?.money || 0).toFixed(2) }}
            </div>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-xl p-12px text-center">
            <div class="mb-4px flex items-center gap-4px text-12px opacity-70">
              <IconDiamond class="text-14px" />
              {{ t('userCenter.score') }}
            </div>
            <div class="text-16px font-bold text-[var(--warning-color)]">
              {{ userInfo?.score || 0 }}
            </div>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-xl p-12px text-center">
            <div class="mb-4px flex items-center gap-4px text-12px opacity-70">
              <IconLevel class="text-14px" />
              {{ t('userCenter.level') }}
            </div>
            <div class="text-16px font-bold text-[var(--info-color)]">
              Lv.{{ userInfo?.level || 0 }}
            </div>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-xl p-12px text-center">
            <div class="mb-4px flex items-center gap-4px text-12px opacity-70">
              <IconCrown class="text-14px" />
              {{ t('userCenter.role') }}
            </div>
            <div class="text-14px font-semibold">
              {{ roleLabel }}
            </div>
          </div>
        </n-grid-item>
      </n-grid>
    </div>

    <!-- 快捷入口 -->
    <div class="px-16px pb-16px">
      <div class="quick-card rounded-xl p-12px">
        <div class="mb-10px text-13px font-semibold opacity-80">
          {{ t('userCenter.quickAccess') }}
        </div>
        <n-grid :x-gap="4" :y-gap="8" :cols="4">
          <n-grid-item v-for="item in quickActions" :key="item.key">
            <n-button
              text
              class="quick-btn h-auto w-full py-8px"
              @click="handleQuickClick(item.key)"
            >
              <template #icon>
                <component :is="item.icon" class="quick-icon mb-4px text-20px" />
              </template>
              <span class="text-12px leading-tight">{{ item.label }}</span>
            </n-button>
          </n-grid-item>
        </n-grid>
      </div>
    </div>

    <!-- 退出登录 -->
    <div class="px-16px pb-20px">
      <n-button
        type="error"
        ghost
        block
        round
        @click="emit('logout')"
      >
        <template #icon>
          <IconLogout />
        </template>
        {{ t('app.loginOut') }}
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.user-center-card {
  background: var(--card-color);
  color: var(--text-color-base);
  box-shadow: var(--box-shadow-1);
}

.avatar-ring {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-suppl) 100%);
}

.user-center-avatar {
  border: 2px solid var(--card-color);
}

.info-grid,
.team-card,
.stat-card,
.quick-card {
  background: var(--body-color);
  transition: all 0.2s ease;
}

.stat-card:hover,
.quick-card .quick-btn:hover {
  background: var(--hover-color);
}

.quick-btn {
  border-radius: 8px;
}

.quick-btn:hover .quick-icon {
  color: var(--primary-color);
}

.quick-btn :deep(.n-button__content) {
  flex-direction: column;
  align-items: center;
}

.quick-btn :deep(.n-button__icon) {
  margin-right: 0;
  margin-bottom: 4px;
}
</style>
