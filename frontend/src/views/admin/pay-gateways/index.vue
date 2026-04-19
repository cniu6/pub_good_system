<template>
  <n-space vertical :size="16">
    <n-card :title="t('route.admin-pay-gateways')">
      <template #header-extra>
        <n-space>
          <TableColumnSelector
            v-model="selectedColumnKeys"
            :options="columnOptions"
            :visible-count="visibleColumnCount"
            :total-count="totalColumnCount"
            :button-label="t('common.showFields')"
            :title="t('common.visibleFields')"
            :hint="t('common.columnVisibilityHint')"
            :reset-label="t('common.restoreDefaultFields')"
            @reset="resetSelectedColumns"
          />
          <n-button type="primary" @click="handleCreate">
            <template #icon><n-icon><icon-park-outline-add-one /></n-icon></template>
            {{ t('adminPayGateways.addGatewayShort') }}
          </n-button>
        </n-space>
      </template>

      <n-space vertical>
        <n-space>
          <n-input v-model:value="keyword" :placeholder="t('adminPayGateways.searchPlaceholder')" clearable style="width: 220px" @keyup.enter="loadList" />
          <n-button type="primary" @click="loadList">{{ t('moneyScore.search') }}</n-button>
        </n-space>

        <n-data-table
          :columns="visibleColumns"
          :data="list"
          :loading="loading"
          :pagination="pagination"
          :scroll-x="tableScrollX"
          striped
          size="small"
          @update:page="(p: number) => { pagination.page = p; loadList() }"
          @update:page-size="(s: number) => { pagination.pageSize = s; pagination.page = 1; loadList() }"
        />
      </n-space>
    </n-card>

    <!-- 新增/编辑弹窗 -->
    <n-modal v-model:show="showModal" preset="card" :title="editingId ? t('adminPayGateways.editGateway') : t('adminPayGateways.addGateway')" style="width: 680px" :mask-closable="false">
      <n-form ref="formRef" :model="form" :rules="formRules" label-placement="left" label-width="100">
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item :label="t('adminPayGateways.gatewayName')" path="name">
              <n-input v-model:value="form.name" :placeholder="t('adminPayGateways.gatewayNamePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.gatewayType')" path="type">
              <n-select v-model:value="form.type" :options="typeOptions" :placeholder="t('adminPayGateways.gatewayTypePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('recharge.paymentMethod')" path="pay_type">
              <n-select v-model:value="form.pay_type" :options="payTypeOptions" :placeholder="t('adminPayGateways.payTypePlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminUsers.status')" path="status">
              <n-switch v-model:value="form.status" :checked-value="1" :unchecked-value="0">
                <template #checked>{{ t('adminUsers.enabled') }}</template>
                <template #unchecked>{{ t('adminUsers.disabled') }}</template>
              </n-switch>
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.description')" path="description">
              <n-input v-model:value="form.description" type="textarea" :placeholder="t('adminPayGateways.descriptionPlaceholder')" :rows="2" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.apiUrl')" path="api_url">
              <n-input v-model:value="form.api_url" :placeholder="t('adminPayGateways.apiUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.merchantId')" path="pid">
              <n-input v-model:value="form.pid" :placeholder="t('adminPayGateways.merchantIdPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.merchantKey')" path="key">
              <n-input v-model:value="form.key" type="password" show-password-on="click" :placeholder="t('adminPayGateways.merchantKeyPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.logoUrl')" path="logo_url">
              <n-input v-model:value="form.logo_url" :placeholder="t('adminPayGateways.logoUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.minAmount')" path="min_amount">
              <n-input-number v-model:value="form.min_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>¥</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.maxAmount')" path="max_amount">
              <n-input-number v-model:value="form.max_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>¥</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.feeRate')" path="fee_rate">
              <n-input-number v-model:value="form.fee_rate" :min="0" :max="100" style="width: 100%">
                <template #suffix>%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.feeMode')" path="fee_mode">
              <n-select v-model:value="form.fee_mode" :options="feeModeOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.minLevel')" path="min_level">
              <n-input-number v-model:value="form.min_level" :min="0" style="width: 100%" :placeholder="t('adminPayGateways.minLevelPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item :label="t('adminPayGateways.sortOrder')" path="sort_order">
              <n-input-number v-model:value="form.sort_order" :min="0" style="width: 100%" :placeholder="t('adminPayGateways.sortOrderPlaceholder')" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item :label="t('adminPayGateways.notifyUrl')" path="notify_url">
              <n-input v-model:value="form.notify_url" :placeholder="t('adminPayGateways.notifyUrlPlaceholder')" />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">{{ t('common.confirm') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { useMessage, useDialog, NTag, NButton, NSpace, NImage } from 'naive-ui'
import type { DataTableColumns, FormRules } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import {
  fetchPayGateways,
  createPayGateway,
  updatePayGateway,
  deletePayGateway,
} from '@/service/api/admin/paygateway'
import type { PayGateway, PayGatewayCreateRequest } from '@/service/api/admin/paygateway'

const message = useMessage()
const dialog = useDialog()
const { t } = useI18n()

const loading = ref(false)
const list = ref<PayGateway[]>([])
const keyword = ref('')
const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
})

// 弹窗
const showModal = ref(false)
const editingId = ref<number | null>(null)
const submitting = ref(false)
const formRef = ref()

const defaultForm = (): PayGatewayCreateRequest => ({
  name: '',
  type: 'epay',
  pay_type: 'alipay',
  description: '',
  status: 1,
  api_url: '',
  pid: '',
  key: '',
  logo_url: '',
  sort_order: 0,
  min_amount: 1,
  max_amount: 10000,
  fee_rate: 0,
  fee_mode: 'add',
  min_level: 0,
  notify_url: '',
})

const form = reactive<PayGatewayCreateRequest>(defaultForm())

const formRules: FormRules = {
  name: [{ required: true, message: t('adminPayGateways.enterName'), trigger: 'blur' }],
  type: [{ required: true, message: t('adminPayGateways.selectType'), trigger: 'change' }],
  pay_type: [{ required: true, message: t('adminPayGateways.selectPayType'), trigger: 'change' }],
}

const typeOptions = [
  { label: t('adminPayGateways.epay'), value: 'epay' },
]

const payTypeOptions = [
  { label: t('recharge.alipay'), value: 'alipay' },
  { label: t('recharge.wechatPay'), value: 'wxpay' },
  { label: t('recharge.qqWallet'), value: 'qqpay' },
  { label: t('recharge.bankCard'), value: 'bank' },
  { label: t('recharge.jdPay'), value: 'jdpay' },
]

const feeModeOptions = [
  { label: t('adminPayGateways.feeModeAdd'), value: 'add' },
  { label: t('adminPayGateways.feeModeInclude'), value: 'include' },
]

const payTypeMap: Record<string, string> = {
  alipay: t('recharge.alipay'),
  wxpay: t('recharge.wechatPay'),
  qqpay: t('recharge.qqWallet'),
  bank: t('recharge.bankCard'),
  jdpay: t('recharge.jdPay'),
}

const columns: DataTableColumns<PayGateway> = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: t('adminPayGateways.logo'),
    key: 'logo_url',
    width: 60,
    render: (row) => {
      if (row.logo_url) {
        return h(NImage, { src: row.logo_url, width: 32, height: 32, objectFit: 'contain', fallbackSrc: '', style: { borderRadius: '4px' } })
      }
      return h('span', { style: { color: '#999', fontSize: '12px' } }, t('recharge.none'))
    },
  },
  {
    title: t('adminPayGateways.gatewayName'),
    key: 'name',
    width: 140,
    ellipsis: { tooltip: true },
  },
  {
    title: t('recharge.paymentMethod'),
    key: 'pay_type',
    width: 90,
    render: (row) => payTypeMap[row.pay_type] || row.pay_type,
  },
  {
    title: t('adminUsers.status'),
    key: 'status',
    width: 70,
    render: (row) => h(NTag, { type: row.status === 1 ? 'success' : 'default', size: 'small', bordered: false }, () => row.status === 1 ? t('adminUsers.enabled') : t('adminUsers.disabled')),
  },
  {
    title: t('adminPayGateways.amountRange'),
    key: 'amount_range',
    width: 140,
    render: (row) => `¥${row.min_amount} - ¥${row.max_amount}`,
  },
  {
    title: t('adminPayGateways.fee'),
    key: 'fee_rate',
    width: 80,
    render: (row) => row.fee_rate > 0 ? `${row.fee_rate}%` : t('recharge.none'),
  },
  {
    title: t('adminPayGateways.minLevel'),
    key: 'min_level',
    width: 80,
    render: (row) => row.min_level > 0 ? `Lv.${row.min_level}` : t('adminPayGateways.unlimited'),
  },
  {
    title: t('adminPayGateways.sortOrder'),
    key: 'sort_order',
    width: 60,
  },
  {
    title: t('moneyScore.actions'),
    key: 'actions',
    width: 140,
    render: (row) => {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, type: 'primary', onClick: () => handleEdit(row) }, () => t('adminUsers.edit')),
        h(NButton, { size: 'small', quaternary: true, type: 'error', onClick: () => handleDelete(row) }, () => t('adminUsers.delete')),
      ])
    },
  },
]

 const selectableColumnOptions = [
   { key: 'id', label: 'ID' },
   { key: 'logo_url', label: t('adminPayGateways.logo') },
   { key: 'name', label: t('adminPayGateways.gatewayName') },
   { key: 'pay_type', label: t('recharge.paymentMethod') },
   { key: 'status', label: t('adminUsers.status') },
   { key: 'amount_range', label: t('adminPayGateways.amountRange') },
   { key: 'fee_rate', label: t('adminPayGateways.fee') },
   { key: 'min_level', label: t('adminPayGateways.minLevel') },
   { key: 'sort_order', label: t('adminPayGateways.sortOrder') },
 ]

 const {
   columnOptions,
   selectedColumnKeys,
   visibleColumns,
   visibleColumnCount,
   totalColumnCount,
   tableScrollX,
   resetSelectedColumns,
 } = useTableColumnVisibility<PayGateway>({
   storageKey: 'admin-pay-gateways-list',
   columns,
   options: selectableColumnOptions,
   minVisibleCount: 1,
   minScrollX: 980,
 })

async function loadList() {
  loading.value = true
  try {
    const res = await fetchPayGateways({ page: pagination.page, page_size: pagination.pageSize, keyword: keyword.value })
    if (res.isSuccess) {
      list.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    }
  } catch {
    message.error(t('adminPayGateways.fetchListFailed'))
  } finally {
    loading.value = false
  }
}

function handleCreate() {
  editingId.value = null
  Object.assign(form, defaultForm())
  showModal.value = true
}

function handleEdit(row: PayGateway) {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    type: row.type,
    pay_type: row.pay_type,
    description: row.description,
    status: row.status,
    api_url: row.api_url,
    pid: row.pid,
    key: row.key,
    logo_url: row.logo_url,
    sort_order: row.sort_order,
    min_amount: row.min_amount,
    max_amount: row.max_amount,
    fee_rate: row.fee_rate,
    fee_mode: row.fee_mode || 'add',
    min_level: row.min_level,
    notify_url: row.notify_url,
  })
  showModal.value = true
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    if (editingId.value) {
      const res = await updatePayGateway(editingId.value, form)
      if (res.isSuccess) {
        message.success(t('adminUsers.updateSuccess'))
        showModal.value = false
        loadList()
      } else {
        message.error(res.message || t('adminUsers.updateFailed'))
      }
    } else {
      const res = await createPayGateway(form)
      if (res.isSuccess) {
        message.success(t('adminUsers.createSuccess'))
        showModal.value = false
        loadList()
      } else {
        message.error(res.message || t('adminUsers.createFailed'))
      }
    }
  } catch {
    message.error(t('adminUsers.operationFailed'))
  } finally {
    submitting.value = false
  }
}

function handleDelete(row: PayGateway) {
  dialog.warning({
    title: t('adminUsers.delete'),
    content: t('adminPayGateways.confirmDeleteContent', { name: row.name }),
    positiveText: t('adminUsers.delete'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        const res = await deletePayGateway(row.id)
        if (res.isSuccess) {
          message.success(t('adminUsers.deleteSuccess'))
          loadList()
        } else {
          message.error(res.message || t('adminUsers.deleteFailed'))
        }
      } catch {
        message.error(t('adminUsers.deleteFailed'))
      }
    },
  })
}

onMounted(() => {
  loadList()
})
</script>
