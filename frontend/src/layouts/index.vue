<script setup lang="ts">
import { useAppStore, useRouteStore } from '@/store'
import {
  BackTop,
  Breadcrumb,
  CollapaseButton,
  FullScreen,
  Logo,
  MobileDrawer,
  Notices,
  Search,
  Setting,
  SettingDrawer,
  TabBar,
  UserCenter,
} from './components'
import Content from './Content.vue'
import { ProLayout, useLayoutMenu } from 'pro-naive-ui'
import { getRuntimeRouteMode } from '@/router/runtime-mode'

const router = useRouter()
const isUserMode = getRuntimeRouteMode() === 'user'

const route = useRoute()
const appStore = useAppStore()
const routeStore = useRouteStore()

const { layoutMode } = storeToRefs(useAppStore())

const {
  layout,
  activeKey,
} = useLayoutMenu({
  mode: layoutMode,
  accordion: true,
  menus: computed(() => routeStore.currentMenus), // 使用 computed 确保异步加载后能触发更新
} as any)

watch(() => route.path, () => {
  activeKey.value = routeStore.activeMenu
}, { immediate: true })

// 移动端抽屉控制
const showMobileDrawer = ref(false)

const sidebarWidth = ref(240)
const sidebarCollapsedWidth = ref(64)

const hasHorizontalMenu = computed(() => ['horizontal', 'mixed-two-column', 'mixed-sidebar'].includes(layoutMode.value))

const hidenCollapaseButton = computed(() => ['horizontal'].includes(layoutMode.value) || appStore.isMobile)
</script>

<template>
  <SettingDrawer />
  <ProLayout
    v-model:collapsed="appStore.collapsed"
    :mode="layoutMode"
    :is-mobile="appStore.isMobile"
    :show-logo="appStore.showLogo && !appStore.isMobile"
    :show-footer="appStore.showFooter"
    :show-tabbar="appStore.showTabs"
    nav-fixed
    show-nav
    show-sidebar
    :nav-height="60"
    :tabbar-height="45"
    :footer-height="40"
    :sidebar-width="sidebarWidth"
    :sidebar-collapsed-width="sidebarCollapsedWidth"
  >
    <template #logo>
      <Logo />
    </template>

    <template #nav-left>
      <template v-if="appStore.isMobile">
        <Logo />
      </template>

      <template v-else>
        <div v-if="!hasHorizontalMenu || !hidenCollapaseButton" class="h-full flex-y-center gap-1 p-x-sm">
          <CollapaseButton v-if="!hidenCollapaseButton" />
          <Breadcrumb v-if="!hasHorizontalMenu" />
        </div>
      </template>
    </template>

    <template #nav-center>
      <div class="h-full flex-y-center gap-1">
        <n-menu v-if="hasHorizontalMenu" v-bind="layout.horizontalMenuProps" />
      </div>
    </template>

    <template #nav-right>
      <div class="h-full flex-y-center gap-1 p-x-xl">
        <!-- 移动端：只显示菜单按钮 -->
        <template v-if="appStore.isMobile">
          <n-button
            quaternary
            @click="showMobileDrawer = true"
          >
            <template #icon>
              <n-icon size="18">
                <icon-park-outline-hamburger-button />
              </n-icon>
            </template>
          </n-button>
        </template>

        <!-- 桌面端：显示完整功能组件 -->
        <template v-else>
          <Search />
          <Notices />
          <FullScreen />
          <DarkModeSwitch />
          <LangsSwitch />
          <Setting />
          <UserCenter />
        </template>
      </div>
    </template>

    <template #sidebar>
      <div class="sidebar-wrapper">
        <n-menu v-bind="layout.verticalMenuProps" :collapsed-width="sidebarCollapsedWidth" />
        <div v-if="isUserMode" class="sidebar-about" :class="{ collapsed: appStore.collapsed }" @click="router.push('/user/about')">
          <icon-park-outline-info class="about-icon" />
          <span v-if="!appStore.collapsed" class="about-text">{{ $t('route.about') }}</span>
        </div>
      </div>
    </template>

    <template #sidebar-extra>
      <n-scrollbar class="flex-[1_0_0]">
        <n-menu v-bind="layout.verticalExtraMenuProps" :collapsed-width="sidebarCollapsedWidth" />
      </n-scrollbar>
    </template>

    <template #tabbar>
      <TabBar />
    </template>

    <template #footer>
      <div class="flex-center h-full">
        {{ appStore.footerText }}
      </div>
    </template>
    <Content />
    <BackTop />
    <SettingDrawer />

    <!-- 移动端功能抽屉 -->
    <MobileDrawer v-model:show="showMobileDrawer">
      <n-menu v-bind="layout.verticalMenuProps" />
    </MobileDrawer>
  </ProLayout>
</template>

<style scoped>
.sidebar-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sidebar-wrapper :deep(.n-menu) {
  flex: 1;
}

.sidebar-about {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  margin: 4px 12px 12px;
  cursor: pointer;
  border-radius: 6px;
  color: var(--n-text-color-3, rgba(194, 194, 194, 0.6));
  font-size: 12px;
  transition: color 0.2s, background-color 0.2s;
}

.sidebar-about:hover {
  color: var(--n-text-color-2, rgba(194, 194, 194, 0.9));
  background-color: var(--n-color-hover, rgba(255, 255, 255, 0.06));
}

.sidebar-about.collapsed {
  justify-content: center;
  margin: 4px 8px 12px;
  padding: 8px;
}

.about-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.about-text {
  white-space: nowrap;
  overflow: hidden;
}
</style>
