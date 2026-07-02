import { Modal, Descriptions, Table, Tag, Spin } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { getActivity } from '@/api/activity'
import { listCampuses } from '@/api/campus'
import type { ActivityStatus, ExecutionRecord, UserBrief } from '@/types'

interface Props {
  open: boolean
  activityId: number | null
  onClose: () => void
}

const statusConfig: Record<ActivityStatus, { color: string; label: string }> = {
  not_started: { color: 'blue', label: '未开始' },
  in_progress: { color: 'green', label: '进行中' },
  ended: { color: 'orange', label: '已结束' },
  settled: { color: 'red', label: '已结算' },
  archived: { color: 'default', label: '已归档' },
}

export default function ActivityDetailModal({ open, activityId, onClose }: Props) {
  const { data: detail, isLoading } = useQuery({
    queryKey: ['activity', activityId],
    queryFn: () => getActivity(activityId!),
    enabled: open && activityId !== null,
  })

  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const campusMap: Record<number, string> = {}
  campuses?.forEach((c) => {
    campusMap[c.id] = c.name
  })

  const contactColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username' },
    { title: '手机号', dataIndex: 'phone' },
    {
      title: '角色',
      dataIndex: 'role',
      width: 120,
      render: (r: string) => {
        const labels: Record<string, string> = {
          hq_admin: '总部管理员',
          campus_operator: '校区操作员',
          activity_contact: '活动联系人',
        }
        return labels[r] || r
      },
    },
  ]

  const execColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '执行次数', dataIndex: 'count', width: 100 },
    { title: '记录人ID', dataIndex: 'recorded_by', width: 100 },
    { title: '记录时间', dataIndex: 'created_at', width: 180 },
  ]

  return (
    <Modal
      title="活动详情"
      open={open}
      onCancel={onClose}
      footer={null}
      width={720}
      destroyOnHidden
    >
      {isLoading ? (
        <div className="flex justify-center py-8">
          <Spin />
        </div>
      ) : detail ? (
        <div>
          <Descriptions
            size="small"
            column={2}
            bordered
            className="mb-6"
            items={[
              { label: 'ID', children: detail.id },
              { label: '活动名称', children: detail.name },
              {
                label: '所属校区',
                children: campusMap[detail.campus_id] || `#${detail.campus_id}`,
              },
              {
                label: '状态',
                children: (
                  <Tag color={statusConfig[detail.status]?.color}>
                    {statusConfig[detail.status]?.label ?? detail.status}
                  </Tag>
                ),
              },
              { label: '计划执行次数', children: detail.planned_executions },
              { label: '已执行次数', children: detail.total_executed },
              { label: '开始日期', children: detail.start_date },
              { label: '结束日期', children: detail.end_date },
              { label: '创建时间', children: detail.created_at },
            ]}
          />

          <h4 className="text-base font-medium mb-2">活动联系人</h4>
          <Table<UserBrief>
            rowKey="id"
            columns={contactColumns}
            dataSource={detail.contacts}
            pagination={false}
            size="small"
            className="mb-4"
            locale={{ emptyText: '暂无联系人' }}
          />

          <h4 className="text-base font-medium mb-2">执行记录</h4>
          <Table<ExecutionRecord>
            rowKey="id"
            columns={execColumns}
            dataSource={detail.executions}
            pagination={false}
            size="small"
            locale={{ emptyText: '暂无执行记录' }}
          />
        </div>
      ) : null}
    </Modal>
  )
}
