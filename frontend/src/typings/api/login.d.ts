/// <reference path="../global.d.ts"/>

namespace Api {
  namespace Login {
    interface RealnameSummary {
      hasVerification: boolean
      id?: number
      status?: 0 | 1 | 2
      realName?: string
      certificateType?: 1 | 2 | 3
      certificateNo?: string
      submittedAt?: number | null
      reviewedAt?: number | null
      rejectReason?: string
    }

    /* 登录返回的用户字段, 该数据是根据用户表扩展而来, 部分字段可能需要覆盖，例如id */
    interface Info extends Entity.User {
      /** 用户id */
      id: number
      /** 用户角色类型 */
      role: Entity.RoleType[]
      /** 访问token */
      accessToken: string
      /** 访问token */
      refreshToken: string
      /** 实名认证摘要 */
      realname?: RealnameSummary
      /** 管理端 TOTP 第二步：为 true 时仅返回 temp_token，尚未正式登录 */
      need_totp?: boolean
      /** TOTP 临时令牌（短时有效） */
      temp_token?: string
    }
  }
}
