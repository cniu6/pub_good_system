<script setup lang="ts">
import { computed, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { NButton, NSpace, type DataTableColumns } from 'naive-ui'
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import { useTableColumnVisibility } from '@/hooks'
import type { SettingDTO } from '@/service/api/admin/settings'
import { addFormRef } from '../composables/settingsState'
import { useAdminSettings } from '../composables/useAdminSettings'

const { t } = useI18n()
const {
  customSettings,
  showAddModal,
  showEditModal,
  addForm,
  editForm,
  addFormRules,
  typeOptions,
  adding,
  savingEdit,
  handleAddSetting,
  handleDeleteSetting,
  handleEditSetting,
  handleSaveSettingEdit,
} = useAdminSettings()

/** 绑定到共享 addFormRef，供 handleAddSetting 里 validate 使用 */
function bindAddFormRef(inst: unknown) {
  addFormRef.value = inst
}

const customColumns: DataTableColumns<SettingDTO> = [
  { title: t('adminSettings.columnKey'), key: 'key' },
  { title: t('adminSettings.columnLabel'), key: 'label' },
  { title: t('adminSettings.columnValue'), key: 'value', ellipsis: { tooltip: true } },
  { title: t('adminSettings.columnType'), key: 'type', width: 80 },
  {
    title: t('adminSettings.columnPublic'),
    key: 'is_public',
    width: 80,
    render: row => row.is_public ? t('adminSettings.yes') : t('adminSettings.no'),
  },
  {
    title: t('adminSettings.columnActions'),
    key: 'actions',
    width: 180,
    render: (row) => {
      return h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleEditSetting(row),
          }, () => t('adminSettings.edit')),
          h(NButton, {
            size: 'small',
            type: 'error',
            text: true,
            onClick: () => handleDeleteSetting(row.key),
          }, () => t('adminSettings.delete')),
        ],
      })
    },
  },
]

const customSelectableColumnOptions = computed(() => [
  { key: 'key', label: t('adminSettings.columnKey') },
  { key: 'label', label: t('adminSettings.columnLabel') },
  { key: 'value', label: t('adminSettings.columnValue') },
  { key: 'type', label: t('adminSettings.columnType') },
  { key: 'is_public', label: t('adminSettings.columnPublic') },
])

const {
  columnOptions: customColumnOptions,
  selectedColumnKeys: customSelectedColumnKeys,
  visibleColumns: customVisibleColumns,
  visibleColumnCount: customVisibleColumnCount,
  totalColumnCount: customTotalColumnCount,
  tableScrollX: customTableScrollX,
  resetSelectedColumns: resetCustomSelectedColumns,
} = useTableColumnVisibility<SettingDTO>({
  storageKey: 'admin-settings-custom-list',
  columns: customColumns,
  options: customSelectableColumnOptions,
  minVisibleCount: 1,
  minScrollX: 760,
})
</script>

<template>
  <n-space vertical :size="16">
    <n-space justify="end">
      <TableColumnSelector
        v-model="customSelectedColumnKeys"
        :options="customColumnOptions"
        :visible-count="customVisibleColumnCount"
        :total-count="customTotalColumnCount"
        :button-label="t('common.showFields')"
        :title="t('common.visibleFields')"
        :hint="t('common.columnVisibilityHint')"
        :reset-label="t('common.restoreDefaultFields')"
        @reset="resetCustomSelectedColumns"
      />
      <n-button type="primary" @click="showAddModal = true">
        <template #icon>
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" style="width: 1em; height: 1em;">
            <path d="M11 11V5h2v6h6v2h-6v6h-2v-6H5v-2z" />
          </svg>
        </template>
        {{ t('adminSettings.addConfigItem') }}
      </n-button>
    </n-space>

    <n-data-table :columns="customVisibleColumns" :data="customSettings" :pagination="false" :bordered="false" :scroll-x="customTableScrollX" />

    <n-modal v-model:show="showAddModal" preset="card" :title="t('adminSettings.addConfigTitle')" style="width: 500px;" :mask-closable="false">
      <n-form :ref="bindAddFormRef" :model="addForm" :rules="addFormRules" label-placement="left" label-width="100px">
        <n-form-item :label="t('adminSettings.configKey')" path="key">
          <n-input v-model:value="addForm.key" :placeholder="t('adminSettings.configKeyPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configValue')" path="value">
          <n-input v-model:value="addForm.value" :placeholder="t('adminSettings.configValuePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configLabel')" path="label">
          <n-input v-model:value="addForm.label" :placeholder="t('adminSettings.configLabelPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configType')" path="type">
          <n-select v-model:value="addForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configDescription')" path="description">
          <n-input v-model:value="addForm.description" :placeholder="t('adminSettings.configDescriptionPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.isPublic')">
          <n-switch v-model:value="addForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">{{ t('adminSettings.publicConfigHint') }}</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="adding" @click="handleAddSetting">{{ t('adminSettings.add') }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showEditModal" preset="card" :title="t('adminSettings.editConfigTitle')" style="width: 520px;" :mask-closable="false">
      <n-form label-placement="left" label-width="100px">
        <n-form-item :label="t('adminSettings.configKey')">
          <n-input v-model:value="editForm.key" disabled />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configValue')">
          <n-input v-model:value="editForm.value" :placeholder="t('adminSettings.configValuePlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configLabel')">
          <n-input v-model:value="editForm.label" :placeholder="t('adminSettings.configLabelPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configType')">
          <n-select v-model:value="editForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.configDescription')">
          <n-input v-model:value="editForm.description" :placeholder="t('adminSettings.configDescriptionPlaceholder')" />
        </n-form-item>
        <n-form-item :label="t('adminSettings.isPublic')">
          <n-switch v-model:value="editForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">{{ t('adminSettings.publicConfigHint') }}</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">{{ t('common.cancel') }}</n-button>
          <n-button type="primary" :loading="savingEdit" @click="handleSaveSettingEdit">{{ t('common.save') }}</n-button>
        </n-space>
      </template>
    </n-modal>
  </n-space>
</template>
