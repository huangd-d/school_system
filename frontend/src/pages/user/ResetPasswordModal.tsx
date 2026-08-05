import { Form, Input, message } from 'antd'
import DraggableModal from '@/components/DraggableModal'
import { useMutation } from '@tanstack/react-query'
import { resetPassword } from '@/api/user'
import type { User } from '@/types'

interface Props {
  open: boolean
  user: User | null
  onClose: () => void
  onSuccess: () => void
}

export default function ResetPasswordModal({ open, user, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const pwdMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { password: string } }) =>
      resetPassword(id, data),
    onSuccess: () => {
      message.success('密码已重置')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  return (
    <DraggableModal
      title={`重置密码 - ${user?.username}`}
      open={open}
      onOk={() => form.submit()}
      onCancel={handleClose}
      confirmLoading={pwdMutation.isPending}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        className="mt-4"
        onFinish={(values: { password: string }) =>
          user && pwdMutation.mutate({ id: user.id, data: values })
        }
      >
        <Form.Item name="password" label="新密码" rules={[{ required: true, message: '请输入新密码' }]}>
          <Input.Password />
        </Form.Item>
      </Form>
    </DraggableModal>
  )
}
