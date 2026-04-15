<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const text = ref('')
// Simulate ajax async content loading
onMounted(() => {
  setTimeout(() => {
    text.value = '<p>' + t('demo.editor.simulateAjax') + '</p>'
  }, 1500)
})

const active = ref(false)
</script>

<template>
  <n-card :title="t('demo.editor.richTitle')">
    <n-space vertical :size="12">
      <n-alert :title="t('demo.editor.quillAlert')" type="success" />
      <n-switch v-model:value="active">
        <template #checked>
          {{ t('demo.editor.disable') }}
        </template>
        <template #unchecked>
          {{ t('demo.editor.enable') }}
        </template>
      </n-switch>
      <n-space :size="12">
        <div class="h-300px">
          <RichTextEditor v-model="text" :disabled="active" />
        </div>
        <div>
          <n-h2>{{ t('demo.editor.vHtmlPreview') }}</n-h2>
          <div v-html="text" />
        </div>
      </n-space>
    </n-space>
  </n-card>
</template>

<style scoped></style>
