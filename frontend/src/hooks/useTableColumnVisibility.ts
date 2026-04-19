import { computed, unref, watch } from 'vue'
import type { MaybeRef } from 'vue'
import { useLocalStorage } from '@vueuse/core'
import type { DataTableColumns } from 'naive-ui'

export interface TableColumnVisibilityOption {
  key: string
  label: string
  defaultVisible?: boolean
}

function resolveColumnKey(column: unknown) {
  const key = (column as { key?: string | number })?.key
  if (typeof key === 'string' || typeof key === 'number')
    return String(key)
  return ''
}

function isSameKeyList(left: string[], right: string[]) {
  if (left.length !== right.length)
    return false
  return left.every((item, index) => item === right[index])
}

function resolveColumnWidth(column: unknown, defaultColumnWidth: number) {
  const widthValue = Number((column as { width?: number | string }).width)
  if (Number.isFinite(widthValue) && widthValue > 0)
    return widthValue

  const minWidthValue = Number((column as { minWidth?: number | string }).minWidth)
  if (Number.isFinite(minWidthValue) && minWidthValue > 0)
    return minWidthValue

  return defaultColumnWidth
}

export function useTableColumnVisibility<T>(params: {
  storageKey: string
  columns: MaybeRef<DataTableColumns<T>>
  options: MaybeRef<TableColumnVisibilityOption[]>
  minVisibleCount?: number
  minScrollX?: number
  defaultColumnWidth?: number
  extraScrollX?: number
}) {
  const minVisibleCount = Math.max(0, params.minVisibleCount ?? 1)
  const minScrollX = Math.max(0, params.minScrollX ?? 0)
  const defaultColumnWidth = Math.max(80, params.defaultColumnWidth ?? 140)
  const extraScrollX = Math.max(0, params.extraScrollX ?? 32)
  const prefixedStorageKey = `${import.meta.env.VITE_STORAGE_PREFIX}table-columns:${params.storageKey}`
  const sourceColumns = computed(() => unref(params.columns))
  const sourceOptions = computed(() => unref(params.options))

  const availableKeys = computed(() => sourceOptions.value.map(item => item.key))
  const defaultSelectedKeys = computed(() => {
    const keys = sourceOptions.value
      .filter(item => item.defaultVisible !== false)
      .map(item => item.key)

    if (keys.length)
      return keys

    return sourceOptions.value.slice(0, Math.max(1, minVisibleCount)).map(item => item.key)
  })

  const storedSelectedKeys = useLocalStorage<string[]>(prefixedStorageKey, [...defaultSelectedKeys.value], {
    writeDefaults: true,
  })

  function normalizeSelectedKeys(value: string[] | null | undefined) {
    if (!sourceOptions.value.length)
      return []

    const availableKeySet = new Set(availableKeys.value)
    const normalizedKeys = Array.from(new Set((Array.isArray(value) ? value : defaultSelectedKeys.value).filter(key => availableKeySet.has(key))))

    if (normalizedKeys.length < minVisibleCount) {
      for (const key of defaultSelectedKeys.value) {
        if (!normalizedKeys.includes(key))
          normalizedKeys.push(key)
        if (normalizedKeys.length >= minVisibleCount)
          break
      }
    }

    if (normalizedKeys.length < minVisibleCount) {
      for (const key of availableKeys.value) {
        if (!normalizedKeys.includes(key))
          normalizedKeys.push(key)
        if (normalizedKeys.length >= minVisibleCount)
          break
      }
    }

    return availableKeys.value.filter(key => normalizedKeys.includes(key))
  }

  watch(
    [availableKeys, defaultSelectedKeys],
    () => {
      const normalizedKeys = normalizeSelectedKeys(storedSelectedKeys.value)
      const currentKeys = Array.isArray(storedSelectedKeys.value) ? storedSelectedKeys.value : []
      if (!isSameKeyList(normalizedKeys, currentKeys))
        storedSelectedKeys.value = normalizedKeys
    },
    { immediate: true },
  )

  const selectedColumnKeys = computed<string[]>({
    get: () => normalizeSelectedKeys(storedSelectedKeys.value),
    set: (value) => {
      storedSelectedKeys.value = normalizeSelectedKeys(value)
    },
  })

  const selectedColumnKeySet = computed(() => new Set(selectedColumnKeys.value))

  const visibleColumns = computed(() => sourceColumns.value.filter((column) => {
    const key = resolveColumnKey(column)
    if (!key)
      return true
    if (!availableKeys.value.includes(key))
      return true
    return selectedColumnKeySet.value.has(key)
  }))

  const tableScrollX = computed(() => {
    const width = visibleColumns.value.reduce((totalWidth, column) => totalWidth + resolveColumnWidth(column, defaultColumnWidth), 0)
    return Math.max(minScrollX, width + extraScrollX)
  })

  const columnOptions = computed(() => sourceOptions.value.map((item) => {
    const checked = selectedColumnKeys.value.includes(item.key)
    return {
      ...item,
      disabled: checked && selectedColumnKeys.value.length <= minVisibleCount,
    }
  }))

  function resetSelectedColumns() {
    storedSelectedKeys.value = [...defaultSelectedKeys.value]
  }

  return {
    columnOptions,
    selectedColumnKeys,
    visibleColumns,
    visibleColumnCount: computed(() => selectedColumnKeys.value.length),
    totalColumnCount: computed(() => sourceOptions.value.length),
    tableScrollX,
    resetSelectedColumns,
  }
}
