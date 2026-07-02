import { useEffect } from 'react'
import { Form, Input, Modal, Select, message } from 'antd'
import { useMutation } from '@tanstack/react-query'
import { createCampus, updateCampus } from '@/api/campus'
import type { Campus, CampusCreateForm, CampusUpdateForm } from '@/types'

const typeLabel = (t: string) => (t === 'hq' ? '总部' : '普通')

interface Props {
  open: boolean
  campus: Campus | null   // null = 新建，非 null = 编辑
  onClose: () => void
  onSuccess: () => void
}

export default function CampusFormModal({ open, campus, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const isEdit = campus !== null

  const createMutation = useMutation({
    mutationFn: createCampus,
    onSuccess: () => {
      message.success('校区创建成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CampusUpdateForm }) =>
      updateCampus(id, data),
    onSuccess: () => {
      message.success('校区更新成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  // 打开弹框时回显数据
  useEffect(() => {
    if (open) {
      if (campus) {
        form.setFieldsValue(campus)
      } else {
        form.resetFields()
      }
    }
  }, [open, campus, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleSubmit = () => {
    form.validateFields().then((values: CampusCreateForm) => {
      if (isEdit) {
        updateMutation.mutate({ id: campus!.id, data: { name: values.name } })
      } else {
        createMutation.mutate(values)
      }
    })
  }

  return (
    <Modal
      title={isEdit ? '编辑校区' : '新建校区'}
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
          rules={[{ required: true, message: '请输入校区名称' }]}
        >
          <Input />
        </Form.Item>
        {/* 新建时可选择类型，编辑时仅显示不可改 */}
        {isEdit ? (
          <Form.Item label="类型">
            <Input value={typeLabel(campus!.type)} disabled />
          </Form.Item>
        ) : (
          <Form.Item
            name="type"
            label="类型"
            rules={[{ required: true, message: '请选择类型' }]}
            initialValue="normal"
          >
            <Select
              options={[
                { label: '总部', value: 'hq' },
                { label: '普通校区', value: 'normal' },
              ]}
            />
          </Form.Item>
        )}
      </Form>
    </Modal>
  )
}
