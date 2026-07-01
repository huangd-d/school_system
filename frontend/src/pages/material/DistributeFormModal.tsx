import { useEffect } from 'react'
import { Descriptions, Form, Input, InputNumber, Modal, Select, message } from 'antd'
import { useMutation } from '@tanstack/react-query'
import { distribute } from '@/api/material'
import type { ActivityListItem, DistributeForm, StockItem } from '@/types'

interface Props {
  open: boolean
  stock: StockItem | null
  activities: ActivityListItem[]
  onClose: () => void
  onSuccess: () => void
}

export default function DistributeFormModal({ open, stock, activities, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const mutation = useMutation({
    mutationFn: distribute,
    onSuccess: () => {
      message.success('派发成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  useEffect(() => {
    if (open) {
      form.resetFields()
    }
  }, [open, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleSubmit = () => {
    if (!stock) return
    form.validateFields().then((values: Omit<DistributeForm, 'stock_id'>) => {
      mutation.mutate({ ...values, stock_id: stock.id })
    })
  }

  if (!stock) return null

  const activityOptions = activities
    .filter((a) => a.status !== 'archived')
    .map((a) => ({
      label: `${a.name} (执行 ${a.total_executed}/${a.planned_executions})`,
      value: a.id,
    }))

  return (
    <Modal
      title="派发物资"
      open={open}
      onOk={handleSubmit}
      onCancel={handleClose}
      confirmLoading={mutation.isPending}
      destroyOnClose
      width={520}
    >
      <Descriptions column={2} size="small" className="mb-4" bordered>
        <Descriptions.Item label="物资名称">{stock.material_name}</Descriptions.Item>
        <Descriptions.Item label="库存余量">{stock.remaining_qty}</Descriptions.Item>
        <Descriptions.Item label="单价">¥{stock.unit_price.toFixed(2)}</Descriptions.Item>
        <Descriptions.Item label="来源">
          {stock.source === 'purchase' ? '采购入库' : '结算回收'}
        </Descriptions.Item>
      </Descriptions>

      <Form form={form} layout="vertical" className="mt-4">
        <Form.Item
          name="activity_id"
          label="目标活动"
          rules={[{ required: true, message: '请选择目标活动' }]}
        >
          <Select
            placeholder="选择活动"
            options={activityOptions}
            showSearch
            optionFilterProp="label"
          />
        </Form.Item>
        <Form.Item
          name="quantity"
          label="派发数量"
          rules={[
            { required: true, message: '请输入派发数量' },
            {
              type: 'number',
              max: stock.remaining_qty,
              message: `派发数量不能超过库存余量 (${stock.remaining_qty})`,
            },
          ]}
        >
          <InputNumber
            min={1}
            max={stock.remaining_qty}
            precision={0}
            placeholder={`1~${stock.remaining_qty}`}
            className="w-full"
          />
        </Form.Item>
        <Form.Item
          name="reason"
          label="派发原因"
          rules={[{ max: 500, message: '原因不能超过500个字符' }]}
        >
          <Input.TextArea rows={2} placeholder="可选，如：新学期开班需要" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
