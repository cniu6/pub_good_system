import { nextTick, ref } from 'vue'
import { i18n } from '@/modules/i18n'
import { defineStore } from 'pinia'
import { useColorMode, useFullscreen, useMediaQuery } from '@vueuse/core'
import type { GlobalThemeOverrides } from 'naive-ui'
import { colord } from 'colord'
import { set } from 'radash'
import { local, setLocale } from '@/utils'
import themeConfig from './theme.json'
import type { ProLayoutMode } from 'pro-naive-ui'

export type TransitionAnimation = '' | 'fade-slide' | 'fade-bottom' | 'fade-scale' | 'zoom-fade' | 'zoom-out'

const { VITE_DEFAULT_LANG, VITE_COPYRIGHT_INFO } = import.meta.env

const docEle = ref(document.documentElement)

const { isFullscreen, toggle } = useFullscreen(docEle)

const { system, store } = useColorMode({
  emitAuto: true,
})

const isMobile = useMediaQuery('(max-width: 700px)')

export const useAppStore = defineStore('app-store', {
  state: () => {
    return {
      footerText: VITE_COPYRIGHT_INFO,
      lang: VITE_DEFAULT_LANG,
      theme: themeConfig as GlobalThemeOverrides,
      primaryColor: themeConfig.common.primaryColor,
      collapsed: false,
      grayMode: false,
      colorWeak: false,
      loadFlag: true,
      showLogo: true,
      showTabs: true,
      showFooter: true,
      showProgress: true,
      showBreadcrumb: true,
      showBreadcrumbIcon: true,
      showWatermark: false,
      showSetting: false,
      transitionAnimation: 'fade-slide' as TransitionAnimation,
      layoutMode: 'vertical' as ProLayoutMode,
    }
  },
  getters: {
    storeColorMode() {
      return store.value
    },
    colorMode() {
      return store.value === 'auto' ? system.value : store.value
    },
    fullScreen() {
      return isFullscreen.value
    },
    isMobile() {
      return isMobile.value
    },
  },
  actions: {
    // 重置所有设置
    resetAlltheme() {
      this.theme = themeConfig
      this.primaryColor = '#18a058'
      this.collapsed = false
      this.grayMode = false
      this.colorWeak = false
      this.loadFlag = true
      this.showLogo = true
      this.showTabs = true
      this.showFooter = true
      this.showBreadcrumb = true
      this.showBreadcrumbIcon = true
      this.showWatermark = false
      this.transitionAnimation = 'fade-slide'
      this.layoutMode = 'vertical'

      // 重置所有配色
      this.setPrimaryColor(this.primaryColor)
    },
    setAppLang(lang: App.lang) {
      setLocale(lang)
      local.set('lang', lang)
      this.lang = lang
      
      // 触发标题更新，为了不循环引用 router，我们通过 document.title 的既有内容进行替换
      // 或者这里只是为了触发一次状态更新，实际的更新逻辑可以放在 router guard 
      // 但 guard 不会因为 store 变化而触发
      
      // 正确的做法：在 main.ts 或者 App.vue 里 watch(locale, () => { 更新标题 })
      // 但这里我们简单点，通过 window.location.reload() 肯定行，但体验不好
      
      // 我们用一种更 hack 的方式：
      // 读取当前的 document.title，尝试分割，替换后缀
      const { t } = i18n.global
      const newAppTitle = t('app.title')
      
      // 更新页面标题
      if (document.title.includes(' - ')) {
        const parts = document.title.split(' - ')
        // 尝试翻译前半部分（页面标题）
        // 这里有个难点：我们不知道前半部分的 key 是什么，只知道现在的文本
        // 如果页面标题是 "Dashboard"，我们如何知道它是 "route.dashboard"？
        // 通常来说，我们无法反向查找。
        
        // 但我们可以利用当前的路由信息重新构建标题
        // 这是一个比较 hack 的方法，但也是最可靠的方法
        
        // 替换后半部分（App 标题）
        parts[parts.length - 1] = newAppTitle
        document.title = parts.join(' - ')
      } else {
        document.title = newAppTitle
      }
      
      // 触发页面重载以更新所有组件的文本（包括面包屑等）
      // 虽然 setLocale 是响应式的，但 document.title 不是
      // 上面的 document.title 更新只能处理后半部分
      
      // 如果我们想要完美更新前半部分，我们需要：
      // 1. 获取当前路由
      // 2. 获取当前路由的 meta.title
      // 3. 重新翻译 meta.title
      
      // 让我们尝试这样做：
      try {
        // 通过 URL hash 获取当前路径 (因为是 hash 模式或 web 模式)
        // 注意：这里假设了简单的路由结构，实际上可能更复杂
        // 最好的方式其实是 reloadPage，但这会刷新内容区域
        
        // 由于我们无法直接访问 router 实例（避免循环依赖），
        // 我们只能做到更新 App Title 部分。
        // 如果用户想要页面标题也更新，通常 router.afterEach 会在下次路由变化时处理。
        // 或者我们可以尝试分发一个自定义事件，让 main.ts 或 App.vue 里的监听器去处理？
        
        // 但这里我们至少可以做到：
        // 如果页面标题是纯文本（未翻译），它不会变。
        // 如果我们能获取到 key...
        
        // 实际上，如果我们在 router guard 里已经处理了 title，
        // 那么这里最简单的就是强制刷新一下路由？
        // router.replace(router.currentRoute.value.fullPath) 
        // 但这需要 router 实例。
        
        // 既然如此，我们只更新 App Title 已经是最优解了，
        // 除非我们能拿到当前的页面 Title Key。
        
        // 还有一种方法：我们可以尝试查找 DOM 中的面包屑或者其他元素？不推荐。
        
        // 让我们采用一个折中方案：
        // 我们已经更新了 i18n locale，Vue 组件（如面包屑、侧边栏）会自动更新。
        // document.title 是浏览器层面的，需要手动更新。
        // 上面的代码已经更新了 App Title。
        // 对于 Page Title，如果用户在路由定义里使用了 key，
        // 那么 router.afterEach 里的逻辑是：t(to.meta.title)
        
        // 问题在于：setLocale 后，afterEach 不会自动触发。
        // 我们可以派发一个事件 'app:locale-changed'
        window.dispatchEvent(new Event('app:locale-changed'))
      } catch (e) {
        console.error(e)
      }
    },
    /* 设置主题色 */
    setPrimaryColor(color: string) {
      const brightenColor = colord(color).lighten(0.05).toHex()
      const darkenColor = colord(color).darken(0.05).toHex()
      set(this.theme, 'common.primaryColor', color)
      set(this.theme, 'common.primaryColorHover', brightenColor)
      set(this.theme, 'common.primaryColorPressed', darkenColor)
      set(this.theme, 'common.primaryColorSuppl', brightenColor)
    },
    setColorMode(mode: 'light' | 'dark' | 'auto') {
      store.value = mode
    },
    /* 切换侧边栏收缩 */
    toggleCollapse() {
      this.collapsed = !this.collapsed
    },
    /* 切换全屏 */
    toggleFullScreen() {
      toggle()
    },
    /**
     * @description: 页面内容重载
     * @param {number} delay - 延迟毫秒数
     * @return {*}
     */
    async reloadPage(delay = 600) {
      this.loadFlag = false
      await nextTick()
      if (delay) {
        setTimeout(() => {
          this.loadFlag = true
        }, delay)
      }
      else {
        this.loadFlag = true
      }
    },
    /* 切换色弱模式 */
    toggleColorWeak() {
      docEle.value.classList.toggle('color-weak')
      this.colorWeak = docEle.value.classList.contains('color-weak')
    },
    /* 切换灰色模式 */
    toggleGrayMode() {
      docEle.value.classList.toggle('gray-mode')
      this.grayMode = docEle.value.classList.contains('gray-mode')
    },
  },
  persist: {
    storage: localStorage,
  },
})
