import { useEffect, useMemo } from 'react'
import { Descriptions, Form, Input, InputNumber, Modal, Tag, message } from 'antd'
import { useMutation } from '@tanstack/react-query'
import { adjustDistribution } from '@/api/material'
import type { AdjustDistributionForm, Distribution, StockItem } from '@/types'

interface Props {
  open: boolean
  distribution: Distribution | null
  stock: StockItem | null
  onClose: () => void
  onSuccess: () => void
}

export default function AdjustDistributionModal({
  open,
  distribution,
  stock,
  onClose,
  onSuccess,
}: Props) {
  const [form] = Form.useForm()

  const mutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: AdjustDistributionForm }) =>
      adjustDistribution(id, data),
    onSuccess: () => {
      message.success('派发调整成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  useEffect(() => {
    if (open && distribution) {
      form.setFieldsValue({ quantity: distribution.quantity, reason: '' })
    }
  }, [open, distribution, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleSubmit = () => {
    if (!distribution) return
    form.validateFields().then((values: AdjustDistributionForm) => {
      mutation.mutate({ id: distribution.id, data: values })
    })
  }

  const newQty = Form.useWatch('quantity', form)
  const diff = useMemo(() => {
    if (!distribution || newQty == null) return 0
    return newQty - distribution.quantity
  }, [newQty, distribution])

  if (!distribution || !stock) return null

  return (
    <Modal
      title="调整派发"
      open={open}
      onOk={handleSubmit}
      onCancel={handleClose}
      confirmLoading={mutation.isPending}
      destroyOnHidden
      width={520}
    >
      <Descriptions column={2} size="small" className="mb-4" bordered>
        <Descriptions.Item label="物资名称">{stock.material_name}</Descriptions.Item>
        <Descriptions.Item label="库存余量">{stock.remaining_qty}</Descriptions.Item>
        <Descriptions.Item label="当前派发量">{distribution.quantity}</Descriptions.Item>
        <Descriptions.Item label="目标活动ID">{distribution.activity_id}</Descriptions.Item>
      </Descriptions>

      <Form form={form} layout="vertical" className="mt-4">
        <Form.Item
          name="quantity"
          label="调整后数量"
          rules={[
            { required: true, message: '请输入调整后数量' },
            {
              validator: (_, value) => {
                if (value != null && value <= 0) {
                  return Promise.reject(new Error('数量必须大于0'))
                }
                if (distribution && value > stock.remaining_qty + distribution.quantity) {
                  return Promise.reject(
                    new Error(`不能超过最大可派发量 (${stock.remaining_qty + distribution.quantity})`),
                  )
                }
                return Promise.resolve()
              },
            },
          ]}
        >
          <InputNumber min={1} precision={0} className="w-full" />
        </Form.Item>

        {diff !== 0 && (
          <div className="mb-4">
            <Tag color={diff > 0 ? 'blue' : 'orange'}>
              {diff > 0
                ? `将追加派发 ${diff} 件`
                : `将退回库存 ${Math.abs(diff)} 件`}
            </Tag>
          </div>
        )}

        <Form.Item
          name="reason"
          label="调整原因"
          rules={[
            { required: true, message: '请输入调整原因' },
            { max: 500, message: '原因不能超过500个字符' },
          ]}
        >
          <Input.TextArea rows={2} placeholder="必填，用于审计追踪" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
