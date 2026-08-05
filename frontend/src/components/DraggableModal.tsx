import { useRef, useState } from 'react'
import type { ReactNode } from 'react'
import Draggable from 'react-draggable'
import type { DraggableData, DraggableEvent } from 'react-draggable'
import { Modal } from 'antd'
import type { ModalProps } from 'antd'

/**
 * 可拖动弹框：按住标题栏可拖动（antd 官方推荐方案：modalRender + react-draggable）。
 *
 * 采用官方 demo 的包裹 div 写法（见 ant.design/components/modal-cn#modal-demo-modal-render）：
 * nodeRef 绑定到自建的 <div>，而不是 cloneElement 注入到 antd 内部元素上。
 * React 19 移除了 findDOMNode，react-draggable 必须使用 nodeRef 模式；包裹 div 的原因是：
 * 1. cloneElement 注入 ref 会整体替换 antd 内部 .ant-modal 的 mergedRef（holderRef / panelRef /
 *    internalRef），破坏 antd 内部机制（动画状态检测、焦点管理等）；
 * 2. 拖动 transform 与 antd zoom 动画的 transform（@keyframes）隔离到不同元素，互不干扰。
 *
 * 注意：不能用 bounds="parent"。antd 6 的 modalRender 目标是 .ant-modal-container（嵌套在
 * .ant-modal 内部），包裹 div 的父节点是 .ant-modal 自身 —— 两者矩形重合时 react-draggable
 * 的 getBoundPosition 会算出 left/right/top/bottom 全为 0，导致完全拖不动。因此按官方 demo
 * 在 onStart 时基于视口（documentElement.clientWidth/Height）动态计算边界。
 */
export default function DraggableModal(props: ModalProps) {
  const nodeRef = useRef<HTMLDivElement>(null)
  const [bounds, setBounds] = useState({ left: 0, top: 0, bottom: 0, right: 0 })

  const onStart = (_event: DraggableEvent, uiData: DraggableData) => {
    const { clientWidth, clientHeight } = window.document.documentElement
    const targetRect = nodeRef.current?.getBoundingClientRect()
    if (!targetRect) {
      return
    }
    setBounds({
      left: -targetRect.left + uiData.x,
      right: clientWidth - (targetRect.right - uiData.x),
      top: -targetRect.top + uiData.y,
      bottom: clientHeight - (targetRect.bottom - uiData.y),
    })
  }

  return (
    <Modal
      {...props}
      styles={
        typeof props.styles === 'function'
          ? props.styles
          : { ...props.styles, header: { cursor: 'move', ...props.styles?.header } }
      }
      modalRender={(modal: ReactNode) => (
        <Draggable
          nodeRef={nodeRef}
          handle=".ant-modal-header"
          bounds={bounds}
          onStart={onStart}
        >
          <div ref={nodeRef}>{modal}</div>
        </Draggable>
      )}
    />
  )
}
