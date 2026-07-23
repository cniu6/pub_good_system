/**
 * 管理端调试终端 API（HTTP 回退；仅 debug ops 开启时可用）
 */
import { request } from '../../http'
import { getAdminApiBase } from './base'

export interface TerminalExecResult {
  output: string
  error?: string
  success: boolean
}

export const adminTerminalApi = {
  exec(cmd: string) {
    return request.Post<Service.ResponseResult<TerminalExecResult>>(
      `${getAdminApiBase()}/debug/terminal/exec`,
      { cmd },
    )
  },
}
