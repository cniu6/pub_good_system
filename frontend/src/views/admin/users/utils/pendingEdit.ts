/** 详情页 → 列表页「打开编辑弹窗」的跨页意图（不用 query，避免 fullPath 变化导致组件重建） */
export const ADMIN_USERS_PENDING_EDIT_KEY = 'admin-users-pending-edit-id'

export function setPendingUserEditId(id: number | string) {
  sessionStorage.setItem(ADMIN_USERS_PENDING_EDIT_KEY, String(id))
}

export function takePendingUserEditId(): number | null {
  const raw = sessionStorage.getItem(ADMIN_USERS_PENDING_EDIT_KEY)
  if (!raw)
    return null
  sessionStorage.removeItem(ADMIN_USERS_PENDING_EDIT_KEY)
  const id = Number(raw)
  return !id || Number.isNaN(id) ? null : id
}
