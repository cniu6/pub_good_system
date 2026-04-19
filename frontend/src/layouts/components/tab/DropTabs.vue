<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import type { TabRouteItem } from '@/store/tab'
import { useTabStore } from '@/store'
import { renderIcon, resolveRouteTitle } from '@/utils'

const tabStore = useTabStore()
const { t } = useI18n()

function isTabRouteItem(option: MenuOption): option is MenuOption & TabRouteItem {
  return typeof option.key === 'string'
    && 'fullPath' in option
    && 'path' in option
    && 'meta' in option
}

function renderDropTabsLabel(option: MenuOption) {
  if (!isTabRouteItem(option))
    return ''
  return resolveRouteTitle(t, option.name, option.meta.title)
}

function renderDropTabsIcon(option: MenuOption) {
  if (!isTabRouteItem(option) || !option.meta.icon)
    return null
  return renderIcon(option.meta.icon)!()
}

const router = useRouter()
function handleDropTabs(key: string, option: MenuOption) {
  if (!isTabRouteItem(option))
    return
  router.push(option.fullPath)
}
</script>

<template>
  <n-dropdown
    :options="tabStore.allTabs"
    :render-label="renderDropTabsLabel"
    :render-icon="renderDropTabsIcon"
    trigger="click"
    size="small"
    key-field="fullPath"
    @select="handleDropTabs"
  >
    <CommonWrapper>
      <icon-park-outline-application-menu />
    </CommonWrapper>
  </n-dropdown>
</template>

<style scoped>

</style>
