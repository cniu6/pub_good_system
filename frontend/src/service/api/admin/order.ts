import { adminPaymentApi } from './payment'

export async function fetchAllPayOrders(params: { page?: number, page_size?: number, user_id?: number, status?: number, keyword?: string }) {
  const res = await adminPaymentApi.listOrders(params)
  if (!res.isSuccess) {
    return res as any
  }

  const list = (res.data?.list || []).map((o) => {
    return {
      id: o.id,
      // 兼容资料/vue 的字段命名
      out_trade_no: o.order_no,
      amount: o.amount,
      status: String(o.status),
      paygateway: o.payment_channel,
      create_time: o.create_time,
    }
  })

  return {
    ...res,
    data: {
      list,
      total: res.data?.total || 0,
    },
  }
}
