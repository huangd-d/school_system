# CLAUDE.md

## 项目概述

教培机构物资管理系统——轻量级、单机部署的物资全生命周期管理。

**核心流程**：总部管理员采购物资 → 配发给各校区活动 → 按不同维度计算摊销成本。

**设计目标**：单文件部署、SQLite 存储、启动即用、Windows 兼容。

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
