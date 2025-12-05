import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue')
      },
      {
        path: 'ci/pipeline',
        name: 'PipelineList',
        component: () => import('@/views/ci/PipelineList.vue')
      },
      {
        path: 'ci/pipeline/:id/builds',
        name: 'PipelineBuilds',
        component: () => import('@/views/ci/builds/BuildList.vue')
      },
      {
        path: 'ci/builds',
        name: 'BuildList',
        component: () => import('@/views/ci/builds/BuildList.vue')
      },
      {
        path: 'ci/pipeline/:id',
        name: 'PipelineDetail',
        component: () => import('@/views/ci/PipelineDetail.vue')
      },
      {
        path: 'ci/builds/:id',
        name: 'BuildDetail',
        component: () => import('@/views/ci/builds/BuildDetail.vue')
      },
      {
        path: 'projects',
        name: 'ProjectsEntry',
        component: () => import('@/views/project/ProjectList.vue')
      },
      { path: 'projects/detail', name: 'ProjectDetail', component: () => import('@/views/project/ProjectDetail.vue') },
      { path: 'projects/overview', name: 'Projects', component: () => import('@/views/project/Projects.vue') },
      { path: 'projects/users', name: 'ProjectUsers', component: () => import('@/views/project/ProjectUsers.vue') },
      { path: 'projects/repositories', name: 'CodeRepositories', component: () => import('@/views/repository/Repositories.vue') },
      { path: 'projects/artifact/registries', name: 'ArtifactRegistries', component: () => import('@/views/artifact/Registries.vue') },
      { path: 'projects/artifact/namespaces', name: 'ArtifactNamespaces', component: () => import('@/views/artifact/Namespaces.vue') },
      { path: 'projects/artifact/repositories', name: 'ArtifactRepositories', component: () => import('@/views/artifact/Repositories.vue') },
      { path: 'projects/artifact/tags', name: 'ArtifactTags', component: () => import('@/views/artifact/Tags.vue') },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/user/Profile.vue')
      },
      {
        path: 'apitoken',
        name: 'ApiToken',
        component: () => import('@/views/user/ApiToken.vue')
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/settings/Settings.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const isAuthenticated = localStorage.getItem('token')
  
  if (to.meta.requiresAuth && !isAuthenticated) {
    // 需要认证但未登录，重定向到登录页
    next({ name: 'Login' })
  } else if (!to.meta.requiresAuth && isAuthenticated && (to.name === 'Login' || to.name === 'Register')) {
    // 已登录但访问登录或注册页，重定向到首页
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router