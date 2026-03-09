<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { NTabs, NTabPane, NInput } from 'naive-ui'

interface Props {
  modelValue: string | Record<string, string>
  langs?: { key: string; label: string }[]
  rows?: number
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  langs: () => [
    { key: 'zhCN', label: '中文' },
    { key: 'enUS', label: 'English' },
  ],
  rows: 3,
  placeholder: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, string>]
}>()

const activeTab = ref(props.langs[0]?.key || 'zhCN')

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

const placeholderFor = computed(() => (lang: string) => {
  const langObj = props.langs.find(l => l.key === lang)
  return props.placeholder || `输入${langObj?.label || lang}备注`
})
</script>

<template>
  <NTabs v-model:value="activeTab" type="line" size="small" animated>
    <NTabPane v-for="lang in langs" :key="lang.key" :name="lang.key" :tab="lang.label">
      <NInput
        :value="i18nData[lang.key] || ''"
        type="textarea"
        :rows="rows"
        :placeholder="placeholderFor(lang.key)"
        @update:value="(v: string) => handleInput(lang.key, v)"
      />
    </NTabPane>
  </NTabs>
</template>
