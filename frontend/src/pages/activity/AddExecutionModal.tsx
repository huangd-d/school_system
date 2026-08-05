import { useEffect } from 'react'
import { Form, InputNumber, message, Descriptions } from 'antd'
import DraggableModal from '@/components/DraggableModal'
import { useMutation } from '@tanstack/react-query'
import { addExecution } from '@/api/activity'
import type { ActivityListItem, AddExecutionForm } from '@/types'
import dayjs from 'dayjs'

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
    // 当前日期未到开始日期时禁止录入
    const today = dayjs().format('YYYY-MM-DD')
    if (today < activity.start_date) {
      message.warning(`活动尚未开始，开始日期为 ${activity.start_date}`)
      return
    }
    addMutation.mutate({ id: activity.id, data: values })
  }

  if (!activity) return null

  const maxCount = activity.planned_executions - activity.total_executed

  return (
    <DraggableModal
      title="录入执行次数"
      open={open}
      onOk={() => form.submit()}
      onCancel={handleClose}
      confirmLoading={addMutation.isPending}
      destroyOnHidden
      width={600}
    >
      <Form form={form} onFinish={handleFinish}>
        <Descriptions size="small" column={1} className="mb-2">
          <Descriptions.Item label="活动名称">{activity.name}</Descriptions.Item>
          <Descriptions.Item label="已执行/未执行/总次数">
            {activity.total_executed} / {maxCount} / {activity.planned_executions}
          </Descriptions.Item>
          <Descriptions.Item label="本次执行次数" style={{ alignItems: 'center' }}>
            <Form.Item
              name="count"
              rules={[
                { required: true, message: '请输入执行次数' },
                { type: 'number', min: 1, message: '必须大于0' },
              ]}
              style={{ marginBottom: 0 }}
            >
              <InputNumber
                min={1}
                max={maxCount}
                style={{ width: '100%' }}
                placeholder={`1 ~ ${maxCount}`}
              />
            </Form.Item>
          </Descriptions.Item>
        </Descriptions>
      </Form>
    </DraggableModal>
  )
}
