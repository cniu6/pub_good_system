<script setup lang="ts">
/**
 * 折叠侧栏：每个根项一个 NDropdown（库定位/主题/动画）。
 * 根级互斥：同一时间只开一个根飞出，避免叠层；
 * 当前根内的 2/3 级子菜单仍由 NDropdown children 正常侧向展开。
 */
import type { DropdownOption, MenuOption } from 'naive-ui'
import type { PropType, VNodeChild } from 'vue'
import { computed, defineComponent, h, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NDropdown, NTooltip } from 'naive-ui'

const props = withDefaults(defineProps<{
  menus: Array<Record<string, any>>
  activeKey?: string | number | null
  collapsedWidth?: number
}>(), {
  activeKey: null,
  collapsedWidth: 64,
})

type FlyoutMenuItem = MenuOption & Record<string, any>

const router = useRouter()

/** 当前打开的根菜单 key；null 表示全关。根级互斥，不会叠多个根飞出 */
const openRootKey = ref<string | null>(null)

const VNodeHost = defineComponent({
  name: 'CollapsedFlyoutVNodeHost',
  props: {
    content: { type: null as unknown as PropType<VNodeChild>, default: null },
  },
  setup(p) {
    return () => p.content as any
  },
})

function vnodeToText(raw: unknown): string {
  if (raw == null)
    return ''
  if (typeof raw === 'string' || typeof raw === 'number')
    return String(raw)
  if (Array.isArray(raw))
    return raw.map(vnodeToText).join('')
  if (typeof raw === 'object') {
    const v = raw as { children?: unknown }
    if (typeof v.children === 'string' || typeof v.children === 'number')
      return String(v.children)
    if (typeof v.children === 'function')
      return vnodeToText(v.children())
    if (Array.isArray(v.children))
      return v.children.map(vnodeToText).join('')
    if (v.children && typeof v.children === 'object' && v.children !== null && 'default' in (v.children as object)) {
      const def = (v.children as { default?: () => unknown }).default
      if (typeof def === 'function')
        return vnodeToText(def())
    }
  }
  return ''
}

function menuKey(item: FlyoutMenuItem): string {
  return String(item.key ?? '')
}

function hasChildren(item: FlyoutMenuItem): boolean {
  return Array.isArray(item.children) && item.children.length > 0
}

function resolveLabelText(item: FlyoutMenuItem): string {
  const label = item.label
  if (typeof label === 'string')
    return label
  if (typeof label === 'function') {
    try {
      return vnodeToText(label()) || menuKey(item)
    }
    catch {
      return menuKey(item)
    }
  }
  return vnodeToText(label) || menuKey(item)
}

function resolveIcon(item: FlyoutMenuItem): VNodeChild {
  const icon = item.icon
  if (!icon)
    return null
  return typeof icon === 'function' ? (icon as () => VNodeChild)() : (icon as VNodeChild)
}

function isActiveKey(key: string): boolean {
  const active = props.activeKey == null ? '' : String(props.activeKey)
  if (!active)
    return false
  return active === key || active.startsWith(`${key}/`)
}

function isRootActive(item: FlyoutMenuItem): boolean {
  const key = menuKey(item)
  if (isActiveKey(key))
    return true
  if (!hasChildren(item))
    return false
  return (item.children as FlyoutMenuItem[]).some(child => isActiveKey(menuKey(child)) || isRootActive(child))
}

/**
 * 每一层飞出面板统一结构：
 * 顶部 = 当前层标题（母项名，偏左一点）
 * 下面 = 可点选的子项（含子菜单的继续侧向展开）
 */
function wrapPanelOptions(title: string, titleKey: string, children: DropdownOption[]): DropdownOption[] {
  const header: DropdownOption = {
    key: `__title_${titleKey}`,
    type: 'render',
    render: () =>
      h(
        'div',
        { class: 'collapsed-flyout-menu__panel-title' },
        title,
      ),
  }
  const divider: DropdownOption = {
    type: 'divider',
    key: `__div_${titleKey}`,
  }
  return [header, divider, ...children]
}

function toDropdownOptions(items: FlyoutMenuItem[]): DropdownOption[] {
  return items.map((item) => {
    const key = menuKey(item)
    const label = resolveLabelText(item)
    const option: DropdownOption = {
      key,
      label,
    }
    if (item.icon)
      option.icon = () => resolveIcon(item) as any
    // 有子级：下一层飞出以「当前项」为标题，下面再列它的子项
    if (hasChildren(item)) {
      option.children = wrapPanelOptions(
        label,
        key,
        toDropdownOptions(item.children as FlyoutMenuItem[]),
      )
    }
    return option
  })
}

/** 根飞出：标题用根菜单名（如「财务中心」/「四级菜单测试」） */
function toRootDropdownOptions(root: FlyoutMenuItem): DropdownOption[] {
  return wrapPanelOptions(
    resolveLabelText(root),
    menuKey(root),
    toDropdownOptions((root.children as FlyoutMenuItem[]) || []),
  )
}

const rootMenus = computed(() => (props.menus || []) as FlyoutMenuItem[])

/**
 * 根飞出更新：
 * - show=true → 立刻切到该根（关掉其它根）
 * - show=false → 仅当关掉的是当前根时才清空（duration 到期）
 */
function onRootShowUpdate(key: string, show: boolean) {
  if (show) {
    openRootKey.value = key
    return
  }
  if (openRootKey.value === key)
    openRootKey.value = null
}

async function onSelect(key: string | number) {
  const path = String(key)
  // 忽略标题/分隔占位 key
  if (!path || path.startsWith('__title_') || path.startsWith('__div_'))
    return
  openRootKey.value = null
  try {
    await router.push(path)
  }
  catch {
    // ignore
  }
}

function onLeafClick(item: FlyoutMenuItem) {
  void onSelect(menuKey(item))
}
</script>

<template>
  <div
    class="collapsed-flyout-menu"
    :style="{ width: `${collapsedWidth}px` }"
  >
    <template v-for="item in rootMenus" :key="menuKey(item)">
      <NDropdown
        v-if="hasChildren(item)"
        :show="openRootKey === menuKey(item)"
        :options="toRootDropdownOptions(item)"
        trigger="hover"
        placement="right-start"
        :delay="80"
        :duration="3000"
        :show-arrow="false"
        animated
        @update:show="(show) => onRootShowUpdate(menuKey(item), show)"
        @select="onSelect"
      >
        <div
          class="collapsed-flyout-menu__icon"
          :class="{
            'is-active': isRootActive(item),
            'is-open': openRootKey === menuKey(item),
          }"
          role="button"
          tabindex="0"
        >
          <span v-if="item.icon" class="collapsed-flyout-menu__icon-glyph">
            <VNodeHost :content="resolveIcon(item)" />
          </span>
          <span v-else class="collapsed-flyout-menu__icon-fallback">
            {{ resolveLabelText(item).slice(0, 1) }}
          </span>
        </div>
      </NDropdown>

      <NTooltip
        v-else
        trigger="hover"
        placement="right"
        :delay="80"
        :duration="400"
      >
        <template #trigger>
          <div
            class="collapsed-flyout-menu__icon"
            :class="{ 'is-active': isRootActive(item) }"
            role="button"
            tabindex="0"
            @click="onLeafClick(item)"
          >
            <span v-if="item.icon" class="collapsed-flyout-menu__icon-glyph">
              <VNodeHost :content="resolveIcon(item)" />
            </span>
            <span v-else class="collapsed-flyout-menu__icon-fallback">
              {{ resolveLabelText(item).slice(0, 1) }}
            </span>
          </div>
        </template>
        {{ resolveLabelText(item) }}
      </NTooltip>
    </template>
  </div>
</template>

<style scoped>
.collapsed-flyout-menu {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 0;
  height: 100%;
  box-sizing: border-box;
}

.collapsed-flyout-menu__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  border-radius: var(--n-border-radius, 3px);
  cursor: pointer;
  color: var(--n-text-color-2, inherit);
  transition: background-color 0.2s var(--n-bezier, ease), color 0.2s var(--n-bezier, ease);
}

.collapsed-flyout-menu__icon:hover,
.collapsed-flyout-menu__icon.is-open,
.collapsed-flyout-menu__icon.is-active {
  color: var(--n-primary-color, #18a058);
  background-color: color-mix(in srgb, var(--n-primary-color, #18a058) 14%, transparent);
}

.collapsed-flyout-menu__icon-glyph {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  line-height: 1;
}

.collapsed-flyout-menu__icon-fallback {
  font-size: 14px;
  font-weight: 600;
}
</style>

<style>
/* Dropdown 内容 Teleport 到 body，标题样式需全局类名 */
.collapsed-flyout-menu__panel-title {
  /* 比带子图标的选项更靠左一点，避免和下面文字完全齐平显得呆板 */
  padding: 6px 12px 2px 10px;
  margin-left: -2px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--n-text-color, rgba(255, 255, 255, 0.9));
  user-select: none;
  pointer-events: none;
}
</style>
