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
