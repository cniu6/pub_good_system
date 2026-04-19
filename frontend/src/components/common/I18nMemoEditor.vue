<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { NTabs, NTabPane, NInput } from 'naive-ui'

const { t } = useI18n()

interface Props {
  modelValue: string | Record<string, string>
  langs?: { key: string; label: string }[]
  rows?: number
  placeholder?: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string>]
}>()

// 使用计算属性提供带翻译的默认值
const langsWithDefaults = computed(() => {
  return props.langs ?? [
    { key: 'zhCN', label: t('common.chinese') },
    { key: 'enUS', label: t('common.english') },
  ]
})

const activeTab = ref(langsWithDefaults.value[0]?.key || 'zhCN')

const i18nData = ref<Record<string, string>>({})

// 初始化：将 modelValue 解析为多语言对象
function parseValue(val: string | Record<string, string>): Record<string, string> {
  if (!val) return {}
  if (typeof val === 'object') return { ...val }
  if (typeof val === 'string' && val.startsWith('{')) {
    try {
      return JSON.parse(val)
    }
    catch {
      return { zhCN: val }
    }
  }
  return { zhCN: val }
}

watch(() => props.modelValue, (val) => {
  i18nData.value = parseValue(val)
}, { immediate: true })

function handleInput(lang: string, text: string) {
  i18nData.value[lang] = text
  emit('update:modelValue', { ...i18nData.value })
}

// rows 默认值
const rowsWithDefault = computed(() => props.rows ?? 3)

const placeholderFor = computed(() => (lang: string) => {
  if (props.placeholder) return props.placeholder
  const langObj = langsWithDefaults.value.find(l => l.key === lang)
  return t('common.enterRemarkFor', { lang: langObj?.label || lang })
})
</script>

<template>
  <NTabs v-model:value="activeTab" type="line" size="small" animated>
    <NTabPane v-for="lang in langsWithDefaults" :key="lang.key" :name="lang.key" :tab="lang.label">
      <NInput
        :value="i18nData[lang.key] || ''"
        type="textarea"
        :rows="rowsWithDefault"
        :placeholder="placeholderFor(lang.key)"
        @update:value="(v: string) => handleInput(lang.key, v)"
      />
    </NTabPane>
  </NTabs>
</template>
