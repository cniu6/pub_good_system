<script setup lang="ts">
import { useAppStore, useRouteStore } from '@/store'

const appStore = useAppStore()
const routeStore = useRouteStore()
</script>

<template>
  <n-el
    class="h-full"
    :class="[
      appStore.layoutMode === 'full-content' ? 'p-0' : 'p-16px',
    ]"
    style="background-color: var(--action-color);"
  >
    <router-view
      v-slot="{ Component, route }"
    >
      <!--
        关键说明：transition(out-in) 的直接子节点不能是 keep-alive（节点身份不变，out-in 会卡在 leave 后不 enter），
        否则用户端侧边栏切换会出现内容空白、刷新又正常。key 挂在外层可切换节点上，并判空 Component。
      -->
      <transition :name="appStore.transitionAnimation" mode="out-in">
        <div
          v-if="Component && appStore.loadFlag"
          :key="route.fullPath"
          class="h-full"
        >
          <keep-alive :include="routeStore.cacheRoutes">
            <component :is="Component" :key="String(route.name || route.fullPath)" />
          </keep-alive>
        </div>
      </transition>
    </router-view>
  </n-el>
</template>
