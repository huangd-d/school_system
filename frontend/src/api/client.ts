import axios from 'axios'
import type { ApiResponse } from '@/types'
import { ErrorCode } from './errorCodes'

// 扩展 axios 请求配置，添加 _retry 标记防止无限刷新
declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _retry?: boolean
  }
}

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

// ── 请求拦截器：附加 JWT ──
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ── 无感刷新队列 ──
let isRefreshing = false
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (error: Error) => void
}> = []

/** 处理刷新队列 — 成功时用新 token 重试全部请求，失败时拒绝全部 */
function processRefreshQueue(error: Error | null, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error || !token) {
      reject(error ?? new Error('token 刷新失败'))
    } else {
      resolve(token)
    }
  })
  failedQueue = []
}

// ── 响应拦截器：统一错误处理 + 无感刷新 ──
client.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse
    if (body.code === 0) return response

    // ── 业务错误 ──
    if (body.code === ErrorCode.Unauthorized) {
      const originalRequest = response.config

      // 刷新接口本身失败 → 直接踢出，防止死循环
      if (originalRequest.url?.includes('/auth/refresh')) {
        // 动态导入 authStore，避免循环依赖
        import('@/stores/authStore').then(({ useAuthStore }) => {
          useAuthStore.getState().logout()
        })
        import('@/router').then(({ router }) => {
          router.navigate('/login', { replace: true })
        })
        return Promise.reject(new Error(body.message || '登录已过期，请重新登录'))
      }

      // 已经重试过一次 → 放弃
      if (originalRequest._retry) {
        import('@/stores/authStore').then(({ useAuthStore }) => {
          useAuthStore.getState().logout()
        })
        import('@/router').then(({ router }) => {
          router.navigate('/login', { replace: true })
        })
        return Promise.reject(new Error(body.message || '登录已过期，请重新登录'))
      }

      // 正在刷新中 → 入队等待
      if (isRefreshing) {
        return new Promise<typeof response>((resolve, reject) => {
          failedQueue.push({
            resolve: (token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(client(originalRequest))
            },
            reject,
          })
        })
      }

      // 开始无感刷新
      isRefreshing = true
      originalRequest._retry = true

      return import('./auth')
        .then(({ refreshToken }) => refreshToken())
        .then((res) => {
          // 刷新成功 — 更新 token 到 store + localStorage
          import('@/stores/authStore').then(({ useAuthStore }) => {
            useAuthStore.getState().setToken(res.token)
          })
          processRefreshQueue(null, res.token)
          // 重试原请求（带上新 token）
          originalRequest.headers.Authorization = `Bearer ${res.token}`
          return client(originalRequest)
        })
        .catch((err) => {
          // 刷新失败 — 清理状态、踢出队列、跳转登录
          processRefreshQueue(err, null)
          import('@/stores/authStore').then(({ useAuthStore }) => {
            useAuthStore.getState().logout()
          })
          import('@/router').then(({ router }) => {
            router.navigate('/login', { replace: true })
          })
          return Promise.reject(err)
        })
        .finally(() => {
          isRefreshing = false
        })
    }

    // 其他业务错误 — 透传后端 message
    return Promise.reject(new Error(body.message || '请求失败'))
  },
  (error) => {
    // 网络 / HTTP 级别错误
    const message = error.response?.data?.message || error.message || '网络错误'
    return Promise.reject(new Error(message))
  },
)

export default client
