---
name: settlement-amortization-formula
description: 结算回收算法和摊销快照计算公式
metadata:
  type: project
---

# 结算回收与摊销计算公式

## 摊销核心公式

```
摊销成本基数 = 活动配发物资总金额 - 已回收物资金额
累计摊销成本 = 摊销成本基数 × (累计执行次数 / 计划总次数)
每日摊销金额 = 摊销成本基数 × (当日执行次数 / 计划总次数)
```

## 结算回收算法

对活动的每条配发记录：

```
usedQty = floor(distribution.quantity × totalExecuted / plannedExecutions)
recoveryQty = distribution.quantity - usedQty
costDeduction = recoveryQty × unitPrice
```

**Why**：按执行进度比例计算已消耗物资量，未消耗的部分回收入库并扣减摊销基数。向下取整保证不超配发量。

**How to apply**：`settlement/service.go` 中 `computeRecoveryItems` 方法实现。Preview 和 Execute 共用同一计算逻辑。

## 摊销快照重算

结算执行和回撤时触发，在单事务中完成：
1. 删除活动全部现有快照
2. 计算摊销基数（配发总额 - 已结算回收总额）
3. 逐日统计执行次数，计算每日摊销额
4. Upsert 快照（利用 `(activity_id, date)` 唯一索引）

**Why**：确保结算/回撤后的摊销数据立即可用，且数据一致性由事务保证。
