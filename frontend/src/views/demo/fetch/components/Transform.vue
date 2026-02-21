<script setup lang="ts">
const emit = defineEmits<{
  update: [data: any] // 具名元组语法
}>()

async function getDictData() {
  // 移除字典相关功能，改为示例数据
  const res = { data: '示例数据' }
  emit('update', res)
}
</script>

<template>
  <n-card title="Transform Data" size="small">
    <n-button @click="getDictData">
      click
    </n-button>
  </n-card>
</template>

<style scoped>

</style>
