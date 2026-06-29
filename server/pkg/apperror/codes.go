package apperror

// ============================================================
//  错误码规范
//  40xxx = 通用系统错误    41xxx = 校区模块
//  42xxx = 账户模块        43xxx = 物资模块
//  44xxx = 活动模块        45xxx = 结算模块
//  46xxx = 报表模块        50xxx = 服务端错误
// ============================================================

// ---------- 通用（40xxx）----------
var (
	ErrInternal     = New(40000, "服务器内部错误")
	ErrInvalidParam = New(40001, "参数校验失败")
	ErrUnauthorized = New(40002, "未登录或登录已过期")
	ErrForbidden    = New(40003, "无权限执行此操作")
	ErrNotFound     = New(40004, "资源不存在")
	ErrConflict     = New(40005, "数据冲突")
)

// ---------- 校区（41xxx）----------
var (
	ErrCampusNotFound     = New(41001, "校区不存在")
	ErrCampusNameEmpty    = New(41002, "校区名称不能为空")
	ErrCampusNameTooLong  = New(41003, "校区名称不能超过100个字符")
	ErrCampusNameDup      = New(41004, "校区名称已存在")
	ErrCampusTypeInvalid  = New(41005, "校区类型无效，只能为 hq 或 normal")
	ErrCampusHQExists     = New(41006, "总部校区已存在，不允许创建第二个总部")
	ErrCampusHQDelete     = New(41007, "总部校区不可删除")
	ErrCampusHasUsers     = New(41008, "该校区下还有账户，请先转移或删除账户")
	ErrCampusHasActivities = New(41009, "该校区下还有活动，请先转移或删除活动")
)

// ---------- 账户（42xxx）----------
var (
	ErrUserNotFound        = New(42001, "账户不存在")
	ErrUserUsernameEmpty   = New(42002, "用户名不能为空")
	ErrUserUsernameTooLong = New(42003, "用户名不能超过50个字符")
	ErrUserUsernameDup     = New(42004, "用户名已存在")
	ErrUserPasswordEmpty   = New(42005, "密码不能为空")
	ErrUserRoleInvalid     = New(42006, "角色无效，只能为 hq_admin、campus_operator 或 activity_contact")
	ErrUserCampusRequired  = New(42007, "必须指定所属校区")
	ErrUserCampusNotFound  = New(42008, "指定的校区不存在")
	ErrUserHQAdminCampus   = New(42009, "总部管理员必须绑定总部校区")
	ErrUserNormalCampus    = New(42010, "校区操作员和活动联系人必须绑定普通校区")
	ErrUserDisableSelf     = New(42011, "不能禁用自己的账户")
	ErrUserDisabled        = New(42012, "账户已禁用")
	ErrUserLoginFailed     = New(42013, "用户名或密码错误")
)

// ---------- 物资（43xxx）----------
// 后续补充

// ---------- 活动（44xxx）----------
// 后续补充

// ---------- 结算（45xxx）----------
// 后续补充

// ---------- 报表（46xxx）----------
// 后续补充
