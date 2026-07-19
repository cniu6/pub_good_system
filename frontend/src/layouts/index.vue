<script setup lang="ts">
import { useAppStore, useAuthStore, useRouteStore } from '@/store'
import {
  BackTop,
  Breadcrumb,
  CollapaseButton,
  CollapsedFlyoutMenu,
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

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const routeStore = useRouteStore()

const { layoutMode } = storeToRefs(useAppStore())
const currentMenus = computed(() => routeStore.menuMode === 'admin' ? routeStore.adminMenus : routeStore.menus)

const {
  layout,
  activeKey,
} = useLayoutMenu({
  mode: layoutMode,
  accordion: true,
  menus: currentMenus,
} as any)

watch(() => route.path, () => {
  activeKey.value = routeStore.activeMenu
}, { immediate: true })

// 移动端抽屉控制
const showMobileDrawer = ref(false)

// 侧边栏宽度：以 store 当前值（含 localStorage 持久化）作为默认参数
const sidebarWidth = computed({
  get: () => appStore.sidebarWidth || 240,
  set: (v: number) => appStore.setSidebarWidth(v),
})
const sidebarCollapsedWidth = 64

const hasHorizontalMenu = computed(() => ['horizontal', 'mixed-two-column', 'mixed-sidebar'].includes(layoutMode.value))

const hidenCollapaseButton = computed(() => ['horizontal'].includes(layoutMode.value) || appStore.isMobile)

const showSidebarResize = computed(() => !appStore.isMobile && !appStore.collapsed && !['horizontal'].includes(layoutMode.value))

let resizing = false
let startX = 0
let startWidth = 240

function onSidebarResizeMove(e: MouseEvent) {
  if (!resizing)
    return
  const delta = e.clientX - startX
  appStore.setSidebarWidth(startWidth + delta)
}

function onSidebarResizeEnd() {
  if (!resizing)
    return
  resizing = false
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  window.removeEventListener('mousemove', onSidebarResizeMove)
  window.removeEventListener('mouseup', onSidebarResizeEnd)
}

function onSidebarResizeStart(e: MouseEvent) {
  if (!showSidebarResize.value)
    return
  e.preventDefault()
  resizing = true
  startX = e.clientX
  startWidth = sidebarWidth.value
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
  window.addEventListener('mousemove', onSidebarResizeMove)
  window.addEventListener('mouseup', onSidebarResizeEnd)
}

onBeforeUnmount(() => {
  onSidebarResizeEnd()
})

// 页面刷新后恢复已登录状态时，也要重新建立 Presence 连接 + 重挂自动刷新 token 定时器
// （否则刷新页面后旧的刷新定时器丢失，access token 到期就只能等 401 时被动刷新）。
onMounted(() => {
  authStore.startPresence()
  authStore.setupAutoRefresh()
})
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
        <!-- 折叠：根级互斥飞出（NDropdown）；展开：原 n-menu -->
        <CollapsedFlyoutMenu
          v-if="appStore.collapsed"
          :menus="currentMenus"
          :active-key="activeKey"
          :collapsed-width="sidebarCollapsedWidth"
        />
        <n-menu
          v-else
          v-bind="layout.verticalMenuProps"
          :collapsed-width="sidebarCollapsedWidth"
        />
        <!-- 右侧拖拽条：动态调整宽度并持久化 -->
        <div
          v-if="showSidebarResize"
          class="sidebar-resize-handle"
          title="拖动调整侧边栏宽度"
          @mousedown="onSidebarResizeStart"
        />
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
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.sidebar-wrapper :deep(.n-menu) {
  flex: 1;
}

.sidebar-resize-handle {
  position: absolute;
  top: 0;
  right: -3px;
  z-index: 20;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  touch-action: none;
}

.sidebar-resize-handle:hover,
.sidebar-resize-handle:active {
  background: color-mix(in srgb, var(--n-primary-color, #18a058) 35%, transparent);
}
</style>
