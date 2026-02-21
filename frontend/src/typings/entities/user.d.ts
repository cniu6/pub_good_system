/// <reference path="../global.d.ts"/>

/** 用户数据库表字段 */
namespace Entity {
  interface User {
    /** 用户id */
    id?: number
    /** 分组id */
    groupId?: number
    /** 用户名 */
    userName?: string
    /** 昵称 */
    nickname?: string
    /** 邮箱 */
    email?: string
    /** 手机号 */
    mobile?: string
    /** 头像 */
    avatar?: string
    /** 背景图 */
    backGround?: string
    /** 性别: 0=未知, 1=男, 2=女 */
    gender?: 0 | 1 | 2
    /** 生日时间戳 */
    birthday?: number | null
    /** 余额 */
    money?: number
    /** 积分 */
    score?: number
    /** 等级 */
    level?: number
    /** 角色 */
    role?: Entity.RoleType
    /** 上次登录时间 */
    lastLoginTime?: number | null
    /** 上次登录IP */
    lastLoginIp?: string
    /** 登录失败次数 */
    loginFailure?: number
    /** 加入IP */
    joinIp?: string
    /** 加入时间 */
    joinTime?: number | null
    /** 签名 */
    motto?: string
    /** 状态: 1=启用, 0=禁用 */
    status?: 0 | 1
    /** API密钥 */
    apikey?: string | null
    /** 偏好语言 */
    language?: string
    /** 国家 */
    country?: string
    /** 当前Token */
    token?: string
    /** 更新时间 */
    updateTime?: number | null
    /** 创建时间 */
    createTime?: number | null
  }

}
