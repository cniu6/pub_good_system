<script setup lang="ts">
import { computed, ref } from 'vue'
import type { TreeOption } from 'naive-ui'

const props = defineProps<{
  loading: boolean
  modelValue: string | null
  tables: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
  'refresh': []
}>()

const { t } = useI18n()
const keyword = ref('')

const treeData = computed<TreeOption[]>(() => [{
  key: 'tables',
  label: t('adminServer.dbTables'),
  children: props.tables
    .filter(name => name.toLowerCase().includes(keyword.value.trim().toLowerCase()))
    .map(name => ({ key: name, label: name, isLeaf: true })),
}])

function handleSelect(keys: Array<string | number>) {
  const value = keys[0]
  emit('update:modelValue', typeof value === 'string' && value !== 'tables' ? value : null)
}
</script>

<template>
  <div class="database-object-tree">
    <NSpace vertical :size="12">
      <NSpace justify="space-between" align="center">
        <NText strong>
          {{ t('adminServer.dbObjects') }}
        </NText>
        <NButton quaternary size="tiny" :loading="loading" @click="emit('refresh')">
          {{ t('common.refresh') }}
        </NButton>
      </NSpace>
      <NInput
        v-model:value="keyword"
        clearable
        :placeholder="t('adminServer.dbSearchTable')"
      />
      <NTree
        block-line
        default-expand-all
        :data="treeData"
        :selected-keys="modelValue ? [modelValue] : []"
        :show-irrelevant-nodes="false"
        :node-props="() => ({ tabindex: 0 })"
        @update:selected-keys="handleSelect"
      />
      <NEmpty v-if="!loading && !tables.length" size="small" :description="t('adminServer.dbEmptyTables')" />
    </NSpace>
  </div>
</template>

<style scoped>
.database-object-tree {
  min-height: 420px;
  padding: 4px;
}
</style>
