import { useState } from 'react'
import { Table, Button, Space, Tag, message } from 'antd'
import { PlusOutlined, EyeOutlined, EditOutlined, PlusCircleOutlined, FolderOpenOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listActivities, archiveActivity } from '@/api/activity'
import { listCampuses } from '@/api/campus'
import { useAuthStore } from '@/stores/authStore'
import type { ActivityListItem, ActivityStatus } from '@/types'
import CreateActivityModal from './CreateActivityModal'
import EditActivityModal from './EditActivityModal'
import AddExecutionModal from './AddExecutionModal'
import ActivityDetailModal from './ActivityDetailModal'

const statusConfig: Record<ActivityStatus, { color: string; label: string }> = {
  not_started: { color: 'blue', label: '未开始' },
  in_progress: { color: 'green', label: '进行中' },
  ended: { color: 'orange', label: '已结束' },
  settled: { color: 'red', label: '已结算' },
  archived: { color: 'default', label: '已归档' },
}

export default function ActivityPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<ActivityListItem | null>(null)
  const [execActivity, setExecActivity] = useState<ActivityListItem | null>(null)
  const [detailId, setDetailId] = useState<number | null>(null)
  const queryClient = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const role = user?.role ?? ''

  const { data: activities, isLoading } = useQuery({
    queryKey: ['activities'],
    queryFn: listActivities,
  })

  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  // 校区 ID → 名称映射
  const campusMap: Record<number, string> = {}
  campuses?.forEach((c) => {
    campusMap[c.id] = c.name
  })

  const refreshList = () => {
    queryClient.invalidateQueries({ queryKey: ['activities'] })
  }

  const archiveMutation = useMutation({
    mutationFn: archiveActivity,
    onSuccess: () => {
      message.success('活动已归档')
      refreshList()
    },
    onError: (e: Error) => {
      message.error(e.message)
    },
  })

  // 是否可编辑（非归档 + hq_admin/campus_operator）
  const canEdit = (status: ActivityStatus) =>
    status !== 'archived' && (role === 'hq_admin' || role === 'campus_operator')

  // 是否可录入执行
  const canExec = (status: ActivityStatus) =>
    (status === 'not_started' || status === 'in_progress') &&
    (role === 'hq_admin' || role === 'activity_contact')

  // 是否可归档
  const canArchive = (status: ActivityStatus) =>
    status === 'settled' && (role === 'hq_admin' || role === 'campus_operator')

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name' },
    {
      title: '校区',
      dataIndex: 'campus_id',
      width: 100,
      render: (id: number) => campusMap[id] || `#${id}`,
    },
    {
      title: '执行进度',
      key: 'progress',
      width: 120,
      render: (_: unknown, r: ActivityListItem) =>
        `${r.total_executed} / ${r.planned_executions}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: ActivityStatus) => {
        const cfg = statusConfig[s]
        return <Tag color={cfg?.color}>{cfg?.label ?? s}</Tag>
      },
    },
    {
      title: '日期',
      key: 'date',
      width: 200,
      render: (_: unknown, r: ActivityListItem) => `${r.start_date} ~ ${r.end_date}`,
    },
    {
      title: '操作',
      key: 'action',
      width: 240,
      render: (_: unknown, r: ActivityListItem) => (
        <Space wrap>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => setDetailId(r.id)}
          >
            详情
          </Button>
          {canEdit(r.status) && (
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => setEditing(r)}
            >
              编辑
            </Button>
          )}
          {canExec(r.status) && (
            <Button
              type="link"
              size="small"
              icon={<PlusCircleOutlined />}
              onClick={() => setExecActivity(r)}
            >
              录入执行
            </Button>
          )}
          {canArchive(r.status) && (
            <Button
              type="link"
              size="small"
              icon={<FolderOpenOutlined />}
              loading={archiveMutation.isPending}
              onClick={() => archiveMutation.mutate(r.id)}
            >
              归档
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">活动管理</h2>
        {(role === 'hq_admin' || role === 'campus_operator') && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
            新建活动
          </Button>
        )}
      </div>

      <Table<ActivityListItem>
        rowKey="id"
        columns={columns}
        dataSource={activities ?? []}
        loading={isLoading}
        pagination={false}
      />

      <CreateActivityModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSuccess={refreshList}
      />

      <EditActivityModal
        open={editing !== null}
        activity={editing}
        onClose={() => setEditing(null)}
        onSuccess={refreshList}
      />

      <AddExecutionModal
        open={execActivity !== null}
        activity={execActivity}
        onClose={() => setExecActivity(null)}
        onSuccess={refreshList}
      />

      <ActivityDetailModal
        open={detailId !== null}
        activityId={detailId}
        onClose={() => setDetailId(null)}
      />
    </div>
  )
}
