<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Delete from './components/Delete.vue'
import DownLoad from './components/DownLoad.vue'
import DownLoadWithProgress from './components/DownLoadWithProgress.vue'
import Env from './components/Env.vue'
import FailedRequest from './components/FailedRequest.vue'
import FailedResponse from './components/FailedResponse.vue'
import FailedResponseWithoutTip from './components/FailedResponseWithoutTip.vue'
import FormPost from './components/FormPost.vue'
import Get from './components/Get.vue'
import NoToken from './components/NoToken.vue'
import Post from './components/Post.vue'
import Put from './components/Put.vue'
import RefreshToken from './components/RefreshToken.vue'
import TokenExpiration from './components/TokenExpiration.vue'
import Transform from './components/Transform.vue'
import UseRequest from './components/UseRequest.vue'

const { t } = useI18n()

const msg = ref()
function handleUpdate(data: any) {
  msg.value = data
}
</script>

<template>
  <n-card :title="t('demo.fetch.title')">
    <n-split direction="horizontal" :max="0.75" :min="0.25">
      <template #1>
        <div class="grid grid-cols-3 gap-2 p-2">
          <Env @update="handleUpdate" />
          <Get @update="handleUpdate" />
          <Post @update="handleUpdate" />
          <FormPost @update="handleUpdate" />
          <Delete @update="handleUpdate" />
          <Put @update="handleUpdate" />
          <UseRequest @update="handleUpdate" />
          <NoToken @update="handleUpdate" />
          <Transform @update="handleUpdate" />
          <DownLoad @update="handleUpdate" />
          <DownLoadWithProgress class="col-span-2" @update="handleUpdate" />
          <RefreshToken @update="handleUpdate" />
          <FailedRequest @update="handleUpdate" />
          <FailedResponse @update="handleUpdate" />
          <FailedResponseWithoutTip @update="handleUpdate" />
          <TokenExpiration @update="handleUpdate" />
        </div>
      </template>
      <template #2>
        <pre class="bg-#eee:30">
      {{ msg }}
      </pre>
      </template>
    </n-split>
  </n-card>
</template>

<style scoped></style>
