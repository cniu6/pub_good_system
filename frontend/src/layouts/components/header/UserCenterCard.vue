<script setup lang="ts">
import IconUser from '~icons/icon-park-outline/user'
import IconSetting from '~icons/icon-park-outline/setting'
import IconDashboard from '~icons/icon-park-outline/dashboard'
import IconWallet from '~icons/icon-park-outline/wallet'
import IconDiamond from '~icons/icon-park-outline/diamond'
import IconLevel from '~icons/icon-park-outline/level'
import IconSwitch from '~icons/icon-park-outline/switch'
import IconCopy from '~icons/icon-park-outline/copy'
import IconPeople from '~icons/icon-park-outline/people'

type QuickAction = 'userCenter' | 'settings' | 'dashboard'

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
const message = useMessage()

const displayName = computed(() => props.userInfo?.nickname || props.userInfo?.userName || t('userCenter.user'))
const displayAccount = computed(() => props.userInfo?.userName || '')

const quickActions = computed(() => [
  { key: 'userCenter' as QuickAction, icon: IconUser, label: t('app.userCenter') },
  { key: 'settings' as QuickAction, icon: IconSetting, label: t('app.setting') },
  { key: 'dashboard' as QuickAction, icon: IconDashboard, label: t('userCenter.dashboard') },
])

const maskedId = computed(() => {
  const id = props.userInfo?.id
  if (!id)
    return ''
  const s = String(id)
  if (s.length <= 1)
    return s
  return `${s[0]}***`
})

const switchOptions = computed(() => [
  { label: t('userCenter.personal'), key: 'personal', disabled: false },
  { label: t('userCenter.noTeamNow'), key: 'team', disabled: true },
])

function copyId() {
  if (!props.userInfo?.id)
    return
  navigator.clipboard?.writeText(String(props.userInfo.id)).then(() => {
    message.success(t('userCenter.idCopied'))
  }).catch(() => {
    message.error(t('userCenter.idCopyFailed'))
  })
}

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
  }
}
</script>

<template>
  <div class="user-center-card w-320px max-w-90vw overflow-hidden rounded-2xl">
    <!-- 头部：左对齐头像 + 名称/ID，右侧空区放团队/个人切换器 -->
    <div class="relative px-16px pt-16px pb-12px">
      <div class="flex items-start gap-12px">
        <div class="avatar-ring flex-shrink-0 rounded-full p-2px">
          <n-avatar
            round
            :size="48"
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

        <div class="min-w-0 flex-1 pt-2px">
          <div class="nickname text-15px font-bold truncate">
            {{ displayName }}
          </div>
          <div v-if="displayAccount" class="mt-2px text-12px opacity-60 truncate">
            @{{ displayAccount }}
          </div>
        </div>

        <n-popover trigger="click" class="!p-0" style="width: auto" placement="bottom-end" arrow-point-to-center>
          <template #trigger>
            <div class="switcher-trigger flex flex-col items-center justify-center rounded-lg p-8px min-w-80px cursor-pointer" :title="t('userCenter.switchTeam')">
              <div class="flex items-center gap-4px">
                <IconPeople class="text-14px" />
                <span class="text-13px font-bold">{{ t('userCenter.personal') }}</span>
              </div>
              <div class="mt-4px flex items-center gap-4px text-11px opacity-70">
                <IconSwitch class="text-11px" />
                <span class="whitespace-nowrap">{{ t('userCenter.switchTeam') }}</span>
              </div>
            </div>
          </template>
          <div class="team-switch-popover w-160px py-6px">
            <div
              v-for="item in switchOptions"
              :key="item.key"
              class="team-switch-item flex items-center justify-between px-12px py-8px text-13px"
              :class="{ 'is-disabled': item.disabled }"
            >
              <span class="flex items-center gap-6px">
                <IconPeople class="text-14px" />
                {{ item.label }}
              </span>
              <span v-if="item.key === 'personal'" class="status-dot rounded-full" />
            </div>
          </div>
        </n-popover>
      </div>

      <div v-if="userInfo?.id" class="mt-12px flex justify-center">
        <n-tooltip trigger="hover">
          <template #trigger>
            <div class="id-row inline-flex cursor-pointer items-center gap-6px rounded-lg px-12px py-6px text-12px" @click="copyId">
              <span class="opacity-70">ID:</span>
              <span class="id-value font-medium opacity-90">{{ maskedId }}</span>
              <IconCopy class="text-12px opacity-70" />
            </div>
          </template>
          {{ t('userCenter.copyId') }}
        </n-tooltip>
      </div>
    </div>

    <!-- 余额 / 积分 / 等级 一排三个 -->
    <div class="px-16px pb-16px">
      <n-grid :x-gap="6" :y-gap="6" :cols="3" responsive="screen">
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-lg p-10px text-center">
            <div class="mb-2px flex items-center gap-4px text-11px opacity-70">
              <IconWallet class="text-13px" />
              {{ t('userCenter.balance') }}
            </div>
            <div class="text-14px font-bold text-[var(--primary-color)]">
              ¥{{ Number(userInfo?.money || 0).toFixed(2) }}
            </div>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-lg p-10px text-center">
            <div class="mb-2px flex items-center gap-4px text-11px opacity-70">
              <IconDiamond class="text-13px" />
              {{ t('userCenter.score') }}
            </div>
            <div class="text-14px font-bold text-[var(--warning-color)]">
              {{ userInfo?.score || 0 }}
            </div>
          </div>
        </n-grid-item>
        <n-grid-item>
          <div class="stat-card flex flex-col items-center justify-center rounded-lg p-10px text-center">
            <div class="mb-2px flex items-center gap-4px text-11px opacity-70">
              <IconLevel class="text-13px" />
              {{ t('userCenter.level') }}
            </div>
            <div class="text-14px font-bold text-[var(--info-color)]">
              Lv.{{ userInfo?.level || 0 }}
            </div>
          </div>
        </n-grid-item>
      </n-grid>
    </div>

    <!-- 快捷入口 -->
    <div class="px-16px pb-12px">
      <div class="quick-card rounded-xl p-12px">
        <n-grid :x-gap="4" :y-gap="8" :cols="3">
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
    <div class="px-16px pb-16px">
      <n-button
        type="error"
        ghost
        block
        round
        @click="emit('logout')"
      >
        <template #icon>
          <icon-park-outline-logout />
        </template>
        {{ t('app.loginOut') }}
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.user-center-card {
  background: var(--body-color);
  color: var(--text-color-base);
}

.avatar-ring {
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-color-suppl) 100%);
}

.user-center-avatar {
  border: 2px solid var(--body-color);
}

.nickname {
  color: var(--primary-color);
}

.id-value {
  font-family: ui-monospace, 'Cascadia Mono', 'Segoe UI Mono', monospace;
}

.id-row {
  background: var(--card-color);
  transition: all 0.2s ease;
}

.id-row:hover {
  background: var(--hover-color);
}

.switcher-trigger {
  background: var(--card-color);
  transition: all 0.2s ease;
}

.switcher-trigger:hover {
  background: var(--hover-color);
}

.team-switch-popover {
  background: var(--card-color);
  border-radius: 8px;
}

.team-switch-item {
  cursor: default;
  transition: background 0.2s;
}

.team-switch-item:not(.is-disabled):hover {
  background: var(--hover-color);
}

.team-switch-item.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.status-dot {
  width: 6px;
  height: 6px;
  background: var(--primary-color);
}

.stat-card,
.quick-card {
  background: var(--card-color);
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

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
