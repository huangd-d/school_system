# server/CLAUDE.md

本文件仅在 Claude 处理 `server/` 目录下的文件时加载。

---

## 架构规约

### 模块内部结构（必须遵守）

每个业务模块固定三个文件，各司其职：

```
module/<name>/
├── handler.go      # HTTP 层：绑定参数 → 调 service → 统一响应 → 不写任何业务逻辑
├── service.go      # 业务层：核心逻辑、校验规则、事务管理、审计日志写入 → 定义 Repository 接口
└── repository.go   # 数据层：实现 Repository 接口 → 纯 GORM 操作 → 不包含业务逻辑
```

**依赖方向**：`handler → service → repository → model`（单向，内层不感知外层）

### 新增模块步骤

1. `mkdir internal/module/<name>`
2. 在 `model/` 下添加模型文件
3. 创建 `handler.go`、`service.go`、`repository.go`
4. 在 `router.go` 中注册路由
5. 在 `pkg/apperror/codes.go` 中按域名添加错误码

**不动任何已有模块代码。**

### Handler 规范

```go
func (h *Handler) XXX(c *gin.Context) {
    // 1. 绑定参数
    var req XXXReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Err(c, apperror.ErrInvalidParam)
        return
    }
    // 2. 调用 service
    data, err := h.svc.XXX(c.Request.Context(), ...)
    // 3. 返回统一响应
    if err != nil { response.Err(c, err); return }
    response.OK(c, data)
}
```

### Service 规范

- 所有方法签名接受 `ctx context.Context` 作为第一个参数
- 返回 `error` 类型必须是 `*apperror.AppError`（用 `apperror.New` / `apperror.Newf` / 已定义变量）
- 金额变动、结算、修正操作必须在 service 内显式写入审计日志
- 需要原子性的操作（摊销重算、结算回收）通过 `db.Transaction()` 包裹

### Repository 规范

- 每个模块的 `service.go` 中定义 `Repository` 接口
- `repository.go` 中 `repository` 结构体（小写，不导出）实现该接口
- `NewRepository(db *gorm.DB) Repository` 返回接口
- 所有方法接受 `ctx context.Context`，内部通过 `r.db.WithContext(ctx)` 执行

---

## 响应规范

**统一格式**：

```json
{ "code": 0, "message": "success", "data": { ... } }
{ "code": 41001, "message": "校区不存在", "data": null }
```

**调用方式**：

```go
response.OK(c, data)           // 成功
response.Err(c, apperror.XXX)  // 业务错误（必须用 codes.go 中定义的变量）
response.Err(c, err)           // 传入 AppError 自动提取 Code，普通 error 按 50000 处理
```

**HTTP 状态码统一返回 200**，前端通过 `code` 字段判断成功（0）或失败（非 0）。

---

## 错误码规范

**集中管理**：`pkg/apperror/codes.go`

**编码规则**：

```
40xxx = 通用系统错误    41xxx = 校区模块
42xxx = 账户模块        43xxx = 物资模块
44xxx = 活动模块        45xxx = 结算模块
46xxx = 报表模块
```

**新增错误码**：在 `codes.go` 对应域名下添加变量，禁止在业务代码中硬编码数字。

```go
// 静态错误（消息固定）
var ErrCampusNotFound = New(41001, "校区不存在")

// 动态错误（消息需拼接）
apperror.New(apperror.ErrCampusNameDup.Code, fmt.Sprintf("校区名称「%s」已存在", name))
```

---

## 代码风格

- **注释**：全部使用中文
- **命名**：Go 公开标识符用 CamelCase，私有用 camelCase，缩写大写（ID、JWT、URL）
- **文件**：UTF-8 编码，gofmt 格式化
- **错误信息**：中文面向用户
