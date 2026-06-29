import client from './client'
import type { ApiResponse, User, UserCreateForm, UserUpdateForm, ResetPasswordForm } from '@/types'

/** 账户列表 — 后端返回平铺数组，无分页 */
export async function listUsers(): Promise<User[]> {
  const res = await client.get<ApiResponse<User[]>>('/users')
  return res.data.data
}

/** 新建账户 */
export async function createUser(data: UserCreateForm): Promise<User> {
  const res = await client.post<ApiResponse<User>>('/users', data)
  return res.data.data
}

/** 编辑账户 — 仅可修改 role 和 campus_id */
export async function updateUser(id: number, data: UserUpdateForm): Promise<User> {
  const res = await client.put<ApiResponse<User>>(`/users/${id}`, data)
  return res.data.data
}

/** 禁用账户 — 对齐后端 PUT /users/:id/disable */
export async function disableUser(id: number): Promise<void> {
  await client.put(`/users/${id}/disable`)
}

/** 重置密码 — 对齐后端 PUT /users/:id/reset-pwd */
export async function resetPassword(id: number, data: ResetPasswordForm): Promise<void> {
  await client.put(`/users/${id}/reset-pwd`, data)
}
