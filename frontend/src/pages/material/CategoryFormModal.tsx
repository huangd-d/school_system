import { useEffect } from 'react'
import { Form, Input, message } from 'antd'
import DraggableModal from '@/components/DraggableModal'
import { useMutation } from '@tanstack/react-query'
import { createCategory, updateCategory } from '@/api/material'
import type { MaterialCategory, CategoryCreateForm } from '@/types'

interface Props {
  open: boolean
  category: MaterialCategory | null
  onClose: () => void
  onSuccess: () => void
}

export default function CategoryFormModal({ open, category, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = category !== null

  const createMutation = useMutation({
    mutationFn: createCategory,
    onSuccess: () => {
      message.success('分类创建成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CategoryCreateForm }) =>
      updateCategory(id, data),
    onSuccess: () => {
      message.success('分类更新成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  useEffect(() => {
    if (open) {
      if (category) {
        form.setFieldsValue(category)
      } else {
        form.resetFields()
      }
    }
  }, [open, category, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleSubmit = () => {
    form.validateFields().then((values: CategoryCreateForm) => {
      if (isEdit) {
        updateMutation.mutate({ id: category!.id, data: values })
      } else {
        createMutation.mutate(values)
      }
    })
  }

  return (
    <DraggableModal
      title={isEdit ? '编辑分类' : '新建分类'}
      open={open}
      onOk={handleSubmit}
      onCancel={handleClose}
      confirmLoading={createMutation.isPending || updateMutation.isPending}
      destroyOnHidden
    >
      <Form form={form} layout="vertical" className="mt-4">
        <Form.Item
          name="name"
          label="名称"
          rules={[
            { required: true, message: '请输入分类名称' },
            { max: 50, message: '分类名称不能超过50个字符' },
          ]}
        >
          <Input placeholder="例如：教材、文具、设备" />
        </Form.Item>
        <Form.Item
          name="note"
          label="备注"
          rules={[{ max: 200, message: '备注不能超过200个字符' }]}
        >
          <Input.TextArea rows={3} placeholder="可选备注" />
        </Form.Item>
      </Form>
    </DraggableModal>
  )
}
