/**
 * 导入导出 URL 辅助（实际下载在页面用 fetch 流式拉 CSV）
 */
import { getAdminApiBase } from './base'

export const adminImportExportApi = {
  usersExportUrl() {
    return `${getAdminApiBase()}/export/users`
  },
  usersTemplateUrl() {
    return `${getAdminApiBase()}/export/users/template`
  },
}
