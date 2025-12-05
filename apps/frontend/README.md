# Vue 3 后台管理系统

这是一个基于 Vue 3 开发的后台管理系统，集成了 user、project 和 code_repository 微服务。

## 技术栈

- Vue 3 (Composition API)
- Vue Router
- Pinia
- Element Plus
- Axios
- Vite

## 项目结构

```
src/
├── api/          # API 接口封装
├── assets/       # 静态资源
├── components/   # 公共组件
├── layouts/      # 布局组件
├── router/       # 路由配置
├── stores/       # 状态管理
├── utils/        # 工具函数
├── views/        # 页面组件
│   ├── auth/     # 认证相关页面
│   ├── dashboard/# 仪表盘
│   ├── error/    # 错误页面
│   ├── project/  # 项目管理
│   ├── repository/# 代码仓库管理
│   └── user/     # 用户管理
├── App.vue       # 根组件
└── main.js       # 入口文件
```

## 功能模块

### 1. 认证模块 (Auth)
- 登录页面 (/login)
- 注册页面 (/register)
- 用户头像组件

### 2. 用户管理模块 (User Management)
- 用户列表展示、搜索、分页
- 新增/编辑/删除用户
- 用户详情页面

### 3. 项目管理模块 (Project Management)
- 项目列表展示、搜索、分页
- 新增/编辑/删除项目
- 项目详情页面

### 4. 代码仓库管理模块 (Code Repository Management)
- 代码仓库列表展示、搜索、分页
- 新增/编辑/删除代码仓库
- 代码仓库详情页面
- 与项目的级联关系处理

## 安装和运行

1. 安装依赖

```bash
npm install
```

2. 开发环境运行

```bash
npm run dev
```

3. 生产环境构建

```bash
npm run build
```

## API 集成

项目集成了三个微服务的 API：

- user 服务：用户管理相关接口
- project 服务：项目管理相关接口
- code_repository 服务：代码仓库管理相关接口

所有 API 请求都通过 `/src/utils/request.js` 封装的 axios 实例发送，并统一处理了请求拦截和响应拦截。

## 路由守卫

项目配置了路由守卫，对需要登录的页面进行访问控制。未登录用户访问需要登录的页面时，会自动跳转到登录页面。

## 状态管理

使用 Pinia 进行状态管理，主要管理用户登录状态和用户信息。

## UI 组件

使用 Element Plus 作为 UI 组件库，提供了丰富的后台管理界面组件。

## 注意事项

1. 确保 API 服务地址正确配置在 `vite.config.js` 中
2. 所有 API 请求都会自动带上 Authorization 头
3. 响应拦截器会统一处理错误状态码和 401 未授权跳转