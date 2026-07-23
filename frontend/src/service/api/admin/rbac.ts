/**
 * 管理端 RBAC API（角色/权限列表 + 给用户分配角色）
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

export interface RbacRole {
  id: number
  code: string
  name: string
  description: string
  create_time: number
}

export interface RbacPermission {
  id: number
  code: string
  name: string
  description: string
  create_time: number
}

export const adminRbacApi = {
  listRoles() {
    return request.Get<Service.ResponseResult<{ list: RbacRole[] }>>(`${getAdminApiBase()}/roles`)
  },
  listPermissions() {
    return request.Get<Service.ResponseResult<{ list: RbacPermission[] }>>(`${getAdminApiBase()}/permissions`)
  },
  listUserRoles(userId: number) {
    return request.Get<Service.ResponseResult<{ list: RbacRole[] }>>(`${getAdminApiBase()}/users/${userId}/roles`)
  },
  assignUserRole(userId: number, data: { role_id?: number, role_code?: string }) {
    return request.Post<Service.ResponseResult<{ user_id: number, role: RbacRole }>>(`${getAdminApiBase()}/users/${userId}/roles`, data)
  },
}
