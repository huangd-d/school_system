---
name: settlement-transaction-guarantee
description: 结算 Execute/Reverse 的单事务保证设计
metadata:
  type: project
---

# 结算事务保证

## 设计决策

结算的 Execute 和 Reverse 操作**必须在单一数据库事务中完成**，确保以下步骤的原子性：

**Execute 事务步骤**：
1. 创建回收 Stock 记录（source="return"）
2. 创建 Settlement 记录
3. 创建 RecoveryItem 记录
4. 更新活动状态为 "settled"
5. 重算摊销快照
6. 写审计日志

**Reverse 事务步骤**：
1. 删除回收 Stock 记录
2. 更新结算状态为 "reversed"
3. 恢复活动状态为 "ended"
4. 重算摊销快照
5. 写审计日志

**Why**：任何一步失败都可能导致数据不一致（如库存已回收但活动状态未更新、快照与结算记录不匹配等）。SQLite 的单写模式天然串行化并发操作。

**How to apply**：`settlement/service.go` 中 Execute/Reverse 通过 `s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {...})` 包裹。事务内使用基于 tx 的临时 Repository 实例。

## 摊销快照重算也在事务内

摊销重算（`recalculateSnapshotsTx`）作为事务的一个步骤执行，不计入独立事务。这确保结算记录与快照数据始终一致。

## 审计日志

所有金额变动操作（Execute/Reverse）必须写审计日志，记录操作人、操作类型、影响金额。审计日志跟随事务成功/失败。
