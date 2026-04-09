import { adminPaymentApi } from './payment'

export async function fetchAllPayOrders(params: { page?: number, page_size?: number, user_id?: number, status?: number, keyword?: string }) {
  return adminPaymentApi.listOrders(params) as any
}
