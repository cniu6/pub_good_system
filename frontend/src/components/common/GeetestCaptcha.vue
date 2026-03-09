<script setup lang="ts">
import { computed, readonly, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { GeetestCaptcha as Vue3Geetest } from 'vue3-geetest'
import type { CaptchaConfig } from 'vue3-geetest'
import { geetestManager } from '@/utils/geetest'
import type { GeetestResult } from '@/utils/geetest'
import { useSettingsStore } from '@/store'

interface Props {
  // 用户信息（可选）
  userInfo?: string
  // 配置覆盖
  config?: Partial<CaptchaConfig>
  // 强制禁用（覆盖全局配置）
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  userInfo: '',
  disabled: false,
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
const settingsStore = useSettingsStore()

// 从后端配置获取极验验证码ID
const geetestCaptchaId = computed(() => settingsStore.geetestCaptchaId)

// 判断极验是否启用（以后端配置为准）
const isGeetestEnabled = computed(() => {
  if (props.disabled)
    return false
  return settingsStore.geetestEnabled
})

// 语言映射
const languageMap: Record<string, NonNullable<CaptchaConfig['language']>> = {
  zhCN: 'zho',
  enUS: 'eng',
}

// 计算当前语言
const currentLanguage = computed(() => {
  const lang = locale.value
  return languageMap[lang] || 'eng'
})

// 监听语言变化，重新加载验证码
watch(currentLanguage, () => {
  captchaKey.value++
})

const defaultNativeButton = {
  width: '100%',
  height: '3rem',
}

const resolvedProduct = computed<NonNullable<CaptchaConfig['product']>>(() => {
  const product = props.config?.product
  if (product === 'bind' || product === 'float' || product === 'popup') {
    return product
  }
  return 'popup'
})

// 极验配置
const captchaConfig = computed<CaptchaConfig>(() => ({
  ...props.config,
  captchaId: geetestCaptchaId.value,
  language: currentLanguage.value,
  product: resolvedProduct.value,
  nativeButton: {
    ...defaultNativeButton,
    ...(props.config?.nativeButton || {}),
  },
}))

// 占位样式
const wrapperStyle = computed(() => {
  // 禁用时隐藏组件
  if (!isGeetestEnabled.value) {
    return {
      display: 'none',
    }
  }

  const product = captchaConfig.value.product
  // bind 模式不需要占位
  if (product === 'bind')
    return {}

  const height = captchaConfig.value.nativeButton?.height || '3rem'
  return {
    minHeight: height,
  }
})

// 模拟的验证结果（禁用时使用）
const mockGeetestResult: GeetestResult = {
  lot_number: 'disabled_geetest_lot_number',
  captcha_output: 'disabled_geetest_output',
  pass_token: 'disabled_geetest_pass_token',
  gen_time: 'disabled_geetest_gen_time',
  captcha_id: 'disabled_geetest_id',
}

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

// 手动显示验证码（或触发模拟成功）
function showCaptcha() {
  // 禁用时直接返回模拟成功结果
  if (!isGeetestEnabled.value) {
    geetestManager.setCaptchaResult(mockGeetestResult)
    isVerified.value = true
    emit('success', mockGeetestResult)
    return
  }

  if (captchaRef.value && typeof captchaRef.value.showCaptcha === 'function') {
    captchaRef.value.showCaptcha()
  }
}

// 重置验证状态
function reset() {
  isVerified.value = false
  geetestManager.clearCaptchaResult()
}

// 暴露方法给父组件
defineExpose({
  showCaptcha,
  reset,
  isReady: readonly(isReady),
  isLoading: readonly(isLoading),
  isVerified: readonly(isVerified),
  isEnabled: readonly(isGeetestEnabled),
})
</script>

<template>
  <div
    :key="captchaKey"
    class="geetest-captcha-wrapper"
    :style="wrapperStyle"
  >
    <!-- 禁用时不渲染极验组件 -->
    <Vue3Geetest
      v-if="isGeetestEnabled"
      :config="captchaConfig"
      @initialized="handleInitialized"
    />
  </div>
</template>

<style scoped>
.geetest-captcha-wrapper {
  width: 100%;
}

.geetest-captcha-wrapper :deep([id^="captcha-"]) {
  width: 100%;
}

.geetest-captcha-wrapper :deep([id^="captcha-"] > *) {
  width: 100%;
}
</style>
