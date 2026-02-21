# 管理端页面 JS 分离实现说明

## 概述

实现了管理端页面的 JS 代码物理隔离，确保普通用户无法通过查看前端源码获取管理端路由和代码结构，提升系统安全性。

## 核心安全机制

### 1. JS 代码物理隔离

通过 Vite 打包配置，将管理端相关代码分离到独立的 chunk：

- **管理端视图组件** → `assets/m/admin-views-[hash].js`
- **管理端路由配置** → `assets/m/admin-core-[hash].js`
- **管理端 API 服务** → `assets/m/admin-api-[hash].js`

### 2. 动态路径配置

管理端路径通过环境变量 `VITE_ADMIN_PATH` 配置，每次打包可以修改，防止被猜测：

```env
# 前端 .env
VITE_ADMIN_PATH=/system-mgr

# 后端 .env
ADMIN_PATH=/system-mgr
```

### 3. 延迟加载机制

管理端路由只有在以下条件满足时才会动态加载：

1. 用户已登录（有 accessToken）
2. 用户角色为管理员（role === 'admin'）
3. 通过路由守卫验证

## 实现细节

### 文件结构

```
frontend/src/
├── router/
│   ├── admin.routes.ts      # 管理端路由定义（会被打包到独立 chunk）
│   ├── admin.loader.ts      # 管理端路由动态加载器
│   └── guard.ts             # 路由守卫（验证管理员权限）
├── views/admin/             # 管理端视图（会被打包到独立 chunk）
│   ├── dashboard/
│   └── users/
└── service/api/admin/       # 管理端 API（会被打包到独立 chunk）
    └── user.ts
```

### 关键代码

#### 1. Vite 打包配置 (`vite.config.ts`)

```typescript
manualChunks(id) {
  // 管理端视图组件 -> admin-views chunk
  if (id.includes('src/views/admin')) {
    return 'admin-views'
  }
  // 管理端路由配置 -> admin-core chunk
  if (id.includes('src/router/admin.routes')) {
    return 'admin-core'
  }
  // 管理端 API 服务 -> admin-api chunk
  if (id.includes('src/service/api/admin')) {
    return 'admin-api'
  }
}
```

#### 2. 路由动态加载 (`store/router/index.ts`)

```typescript
async initAuthRoute() {
  // ... 初始化普通路由
  
  // 检查用户是否为管理员
  const role = local.get('role')
  const hasAdminRole = roles.includes('admin') || role === 'admin'

  if (hasAdminRole) {
    // 动态加载管理端路由（会被打包到独立的 chunk）
    const adminRoutes = await loadAdminRoutes()
    adminRoutes.forEach(route => {
      router.addRoute(route)
    })
  }
}
```

#### 3. 路由守卫 (`router/guard.ts`)

```typescript
// 判断是否是管理端路由
const adminPath = import.meta.env.VITE_ADMIN_PATH || '/admin'
const isAdminRoute = to.path.startsWith(adminPath)

// 处理管理端路由访问权限
if (isAdminRoute && (!isLogin || !hasAdminRole)) {
  next({ path: '/login', query: { redirect: to.fullPath } })
  return
}
```

## 当前实现的管理端功能

### 已实现页面

1. **仪表盘** (`/admin/dashboard`)
   - 管理端首页

2. **用户管理** (`/admin/users`)
   - 用户列表（分页、搜索、筛选）
   - 创建用户
   - 编辑用户
   - 删除用户
   - 查看用户详情

### API 接口

管理端 API 统一使用 `/api/v1/admin` 前缀：

- `GET /api/v1/admin/users` - 用户列表
- `GET /api/v1/admin/users/:id` - 用户详情
- `POST /api/v1/admin/users` - 创建用户
- `PUT /api/v1/admin/users/:id` - 更新用户
- `DELETE /api/v1/admin/users/:id` - 删除用户
- `PUT /api/v1/admin/users/:id/status` - 更新用户状态

## 使用说明

### 1. 环境变量配置

**前端配置** (`frontend/.env`):

```env
VITE_ADMIN_PATH=/system-mgr
VITE_BASE_URL=/
```

**后端配置** (`.env`):

```env
ADMIN_PATH=/system-mgr
```

> ⚠️ **安全提示**: 每次部署时建议修改 `VITE_ADMIN_PATH` 和 `ADMIN_PATH` 的值，使用不易猜测的路径。

### 2. 打包验证

打包后检查 `dist/assets/m/` 目录，应该包含：

- `admin-views-[hash].js`
- `admin-core-[hash].js`
- `admin-api-[hash].js`

这些文件只有在管理员登录后才会被加载。

### 3. 访问管理端

1. 使用管理员账号登录
2. 访问管理端路径（如 `/system-mgr`）
3. 系统会自动加载管理端路由和组件

## 安全优势

1. **代码隔离**: 普通用户无法通过查看 JS 源码获取管理端路由信息
2. **路径隐藏**: 管理端路径可通过环境变量动态配置
3. **延迟加载**: 只有管理员登录后才加载管理端代码，减少初始包体积
4. **权限验证**: 多层权限验证（路由守卫 + 动态加载）

## 后续扩展

如需添加新的管理端页面：

1. 在 `frontend/src/views/admin/` 创建页面组件
2. 在 `frontend/src/router/admin.routes.ts` 添加路由配置
3. 在 `frontend/src/service/api/admin/` 添加 API 服务（如需要）

所有管理端相关代码会自动被打包到独立的 chunk 中。

## 注意事项

1. **环境变量同步**: 前端和后端的 `ADMIN_PATH` 必须保持一致
2. **路径修改**: 修改 `VITE_ADMIN_PATH` 后需要重新打包前端
3. **权限检查**: 确保后端 API 也有相应的管理员权限验证
4. **路由命名**: 管理端路由名称建议使用 `admin-` 前缀，便于识别

