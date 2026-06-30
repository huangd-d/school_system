// ==================== 通用类型 ====================

/** 统一响应格式 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// ==================== 枚举 ====================

/** 用户角色 */
export type Role = 'hq_admin' | 'campus_operator' | 'activity_contact'

/** 校区类型 — 对齐后端: hq=总部, normal=普通 */
export type CampusType = 'hq' | 'normal'

// ==================== 认证 ====================

export interface LoginRequest {
  username: string
  password: string
}

/** 登录响应 — 对齐后端 auth/LoginResp */
export interface LoginResponse {
  token: string
  user: UserInfo
}

/** 当前用户信息 — 对齐后端 LoginResp.User */
export interface UserInfo {
  id: number
  username: string
  phone: string
  role: Role
  campus_id: number
}

/** 刷新 token 响应 — 对齐后端 */
export interface RefreshResponse {
  token: string
}

// ==================== 校区 ====================

/** 校区 — 对齐后端 model.Campus + handler.CampusResp */
export interface Campus {
  id: number
  name: string
  type: CampusType
}

/** 创建校区请求 — 对齐后端 CreateCampusReq */
export interface CampusCreateForm {
  name: string
  type: CampusType
}

/** 编辑校区请求 — 对齐后端 UpdateCampusReq（仅名称可改） */
export interface CampusUpdateForm {
  name: string
}

// ==================== 账户 ====================

/** 账户 — 对齐后端 user.UserResp */
export interface User {
  id: number
  username: string
  phone: string
  role: Role
  campus_id: number
  status: 'active' | 'disabled'
  created_at: string
}

/** 新建账户请求 — 对齐后端 CreateUserReq */
export interface UserCreateForm {
  username: string
  password: string
  phone: string
  role: Role
  campus_id: number
}

/** 编辑账户请求 — 对齐后端 UpdateUserReq */
export interface UserUpdateForm {
  username?: string
  phone?: string
  role: Role
  campus_id: number
}

/** 重置密码请求 — 对齐后端 ResetPasswordReq */
export interface ResetPasswordForm {
  password: string
}
