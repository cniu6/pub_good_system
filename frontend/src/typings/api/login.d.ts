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
      /**
       * 本次会话签发的 JWT auth_guard（user/admin）。
       * 必须与后端 LoginResult.authGuard 一致；刷新与管理端准入以此为准，不可用 role 猜测。
       */
      authGuard: Entity.AuthGuardType
      /** 访问token */
      accessToken: string
      /** 访问token */
      refreshToken: string
      /** 实名认证摘要 */
      realname?: RealnameSummary
    }
  }
}
