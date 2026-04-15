<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <n-space vertical>
    <n-card :title="t('demo.icons.iconPicker')">
      <icon-select />
    </n-card>
    <n-card :title="t('demo.icons.autoImportIcons')">
      <div>
        {{ t('demo.icons.normal') }}：<icon-park-outline-apple />
      </div>
      <div>
        {{ t('demo.icons.large') }}：<icon-park-outline-apple class="text-2em" />
      </div>
      <div>
        {{ t('demo.icons.extraLarge') }}：<icon-park-outline-apple class="text-4em" />
      </div>
    </n-card>
    <n-card :title="t('demo.icons.autoImportSvg')">
      <div>
        {{ t('demo.icons.normal') }}：<svg-icons-cool />
      </div>
      <div>
        {{ t('demo.icons.large') }}：<svg-icons-cool class="text-2em" />
      </div>
      <div>
        {{ t('demo.icons.extraLarge') }}：<svg-icons-cool class="text-4em" />
      </div>
      <div>
        {{ t('demo.icons.novaIconLoad') }}：<nova-icon icon="local:cool" />
      </div>
    </n-card>
  </n-space>
</template>
