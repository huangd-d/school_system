import { useEffect } from 'react'
import { Form, Input, Select, message } from 'antd'
import DraggableModal from '@/components/DraggableModal'
import { useMutation } from '@tanstack/react-query'
import { createUser } from '@/api/user'
import { listCampuses } from '@/api/campus'
import { useQuery } from '@tanstack/react-query'
import type { UserCreateForm, Role } from '@/types'

const roleOptions: { label: string; value: Role }[] = [
  { label: '总部管理员', value: 'hq_admin' },
  { label: '校区操作员', value: 'campus_operator' },
  { label: '活动联系人', value: 'activity_contact' },
]

interface Props {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function CreateUserModal({ open, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const createMutation = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      message.success('账户创建成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  // 打开弹框时重置表单
  useEffect(() => {
    if (open) {
      form.resetFields()
    }
  }, [open, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  return (
    <DraggableModal
      title="新建账户"
      open={open}
      onOk={() => form.submit()}
      onCancel={handleClose}
      confirmLoading={createMutation.isPending}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        className="mt-4"
        onFinish={(values: UserCreateForm) => createMutation.mutate(values)}
      >
        <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="password" label="密码" rules={[{ required: true }]}>
          <Input.Password />
        </Form.Item>
        <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }]}>
          <Input maxLength={20} />
        </Form.Item>
        <Form.Item name="role" label="角色" rules={[{ required: true }]} initialValue="campus_operator">
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
    </DraggableModal>
  )
}
