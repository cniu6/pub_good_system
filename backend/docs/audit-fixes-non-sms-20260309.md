# FST 非短信审计修复记录

日期：2026-03-09

## 本轮修复范围

本轮仅处理非短信问题，覆盖：

- 嵌入模式启动初始化不一致
- 系统设置分类与运行时配置同步不一致
- 注册开关未在后端生效
- 注册验证码过早消费
- `AppMode` 直接字符串判断导致的环境识别不一致
- 操作日志缺少用户名
- 用户中心静默失败
- 废弃认证控制器与新控制器逻辑漂移

未处理范围：短信发送能力本身的缺口与 Provider 未实现问题。

## 已修改文件

### 1. `backend/cmd/main_embedded.go`

修复内容：

- 补齐 `models.InitSystemSettingsTable()`
- 补齐 `services.InitSettingsService()`
- 嵌入模式插件初始化改为与主启动同一套 `plugins.Manager` 流程
- 保持 `demo` 插件在嵌入模式下可注册，避免编译期未使用导入

效果：

- `embedded` 模式下不再缺失 `system_settings` 初始化
- 数据库中的系统设置缓存可在嵌入模式正常加载
- 插件加载路径与普通启动更加一致

### 2. `backend/app/controllers/admin/settings_controller.go`

修复内容：

- `CategoryLabelMap` 增加 `payment`
- `isValidCategory` 增加 `payment`
- 创建配置时的分类错误提示补齐 `payment` / `sms`

效果：

- 默认系统设置中的 `payment` 分类可以在后台被完整识别与展示
- 分类校验与默认数据保持一致

### 3. `backend/app/services/settings_service.go`

修复内容：

- `InitSettingsService()` 在刷新缓存后同步运行时配置
- 新增 `ApplyGlobalRuntimeConfig()`，将数据库设置同步到 `config.GlobalConfig`

同步项包括：

- Geetest 运行时配置
- 邮箱/短信验证码开关
- 短信服务基础配置

效果：

- 启动时即可把数据库配置同步到运行时配置对象
- 避免只加载缓存、不更新运行时配置造成的行为不一致

### 4. `backend/app/services/sms_service.go`

修复内容：

- `InitSMSService()` 初始化后立即读取当前有效配置并 `SetConfig`

效果：

- 启动后 `GlobalSMSService` 可立即使用数据库中的当前配置
- 不再依赖管理员再次保存设置才能把运行时短信配置刷新进去

说明：

- 这里修的是“启动配置同步”问题，不涉及短信 Provider 能力本身的实现

### 5. `backend/app/controllers/public/auth_controller.go`

修复内容：

- 新增 `registrationAllowed()`
- 新增 `isNonProductionMode()`
- `Register` 增加 `allow_register` 后端校验
- `SendRegisterCode` 增加 `allow_register` 后端校验
- 将注册验证码消费移动到用户名/邮箱/极验校验之后
- 注册成功后再清理该邮箱注册验证码
- 登录失败调试输出改用统一非生产环境判断
- 登录成功后创建 `user_sessions` 失败时不再静默吞错，而是返回错误
- 重置密码邮件相关非生产分支改用统一环境判断

效果：

- 当前端显示“禁止注册”时，后端也会真正拒绝注册与注册验证码申请
- 用户因为用户名格式、邮箱重复、极验失败等本地校验失败时，不会白白消耗验证码
- `APP_MODE=prod` 不再被误判为非生产
- 避免返回一个无法通过会话校验的“假成功登录”结果

### 6. `backend/app/controllers/auth_controller.go`

修复内容：

- 对废弃认证控制器补充 `registrationAllowed()` 与 `isNonProductionMode()`
- 同步注册开关校验
- 同步注册验证码消费顺序
- 同步环境判断逻辑

效果：

- 即使该控制器当前已废弃，也不会继续与 `public.AuthController` 产生明显逻辑漂移
- 降低未来误接回旧控制器时引入回归的风险

### 7. `backend/internal/middleware/auth.go`

修复内容：

- 认证通过后额外查询用户，并把 `username` 放入 Gin Context

效果：

- 下游操作日志中间件可以获取到用户名
- 已认证操作日志不再只有 `user_id` 没有 `username`

### 8. `backend/app/controllers/user/profile_controller.go`

修复内容：

- `UpdateSettings` 中更新语言失败时改为返回错误
- `GetSessions` 查询失败时改为返回错误，不再伪装成空数组成功

效果：

- 用户设置更新失败不再静默
- 会话列表数据库异常时可被正确暴露，便于排查

## 验证情况

执行：

```bash
go test ./...
```

结果：

- `app/controllers`、`app/models`、`app/services`、`internal/*`、`routes` 等包可通过当前代码层编译
- 失败点集中在 `fst/backend/cmd`

失败原因：

- 启动入口测试/构建过程中会触发配置与数据库初始化
- 当前本地环境未提供可用 MySQL 凭据，报错为：`Access denied for user 'root'@'localhost'`

结论：

- 本轮修改未暴露新的业务包编译错误
- 当前剩余阻塞是本地数据库环境依赖，不是这批补丁本身的语法/类型错误

## 本轮未处理项

- 短信 Provider 的真实实现缺失（Aliyun / Tencent 仍为 TODO）
- 短信发送失败仍返回成功的业务问题未处理（按本轮范围排除）

## 建议后续动作

- 在具备测试数据库配置的环境下，再执行一次完整 `go test ./...`
- 若后续确认彻底不再使用旧认证控制器，可进一步删除或瘦身 `backend/app/controllers/auth_controller.go`
