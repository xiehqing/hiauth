# HiAuth 前端需求文档

## 1. 项目概述

HiAuth 是一个面向后台管理系统的认证与权限管理平台，核心能力包括用户登录、用户管理、角色管理、部门管理、菜单管理、角色授权和系统配置管理。

前端需要提供一套后台管理界面，支持管理员完成账号维护、组织架构维护、角色权限分配、系统参数配置等日常操作，并根据登录用户拥有的菜单和权限动态生成导航与操作入口。

## 2. 建设目标

1. 实现统一登录入口，支持 RSA-OAEP-SHA256 密码加密传输。
2. 登录成功后，根据后端返回的菜单、角色、权限生成当前用户可访问的系统界面。
3. 实现用户、角色、部门、菜单、系统配置的完整管理能力。
4. 支持角色绑定菜单权限，用户通过角色获得可访问菜单和操作权限。
5. 提供清晰的表单校验、错误提示、删除确认和空状态展示。
6. 所有接口返回提示、页面文案、校验提示均以中文展示。

## 3. 技术建议

前端框架建议使用 Vue 3 或 React，配合成熟后台管理组件库实现。

推荐技术组合：

| 类型 | 建议 |
| --- | --- |
| 框架 | Vue 3 + Vite  |
| UI 组件 | Element Plus  |
| 路由 | Vue Router / |
| 状态管理 | Pinia  |
| 请求库 | Axios |
| 加密库 | node-forge |
| 表格 | 组件库 Table |
| 树形控件 | 组件库 Tree |
| 权限控制 | 路由守卫 + 按钮权限指令/组件 |

HTTP 环境下浏览器 `window.crypto.subtle` 不可用，登录密码 RSA-OAEP-SHA256 加密建议使用 `node-forge` 实现。生产环境仍建议启用 HTTPS。

## 4. 全局交互规范

### 4.1 接口基础信息

接口统一前缀：

```text
/api/v1
```

登录相关接口不需要登录态，其余业务接口均需要在请求头中携带 Token：

```http
Authorization: Bearer <accessToken>
```

前端存储 Token 时建议只存储后端返回的 `accessToken` 原值，请求拦截器中统一拼接 `Bearer ` 前缀。

### 4.2 通用响应处理

前端应封装统一请求层，集中处理以下场景：

1. 请求成功：读取响应中的业务数据并返回给页面。
2. 参数错误：展示后端返回的中文错误信息。
3. 未登录或 Token 失效：清理本地登录信息并跳转登录页。
4. 系统异常：展示“系统异常，请稍后重试”或后端返回消息。
5. 删除、新增、编辑成功后：展示成功提示并刷新当前列表。

### 4.3 通用分页参数

支持分页的列表页面统一使用以下查询参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| pageNo | number | 当前页 |
| pageSize | number | 每页条数 |
| keyword | string | 搜索关键字 |
| sortField | string | 排序字段 |
| sortOrder | string | 排序方向 |

部门列表不使用分页。

### 4.4 通用字段

实体通常包含以下基础字段，页面可根据实际响应展示：

| 字段 | 说明 |
| --- | --- |
| id | 主键 ID |
| createdAt | 创建时间 |
| updatedAt | 更新时间 |
| createdBy | 创建人 |
| updatedBy | 更新人 |

## 5. 登录与认证

### 5.1 登录页

页面路径建议：

```text
/login
```

页面元素：

1. 系统名称。
2. 用户名输入框。
3. 密码输入框。
4. 登录按钮。
5. 登录中状态。
6. 错误提示区域。

表单校验：

| 字段 | 校验 |
| --- | --- |
| username | 必填 |
| password | 必填 |

登录流程：

1. 页面初始化时调用 `GET /api/v1/auth/encrypt-config`。
2. 如果返回 `enabled=false`，登录时直接提交明文密码。
3. 如果返回 `enabled=true`，使用返回的 `publicKey` 对密码执行 RSA-OAEP-SHA256 加密。
4. 调用 `POST /api/v1/auth/login`。
5. 登录成功后保存 `accessToken`、用户信息、角色、权限、菜单。
6. 根据菜单生成动态路由和侧边栏。
7. 跳转到首页或第一个可访问菜单。

登录请求：

```json
{
  "username": "admin",
  "password": "加密后的密码或明文密码",
  "device": "web"
}
```

登录成功核心响应：

| 字段 | 说明 |
| --- | --- |
| accessToken | 访问令牌 |
| tokenType | Token 类型，当前为 Bearer |
| user | 当前用户信息 |
| roles | 当前用户角色标识数组 |
| permissions | 当前用户权限标识数组 |
| menus | 当前用户可访问菜单树 |
| department | 当前用户所属部门 |

### 5.2 RSA 前端加密

HTTP 环境下建议使用 `node-forge`：

```js
import forge from 'node-forge'

function normalizePublicKey(publicKey) {
  const key = publicKey.trim()
  if (key.includes('-----BEGIN PUBLIC KEY-----')) {
    return key
  }
  return [
    '-----BEGIN PUBLIC KEY-----',
    key,
    '-----END PUBLIC KEY-----'
  ].join('\n')
}

export function encryptPasswordRSAOAEP(password, publicKey) {
  const publicKeyPem = normalizePublicKey(publicKey)
  const rsaPublicKey = forge.pki.publicKeyFromPem(publicKeyPem)
  const encryptedBytes = rsaPublicKey.encrypt(password, 'RSA-OAEP', {
    md: forge.md.sha256.create(),
    mgf1: {
      md: forge.md.sha256.create()
    }
  })
  return forge.util.encode64(encryptedBytes)
}
```

### 5.3 当前用户

接口：

```text
GET /api/v1/auth/me
```

使用场景：

1. 刷新页面后恢复登录状态。
2. 获取最新用户信息、角色、权限、菜单。
3. 校验 Token 是否仍然有效。

### 5.4 退出登录

接口：

```text
POST /api/v1/auth/logout
```

交互要求：

1. 用户点击退出登录时弹出确认框。
2. 调用接口成功后清理本地 Token、用户信息、路由缓存。
3. 跳转登录页。
4. 接口失败时也可允许本地退出，避免用户被卡住。

## 6. 布局与导航

### 6.1 后台主布局

主布局建议包含：

1. 顶部栏：系统名称、当前用户、部门、退出登录。
2. 侧边栏：根据当前用户 `menus` 动态生成。
3. 内容区：承载业务页面。
4. 面包屑：根据当前路由展示。
5. Tab 页签：可选。

### 6.2 动态菜单

菜单数据来源：

```text
登录响应 menus，树结构
GET /api/v1/auth/me 响应 menus，树结构
```

菜单字段：

| 字段 | 说明 |
| --- | --- |
| id | 菜单 ID |
| type | 类型，0 分组，1 菜单，2 按钮 |
| parentId | 上级菜单 ID，0 表示顶级 |
| name | 菜单名称 |
| route | 前端路由 |
| icon | 图标 |
| sort | 排序 |
| show | 是否显示，0 隐藏，1 显示 |

前端处理规则：

1. 将 `type=0` 的数据作为菜单分组渲染。
2. 只将 `type=1` 且 `show=1` 的数据渲染为可点击菜单。
3. `type=2` 的数据用于按钮权限控制。
4. `show=0` 的菜单不在侧边栏展示，但可以作为可访问路由。
5. 菜单按 `sort` 升序排列。
6. 菜单树根据 `parentId` 组装。

## 7. 首页

登录后的首页建议展示系统概览信息：

1. 当前登录用户。
2. 当前所属部门。
3. 当前拥有角色。
4. 常用入口：用户管理、角色管理、部门管理、菜单管理、系统配置。
5. 系统通知区域，可展示项目说明、版本信息或管理员公告。

如果当前用户没有任何可展示菜单，应展示“暂无可访问菜单，请联系管理员配置权限”。

## 8. 用户管理

页面路径建议：

```text
/system/users
```

### 8.1 列表页

接口：

```text
GET /api/v1/users
```

查询条件：

| 字段 | 类型 | 控件 | 说明 |
| --- | --- | --- | --- |
| keyword | string | 输入框 | 用户名、昵称等关键字 |
| status | number | 下拉框 | 0 禁用，1 启用 |
| roleId | number | 下拉框 | 角色过滤 |
| departmentId | number | 树选择 | 部门过滤，包含子部门用户 |
| pageNo | number | 分页 | 当前页 |
| pageSize | number | 分页 | 每页条数 |

列表字段：

| 字段 | 说明 |
| --- | --- |
| username | 用户名 |
| nickname | 昵称 |
| phone | 手机号 |
| email | 邮箱 |
| department.name | 所属部门 |
| roles | 角色 |
| status | 状态 |
| createdAt | 创建时间 |
| 操作 | 查看、编辑、删除 |

交互要求：

1. 状态用标签展示：启用、禁用。
2. 角色多值用 Tag 展示。
3. 部门过滤使用部门树下拉。
4. 根据部门过滤时，后端会返回该部门及其子部门下的用户。
5. 删除用户为硬删除，操作前必须二次确认。

### 8.2 新增用户

接口：

```text
POST /api/v1/users
```

表单字段：

| 字段 | 控件 | 必填 | 说明 |
| --- | --- | --- | --- |
| username | 输入框 | 是 | 用户名 |
| password | 密码框 | 是 | 初始密码 |
| nickname | 输入框 | 是 | 昵称 |
| phone | 输入框 | 否 | 手机号 |
| email | 输入框 | 否 | 邮箱 |
| status | 单选/开关 | 否 | 默认启用 |
| roleIds | 多选下拉 | 否 | 可绑定多个角色 |
| departmentId | 树选择 | 否 | 单选部门 |

前端校验：

1. 用户名必填。
2. 用户名必须以字母开头。
3. 用户名长度 3-32 位。
4. 用户名仅支持字母、数字、下划线、点、短横线。
5. 密码必填，最小长度默认 8 位，实际以系统配置 `security.password.min_length` 为准。
6. 昵称必填。
7. 邮箱格式合法时再提交。
8. 手机号格式合法时再提交。
9. 部门只能单选。

提交示例：

```json
{
  "username": "zhangsan",
  "password": "12345678",
  "nickname": "张三",
  "phone": "13800000000",
  "email": "zhangsan@example.com",
  "status": 1,
  "roleIds": [1, 2],
  "departmentId": 3
}
```

### 8.3 编辑用户

接口：

```text
PUT /api/v1/users/{id}
```

编辑规则：

1. 用户名允许编辑，但必须满足用户名规则。
2. 密码为空时表示不修改密码。
3. 角色可重新多选。
4. 部门仍然只能单选。
5. 状态可切换启用/禁用。

### 8.4 用户详情

接口：

```text
GET /api/v1/users/{id}
```

详情页或抽屉展示用户基础信息、所属部门、角色、状态、创建时间、更新时间。

## 9. 角色管理

页面路径建议：

```text
/system/roles
```

### 9.1 角色列表

接口：

```text
GET /api/v1/roles
```

查询条件：

| 字段 | 控件 | 说明 |
| --- | --- | --- |
| keyword | 输入框 | 角色名称或标识关键字 |
| builtIn | 下拉框 | 0 自定义角色，1 内置角色 |
| pageNo | 分页 | 当前页 |
| pageSize | 分页 | 每页条数 |

列表字段：

| 字段 | 说明 |
| --- | --- |
| displayName | 角色名称 |
| name | 角色标识 |
| description | 描述 |
| builtIn | 是否内置 |
| createdAt | 创建时间 |
| 操作 | 查看、编辑、授权、删除 |

交互规则：

1. 内置角色使用醒目的标签展示。
2. 内置角色不允许删除，删除按钮置灰或隐藏。
3. 编辑角色时不允许修改 `name`。
4. 创建角色时不支持设置是否内置，默认创建为自定义角色。

### 9.2 新增角色

接口：

```text
POST /api/v1/roles
```

字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| displayName | 是 | 角色名称 |
| name | 是 | 角色标识 |
| description | 否 | 描述 |

### 9.3 编辑角色

接口：

```text
PUT /api/v1/roles/{id}
```

字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| displayName | 是 | 角色名称 |
| description | 否 | 描述 |

角色标识 `name` 不在编辑表单中开放修改。

### 9.4 角色授权

获取角色已授权菜单：

```text
GET /api/v1/roles/{id}/menus
```

保存角色授权菜单：

```text
PUT /api/v1/roles/{id}/menus
```

提交示例：

```json
{
  "menuIds": [1, 2, 3, 10, 11]
}
```

交互要求：

1. 使用菜单树展示所有菜单和按钮权限。
2. 支持父子联动勾选。
3. 支持半选状态。
4. 保存前提示“确认保存当前角色权限配置吗？”。
5. 保存成功后提示“授权成功”。

### 9.5 角色下拉

接口：

```text
GET /api/v1/roles/options
```

使用场景：

1. 用户新增/编辑中的角色多选。
2. 用户列表中的角色过滤。

## 10. 部门管理

页面路径建议：

```text
/system/departments
```

### 10.1 部门列表

接口：

```text
GET /api/v1/departments
```

查询条件：

| 字段 | 控件 | 说明 |
| --- | --- | --- |
| keyword | 输入框 | 部门名称关键字 |
| parentId | 树选择 | 上级部门过滤 |

响应结构为树结构。

交互规则：

1. 使用树形表格展示部门。
2. 查询关键字命中非顶级部门时，需要展示其上级链路。
3. 按 `parentId` 过滤非顶级部门时，也需要展示其上级链路。
4. 部门列表不分页。
5. 删除部门前必须确认。
6. 如果部门有关联用户或子部门，后端可能因为数据关联限制拒绝删除，前端展示后端中文错误即可。

### 10.2 新增部门

接口：

```text
POST /api/v1/departments
```

字段：

| 字段 | 控件 | 必填 | 说明 |
| --- | --- | --- | --- |
| parentId | 树选择 | 否 | 上级部门，0 表示顶级 |
| name | 输入框 | 是 | 部门名称 |
| sort | 数字输入 | 否 | 排序 |

### 10.3 编辑部门

接口：

```text
PUT /api/v1/departments/{id}
```

规则：

1. 部门名称必填。
2. 上级部门不能选择自己。
3. 建议前端也禁止选择自己的子部门作为上级，避免形成循环。

### 10.4 部门下拉

接口：

```text
GET /api/v1/departments/options
```

使用场景：

1. 用户新增/编辑中的部门单选。
2. 用户列表中的部门过滤。
3. 部门新增/编辑中的上级部门选择。

## 11. 菜单管理

页面路径建议：

```text
/system/menus
```

### 11.1 菜单列表

接口：

```text
GET /api/v1/menus
```

查询条件：

| 字段 | 控件 | 说明 |
| --- | --- | --- |
| keyword | 输入框 | 菜单名称或路由 |
| type | 下拉框 | 1 菜单，2 按钮 |
| parentId | 树选择 | 上级菜单 |
| show | 下拉框 | 0 隐藏，1 显示 |
| pageNo | 分页 | 当前页 |
| pageSize | 分页 | 每页条数 |

列表字段：

| 字段 | 说明 |
| --- | --- |
| name | 菜单名称 |
| type | 类型 |
| route | 路由 |
| icon | 图标 |
| show | 是否显示 |
| sort | 排序 |
| 操作 | 查看、编辑、删除 |

### 11.2 新增菜单

接口：

```text
POST /api/v1/menus
```

字段：

| 字段 | 控件 | 必填 | 说明 |
| --- | --- | --- | --- |
| type | 单选 | 否 | 0 分组，1 菜单，2 按钮；不传默认为分组 |
| parentId | 树选择 | 否 | 上级菜单，0 表示顶级 |
| name | 输入框 | 是 | 菜单名称 |
| route | 输入框 | 否 | 前端路由或权限标识 |
| icon | 输入框/图标选择器 | 否 | 图标 |
| show | 开关 | 否 | 是否显示；分组类型无需填写 |
| sort | 数字输入 | 否 | 排序 |

说明：

1. 分组类型用于侧边栏目录分组，创建分组时仅需要填写菜单名称。
2. 菜单类型用于页面导航。
3. 按钮类型用于操作权限控制。
4. 按钮权限的 `route` 可作为权限标识使用，例如 `system:user:create`。

### 11.3 编辑菜单

接口：

```text
PUT /api/v1/menus/{id}
```

规则：

1. 菜单名称必填。
2. 菜单类型只能为 0、1 或 2。
3. 显示状态只能为 0 或 1。
4. 上级菜单不能选择自己。

## 12. 系统配置

页面路径建议：

```text
/system/configs
```

### 12.1 配置列表

接口：

```text
GET /api/v1/system-configs
```

查询条件：

| 字段 | 控件 | 说明 |
| --- | --- | --- |
| keyword | 输入框 | 配置键、配置名称关键字 |
| group | 输入框/下拉框 | 配置分组 |
| enabled | 下拉框 | 0 禁用，1 启用 |
| pageNo | 分页 | 当前页 |
| pageSize | 分页 | 每页条数 |

列表字段：

| 字段 | 说明 |
| --- | --- |
| key | 配置键 |
| name | 配置名称 |
| value | 配置值 |
| valueType | 值类型 |
| group | 分组 |
| enabled | 是否启用 |
| sort | 排序 |
| description | 描述 |
| 操作 | 查看、编辑、删除 |

### 12.2 新增/编辑配置

新增接口：

```text
POST /api/v1/system-configs
```

编辑接口：

```text
PUT /api/v1/system-configs/{id}
```

字段：

| 字段 | 控件 | 必填 | 说明 |
| --- | --- | --- | --- |
| key | 输入框 | 是 | 配置键 |
| value | 动态输入 | 否 | 配置值 |
| name | 输入框 | 是 | 配置名称 |
| valueType | 下拉框 | 是 | string、number、bool、json |
| group | 输入框/下拉框 | 否 | 默认 default |
| description | 文本域 | 否 | 描述 |
| enabled | 开关 | 是 | 是否启用 |
| sort | 数字输入 | 否 | 排序 |

前端校验：

1. 配置键必填。
2. 配置键必须以字母开头。
3. 配置键仅支持字母、数字、下划线、点、冒号、短横线。
4. 配置名称必填。
5. 值类型只能为 `string`、`number`、`bool`、`json`。
6. `number` 类型应校验为数字。
7. `bool` 类型建议使用开关或 true/false 下拉。
8. `json` 类型建议使用代码编辑器，并在提交前校验 JSON 格式。

### 12.3 常用系统配置项

| 配置键 | 类型 | 分组 | 说明 |
| --- | --- | --- | --- |
| site.title | string | site | 系统标题 |
| site.logo | string | site | 系统 Logo 地址 |
| site.favicon | string | site | 浏览器图标地址 |
| site.copyright | string | site | 版权信息 |
| security.password.min_length | number | security | 用户密码最小长度 |
| security.login.encrypt.enabled | bool | security | 是否启用登录密码 RSA 加密 |
| security.login.rsa.public_key | string | security | 登录加密 RSA 公钥，前端使用 |
| security.login.rsa.private_key | string | security | 登录解密 RSA 私钥，仅后端使用 |
| security.login.max_attempts | number | security | 登录失败最大次数 |
| security.login.locked_minutes | number | security | 登录失败锁定分钟数 |
| security.login.concurrent.enabled | bool | security | 是否允许并发登录 |
| security.token.expire_minutes | number | security | Token 过期分钟数 |
| security.token.jwt_secret_key | string | security | JWT 密钥 |
| security.token.storage | json | security | Token 存储配置 |

敏感配置展示要求：

1. 私钥、JWT 密钥、Redis 密码等敏感值默认脱敏展示。
2. 编辑时提供“显示/隐藏”按钮。
3. 列表页不直接完整展示敏感值。
4. 删除和修改敏感配置时需要二次确认。

### 12.4 已启用配置

接口：

```text
GET /api/v1/system-configs/enabled
GET /api/v1/system-configs/enabled-map
```

使用场景：

1. 获取站点标题、Logo、版权信息。
2. 根据系统配置调整前端功能表现。
3. 管理端配置预览。

## 13. 权限控制

### 13.1 路由权限

前端应根据当前用户菜单控制可访问路由：

1. 用户登录后动态注册菜单路由。
2. 访问不存在或无权限路由时跳转 403 页面。
3. 页面刷新时通过 `GET /api/v1/auth/me` 恢复用户菜单。
4. 菜单为空时跳转无权限页面。

### 13.2 按钮权限

按钮权限来源于 `permissions` 或 `menus` 中 `type=2` 的数据。

建议封装权限判断方法：

```js
function hasPermission(permission) {
  return userStore.permissions.includes(permission)
}
```

页面按钮控制：

1. 新增按钮。
2. 编辑按钮。
3. 删除按钮。
4. 授权按钮。
5. 查看详情按钮。

如果当前后端菜单按钮暂未统一权限标识规则，建议使用按钮菜单的 `route` 字段作为权限标识。

## 14. 字典与枚举

### 14.1 用户状态

| 值 | 文案 |
| --- | --- |
| 0 | 禁用 |
| 1 | 启用 |

### 14.2 菜单类型

| 值 | 文案 |
| --- | --- |
| 0 | 分组 |
| 1 | 菜单 |
| 2 | 按钮 |

### 14.3 菜单显示状态

| 值 | 文案 |
| --- | --- |
| 0 | 隐藏 |
| 1 | 显示 |

### 14.4 角色类型

| 值 | 文案 |
| --- | --- |
| 0 | 自定义角色 |
| 1 | 内置角色 |

### 14.5 系统配置值类型

| 值 | 文案 |
| --- | --- |
| string | 字符串 |
| number | 数字 |
| bool | 布尔 |
| json | JSON |

### 14.6 系统配置启用状态

| 值 | 文案 |
| --- | --- |
| 0 | 禁用 |
| 1 | 启用 |

## 15. 页面清单

| 页面 | 路径建议 | 说明 |
| --- | --- | --- |
| 登录页 | /login | 用户登录 |
| 首页 | /dashboard | 系统概览 |
| 用户管理 | /system/users | 用户 CRUD |
| 角色管理 | /system/roles | 角色 CRUD、角色授权 |
| 部门管理 | /system/departments | 部门树管理 |
| 菜单管理 | /system/menus | 菜单与按钮权限管理 |
| 系统配置 | /system/configs | 系统变量配置 |
| 个人信息 | /account/profile | 当前用户信息展示 |
| 403 页面 | /403 | 无权限访问 |
| 404 页面 | /404 | 页面不存在 |

## 16. 接口清单

### 16.1 认证接口

| 方法 | 地址 | 说明 | 登录态 |
| --- | --- | --- | --- |
| GET | /api/v1/health | 健康检查 | 否 |
| GET | /api/v1/auth/encrypt-config | 获取登录加密配置 | 否 |
| POST | /api/v1/auth/login | 登录 | 否 |
| POST | /api/v1/auth/logout | 退出登录 | 是 |
| GET | /api/v1/auth/me | 当前用户信息 | 是 |

### 16.2 用户接口

| 方法 | 地址 | 说明 |
| --- | --- | --- |
| GET | /api/v1/users | 用户列表 |
| POST | /api/v1/users | 新增用户 |
| GET | /api/v1/users/{id} | 用户详情 |
| PUT | /api/v1/users/{id} | 编辑用户 |
| DELETE | /api/v1/users/{id} | 删除用户 |

### 16.3 角色接口

| 方法 | 地址 | 说明 |
| --- | --- | --- |
| GET | /api/v1/roles | 角色列表 |
| GET | /api/v1/roles/options | 角色下拉 |
| POST | /api/v1/roles | 新增角色 |
| GET | /api/v1/roles/{id} | 角色详情 |
| PUT | /api/v1/roles/{id} | 编辑角色 |
| DELETE | /api/v1/roles/{id} | 删除角色 |
| GET | /api/v1/roles/{id}/menus | 获取角色菜单 |
| PUT | /api/v1/roles/{id}/menus | 保存角色菜单 |

### 16.4 部门接口

| 方法 | 地址 | 说明 |
| --- | --- | --- |
| GET | /api/v1/departments | 部门树列表 |
| GET | /api/v1/departments/options | 部门下拉树 |
| POST | /api/v1/departments | 新增部门 |
| GET | /api/v1/departments/{id} | 部门详情 |
| PUT | /api/v1/departments/{id} | 编辑部门 |
| DELETE | /api/v1/departments/{id} | 删除部门 |

### 16.5 菜单接口

| 方法 | 地址 | 说明 |
| --- | --- | --- |
| GET | /api/v1/menus | 菜单列表 |
| POST | /api/v1/menus | 新增菜单 |
| GET | /api/v1/menus/{id} | 菜单详情 |
| PUT | /api/v1/menus/{id} | 编辑菜单 |
| DELETE | /api/v1/menus/{id} | 删除菜单 |

### 16.6 系统配置接口

| 方法 | 地址 | 说明 |
| --- | --- | --- |
| GET | /api/v1/system-configs | 配置列表 |
| POST | /api/v1/system-configs | 新增配置 |
| GET | /api/v1/system-configs/enabled | 启用配置列表 |
| GET | /api/v1/system-configs/enabled-map | 启用配置键值映射 |
| GET | /api/v1/system-configs/by-key/{key} | 根据配置键查询 |
| GET | /api/v1/system-configs/{id} | 配置详情 |
| PUT | /api/v1/system-configs/{id} | 编辑配置 |
| DELETE | /api/v1/system-configs/{id} | 删除配置 |

## 17. 异常与提示

前端需要重点处理以下后端提示：

| 场景 | 建议前端处理 |
| --- | --- |
| 用户名或密码错误 | 登录页错误提示 |
| 用户已被禁用 | 登录页错误提示 |
| 用户已被临时锁定 | 展示锁定剩余时间或后端提示 |
| 请先登录 | 清理状态并跳转登录页 |
| 登录状态已失效，请重新登录 | 清理状态并跳转登录页 |
| 内置角色不允许删除 | 角色列表提示 |
| 数据已存在，请检查唯一字段 | 表单错误提示 |
| 数据关联不合法 | 表单或删除错误提示 |
| 记录不存在 | 返回列表或展示空状态 |

## 18. 删除交互

所有删除操作均为硬删除，前端必须提供二次确认。

确认文案建议：

| 对象 | 文案 |
| --- | --- |
| 用户 | 确认删除该用户吗？删除后不可恢复。 |
| 角色 | 确认删除该角色吗？删除后不可恢复。 |
| 部门 | 确认删除该部门吗？删除后不可恢复。 |
| 菜单 | 确认删除该菜单吗？删除后不可恢复。 |
| 系统配置 | 确认删除该系统配置吗？删除后不可恢复。 |

对于内置角色，前端应禁用删除按钮并展示原因。

## 19. 安全要求

1. 登录密码根据后端加密配置决定是否 RSA 加密。
2. Token 不应拼接在 URL 中。
3. 私钥类系统配置默认脱敏展示。
4. 退出登录时清理 Token、用户信息、动态路由和权限缓存。
5. 401 或登录失效时统一跳转登录页。
6. 前端不要保存用户明文密码。
7. HTTP 环境下 RSA 只能避免密码明文出现在请求体中，无法替代 HTTPS。

## 20. 联调重点

1. 登录加密配置启用和关闭两种模式都需要验证。
2. RSA-OAEP-SHA256 前端加密结果必须能够被后端私钥解密。
3. 用户新增密码入库由后端 MD5 处理，前端新增用户时不需要 MD5。
4. 用户登录失败超过限制后，非 admin 用户会被临时锁定。
5. admin 用户不做登录错误次数锁定。
6. 部门列表和部门下拉均为树结构。
7. 用户部门选择必须为单选。
8. 用户按部门过滤时需要验证子部门用户也会被查出。
9. 角色编辑时不允许修改角色标识。
10. 内置角色不允许删除。
11. 系统配置 JSON 类型需要验证 JSON 格式。
12. Token 请求头必须使用 `Authorization: Bearer <accessToken>`。

## 21. 验收标准

1. 用户可以完成登录、退出、刷新恢复登录态。
2. 登录后侧边栏根据当前用户菜单动态展示。
3. 无权限路由不可访问。
4. 用户管理支持查询、新增、编辑、查看、删除。
5. 用户新增和编辑支持角色多选、部门单选。
6. 角色管理支持查询、新增、编辑、查看、删除和菜单授权。
7. 部门管理以树结构展示，支持新增、编辑、删除和过滤。
8. 菜单管理支持菜单和按钮权限维护。
9. 系统配置支持不同值类型的新增、编辑、查询、删除。
10. 所有表单都有必要的前端校验。
11. 所有删除操作都有二次确认。
12. 接口错误信息能够以中文展示给用户。
13. 登录密码 RSA 加密在 HTTP 环境下可用。
14. 页面刷新后权限、菜单和用户状态能够正确恢复。
