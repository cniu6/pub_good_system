<script setup lang="ts">
/**
 * 大厂式手机号输入：国家/区号选择 + 本地号码
 * - 仅大陆模式：锁定 CN，只填 11 位
 * - 国际模式：可选国家，区号自动带上，提交时合成 E.164
 */
import type { SelectOption } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useSettingsStore } from '@/store'
import { countryFromLanguage, DEFAULT_DIAL_COUNTRY_CODE, DIAL_COUNTRIES, dialCountryLabel, getDialCountry } from '@/constants/dialCodes'
import { composeE164, normalizeAndValidateMobile, splitMobileToCountryNational } from '@/utils/phone'
import { fetchPhoneCountryDetect } from '@/service/api/geo'

const props = withDefaults(defineProps<{
  modelValue?: string
  disabled?: boolean
  /** 为空时自动跟随 settingsStore.mobileCnOnly */
  cnOnly?: boolean | null
  /** 是否主动探测默认国家（打开弹窗时建议 true） */
  autoDetect?: boolean
}>(), {
  modelValue: '',
  disabled: false,
  cnOnly: null,
  autoDetect: true,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t, locale } = useI18n()
const settingsStore = useSettingsStore()

const effectiveCnOnly = computed(() => props.cnOnly ?? settingsStore.mobileCnOnly)
const isZh = computed(() => String(locale.value || '').toLowerCase().startsWith('zh'))

const countryCode = ref(DEFAULT_DIAL_COUNTRY_CODE)
const national = ref('')
const detecting = ref(false)
let syncingFromParent = false
let userPickedCountry = false

const countryOptions = computed<SelectOption[]>(() => {
  const list = effectiveCnOnly.value
    ? DIAL_COUNTRIES.filter(c => c.code === 'CN')
    : DIAL_COUNTRIES
  return list.map(c => ({
    label: dialCountryLabel(c, isZh.value),
    value: c.code,
  }))
})

const dialPrefix = computed(() => {
  const c = getDialCountry(countryCode.value)
  return c ? `+${c.dialCode}` : '+1'
})

const hintText = computed(() => {
  if (effectiveCnOnly.value)
    return t('profile.phoneHintCN')
  return t('profile.phoneHintCountrySelect')
})

function emitComposed() {
  if (syncingFromParent)
    return
  const dial = getDialCountry(countryCode.value)?.dialCode || '1'
  if (effectiveCnOnly.value) {
    const n = national.value.replace(/\D/g, '')
    emit('update:modelValue', n)
    return
  }
  if (!national.value.trim()) {
    emit('update:modelValue', '')
    return
  }
  emit('update:modelValue', composeE164(dial, national.value))
}

function applySplit(raw: string, fallbackCode: string) {
  syncingFromParent = true
  const { country, national: nat } = splitMobileToCountryNational(raw, fallbackCode)
  countryCode.value = effectiveCnOnly.value ? 'CN' : country.code
  national.value = nat
  nextTick(() => {
    syncingFromParent = false
  })
}

async function detectDefaultCountry() {
  if (!props.autoDetect || userPickedCountry || national.value)
    return
  if (effectiveCnOnly.value) {
    countryCode.value = 'CN'
    return
  }
  detecting.value = true
  try {
    const lang = String(locale.value || 'zhCN')
    const res: any = await fetchPhoneCountryDetect(lang)
    const code = res?.data?.country_code
      || (res as any)?.country_code
    if (res?.isSuccess !== false && code && getDialCountry(code)) {
      countryCode.value = code
      return
    }
  }
  catch {
    // 静默：走语言保底
  }
  finally {
    detecting.value = false
  }
  countryCode.value = countryFromLanguage(String(locale.value || ''))
}

watch(
  () => props.modelValue,
  (v) => {
    const composed = effectiveCnOnly.value
      ? national.value.replace(/\D/g, '')
      : (national.value.trim()
          ? composeE164(getDialCountry(countryCode.value)?.dialCode || '1', national.value)
          : '')
    if ((v || '') === composed)
      return
    applySplit(v || '', countryCode.value || countryFromLanguage(String(locale.value || '')))
  },
  { immediate: true },
)

watch(effectiveCnOnly, (cn) => {
  if (cn) {
    countryCode.value = 'CN'
    emitComposed()
  }
  else if (!userPickedCountry && !national.value) {
    detectDefaultCountry()
  }
})

watch([countryCode, national], () => emitComposed())

onMounted(() => {
  if (!props.modelValue)
    detectDefaultCountry()
})

function onCountryUpdate(code: string) {
  userPickedCountry = true
  countryCode.value = code
}

/** 供父组件在提交前拿到规范化结果；非法返回 null */
function getNormalized(): string | null {
  const raw = props.modelValue || ''
  if (!raw.trim() && !national.value.trim())
    return ''
  return normalizeAndValidateMobile(
    effectiveCnOnly.value ? national.value : (props.modelValue || composeE164(getDialCountry(countryCode.value)?.dialCode || '1', national.value)),
    effectiveCnOnly.value,
  )
}

defineExpose({ getNormalized, detectDefaultCountry })
</script>

<template>
  <n-space vertical :size="4" style="width: 100%;">
    <n-input-group>
      <n-select
        :value="countryCode"
        :options="countryOptions"
        :disabled="disabled || effectiveCnOnly"
        :loading="detecting"
        filterable
        style="width: 180px; flex-shrink: 0;"
        :consistent-menu-width="false"
        @update:value="onCountryUpdate"
      />
      <n-input
        :value="national"
        :disabled="disabled"
        :placeholder="effectiveCnOnly ? t('profile.enterNewPhone') : t('profile.enterNationalNumber')"
        @update:value="(v: string) => { national = v }"
      >
        <template #prefix>
          <n-text depth="3" style="white-space: nowrap;">
            {{ dialPrefix }}
          </n-text>
        </template>
      </n-input>
    </n-input-group>
    <n-text depth="3" style="font-size: 12px;">
      {{ hintText }}
    </n-text>
  </n-space>
</template>
