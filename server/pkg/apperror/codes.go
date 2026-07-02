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
	ErrUserPhoneEmpty      = New(42014, "手机号不能为空")
	ErrUserPhoneTooLong    = New(42015, "手机号不能超过20个字符")
)

// ---------- 物资（43xxx）----------
var (
	ErrMaterialCategoryNameEmpty   = New(43001, "分类名称不能为空")
	ErrMaterialCategoryNameTooLong = New(43002, "分类名称不能超过50个字符")
	ErrMaterialCategoryNameDup     = New(43003, "分类名称已存在")
	ErrMaterialCategoryNotFound    = New(43004, "物资分类不存在")
	ErrMaterialCategoryHasPurchase = New(43005, "该分类下有采购记录，无法删除")
	ErrMaterialNameEmpty           = New(43006, "物资名称不能为空")
	ErrMaterialCategoryNotExist    = New(43007, "指定的物资分类不存在")
	ErrMaterialQuantityInvalid     = New(43008, "采购数量必须为正整数")
	ErrMaterialAmountInvalid       = New(43009, "总金额必须大于0")
	ErrMaterialStockNotFound       = New(43010, "库存记录不存在")
	ErrMaterialStockInsufficient   = New(43011, "库存余量不足")
	ErrMaterialDistQuantityInvalid = New(43012, "派发数量必须为正整数")
	ErrMaterialActivityNotFound    = New(43013, "指定活动不存在")
	ErrMaterialDistNotFound        = New(43014, "派发记录不存在")
	ErrMaterialPermDenied          = New(43015, "仅总部管理员可操作物资")
	ErrMaterialAdjQuantityZero     = New(43016, "调整数量必须大于0")
)

// ---------- 活动（44xxx）----------
var (
	ErrActivityNotFound               = New(44001, "活动不存在")
	ErrActivityNameEmpty              = New(44002, "活动名称不能为空")
	ErrActivityNameTooLong            = New(44003, "活动名称不能超过200个字符")
	ErrActivityCampusNotFound          = New(44004, "指定校区不存在")
	ErrActivityPlannedExecInvalid      = New(44005, "计划执行次数必须大于0")
	ErrActivityDateInvalid             = New(44006, "结束日期必须晚于开始日期")
	// 44007 已废弃（联系人不再限制同一校区）
	ErrActivityContactsNotFound = New(44008, "指定的活动联系人不存在")
	ErrActivityArchivedCannotModify    = New(44009, "已归档的活动不支持修改")
	ErrActivityNotSettled             = New(44010, "只有已结算的活动才能归档")
	ErrActivityExecCountInvalid        = New(44011, "执行次数必须大于0")
	ErrActivityExecExceedPlanned       = New(44012, "累计执行次数将超出计划次数")
	ErrActivityStatusNoExec            = New(44013, "当前活动状态不允许记录执行")
	ErrActivityPermissionDenied        = New(44014, "无权限执行此操作")
	ErrActivityPlannedExecBelowExecuted = New(44015, "计划执行次数不能小于已执行次数")
	ErrActivityNotContactPerson        = New(44016, "仅活动联系人或总部管理员可记录执行")
	ErrActivityCampusMismatch         = New(44017, "校区操作员只能创建本校区的活动")
)

// ---------- 结算（45xxx）----------
// 后续补充

// ---------- 报表（46xxx）----------
// 后续补充
