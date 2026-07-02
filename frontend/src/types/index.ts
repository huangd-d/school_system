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

// ==================== 活动 ====================

/** 活动状态 */
export type ActivityStatus = 'not_started' | 'in_progress' | 'ended' | 'settled' | 'archived'

/** 活动基础响应（Create/Update 返回） */
export interface Activity {
  id: number
  name: string
  campus_id: number
  planned_executions: number
  start_date: string
  end_date: string
  status: ActivityStatus
  created_by: number
  created_at: string
  updated_at: string
}

/** 活动列表项（List 返回） */
export interface ActivityListItem extends Activity {
  contact_ids: number[]
  total_executed: number
}

/** 活动详情（Detail 返回） */
export interface ActivityDetail extends Activity {
  contacts: UserBrief[]
  executions: ExecutionRecord[]
  total_executed: number
}

/** 联系人简要信息 */
export interface UserBrief {
  id: number
  username: string
  phone: string
  role: Role
}

/** 执行记录 */
export interface ExecutionRecord {
  id: number
  activity_id: number
  count: number
  recorded_by: number
  created_at: string
}

/** 新建活动请求 */
export interface ActivityCreateForm {
  name: string
  campus_id: number
  contact_ids?: number[]
  planned_executions: number
  start_date: string
  end_date: string
}

/** 编辑活动请求（全部可选，部分更新） */
export interface ActivityUpdateForm {
  name?: string
  contact_ids?: number[]
  planned_executions?: number
}

/** 录入执行次数请求 */
export interface AddExecutionForm {
  count: number
}

// ==================== 通用分页 ====================

export interface PaginationParams {
  page?: number
  page_size?: number
}

export interface PaginatedData<T> {
  list: T[]
  total: number
}

// ==================== 物资分类 ====================

/** 物资分类 — 对齐后端 CategoryResp */
export interface MaterialCategory {
  id: number
  name: string
  note: string
  created_at: string
}

/** 新建/编辑分类请求 — 对齐后端 CreateCategoryReq / UpdateCategoryReq */
export interface CategoryCreateForm {
  name: string
  note?: string
}

// ==================== 采购 ====================

/** 采购单 — 对齐后端 PurchaseOrderResp */
export interface PurchaseOrder {
  id: number
  material_name: string
  category_id: number
  quantity: number
  total_amount: number
  unit_price: number
  notes: string
  purchaser_id: number
  created_at: string
}

/** 采购请求 — 对齐后端 PurchaseReq */
export interface PurchaseForm {
  material_name: string
  category_id: number
  quantity: number
  total_amount: number
  notes?: string
}

// ==================== 库存 ====================

/** 库存项 — 对齐后端 StockResp */
export interface StockItem {
  id: number
  purchase_order_id: number
  category_id: number
  material_name: string
  total_quantity: number
  remaining_qty: number
  unit_price: number
  source: string
  created_at: string
  updated_at: string
}

/** 库存查询参数 — 对齐后端 ListStock 的 query params */
export interface StockQuery extends PaginationParams {
  category_id?: number
  keyword?: string
}

// ==================== 派发 ====================

/** 派发记录 — 对齐后端 DistributionResp */
export interface Distribution {
  id: number
  stock_id: number
  activity_id: number
  quantity: number
  operator_id: number
  reason: string
  created_at: string
}

/** 派发请求 — 对齐后端 DistributeReq */
export interface DistributeForm {
  stock_id: number
  activity_id: number
  quantity: number
  reason?: string
}

/** 调整派发请求 — 对齐后端 AdjustDistributionReq */
export interface AdjustDistributionForm {
  quantity: number
  reason: string
}

/** 派发记录（含物资名和活动名） — 对齐后端 DistributionRecordResp */
export interface DistributionRecord {
  id: number
  stock_id: number
  material_name: string
  activity_id: number
  activity_name: string
  quantity: number
  operator_id: number
  reason: string
  created_at: string
}

/** 派发记录查询参数 */
export interface DistributionQuery extends PaginationParams {
  activity_id?: number
  keyword?: string
  start_date?: string
  end_date?: string
}
