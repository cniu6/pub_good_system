<template>
  <n-space vertical :size="16">
    <n-card title="支付通道管理">
      <template #header-extra>
        <n-button type="primary" @click="handleCreate">
          <template #icon><n-icon><icon-park-outline-add-one /></n-icon></template>
          新增通道
        </n-button>
      </template>

      <n-space vertical>
        <n-space>
          <n-input v-model:value="keyword" placeholder="搜索通道名称" clearable style="width: 220px" @keyup.enter="loadList" />
          <n-button type="primary" @click="loadList">搜索</n-button>
        </n-space>

        <n-data-table
          :columns="columns"
          :data="list"
          :loading="loading"
          :pagination="pagination"
          striped
          size="small"
          @update:page="(p: number) => { pagination.page = p; loadList() }"
          @update:page-size="(s: number) => { pagination.pageSize = s; pagination.page = 1; loadList() }"
        />
      </n-space>
    </n-card>

    <!-- 新增/编辑弹窗 -->
    <n-modal v-model:show="showModal" preset="card" :title="editingId ? '编辑支付通道' : '新增支付通道'" style="width: 680px" :mask-closable="false">
      <n-form ref="formRef" :model="form" :rules="formRules" label-placement="left" label-width="100">
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item label="通道名称" path="name">
              <n-input v-model:value="form.name" placeholder="如：支付宝快捷" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="通道类型" path="type">
              <n-select v-model:value="form.type" :options="typeOptions" placeholder="选择通道类型" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="支付方式" path="pay_type">
              <n-select v-model:value="form.pay_type" :options="payTypeOptions" placeholder="选择支付方式" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="状态" path="status">
              <n-switch v-model:value="form.status" :checked-value="1" :unchecked-value="0">
                <template #checked>启用</template>
                <template #unchecked>禁用</template>
              </n-switch>
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="描述/提示" path="description">
              <n-input v-model:value="form.description" type="textarea" placeholder="显示在用户端的通道描述信息" :rows="2" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="API 地址" path="api_url">
              <n-input v-model:value="form.api_url" placeholder="易支付网关地址，如 https://pay.example.com/" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="商户ID" path="pid">
              <n-input v-model:value="form.pid" placeholder="商户PID" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="商户密钥" path="key">
              <n-input v-model:value="form.key" type="password" show-password-on="click" placeholder="商户Key" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="Logo地址" path="logo_url">
              <n-input v-model:value="form.logo_url" placeholder="通道Logo图片URL（可选）" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="最低金额" path="min_amount">
              <n-input-number v-model:value="form.min_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>¥</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="最高金额" path="max_amount">
              <n-input-number v-model:value="form.max_amount" :min="0" :precision="2" style="width: 100%">
                <template #prefix>¥</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="手续费率" path="fee_rate">
              <n-input-number v-model:value="form.fee_rate" :min="0" :max="100" style="width: 100%">
                <template #suffix>%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="手续费模式" path="fee_mode">
              <n-select v-model:value="form.fee_mode" :options="feeModeOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="最低等级" path="min_level">
              <n-input-number v-model:value="form.min_level" :min="0" style="width: 100%" placeholder="0=不限制" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="排序" path="sort_order">
              <n-input-number v-model:value="form.sort_order" :min="0" style="width: 100%" placeholder="越小越靠前" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="回调地址" path="notify_url">
              <n-input v-model:value="form.notify_url" placeholder="自定义回调地址（留空使用全局）" />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h } from 'vue'
import { useMessage, useDialog, NTag, NButton, NSpace, NImage } from 'naive-ui'
import type { DataTableColumns, FormRules } from 'naive-ui'
import {
  fetchPayGateways,
  createPayGateway,
  updatePayGateway,
  deletePayGateway,
} from '@/service/api/admin/paygateway'
import type { PayGateway, PayGatewayCreateRequest } from '@/service/api/admin/paygateway'

const message = useMessage()
const dialog = useDialog()

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
  name: [{ required: true, message: '请输入通道名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择通道类型', trigger: 'change' }],
  pay_type: [{ required: true, message: '请选择支付方式', trigger: 'change' }],
}

const typeOptions = [
  { label: '易支付 (Epay)', value: 'epay' },
]

const payTypeOptions = [
  { label: '支付宝', value: 'alipay' },
  { label: '微信支付', value: 'wxpay' },
  { label: 'QQ钱包', value: 'qqpay' },
  { label: '银行卡', value: 'bank' },
  { label: '京东支付', value: 'jdpay' },
]

const feeModeOptions = [
  { label: '加收（用户多付）', value: 'add' },
  { label: '包含（到账减少）', value: 'include' },
]

const payTypeMap: Record<string, string> = {
  alipay: '支付宝',
  wxpay: '微信支付',
  qqpay: 'QQ钱包',
  bank: '银行卡',
  jdpay: '京东支付',
}

const columns: DataTableColumns<PayGateway> = [
  {
    title: 'ID',
    key: 'id',
    width: 60,
  },
  {
    title: 'Logo',
    key: 'logo_url',
    width: 60,
    render: (row) => {
      if (row.logo_url) {
        return h(NImage, { src: row.logo_url, width: 32, height: 32, objectFit: 'contain', fallbackSrc: '', style: { borderRadius: '4px' } })
      }
      return h('span', { style: { color: '#999', fontSize: '12px' } }, '无')
    },
  },
  {
    title: '通道名称',
    key: 'name',
    width: 140,
    ellipsis: { tooltip: true },
  },
  {
    title: '支付方式',
    key: 'pay_type',
    width: 90,
    render: (row) => payTypeMap[row.pay_type] || row.pay_type,
  },
  {
    title: '状态',
    key: 'status',
    width: 70,
    render: (row) => h(NTag, { type: row.status === 1 ? 'success' : 'default', size: 'small', bordered: false }, () => row.status === 1 ? '启用' : '禁用'),
  },
  {
    title: '金额范围',
    key: 'amount_range',
    width: 140,
    render: (row) => `¥${row.min_amount} - ¥${row.max_amount}`,
  },
  {
    title: '手续费',
    key: 'fee_rate',
    width: 80,
    render: (row) => row.fee_rate > 0 ? `${row.fee_rate}%` : '无',
  },
  {
    title: '最低等级',
    key: 'min_level',
    width: 80,
    render: (row) => row.min_level > 0 ? `Lv.${row.min_level}` : '不限',
  },
  {
    title: '排序',
    key: 'sort_order',
    width: 60,
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    render: (row) => {
      return h(NSpace, { size: 4 }, () => [
        h(NButton, { size: 'small', quaternary: true, type: 'primary', onClick: () => handleEdit(row) }, () => '编辑'),
        h(NButton, { size: 'small', quaternary: true, type: 'error', onClick: () => handleDelete(row) }, () => '删除'),
      ])
    },
  },
]

async function loadList() {
  loading.value = true
  try {
    const res = await fetchPayGateways({ page: pagination.page, page_size: pagination.pageSize, keyword: keyword.value })
    if (res.isSuccess) {
      list.value = res.data?.list || []
      pagination.itemCount = res.data?.total || 0
    }
  } catch {
    message.error('获取通道列表失败')
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
        message.success('更新成功')
        showModal.value = false
        loadList()
      } else {
        message.error(res.message || '更新失败')
      }
    } else {
      const res = await createPayGateway(form)
      if (res.isSuccess) {
        message.success('创建成功')
        showModal.value = false
        loadList()
      } else {
        message.error(res.message || '创建失败')
      }
    }
  } catch {
    message.error('操作失败')
  } finally {
    submitting.value = false
  }
}

function handleDelete(row: PayGateway) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除通道「${row.name}」吗？`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await deletePayGateway(row.id)
        if (res.isSuccess) {
          message.success('删除成功')
          loadList()
        } else {
          message.error(res.message || '删除失败')
        }
      } catch {
        message.error('删除失败')
      }
    },
  })
}

onMounted(() => {
  loadList()
})
</script>
