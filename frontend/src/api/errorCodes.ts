/**
 * 后端业务错误码常量
 * 与 server/pkg/apperror/codes.go 严格对应，用于前端分支判断
 * 未列出的错误码走透传（后端返回的 message 即可）
 */
export const ErrorCode = {
  // ── 通用 40xxx ──
  /** 服务器内部错误 */
  Internal: 40000,
  /** 参数校验失败 */
  InvalidParam: 40001,
  /** 未登录或登录已过期 */
  Unauthorized: 40002,
  /** 无权限执行此操作 */
  Forbidden: 40003,
  /** 资源不存在 */
  NotFound: 40004,
  /** 数据冲突 */
  Conflict: 40005,
} as const

export type ErrorCodeValue = (typeof ErrorCode)[keyof typeof ErrorCode]
