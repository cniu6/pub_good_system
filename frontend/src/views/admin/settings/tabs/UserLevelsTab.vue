<script setup lang="ts">
/**
 * 系统配置内页：用户等级能力（API Key / 充值 / 提现）
 */
import type { DataTableColumns } from 'naive-ui'
import type { UserLevelCap } from '@/service/api/admin/user-level'
import { NSwitch } from 'naive-ui'
import { withSubmitLock } from '@/hooks'
import { adminApi } from '@/service/api/admin'

const { t } = useI18n()
const message = useMessage()

const levels = ref<UserLevelCap[]>([])
const loading = ref(false)
/** 等级开关写操作串行锁，避免连翻并发改权限 */
const savingLevel = ref(false)

async function loadLevels() {
  loading.value = true
  try {
    const res = await adminApi.userLevel.list()
    if (res.isSuccess && res.data)
      levels.value = res.data.list || []
    else
      message.error(res.message || t('adminUserLevels.fetchFailed'))
  }
  catch {
    message.error(t('adminUserLevels.fetchFailed'))
  }
  finally {
    loading.value = false
  }
}

async function saveLevel(row: UserLevelCap) {
  await withSubmitLock(savingLevel, async () => {
    const res = await adminApi.userLevel.update({
      level: row.level,
      name: row.name,
      allow_api_key: row.allow_api_key,
      allow_recharge: row.allow_recharge,
      allow_withdraw: row.allow_withdraw,
      menu_flags: row.menu_flags || '{}',
    })
    if (res.isSuccess) {
      message.success(t('adminUserLevels.saveSuccess'))
      if (res.data?.item) {
        const idx = levels.value.findIndex(l => l.level === row.level)
        if (idx >= 0)
          levels.value[idx] = res.data.item
      }
    }
    else {
      message.error(res.message || t('adminUserLevels.actionFailed'))
      await loadLevels()
    }
  })
}

function onLevelSwitch(row: UserLevelCap, key: 'allow_api_key' | 'allow_recharge' | 'allow_withdraw', v: boolean) {
  if (savingLevel.value)
    return
  row[key] = v
  void saveLevel(row)
}

const columns = computed<DataTableColumns<UserLevelCap>>(() => [
  { title: t('adminUserLevels.level'), key: 'level', width: 80 },
  { title: t('adminUserLevels.name'), key: 'name', width: 160 },
  {
    title: t('adminUserLevels.allowApiKey'),
    key: 'allow_api_key',
    width: 120,
    render: row => h(NSwitch, {
      value: row.allow_api_key,
      disabled: savingLevel.value,
      onUpdateValue: (v: boolean) => onLevelSwitch(row, 'allow_api_key', v),
    }),
  },
  {
    title: t('adminUserLevels.allowRecharge'),
    key: 'allow_recharge',
    width: 120,
    render: row => h(NSwitch, {
      value: row.allow_recharge,
      disabled: savingLevel.value,
      onUpdateValue: (v: boolean) => onLevelSwitch(row, 'allow_recharge', v),
    }),
  },
  {
    title: t('adminUserLevels.allowWithdraw'),
    key: 'allow_withdraw',
    width: 120,
    render: row => h(NSwitch, {
      value: row.allow_withdraw,
      disabled: savingLevel.value,
      onUpdateValue: (v: boolean) => onLevelSwitch(row, 'allow_withdraw', v),
    }),
  },
])

onMounted(() => {
  void loadLevels()
})
</script>

<template>
  <n-space vertical>
    <n-alert type="info" :bordered="false">
      {{ t('adminUserLevels.hint') }}
    </n-alert>
    <n-data-table :columns="columns" :data="levels" :loading="loading" :bordered="false" />
  </n-space>
</template>
