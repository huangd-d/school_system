import { useEffect } from 'react'
import { Form, Input, Modal, Select, message } from 'antd'
import { useMutation, useQuery } from '@tanstack/react-query'
import { updateUser } from '@/api/user'
import { listCampuses } from '@/api/campus'
import type { User, UserUpdateForm, Role } from '@/types'

const roleOptions: { label: string; value: Role }[] = [
  { label: '总部管理员', value: 'hq_admin' },
  { label: '校区操作员', value: 'campus_operator' },
  { label: '活动联系人', value: 'activity_contact' },
]

interface Props {
  open: boolean
  user: User | null
  onClose: () => void
  onSuccess: () => void
}

export default function EditUserModal({ open, user, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UserUpdateForm }) =>
      updateUser(id, data),
    onSuccess: () => {
      message.success('账户更新成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  // 打开弹框时回显数据
  useEffect(() => {
    if (open && user) {
      form.setFieldsValue({
        username: user.username,
        phone: user.phone,
        role: user.role,
        campus_id: user.campus_id,
      })
    }
  }, [open, user, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  return (
    <Modal
      title={`编辑账户 - ${user?.username}`}
      open={open}
      onOk={() => form.submit()}
      onCancel={handleClose}
      confirmLoading={updateMutation.isPending}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        className="mt-4"
        onFinish={(values: UserUpdateForm) => {
          if (user) {
            updateMutation.mutate({ id: user.id, data: values })
          }
        }}
      >
        <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }]}>
          <Input maxLength={20} />
        </Form.Item>
        <Form.Item name="role" label="角色" rules={[{ required: true }]}>
          <Select options={roleOptions} />
        </Form.Item>
        <Form.Item name="campus_id" label="校区" rules={[{ required: true, message: '请选择校区' }]}>
          <Select
            showSearch
            optionFilterProp="label"
            placeholder="请选择校区"
            options={campuses?.map(c => ({ label: c.name, value: c.id })) ?? []}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
