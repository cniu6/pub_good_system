<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store'
import { fetchDashboard } from '@/service'
import Chart from './components/chart.vue'

const authStore = useAuthStore()
const { t } = useI18n()
const userInfo = computed(() => authStore.userInfo)

const loading = ref(false)
const dashboardData = ref<any>(null)

const stats = computed(() => dashboardData.value?.stats || {})
const announcements = computed(() => dashboardData.value?.announcements || [])

async function loadDashboard() {
  loading.value = true
  try {
    const response = await fetchDashboard()
    if (response.isSuccess && response.data) {
      dashboardData.value = response.data
    }
  }
  catch (error) {
    if (import.meta.env.DEV)
      console.error('[workbench] fetch dashboard failed', error)
  }
  finally {
    loading.value = false
  }
}

const announcementTagType: Record<string, 'info' | 'success' | 'warning' | 'error'> = {
  info: 'info',
  success: 'success',
  warning: 'warning',
  error: 'error',
}

const announcementTagLabel: Record<string, string> = {
  info: t('workbench.announcementInfo'),
  success: t('workbench.announcementSuccess'),
  warning: t('workbench.announcementWarning'),
  error: t('workbench.announcementError'),
}

const router = useRouter()
function goToUserCenter() {
  router.push('/user/account/user-center')
}

onMounted(() => {
  loadDashboard()
})
</script>

<template>
  <n-spin :show="loading">
    <n-grid
      :x-gap="16"
      :y-gap="16"
      :cols="3"
      item-responsive
      responsive="screen"
    >
      <!-- 左侧主要内容区 -->
      <n-gi span="3 m:2">
        <n-space vertical :size="16">
          <!-- 图表区域 -->
          <n-card style="--n-padding-left: 0;">
            <Chart />
          </n-card>

          <!-- 统计卡片区域 -->
          <n-grid
            :x-gap="16"
            :y-gap="16"
            :cols="4"
            item-responsive
            responsive="screen"
          >
            <n-gi span="2 l:1">
              <n-card>
                <n-thing>
                  <template #avatar>
                    <n-el>
                      <n-icon-wrapper :size="46" color="var(--success-color)" :border-radius="999">
                        <nova-icon :size="26" icon="icon-park-outline:finance" />
                      </n-icon-wrapper>
                    </n-el>
                  </template>
                  <template #header>
                    <n-statistic :label="t('workbench.balance')">
                      <template #prefix>
                        ¥
                      </template>
                      <n-number-animation show-separator :from="0" :to="Number(stats.money) || 0" :precision="2" />
                    </n-statistic>
                  </template>
                </n-thing>
              </n-card>
            </n-gi>
            <n-gi span="2 l:1">
              <n-card>
                <n-thing>
                  <template #avatar>
                    <n-el>
                      <n-icon-wrapper :size="46" color="var(--success-color)" :border-radius="999">
                        <nova-icon :size="26" icon="icon-park-outline:star" />
                      </n-icon-wrapper>
                    </n-el>
                  </template>
                  <template #header>
                    <n-statistic :label="t('workbench.score')">
                      <n-number-animation show-separator :from="0" :to="stats.score || 0" />
                    </n-statistic>
                  </template>
                </n-thing>
              </n-card>
            </n-gi>
            <n-gi span="2 l:1">
              <n-card>
                <n-thing>
                  <template #avatar>
                    <n-el>
                      <n-icon-wrapper :size="46" color="var(--success-color)" :border-radius="999">
                        <nova-icon :size="26" icon="icon-park-outline:user" />
                      </n-icon-wrapper>
                    </n-el>
                  </template>
                  <template #header>
                    <n-statistic :label="t('workbench.loginCount')">
                      <n-number-animation show-separator :from="0" :to="stats.loginCount || 0" />
                      <template #suffix>
                        {{ t('workbench.times') }}
                      </template>
                    </n-statistic>
                  </template>
                </n-thing>
              </n-card>
            </n-gi>
            <n-gi span="2 l:1">
              <n-card>
                <n-thing>
                  <template #avatar>
                    <n-el>
                      <n-icon-wrapper :size="46" color="var(--success-color)" :border-radius="999">
                        <nova-icon :size="26" icon="icon-park-outline:time" />
                      </n-icon-wrapper>
                    </n-el>
                  </template>
                  <template #header>
                    <n-statistic :label="t('workbench.daysJoined')">
                      <n-number-animation :from="0" :to="stats.daysJoined || 0" />
                      <template #suffix>
                        {{ t('workbench.days') }}
                      </template>
                    </n-statistic>
                  </template>
                </n-thing>
              </n-card>
            </n-gi>
          </n-grid>

          <!-- 快捷操作 -->
          <n-card :title="t('workbench.quickActions')">
            <n-space>
              <n-button type="primary" @click="goToUserCenter">
                <template #icon>
                  <nova-icon icon="icon-park-outline:edit" />
                </template>
                {{ t('workbench.editProfile') }}
              </n-button>
              <n-button @click="router.push('/user/account/user-center')">
                <template #icon>
                  <nova-icon icon="icon-park-outline:setting-one" />
                </template>
                {{ t('workbench.accountSettings') }}
              </n-button>
            </n-space>
          </n-card>
        </n-space>
      </n-gi>

      <!-- 右侧边栏 -->
      <n-gi span="3 m:1">
        <n-space vertical :size="16">
          <!-- 用户欢迎卡片 -->
          <n-card>
            <n-flex align="center" :size="16">
              <n-avatar
                round
                :size="56"
                :src="userInfo?.avatar"
                :img-props="{ referrerpolicy: 'no-referrer' }"
              />
              <div>
                <n-h4 style="margin: 0;">
                  {{ userInfo?.nickname || userInfo?.userName || t('workbench.user') }}，{{ t('workbench.welcomeBack') }}
                </n-h4>
                <n-text depth="3">
                  {{ userInfo?.role?.includes('admin') ? t('workbench.admin') : t('workbench.normalUser') }} · {{ t('workbench.level') }} {{ stats.level || 0 }}
                </n-text>
              </div>
            </n-flex>
          </n-card>

          <!-- 公告 -->
          <n-card :title="t('workbench.announcements')">
            <n-list>
              <n-list-item v-for="item in announcements" :key="item.id">
                <template #prefix>
                  <n-tag
                    :bordered="false"
                    :type="announcementTagType[item.type] || 'info'"
                    size="small"
                  >
                    {{ announcementTagLabel[item.type] || t('workbench.announcementDefault') }}
                  </n-tag>
                </template>
                <n-tooltip trigger="hover">
                  <template #trigger>
                    <n-button text>
                      {{ item.title }}
                    </n-button>
                  </template>
                  {{ item.content }}
                </n-tooltip>
              </n-list-item>
              <n-empty v-if="announcements.length === 0" :description="t('workbench.noAnnouncements')" />
            </n-list>
          </n-card>

          <!-- 账户概览 -->
          <n-grid :x-gap="16" :y-gap="16" :cols="2">
            <n-gi :span="1">
              <n-card>
                <n-flex vertical align="center">
                  <n-text depth="3">
                    {{ t('workbench.level') }}
                  </n-text>
                  <n-icon-wrapper :size="46" :border-radius="999">
                    <nova-icon :size="26" icon="icon-park-outline:level" />
                  </n-icon-wrapper>
                  <n-text strong class="text-2xl">
                    Lv.{{ stats.level || 0 }}
                  </n-text>
                </n-flex>
              </n-card>
            </n-gi>
            <n-gi :span="1">
              <n-card>
                <n-flex vertical align="center">
                  <n-text depth="3">
                    {{ t('workbench.score') }}
                  </n-text>
                  <n-el>
                    <n-icon-wrapper :size="46" color="var(--warning-color)" :border-radius="999">
                      <nova-icon :size="26" icon="icon-park-outline:star" />
                    </n-icon-wrapper>
                  </n-el>
                  <n-text strong class="text-2xl">
                    {{ stats.score || 0 }}
                  </n-text>
                </n-flex>
              </n-card>
            </n-gi>
          </n-grid>
        </n-space>
      </n-gi>
    </n-grid>
  </n-spin>
</template>

<style scoped></style>
