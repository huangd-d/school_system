import { useEffect } from 'react'
import { Form, Input, InputNumber, Modal, Select, DatePicker, message } from 'antd'
import { useMutation, useQuery } from '@tanstack/react-query'
import { createActivity } from '@/api/activity'
import { listCampuses } from '@/api/campus'
import { listUsers } from '@/api/user'
import type { ActivityCreateForm } from '@/types'
import dayjs from 'dayjs'

interface Props {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function CreateActivityModal({ open, onClose, onSuccess }: Props) {
  const [form] = Form.useForm()
  const selectedCampusId: number | undefined = Form.useWatch('campus_id', form)

  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const { data: users } = useQuery({
    queryKey: ['users'],
    queryFn: listUsers,
  })

  // 根据所选校区过滤联系人候选 不需要根据校区过滤，一般都是团队合作。
  const contactOptions = (users ?? [])
    .map((u) => ({
      label: `${u.username} (${u.phone})`,
      value: u.id,
    }))
    // .filter((u) => !selectedCampusId || u.campus_id === selectedCampusId)

  const createMutation = useMutation({
    mutationFn: createActivity,
    onSuccess: () => {
      message.success('活动创建成功')
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

  // 校区变更时清空已选联系人
  useEffect(() => {
    if (selectedCampusId) {
      form.setFieldValue('contact_ids', undefined)
    }
  }, [selectedCampusId, form])

  const handleClose = () => {
    form.resetFields()
    onClose()
  }

  const handleFinish = (values: Record<string, unknown>) => {
    const data: ActivityCreateForm = {
      name: values.name as string,
      campus_id: values.campus_id as number,
      contact_ids: values.contact_ids as number[] | undefined,
      planned_executions: values.planned_executions as number,
      start_date: (values.start_date as dayjs.Dayjs).format('YYYY-MM-DD'),
      end_date: (values.end_date as dayjs.Dayjs).format('YYYY-MM-DD'),
    }
    createMutation.mutate(data)
  }

  return (
    <Modal
      title="新建活动"
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
        onFinish={handleFinish}
      >
        <Form.Item
          name="name"
          label="活动名称"
          rules={[
            { required: true, message: '请输入活动名称' },
            { max: 200, message: '名称不能超过200个字符' },
          ]}
        >
          <Input />
        </Form.Item>

        <Form.Item
          name="campus_id"
          label="所属校区"
          rules={[{ required: true, message: '请选择校区' }]}
        >
          <Select
            showSearch
            optionFilterProp="label"
            placeholder="请选择校区"
            options={campuses?.map((c) => ({ label: c.name, value: c.id })) ?? []}
          />
        </Form.Item>

        <Form.Item
          name="contact_ids"
          label="活动联系人"
          tooltip="可选多个，支持跨校区协作"
        >
          <Select
            mode="multiple"
            showSearch
            optionFilterProp="label"
            placeholder="请选择联系人（可选）"
            options={contactOptions}
            disabled={!selectedCampusId}
          />
        </Form.Item>

        <Form.Item
          name="planned_executions"
          label="计划执行次数"
          rules={[
            { required: true, message: '请输入计划执行次数' },
            { type: 'number', min: 1, message: '必须大于0' },
          ]}
        >
          <InputNumber min={1} className="w-full" />
        </Form.Item>

        <Form.Item
          name="start_date"
          label="开始日期"
          rules={[{ required: true, message: '请选择开始日期' }]}
        >
          <DatePicker className="w-full" format="YYYY-MM-DD" />
        </Form.Item>

        <Form.Item
          name="end_date"
          label="结束日期"
          rules={[{ required: true, message: '请选择结束日期' }]}
        >
          <DatePicker className="w-full" format="YYYY-MM-DD" />
        </Form.Item>
      </Form>
    </Modal>
  )
}
