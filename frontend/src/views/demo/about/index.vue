<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import lib from '@/../package.json'

const { t } = useI18n()
</script>

<template>
  <n-space vertical>
    <n-card :title="t('demo.about.aboutTitle')">
      {{ t('demo.about.aboutDesc') }}
    </n-card>
    <n-card :title="t('demo.about.infoTitle')">
      <n-descriptions
        label-placement="left"
        bordered
        :column="2"
      >
        <n-descriptions-item :label="t('app.github')">
          <n-button
            text
            tag="a"
            href="https://github.com/chansee97/nova-admin"
            target="_blank"
            type="primary"
          >
            {{ t('app.github') }}
          </n-button>
        </n-descriptions-item>
        <n-descriptions-item :label="t('app.gitee')">
          <n-button
            text
            tag="a"
            href="https://gitee.com/chansee97/nova-admin"
            target="_blank"
            type="primary"
          >
            {{ t('app.gitee') }}
          </n-button>
        </n-descriptions-item>
        <n-descriptions-item :label="t('demo.about.previewUrl')">
          <n-button
            text
            tag="a"
            href="https://admin-nova.vercel.app/"
            target="_blank"
            type="primary"
          >
            {{ t('demo.about.previewUrl') }}
          </n-button>
        </n-descriptions-item>
        <n-descriptions-item :label="t('demo.about.version')">
          <n-tag :bordered="false">
            {{ lib.version }}
          </n-tag>
        </n-descriptions-item>
      </n-descriptions>
    </n-card>
    <n-card :title="t('demo.about.prodDeps')">
      <n-descriptions
        label-placement="left"
        bordered
        :column="4"
      >
        <n-descriptions-item
          v-for="(item, key, index) in lib.dependencies"
          :key="index"
          :label="key"
        >
          {{ item }}
        </n-descriptions-item>
      </n-descriptions>
    </n-card>
    <n-card :title="t('demo.about.devDeps')">
      <n-descriptions
        label-placement="left"
        bordered
        :column="4"
      >
        <n-descriptions-item
          v-for="(item, key, index) in lib.devDependencies"
          :key="index"
          :label="key"
        >
          {{ item }}
        </n-descriptions-item>
      </n-descriptions>
    </n-card>
  </n-space>
</template>

<style scoped></style>
