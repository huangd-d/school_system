import { create } from 'zustand'
import type { UserInfo, LoginRequest } from '@/types'
import { login as loginApi } from '@/api/auth'

const TOKEN_KEY = 'token'
const USER_KEY = 'user'

interface AuthState {
  user: UserInfo | null
  token: string | null
  loading: boolean
  /** 登录 */
  login: (req: LoginRequest) => Promise<void>
  /** 恢复会话 — 从 localStorage 读取，无需调接口 */
  restore: () => void
  /** 退出登录 */
  logout: () => void
}

/** 从 localStorage 恢复用户信息 */
function loadUser(): UserInfo | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as UserInfo) : null
  } catch {
    return null
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  user: loadUser(),
  token: localStorage.getItem(TOKEN_KEY),
  loading: false,

  login: async (req) => {
    set({ loading: true })
    try {
      const res = await loginApi(req)
      localStorage.setItem(TOKEN_KEY, res.token)
      localStorage.setItem(USER_KEY, JSON.stringify(res.user))
      set({ token: res.token, user: res.user, loading: false })
    } catch (e) {
      set({ loading: false })
      throw e
    }
  },

  restore: () => {
    // 纯客户端恢复，无需网络请求（后端无 profile 接口）
    set({ user: loadUser(), token: localStorage.getItem(TOKEN_KEY) })
  },

  logout: () => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    set({ token: null, user: null })
  },
}))
