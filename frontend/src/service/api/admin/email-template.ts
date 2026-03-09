/**
 * Admin email template API.
 */
import { request } from '../../http'

// Admin API path is fixed at /admin to match backend routes.
const ADMIN_PATH = '/admin'
const BASE_URL = `/api/v1${ADMIN_PATH}/email-templates`

export interface EmailTemplate {
  id: number
  name: string
  lang: string
  title: string
  subject: string
  content: string
  description: string
  variables: string
  status: number
  created_at: string
  updated_at: string
}

export const adminEmailTemplateApi = {
  list() {
    return request.Get<Service.ResponseResult<EmailTemplate[]>>(BASE_URL)
  },

  detail(id: number) {
    return request.Get<Service.ResponseResult<EmailTemplate>>(`${BASE_URL}/${id}`)
  },

  update(id: number, data: {
    subject: string
    content: string
    description?: string
    status?: number
  }) {
    return request.Put<Service.ResponseResult<{ message: string }>>(`${BASE_URL}/${id}`, data)
  },

  preview(id: number, data: {
    content: string
    vars?: Record<string, any>
  }) {
    return request.Post<Service.ResponseResult<{ subject: string; content: string; wrapped: string }>>(`${BASE_URL}/${id}/preview`, data)
  },

  reset(id: number) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`${BASE_URL}/${id}/reset`, {})
  },

  sendTest(data: { to: string; subject?: string; content?: string; template_id?: number }) {
    return request.Post<Service.ResponseResult<{ message: string }>>(`/api/v1${ADMIN_PATH}/email-send-test`, data)
  },
}

// Backward-compatible exports.
export const fetchEmailTemplateList = () => adminEmailTemplateApi.list()
export const fetchEmailTemplateDetail = (id: number) => adminEmailTemplateApi.detail(id)
export const fetchUpdateEmailTemplate = (id: number, data: Parameters<typeof adminEmailTemplateApi.update>[1]) => adminEmailTemplateApi.update(id, data)
export const fetchPreviewEmailTemplate = (id: number, data: Parameters<typeof adminEmailTemplateApi.preview>[1]) => adminEmailTemplateApi.preview(id, data)
export const fetchResetEmailTemplate = (id: number) => adminEmailTemplateApi.reset(id)
