<script setup lang="ts">
import { computed } from 'vue'
import type { TableColumnVisibilityOption } from '@/hooks'

interface ColumnSelectorOption extends TableColumnVisibilityOption {
  disabled?: boolean
}

const props = withDefaults(defineProps<{
  modelValue: string[]
  options: ColumnSelectorOption[]
  visibleCount: number
  totalCount: number
  buttonLabel: string
  title: string
  resetLabel: string
  hint?: string
}>(), {
  hint: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
  'reset': []
}>()

const selectedValues = computed({
  get: () => props.modelValue,
  set: (value: string[]) => emit('update:modelValue', value),
})
</script>

<template>
  <NPopover trigger="click" placement="bottom-end">
    <template #trigger>
      <NButton size="small">
        {{ buttonLabel }} ({{ visibleCount }}/{{ totalCount }})
      </NButton>
    </template>
    <NSpace vertical size="small" style="width: 260px;">
      <NText strong>
        {{ title }}
      </NText>
      <NCheckboxGroup v-model:value="selectedValues">
        <NSpace vertical size="small">
          <NCheckbox
            v-for="item in options"
            :key="item.key"
            :value="item.key"
            :disabled="item.disabled"
          >
            {{ item.label }}
          </NCheckbox>
        </NSpace>
      </NCheckboxGroup>
      <NText v-if="hint" depth="3">
        {{ hint }}
      </NText>
      <NSpace justify="end">
        <NButton size="tiny" @click="emit('reset')">
          {{ resetLabel }}
        </NButton>
      </NSpace>
    </NSpace>
  </NPopover>
</template>
