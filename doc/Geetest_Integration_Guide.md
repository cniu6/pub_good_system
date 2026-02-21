# Geetest Integration Guide (极验验证码集成指南)

## 1. 问题描述
在使用 `vue3-geetest` 组件时，如果出现以下警告：
```
[Vue warn]: injection "geetest-config" not found.
```
这表明 `vue3-geetest` 插件未在 Vue 应用中正确安装（未调用 `app.use(Geetest, ...)`），导致组件内部无法注入全局配置。

## 2. 正确集成步骤

### 2.1 安装插件 (main.ts)
在入口文件 `main.ts` 中，必须使用 `app.use()` 注册 Geetest 插件，并提供全局配置（如 `captchaId`）。

```typescript
// main.ts
import { createApp } from 'vue'
import App from './App.vue'
import { Geetest } from 'vue3-geetest'

async function setupApp() {
  const app = createApp(App)

  // ... 安装 Router, Pinia 等 ...

  // 1. 准备极验配置
  const geetestConfig = {
    // 优先从环境变量获取，否则使用默认值
    captchaId: import.meta.env.VITE_GEETEST_CAPTCHA_ID || "YOUR_CAPTCHA_ID",
    language: "zh-cn",
    product: "popup",
  }

  // 2. 注册极验插件 (解决 injection not found 问题)
  app.use(Geetest, {
    captchaId: geetestConfig.captchaId,
    language: geetestConfig.language,
    product: geetestConfig.product,
  })

  // 3. (可选) 挂载全局属性，方便在其他地方访问配置
  app.config.globalProperties.$geetestConfig = {
    geetest_enabled: "true",
    geetest_captcha_id: geetestConfig.captchaId,
  }

  app.mount('#app')
}

setupApp()
```

### 2.2 组件封装 (GeetestCaptcha.vue)
封装组件时，可以复用全局配置，也可以通过 props 覆盖配置。

```typescript
// GeetestCaptcha.vue
<script setup lang="ts">
import { GeetestCaptcha as Vue3Geetest } from 'vue3-geetest'
import { computed } from 'vue'

const props = defineProps<{ config?: any }>()

// 这里可以定义组件特有的配置，会与全局配置合并（具体行为取决于 vue3-geetest 实现）
const config = computed(() => ({
  nativeButton: { width: '100%', height: '3rem' },
  ...props.config
}))

// ... handleInitialized 等逻辑 ...
</script>

<template>
  <Vue3Geetest :config="config" @initialized="handleInitialized" />
</template>
```

## 3. 环境变量配置 (.env)
建议将敏感配置放入 `.env` 文件：
```properties
VITE_GEETEST_CAPTCHA_ID=your_captcha_id_here
VITE_GEETEST_ENABLED=true
```

## 4. 常见错误
*   **未安装插件**: 直接使用 `<Vue3Geetest>` 组件而没有在 `main.ts` 中 `app.use(Geetest)`，会报 injection 错误。
*   **配置缺失**: `captchaId` 是必须的。
