# CLAUDE.md

## 项目概述

教培机构物资管理系统——轻量级、单机部署的物资全生命周期管理。

**核心流程**：总部管理员采购物资 → 配发给各校区活动 → 按不同维度计算摊销成本。

**设计目标**：单文件部署、SQLite 存储、启动即用、Windows 兼容。

---

## 技术栈

### 后端

| 库 | 用途 |
|----|------|
| Go 1.21+ | 语言 |
| Gin | HTTP 框架 |
| GORM + `gorm.io/driver/sqlite` | ORM（底层 `modernc.org/sqlite`，纯 Go，无需 CGO） |
| `github.com/golang-jwt/jwt/v5` | JWT 认证（有效期 1 天） |
| `github.com/go-playground/validator` | 请求参数校验 |
| `github.com/spf13/viper` | 配置管理（环境变量 > 配置文件 > 默认值） |
| `github.com/gin-contrib/cors` | 跨域 |
| `golang.org/x/crypto` | 密码 bcrypt 哈希 |
| `go.uber.org/zap` | 结构化日志 |
| `gopkg.in/natefinch/lumberjack.v2` | 日志文件切割 |
| `github.com/robfig/cron/v3` | 定时备份 SQLite（每天凌晨3点） |

### 前端

| 库 | 用途 |
|----|------|
| React 19 + Vite 8 + TypeScript | 框架 |
| Ant Design 6.5 | UI 组件库 |
| `@tanstack/react-query` | 服务端状态管理 |
| Zustand | UI 状态管理 |
| react-router v7 | 路由 |
| recharts | 报表图表 |
| axios | HTTP 请求 + 拦截器 |
| Tailwind CSS | 自定义样式 |

---

## 目录结构

```
server/
├── cmd/server/main.go              # 入口（配置→日志→DB→迁移→种子→路由→启动）
├── internal/
│   ├── config/config.go            # Viper 配置加载
│   ├── database/
│   │   ├── sqlite.go               # GORM 连接 + WAL 模式
│   │   └── backup.go               # 定时备份
│   ├── model/                      # 数据模型（全部 GORM 模型）
│   │   ├── campus.go               #   校区
│   │   ├── user.go                 #   账户
│   │   ├── material.go             #   物资分类 + 采购单 + 库存
│   │   ├── activity.go             #   活动 + 活动联系人关联
│   │   ├── distribution.go         #   派发记录
│   │   ├── execution.go            #   执行记录
│   │   ├── amortization.go         #   摊销快照
│   │   ├── settlement.go           #   结算 + 回收明细
│   │   └── audit.go                #   审计日志
│   ├── module/                     # 业务模块（每个模块一个目录，高内聚低耦合）
│   │   ├── auth/                   #   登录、JWT 签发/验证
│   │   ├── campus/                 #   校区 CRUD（已实现）
│   │   ├── user/                   #   账户管理
│   │   ├── material/               #   采购 + 库存 + 派发
│   │   ├── activity/               #   活动 + 执行次数
│   │   ├── settlement/             #   结算 + 回撤
│   │   └── report/                 #   摊销快照 + 四维报表
│   ├── middleware/
│   │   ├── auth.go                 #   JWT 鉴权
│   │   ├── cors.go                 #   跨域
│   │   ├── logger.go               #   请求日志
│   │   ├── recovery.go             #   异常恢复
│   │   └── datascope.go            #   校区数据隔离
│   └── router/router.go            # 路由注册 + 中间件链
└── pkg/
    ├── response/response.go        # 统一响应 {code, message, data}
    ├── apperror/
    │   ├── error.go                # AppError 类型 + New/Newf 工厂函数
    │   └── codes.go                # 所有业务错误码集中定义
    └── pagination/page.go          # 分页工具
```

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

## 前端规约

### 弹框组件分离（必须遵守）

所有弹框（Modal/Drawer/Popconfirm 等弹窗类组件）**禁止**与页面组件写在同一个文件中，必须抽离为独立文件。

**目录结构**：

```
pages/<name>/
├── <Name>Page.tsx          # 页面主文件：表格、搜索、工具栏
├── Create<Name>Modal.tsx   # 新建弹框
├── Edit<Name>Modal.tsx     # 编辑弹框
└── ...                     # 其他弹框（重置密码/确认删除等）
```

**组件接口规范**：每个弹框组件通过 props 控制显隐和接收上下文数据，通过回调通知父组件刷新。

```tsx
// 新建类弹框 — 无需外部数据
interface Props {
  open: boolean
  onClose: () => void       // 关闭弹框
  onSuccess: () => void     // 操作成功后刷新列表
}

// 编辑/操作类弹框 — 需要操作目标
interface Props {
  open: boolean
  user: User                // 操作目标（用具体业务实体替代）
  onClose: () => void
  onSuccess: () => void
}
```

**职责划分**：
- 弹框组件：拥有自己的 Form 实例、mutation 逻辑、提交处理
- 页面组件：只负责表格渲染、数据查询、控制弹框开闭

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

## 鉴权与权限

### 路由分明

```
/api/v1/auth/login      # 公开
/api/v1/auth/refresh    # 公开
/api/v1/*               # 需登录（其余全部）
```

### 中间件链

```
请求 → recovery → logger → cors → auth → datascope → handler
```

- **auth**：解析 JWT，注入 `user_id`、`campus_id`、`role` 到 context
- **datascope**：数据隔离标记（具体过滤逻辑在 service 层按角色实现）

### 三种角色

| 角色 | 常量 | 数据范围 |
|------|------|---------|
| 总部管理员 | `hq_admin` | 全平台 |
| 校区操作员 | `campus_operator` | 仅所属校区 |
| 活动联系人 | `activity_contact` | 仅关联活动 |

---

## 数据模型要点

- **校区**：总部节点内置不可删除，新建默认 normal
- **账户**：总部管理员绑定总部校区，密码由管理员设置无复杂度要求
- **采购单**：一单一品，采购不绑定活动，单价=总金额÷数量
- **库存**：`total_quantity`（入库总量） + `remaining_qty`（可派发余量）
- **派发**：一种物资可派发多个活动，累计不超库存余量
- **活动**：状态自动流转（日期判定），支持多联系人，结算可回撤
- **摊销快照**：按天粒度，历史修正时同一事务内重算

---

## 设计决策（已确认的非默认规则）

以下规则与初始产品规格不同，是用户明确确认过的改动，**新代码必须遵守**：

- **活动联系人跨校区**：联系人不受校区限制。后端不校验 `CampusID` 匹配（`validateContacts` 只查存在性），前端 `CreateActivityModal` / `EditActivityModal` 不下拉过滤。详见 memory `[[contacts-no-campus-binding]]`。
- **审计日志**：所有金额变动操作必须记录，在 service 层统一写入
- **user / campus 模块仅总部管理员可写**：校区和账户的创建、编辑、删除、禁用、重置密码等写操作仅 `hq_admin` 可执行。后端 handler 通过 `checkHQAdmin` 统一拦截（返回 40003），前端 AppLayout 按角色过滤菜单，CampusPage / UserPage 有页面级守卫。List 查询不受限制。详见 memory `[[hq-admin-write-permission]]`。
- **结算回收算法**：按执行进度比例计算已消耗量（`usedQty = floor(qty × totalExecuted / plannedExecutions)`），未消耗部分回收入库并扣减摊销基数。详见 memory `[[settlement-amortization-formula]]`。
- **结算单事务保证**：Execute 和 Reverse 操作在单事务中完成（库存变更 + 结算记录 + 摊销重算 + 审计日志），确保原子性和数据一致性。详见 memory `[[settlement-transaction-guarantee]]`。

---

## 行为规则

当用户确认一个**改变了原有业务规则**的决策时，你必须同时完成以下 4 步，缺一不可：

1. **改代码** — 前后端同步修改，不只在一端打补丁
2. **加测试** — 把新规则写成自动化测试（可执行规约，改了就跑不过）
3. **写 memory** — 在 `memory/` 目录下创建 memory 文件，记录 Why + How to apply
4. **更新 CLAUDE.md** — 如果是全局规则，加入上方「设计决策」节

---

## 代码风格

- **注释**：全部使用中文
- **命名**：Go 公开标识符用 CamelCase，私有用 camelCase，缩写大写（ID、JWT、URL）
- **文件**：UTF-8 编码，gofmt 格式化
- **错误信息**：中文面向用户

---

## 构建与运行

```bash
# 开发
cd server && go run ./cmd/server/

# 编译
cd server && go build -o school-system.exe ./cmd/server/

# 配置（可选）
# 通过环境变量或 config.yaml 覆盖默认值
SERVER_PORT=8080
DB_PATH=./data/school.db
JWT_SECRET=your-secret
JWT_EXPIRE_HOUR=24
CORS_ORIGINS=http://localhost:5173
```

---

## 当前进度

### 后端

| 模块 | 状态 | 说明 |
|------|:----:|------|
| 项目骨架 | ✅ | 配置、日志、DB 连接、AutoMigrate、种子数据、优雅关闭 |
| 中间件链 | ✅ | auth / cors / logger / recovery / datascope |
| 统一响应 + 错误码 | ✅ | `response.OK` / `response.Err`，40xxx~44xxx 已定义 |
| auth | ✅ | 登录、JWT 签发/验证、Token 刷新（无 repository，直接读 DB） |
| campus | ✅ | 完整 CRUD + 业务校验 + 4 个测试文件 |
| user | ✅ | CRUD + 禁用/启用/重置密码 + 4 个测试文件 |
| material | ✅ | 分类 CRUD、采购入库、库存查询、派发/调整、派发记录查询 + 4 个测试文件 |
| activity | ✅ | CRUD + 状态自动流转 + 多联系人 + 执行次数 + 归档 + 4 个测试文件 |
| settlement | ✅ | Preview / Execute / Reverse + 摊销快照重算 + 审计日志 + 测试 |
| report | ✅ | 四个维度查询（ByActivity/ByDateRange/ByCampus/ByCategory）+ 聚合 + 测试 |

### 前端

| 页面 | 状态 | 说明 |
|------|:----:|------|
| 框架搭建 | ✅ | Vite + React 19 + Ant Design + React Query + Zustand + Router |
| API 层 | ✅ | axios 实例 + 拦截器 + 10 个模块 API 文件 + 错误码映射 |
| 登录页 | ✅ | LoginPage |
| 仪表盘 | ✅ | DashboardPage |
| 校区管理 | ✅ | CampusPage + CampusFormModal |
| 账户管理 | ✅ | UserPage + Create/Edit/ResetPassword Modal |
| 物资管理 | ✅ | MaterialPage + 分类/采购库存/派发/派发记录四个 Tab + 各 Form Modal |
| 活动管理 | ✅ | ActivityPage + Create/Edit/Detail/AddExecution Modal |
| 结算管理 | ✅ | SettlementPage + SettlementPreviewModal + SettlementHistoryModal |
| 报表页面 | ✅ | ReportPage（按活动/日期/校区/品类四维报表 + Recharts 图表） |

---

## 参考文档

- 产品需求规格说明书：`产品需求规格说明书.md`
- 后端架构详见计划文件：`C:\Users\60900\.claude\plans\1-goofy-abelson.md`
