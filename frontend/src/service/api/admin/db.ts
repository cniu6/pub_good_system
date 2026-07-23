/**
 * 管理端数据库控制台 API
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

function dbUrl(path = '') {
  return `${getAdminApiBase()}/db${path}`
}

export interface DbInfo {
  driver: string
  backup_supported: boolean
  write_enabled: boolean
}

export interface DbTableRowsResult {
  table: string
  columns: string[]
  rows: Record<string, unknown>[]
  total: number
  page: number
  page_size: number
}

export interface DbSqlResult {
  columns: string[]
  rows: Record<string, unknown>[]
  row_count?: number
  rows_affected?: number
  truncated?: boolean
  duration_ms?: number
}

export interface DbColumnMeta {
  name: string
  type: string
  nullable: boolean
  default_value: string
  comment: string
  primary_key: boolean
  auto_increment: boolean
}

export interface DbIndexMeta {
  name: string
  columns: string[]
  unique: boolean
  primary_key: boolean
}

export interface DbForeignKeyMeta {
  name: string
  columns: string[]
  ref_table: string
  ref_columns: string[]
}

export interface DbTableMeta {
  table: string
  comment: string
  columns: DbColumnMeta[]
  indexes: DbIndexMeta[]
  foreign_keys: DbForeignKeyMeta[]
}

export interface DbTableDdlResult {
  table: string
  ddl: string
}

export interface DbRowWriteResult {
  rows_affected: number
}

export const adminDbApi = {
  info() {
    return request.Get<Service.ResponseResult<DbInfo>>(dbUrl('/info'))
  },
  tables() {
    return request.Get<Service.ResponseResult<{ list: string[] }>>(dbUrl('/tables'))
  },
  tableRows(name: string, params?: { page?: number, page_size?: number }) {
    return request.Get<Service.ResponseResult<DbTableRowsResult>>(
      dbUrl(`/tables/${encodeURIComponent(name)}/rows`),
      { params },
    )
  },
  tableMeta(name: string) {
    return request.Get<Service.ResponseResult<DbTableMeta>>(
      dbUrl(`/tables/${encodeURIComponent(name)}/meta`),
    )
  },
  tableDdl(name: string) {
    return request.Get<Service.ResponseResult<DbTableDdlResult>>(
      dbUrl(`/tables/${encodeURIComponent(name)}/ddl`),
    )
  },
  createTableRow(name: string, values: Record<string, unknown>) {
    return request.Post<Service.ResponseResult<DbRowWriteResult>>(
      dbUrl(`/tables/${encodeURIComponent(name)}/rows`),
      { values },
    )
  },
  updateTableRow(name: string, primaryKey: Record<string, unknown>, values: Record<string, unknown>) {
    return request.Patch<Service.ResponseResult<DbRowWriteResult>>(
      dbUrl(`/tables/${encodeURIComponent(name)}/rows`),
      { primary_key: primaryKey, values },
    )
  },
  deleteTableRow(name: string, primaryKey: Record<string, unknown>) {
    return request.Delete<Service.ResponseResult<DbRowWriteResult>>(
      dbUrl(`/tables/${encodeURIComponent(name)}/rows`),
      { primary_key: primaryKey },
    )
  },
  execSql(data: { sql: string, allow_write?: boolean }) {
    return request.Post<Service.ResponseResult<DbSqlResult>>(dbUrl('/sql'), {
      sql: data.sql,
      allow_write: Boolean(data.allow_write),
    })
  },
  /** 返回备份下载完整 URL（需自行带 Authorization 拉文件） */
  backupUrl() {
    return dbUrl('/backup')
  },
}
