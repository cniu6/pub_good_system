<template>
  <div class="admin-user-detail">
    <n-card :title="`用户详情 - ${user?.username || ''}`">
      <n-descriptions bordered :column="2">
        <n-descriptions-item label="ID">{{ user?.id }}</n-descriptions-item>
        <n-descriptions-item label="用户名">{{ user?.username }}</n-descriptions-item>
        <n-descriptions-item label="昵称">{{ user?.nickname }}</n-descriptions-item>
        <n-descriptions-item label="邮箱">{{ user?.email }}</n-descriptions-item>
<n-descriptions-item label="手机">{{ user?.mobile || '-' }}</n-descriptions-item>
        <n-descriptions-item label="状态">
          <n-tag :type="user?.status === 1 ? 'success' : 'error'">
            {{ user?.status === 1 ? '启用' : '禁用' }}
          </n-tag>
        </n-descriptions-item>
      </n-descriptions>

      <n-space class="mt-4">
        <n-button @click="$router.back()">返回</n-button>
        <n-button type="primary" @click="handleEdit">编辑</n-button>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { NCard, NDescriptions, NDescriptionsItem, NTag, NSpace, NButton, useMessage } from 'naive-ui'
import { adminApi } from '@/service/api/admin'

const route = useRoute()
const message = useMessage()
const user = ref<any>(null)

async function fetchUser() {
  const id = route.params.id as string
  try {
    const res = await adminApi.user.detail(Number(id))
    user.value = res.data?.user
  } catch (error) {
    message.error('获取用户详情失败')
  }
}

function handleEdit() {
  message.info('编辑用户')
}

onMounted(() => {
  fetchUser()
})
</script>

<style scoped>
.admin-user-detail {
  padding: 16px;
}
.mt-4 {
  margin-top: 16px;
}
</style>
