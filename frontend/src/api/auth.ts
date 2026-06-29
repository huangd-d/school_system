import client from './client'
import type { ApiResponse, LoginRequest, LoginResponse, RefreshResponse } from '@/types'

/** 登录 */
export async function login(req: LoginRequest): Promise<LoginResponse> {
  const res = await client.post<ApiResponse<LoginResponse>>('/auth/login', req)
  return res.data.data
}

/** 刷新 token — 后端仅返回 { token }，不返回完整用户信息 */
export async function refreshToken(): Promise<RefreshResponse> {
  const res = await client.post<ApiResponse<RefreshResponse>>('/auth/refresh')
  return res.data.data
}
