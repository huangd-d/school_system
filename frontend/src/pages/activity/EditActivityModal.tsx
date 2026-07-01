import { useEffect } from 'react'
import { Form, Input, InputNumber, Modal, Select, message } from 'antd'
import { useMutation, useQuery } from '@tanstack/react-query'
import { updateActivity } from '@/api/activity'
import { listUsers } from '@/api/user'
import type { ActivityListItem, ActivityUpdateForm } from '@/types'

interface Props {
  open: boolean
  activity: ActivityListItem | null
  onClose: () => void
  onSuccess: () => void
}

export default function EditActivityModal({ open, activity, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const { data: users } = useQuery({
    queryKey: ['users'],
    queryFn: listUsers,
  })

  // 根据活动校区过滤联系人候选
  const contactOptions = (users ?? [])
    .filter((u) => !activity || u.campus_id === activity.campus_id)
    .map((u) => ({
      label: `${u.username} (${u.phone})`,
      value: u.id,
    }))

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: ActivityUpdateForm }) =>
      updateActivity(id, data),
    onSuccess: () => {
      message.success('活动更新成功')
      onSuccess()
      handleClose()
    },
    onError: (e: Error) => message.error(e.message),
  })

  // 回显数据
  useEffect(() => {
    if (open && activity) {
      form.setFieldsValue({
        name: activity.name,
        contact_ids: activity.contact_ids,
        planned_executions: activity.planned_executions,
      })
    }
  }, [open, activity, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleFinish = (values: Record<string, unknown>) => {
    if (!activity) return
    const data: ActivityUpdateForm = {}
    const newName = values.name as string | undefined
    const newContacts = values.contact_ids as number[] | undefined
    const newPlanned = values.planned_executions as number | undefined

    // 仅发送与现有值不同的字段
    if (newName && newName !== activity.name) {
      data.name = newName
    }
    if (newContacts !== undefined) {
      data.contact_ids = newContacts
    }
    if (newPlanned !== undefined && newPlanned !== activity.planned_executions) {
      data.planned_executions = newPlanned
    }

    if (Object.keys(data).length === 0) {
      message.info('没有需要修改的内容')
      handleClose()
      return
    }

    updateMutation.mutate({ id: activity.id, data })
  }

  if (!activity) return null

  return (
    <Modal
      title="编辑活动"
      open={open}
      onOk={() => form.submit()}
      onCancel={handleClose}
      confirmLoading={updateMutation.isPending}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        className="mt-4"
        onFinish={handleFinish}
      >
        <Form.Item name="name" label="活动名称">
          <Input placeholder="留空则不修改" maxLength={200} />
        </Form.Item>

        <Form.Item name="contact_ids" label="活动联系人">
          <Select
            mode="multiple"
            showSearch
            optionFilterProp="label"
            placeholder="留空则不修改"
            options={contactOptions}
          />
        </Form.Item>

        <Form.Item
          name="planned_executions"
          label="计划执行次数"
          rules={[
            { type: 'number', min: 1, message: '必须大于0' },
          ]}
          extra={`已执行 ${activity.total_executed} 次，修改后不能小于此值`}
        >
          <InputNumber
            min={activity.total_executed}
            className="w-full"
            placeholder="留空则不修改"
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
