import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import fs from 'fs'

// https://vitejs.dev/config/
// 动态提供 /workflows/* 与 /workflows/index.json，优先读取仓库根目录 .workflows
const workflowsPlugin = {
  name: 'dev-workflows-middleware',
  configureServer(server) {
    const publicDir = resolve(__dirname, 'public/workflows')
    const repoDir = resolve(__dirname, '../../..', '.workflows')

    server.middlewares.use((req, res, next) => {
      try {
        const url = req.url || ''
        // 动态索引
        if (url === '/workflows/index.json') {
          const names = new Set()
          try {
            if (fs.existsSync(repoDir)) {
              for (const fn of fs.readdirSync(repoDir)) {
                if (/\.(ya?ml)$/i.test(fn)) names.add(fn)
              }
            }
          } catch (_) { }
          try {
            if (fs.existsSync(publicDir)) {
              for (const fn of fs.readdirSync(publicDir)) {
                if (/\.(ya?ml)$/i.test(fn)) names.add(fn)
              }
            }
          } catch (_) { }
          const list = Array.from(names).sort().map((fn) => ({ label: fn, path: `/workflows/${fn}` }))
          res.setHeader('Content-Type', 'application/json')
          res.statusCode = 200
          res.end(JSON.stringify(list))
          return
        }
        // 提供 YAML 内容：优先 .workflows，其次 public/workflows
        if (url.startsWith('/workflows/')) {
          const name = decodeURIComponent(url.replace('/workflows/', ''))
          if (!/\.(ya?ml)$/i.test(name)) return next()
          const candidates = [resolve(repoDir, name), resolve(publicDir, name)]
          for (const p of candidates) {
            try {
              if (fs.existsSync(p)) {
                const buf = fs.readFileSync(p)
                res.setHeader('Content-Type', 'text/yaml; charset=utf-8')
                res.statusCode = 200
                res.end(buf)
                return
              }
            } catch (_) { }
          }
          return next()
        }
        next()
      } catch (_) { next() }
    })
  }
}

export default defineConfig({
  plugins: [vue(), workflowsPlugin],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5175,
    fs: { allow: [resolve(__dirname, '../../..')] },
    proxy: {
      // 将以这些前缀开头的请求代理到本地后端
      '/user_service': { target: 'http://localhost:31080', changeOrigin: true },
      '/project_service': { target: 'http://localhost:31080', changeOrigin: true },
      '/code_repository_service': { target: 'http://localhost:31080', changeOrigin: true },
      '/artifact_service': { target: 'http://localhost:31080', changeOrigin: true },
      '/ci_service/api/v1/executor/ws': { target: 'ws://localhost:31080', changeOrigin: true, ws: true },
      '/ci_service': { target: 'http://localhost:31080', changeOrigin: true, ws: true }
    }
  }
})