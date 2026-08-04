# frontend/CLAUDE.md

本文件仅在 Claude 处理 `frontend/` 目录下的文件时加载。

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
