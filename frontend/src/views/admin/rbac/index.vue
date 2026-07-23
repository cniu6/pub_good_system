/**
 * RBAC 角色权限查看 + 按用户 ID 分配角色（单组织 MVP）
 */
<script setup lang="ts">
import { h, onMounted, reactive, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NTag, useMessage } from 'naive-ui'
import { adminRbacApi } from '@/service/api/admin/rbac'
import type { RbacPermission, RbacRole } from '@/service/api/admin/rbac'

const { t } = useI18n()
const message = useMessage()
const loading = ref(false)
const roles = ref<RbacRole[]>([])
const permissions = ref<RbacPermission[]>([])
const userRoles = ref<RbacRole[]>([])

const assignForm = reactive({
  user_id: null as number | null,
  role_id: null as number | null,
})

const roleOptions = computed(() => roles.value.map(r => ({
  label: `${r.name} (${r.code})`,
  value: r.id,
})))

async function fetchAll() {
  loading.value = true
  try {
    const [rRes, pRes] = await Promise.all([
      adminRbacApi.listRoles(),
      adminRbacApi.listPermissions(),
    ])
    if (rRes.isSuccess && rRes.data)
      roles.value = rRes.data.list || []
    if (pRes.isSuccess && pRes.data)
      permissions.value = pRes.data.list || []
  }
  catch {
    message.error(t('adminRbac.fetchFailed'))
  }
  finally {
    loading.value = false
  }
}

async function loadUserRoles() {
  if (!assignForm.user_id) {
    message.warning(t('adminRbac.userIdRequired'))
    return
  }
  const res = await adminRbacApi.listUserRoles(assignForm.user_id)
  if (res.isSuccess && res.data)
    userRoles.value = res.data.list || []
  else
    message.error(res.message || t('adminRbac.fetchFailed'))
}

async function assignRole() {
  if (!assignForm.user_id || !assignForm.role_id) {
    message.warning(t('adminRbac.assignRequired'))
    return
  }
  const res = await adminRbacApi.assignUserRole(assignForm.user_id, { role_id: assignForm.role_id })
  if (res.isSuccess) {
    message.success(t('adminRbac.assignSuccess'))
    loadUserRoles()
  }
  else {
    message.error(res.message || t('adminRbac.actionFailed'))
  }
}

const roleColumns: DataTableColumns<RbacRole> = [
  { title: 'ID', key: 'id', width: 70 },
  { title: t('adminRbac.code'), key: 'code', width: 140 },
  { title: t('adminRbac.name'), key: 'name', width: 140 },
  { title: t('adminRbac.description'), key: 'description', ellipsis: { tooltip: true } },
]

const permColumns: DataTableColumns<RbacPermission> = [
  { title: 'ID', key: 'id', width: 70 },
  {
    title: t('adminRbac.code'),
    key: 'code',
    width: 160,
    render: row => h(NTag, { size: 'small', type: 'info' }, { default: () => row.code }),
  },
  { title: t('adminRbac.name'), key: 'name', width: 140 },
  { title: t('adminRbac.description'), key: 'description', ellipsis: { tooltip: true } },
]

onMounted(fetchAll)
</script>

<template>
  <n-space vertical :size="16">
    <n-card :title="t('adminRbac.assignTitle')" :bordered="false">
      <n-alert type="warning" class="mb-12px">
        {{ t('adminRbac.mvpScopeHint') }}
      </n-alert>
      <n-space align="center" class="mb-12px">
        <n-input-number v-model:value="assignForm.user_id" :min="1" :placeholder="t('adminRbac.userId')" />
        <n-select
          v-model:value="assignForm.role_id"
          :options="roleOptions"
          :placeholder="t('adminRbac.selectRole')"
          style="width: 240px"
          clearable
        />
        <n-button @click="loadUserRoles">
          {{ t('adminRbac.loadUserRoles') }}
        </n-button>
        <n-button type="primary" @click="assignRole">
          {{ t('adminRbac.assign') }}
        </n-button>
      </n-space>
      <n-space>
        <n-tag v-for="r in userRoles" :key="r.id" type="success">
          {{ r.name }} ({{ r.code }})
        </n-tag>
        <span v-if="!userRoles.length" class="opacity-60">{{ t('adminRbac.noUserRoles') }}</span>
      </n-space>
    </n-card>

    <n-card :title="t('adminRbac.rolesTitle')" :bordered="false" :loading="loading">
      <n-data-table :columns="roleColumns" :data="roles" :bordered="false" />
    </n-card>

    <n-card :title="t('adminRbac.permsTitle')" :bordered="false" :loading="loading">
      <n-alert type="info" class="mb-12px">
        {{ t('adminRbac.hint') }}
      </n-alert>
      <n-data-table :columns="permColumns" :data="permissions" :bordered="false" />
    </n-card>
  </n-space>
</template>
