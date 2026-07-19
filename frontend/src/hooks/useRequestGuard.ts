/**
 * 列表/搜索请求竞态保护：快速翻页或连续搜索时，丢弃过期响应，避免旧数据覆盖新结果。
 *
 * 用法：
 *   const { begin, isLatest } = useRequestGuard()
 *   async function fetchList() {
 *     const token = begin()
 *     loading.value = true
 *     try {
 *       const res = await api.list(...)
 *       if (!isLatest(token)) return
 *       list.value = res.data.list
 *     } finally {
 *       if (isLatest(token)) loading.value = false
 *     }
 *   }
 */
export function useRequestGuard() {
  let seq = 0

  /** 发起新请求前调用，返回本次请求 token */
  function begin() {
    seq += 1
    return seq
  }

  /** 响应回来时判断是否仍是最新一次请求 */
  function isLatest(token: number) {
    return token === seq
  }

  return { begin, isLatest }
}
