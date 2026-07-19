/**
 * Admin SMS template API.
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

function baseUrl() { return `${getAdminApiBase()}/sms-templates` }

export interface SMSTemplate {
  id: number
  name: string
  lang: string
  sign_name: string
  content: string
  description: string
  variables: string
  status: number
  created_at: string
  updated_at: string
}

export const adminSMSTemplateApi = {
  list() {
    return request.Get<Service.ResponseResult<SMSTemplate[]>>(baseUrl())
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<SMSTemplate>>(`${baseUrl()}/${id}`)
  },

  update(id: number, data: {
    sign_name?: string
    content: string
    description?: string
    status?: number
  }) {
    return request.Put<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/${id}`, data)
  },

  preview(id: number, data: {
    content: string
    vars?: Record<string, any>
  }) {
    return request.Post<Service.ResponseResult<{ content: string; sign_name: string }>>(`${baseUrl()}/${id}/preview`, data)
  },

  reset(id: number) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${baseUrl()}/${id}/reset`, {})
  },
}

export const fetchSMSTemplateList = () => adminSMSTemplateApi.list()
export const fetchUpdateSMSTemplate = (id: number, data: Parameters<typeof adminSMSTemplateApi.update>[1]) => adminSMSTemplateApi.update(id, data)
export const fetchPreviewSMSTemplate = (id: number, data: Parameters<typeof adminSMSTemplateApi.preview>[1]) => adminSMSTemplateApi.preview(id, data)
export const fetchResetSMSTemplate = (id: number) => adminSMSTemplateApi.reset(id)
