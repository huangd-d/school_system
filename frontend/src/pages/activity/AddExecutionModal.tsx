import { useEffect } from 'react'
import { Form, InputNumber, Modal, message, Descriptions } from 'antd'
import { useMutation } from '@tanstack/react-query'
import { addExecution } from '@/api/activity'
import type { ActivityListItem, AddExecutionForm } from '@/types'

interface Props {
  open: boolean
  activity: ActivityListItem | null
  onClose: () => void
  onSuccess: () => void
}

export default function AddExecutionModal({ open, activity, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()

  const addMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: AddExecutionForm }) =>
      addExecution(id, data),
    onSuccess: () => {
      message.success('执行次数录入成功')
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

  const handleFinish = (values: { count: number }) => {
    if (!activity) return
    addMutation.mutate({ id: activity.id, data: values })
  }

  if (!activity) return null

  const maxCount = activity.planned_executions - activity.total_executed

  return (
    <Modal
      title="录入执行次数"
      open={open}
      onOk={() => form.submit()}
      onCancel={handleClose}
      confirmLoading={addMutation.isPending}
      destroyOnClose
    >
      <Descriptions
        size="small"
        column={1}
        className="mb-4"
        items={[
          { label: '活动名称', children: activity.name },
          { label: '计划次数', children: activity.planned_executions },
          { label: '已执行次数', children: activity.total_executed },
          { label: '剩余可录', children: maxCount },
        ]}
      />

      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
      >
        <Form.Item
          name="count"
          label="本次执行次数"
          rules={[
            { required: true, message: '请输入执行次数' },
            { type: 'number', min: 1, message: '必须大于0' },
          ]}
        >
          <InputNumber
            min={1}
            max={maxCount}
            className="w-full"
            placeholder={`1 ~ ${maxCount}`}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
