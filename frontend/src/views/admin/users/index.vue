<template>
  <div class="admin-users">
    <n-card title="用户管理">
      <!-- 搜索栏 -->
      <n-space class="mb-4">
        <n-input v-model:value="searchForm.keyword" placeholder="搜索用户名/邮箱" clearable style="width: 200px" />
        <n-select v-model:value="searchForm.status" :options="statusOptions" placeholder="状态" clearable style="width: 120px" />
        <n-button type="primary" @click="handleSearch">搜索</n-button>
        <n-button @click="handleReset">重置</n-button>
        <n-button type="success" @click="handleCreate">新增用户</n-button>
      </n-space>

      <!-- 用户表格 -->
      <n-data-table
        :columns="columns"
        :data="userList"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />

      <!-- 用户表单弹窗 -->
      <n-modal v-model:show="showModal" :title="modalTitle" preset="card" style="width: 600px">
        <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" label-width="80px">
          <n-form-item label="用户名" path="username">
            <n-input v-model:value="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
          </n-form-item>
          <n-form-item v-if="!isEdit" label="密码" path="password">
            <n-input v-model:value="formData.password" type="password" placeholder="请输入密码" show-password-on="click" />
          </n-form-item>
          <n-form-item label="邮箱" path="email">
            <n-input v-model:value="formData.email" placeholder="请输入邮箱" />
          </n-form-item>
          <n-form-item label="昵称" path="nickname">
            <n-input v-model:value="formData.nickname" placeholder="请输入昵称" />
          </n-form-item>
<n-form-item label="手机" path="mobile">
            <n-input v-model:value="formData.mobile" placeholder="请输入手机号" />
          </n-form-item>
          <n-form-item label="状态" path="status">
            <n-radio-group v-model:value="formData.status">
              <n-radio :value="1">启用</n-radio>
              <n-radio :value="0">禁用</n-radio>
            </n-radio-group>
          </n-form-item>
        </n-form>
        <template #footer>
          <n-space justify="end">
            <n-button @click="showModal = false">取消</n-button>
            <n-button type="primary" @click="handleSubmit" :loading="submitting">确定</n-button>
          </n-space>
        </template>
      </n-modal>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  NCard, NSpace, NInput, NSelect, NButton, NDataTable, NTag, NModal, NForm, NFormItem,
  NRadioGroup, NRadio, useMessage, useDialog
} from 'naive-ui'
import type { DataTableColumns, FormInst } from 'naive-ui'
import { adminApi } from '@/service/api/admin'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const submitting = ref(false)
const userList = ref<any[]>([])
const showModal = ref(false)
const formRef = ref<FormInst | null>(null)
const isEdit = ref(false)
const currentUserId = ref<number | null>(null)

const searchForm = reactive({
  keyword: '',
  status: null as number | null,
  page: 1,
  page_size: 10,
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

const formData = reactive({
  username: '',
  password: '',
  email: '',
  nickname: '',
  mobile: '',
  status: 1,
})

const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度为 3-20 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 个字符', trigger: 'blur' },
  ],
email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
}

const statusOptions = [
  { label: '启用', value: 1 },
  { label: '禁用', value: 0 },
]

const modalTitle = computed(() => isEdit.value ? '编辑用户' : '新增用户')

const columns: DataTableColumns<any> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '用户名', key: 'username' },
  { title: '昵称', key: 'nickname' },
  { title: '邮箱', key: 'email' },
{
    title: '状态',
    key: 'status',
    render(row) {
      return h(NTag, { type: row.status === 1 ? 'success' : 'error' }, () => row.status === 1 ? '启用' : '禁用')
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 250,
    render(row) {
      return h(NSpace, {}, () => [
        h(NButton, { size: 'small', onClick: () => handleView(row) }, () => '查看'),
        h(NButton, { size: 'small', type: 'primary', onClick: () => handleEdit(row) }, () => '编辑'),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, () => '删除'),
      ])
    },
  },
]

async function fetchUsers() {
  loading.value = true
  try {
    // 过滤掉 null/undefined 的参数，避免发送 "null" 字符串导致后端解析失败
    const params = Object.fromEntries(
      Object.entries(searchForm).filter(([_, v]) => v !== null && v !== undefined)
    )
    const res = await adminApi.user.list(params)
    if (res.code === 200) {
      userList.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
      pagination.page = res.data?.page || 1
      pagination.pageSize = res.data?.page_size || 10
    } else {
      message.error(res.message || '获取用户列表失败')
    }
  } catch (error) {
    message.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  searchForm.page = 1
  pagination.page = 1
  fetchUsers()
}

function handleReset() {
  searchForm.keyword = ''
  searchForm.status = null
  searchForm.page = 1
  pagination.page = 1
  fetchUsers()
}

function handlePageChange(page: number) {
  searchForm.page = page
  pagination.page = page
  fetchUsers()
}

function handlePageSizeChange(pageSize: number) {
  searchForm.page_size = pageSize
  pagination.pageSize = pageSize
  pagination.page = 1
  searchForm.page = 1
  fetchUsers()
}

function resetForm() {
  formData.username = ''
  formData.password = ''
  formData.email = ''
  formData.nickname = ''
  formData.mobile = ''
  formData.role = 'user'
  formData.status = 1
  currentUserId.value = null
  isEdit.value = false
  formRef.value?.restoreValidation()
}

function handleCreate() {
  resetForm()
  showModal.value = true
}

function handleView(row: any) {
  router.push(`/admin/users/${row.id}`)
}

async function handleEdit(row: any) {
  resetForm()
  isEdit.value = true
  currentUserId.value = row.id
  
  try {
    const res = await adminApi.user.detail(row.id)
    if (res.code === 200 && res.data?.user) {
      const user = res.data.user
      formData.username = user.username || ''
      formData.email = user.email || ''
      formData.nickname = user.nickname || ''
      formData.mobile = user.mobile || ''
      formData.role = user.role || 'user'
      formData.status = user.status ?? 1
    }
    showModal.value = true
  } catch (error) {
    message.error('获取用户信息失败')
  }
}

async function handleSubmit() {
  if (!formRef.value) return
  
  await formRef.value.validate(async (errors) => {
    if (errors) return
    
    submitting.value = true
    try {
      if (isEdit.value && currentUserId.value) {
        // 编辑用户
        const res = await adminApi.user.update(currentUserId.value, {
          email: formData.email,
          nickname: formData.nickname,
          mobile: formData.mobile,
          role: formData.role,
          status: formData.status,
        })
        if (res.code === 200) {
          message.success('更新用户成功')
          showModal.value = false
          fetchUsers()
        } else {
          message.error(res.message || '更新用户失败')
        }
      } else {
        // 创建用户
        const res = await adminApi.user.create({
          username: formData.username,
          password: formData.password,
          email: formData.email,
          nickname: formData.nickname,
          mobile: formData.mobile,
          role: formData.role,
          status: formData.status,
        })
        if (res.code === 200) {
          message.success('创建用户成功')
          showModal.value = false
          fetchUsers()
        } else {
          message.error(res.message || '创建用户失败')
        }
      }
    } catch (error: any) {
      message.error(error.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

function handleDelete(row: any) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除用户 "${row.username}" 吗？此操作不可恢复。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await adminApi.user.delete(row.id)
        if (res.code === 200) {
          message.success('删除用户成功')
          fetchUsers()
        } else {
          message.error(res.message || '删除用户失败')
        }
      } catch (error) {
        message.error('删除用户失败')
      }
    },
  })
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.admin-users {
  padding: 16px;
}
.mb-4 {
  margin-bottom: 16px;
}
</style>
