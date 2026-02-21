<script setup lang="ts">
import { computed, inject, readonly, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { GeetestCaptcha as Vue3Geetest } from 'vue3-geetest'
import { geetestManager } from '@/utils/geetest'
import type { GeetestResult } from '@/utils/geetest'

interface Props {
  // 用户信息（可选）
  userInfo?: string
  // 配置覆盖
  config?: Record<string, any>
}

const props = withDefaults(defineProps<Props>(), {
  userInfo: '',
})

const emit = defineEmits<{
  success: [result: GeetestResult]
  error: [error: any]
  ready: []
}>()

// 验证码实例
const captchaRef = ref()

// 验证码状态
const isReady = ref(false)
const isLoading = ref(true)
const isVerified = ref(false)
const captchaKey = ref(0)

const { locale } = useI18n()

const globalGeetestConfig = inject<any>('geetestConfig', {})

// 语言映射
const languageMap: Record<string, string> = {
  zhCN: 'zho',
  enUS: 'eng',
}

// 计算当前语言
const currentLanguage = computed(() => {
  const lang = locale.value
  return languageMap[lang] || globalGeetestConfig.language || 'eng'
})

// 监听语言变化，重新加载验证码
watch(currentLanguage, () => {
  captchaKey.value++
})

// 极验配置
const captchaConfig = computed(() => ({
  ...globalGeetestConfig,
  language: currentLanguage.value,
  ...props.config,
}))

// 占位样式
const wrapperStyle = computed(() => {
  const product = captchaConfig.value.product
  // bind 模式不需要占位
  if (product === 'bind')
    return {}

  const height = captchaConfig.value.nativeButton?.height || '3rem'
  return {
    minHeight: height,
  }
})

// 处理验证码初始化完成
function handleInitialized(captcha: any) {
  captchaRef.value = captcha
  isReady.value = true
  isLoading.value = false
  emit('ready')

  // 绑定事件
  captcha
    .onReady(() => {})
    .onSuccess(() => {
      try {
        const result = captcha.getValidate()
        if (!result) {
          emit('error', new Error('Failed to get validation result'))
          return
        }

        geetestManager.setCaptchaResult(result)
        isVerified.value = true
        emit('success', result)
      }
      catch {
        emit('error', new Error('Failed to handle validation result'))
      }
    })
    .onError((error: any) => {
      emit('error', error || new Error('Verification failed, please try again'))
    })
}

// 手动显示验证码
function showCaptcha() {
  if (captchaRef.value && typeof captchaRef.value.showCaptcha === 'function') {
    captchaRef.value.showCaptcha()
  }
}

// 暴露方法给父组件
defineExpose({
  showCaptcha,
  isReady: readonly(isReady),
  isLoading: readonly(isLoading),
  isVerified: readonly(isVerified),
})
</script>

<template>
  <div
    :key="captchaKey"
    class="geetest-captcha-wrapper"
    :style="wrapperStyle"
  >
    <Vue3Geetest
      :config="captchaConfig"
      @initialized="handleInitialized"
    />
  </div>
</template>

<style scoped>
.geetest-captcha-wrapper {
  width: 100%;
}
</style>
