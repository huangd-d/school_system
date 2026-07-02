import { useEffect } from 'react'
import { Form, Input, InputNumber, Modal, Select, message } from 'antd'
import { useMutation } from '@tanstack/react-query'
import { createPurchase } from '@/api/material'
import type { MaterialCategory, PurchaseForm } from '@/types'

interface Props {
  open: boolean
  categories: MaterialCategory[]
  onClose: () => void
  onSuccess: () => void
}

export default function PurchaseFormModal({ open, categories, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const mutation = useMutation({
    mutationFn: createPurchase,
    onSuccess: () => {
      message.success('采购入库成功')
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
    form.validateFields().then((values: PurchaseForm) => {
      mutation.mutate(values)
    })
  }

  const categoryOptions = categories.map((c) => ({ label: c.name, value: c.id }))

  return (
    <Modal
      title="新建采购"
      open={open}
      onOk={handleSubmit}
      onCancel={handleClose}
      confirmLoading={mutation.isPending}
      destroyOnHidden
      width={520}
    >
      <Form form={form} layout="vertical" className="mt-4">
        <Form.Item
          name="material_name"
          label="物资名称"
          rules={[
            { required: true, message: '请输入物资名称' },
            { max: 200, message: '物资名称不能超过200个字符' },
          ]}
        >
          <Input placeholder="例如：语文教材" />
        </Form.Item>
        <Form.Item
          name="category_id"
          label="物资分类"
          rules={[{ required: true, message: '请选择物资分类' }]}
        >
          <Select
            placeholder="选择分类"
            options={categoryOptions}
            showSearch
            optionFilterProp="label"
          />
        </Form.Item>
        <Form.Item
          name="quantity"
          label="采购数量"
          rules={[{ required: true, message: '请输入数量' }]}
        >
          <InputNumber min={1} precision={0} placeholder="正整数" className="w-full" />
        </Form.Item>
        <Form.Item
          name="total_amount"
          label="总金额 (¥)"
          rules={[{ required: true, message: '请输入总金额' }]}
        >
          <InputNumber
            min={0.01}
            step={0.01}
            precision={2}
            prefix="¥"
            placeholder="大于0"
            className="w-full"
          />
        </Form.Item>
        <Form.Item
          name="notes"
          label="备注"
          rules={[{ max: 500, message: '备注不能超过500个字符' }]}
        >
          <Input.TextArea rows={3} placeholder="可选备注" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
