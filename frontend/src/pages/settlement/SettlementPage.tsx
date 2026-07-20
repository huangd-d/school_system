import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Table, Button, Space, Tag } from 'antd'
import { DollarOutlined, HistoryOutlined } from '@ant-design/icons'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { listActivities } from '@/api/activity'
import { useAuthStore } from '@/stores/authStore'
import type { ActivityListItem, ActivityStatus } from '@/types'
import SettlementPreviewModal from './SettlementPreviewModal'
import SettlementHistoryModal from './SettlementHistoryModal'

const statusMap: Record<ActivityStatus, { label: string; color: string }> = {
  not_started: { label: '未开始', color: 'default' },
  in_progress: { label: '进行中', color: 'processing' },
  ended: { label: '已结束', color: 'warning' },
  settled: { label: '已结算', color: 'success' },
  archived: { label: '已归档', color: 'default' },
}

export default function SettlementPage() {
  const navigate = useNavigate()
  const role = useAuthStore((s) => s.user?.role)
  const queryClient = useQueryClient()
  const [previewOpen, setPreviewOpen] = useState(false)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [selectedActivity, setSelectedActivity] = useState<ActivityListItem | null>(null)

  const { data: activities, isLoading } = useQuery({
    queryKey: ['activities'],
    queryFn: listActivities,
    enabled: !!role && role === 'hq_admin',
  })

  // 筛选可结算/已结算的活动
  const settleActivities = activities?.filter(
    (a) => a.status === 'ended' || a.status === 'settled',
  ) ?? []

  const openPreview = (record: ActivityListItem) => {
    setSelectedActivity(record)
    setPreviewOpen(true)
  }

  const openHistory = (record: ActivityListItem) => {
    setSelectedActivity(record)
    setHistoryOpen(true)
  }

  const handleSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ['activities'] })
  }

  // 仅 hq_admin 可访问
  if (role && role !== 'hq_admin') {
    navigate('/dashboard', { replace: true })
    return null
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '活动名称', dataIndex: 'name' },
    {
      title: '执行进度',
      key: 'progress',
      width: 150,
      render: (_: unknown, r: ActivityListItem) =>
        `${r.total_executed ?? 0} / ${r.planned_executions}`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: ActivityStatus) => (
        <Tag color={statusMap[s]?.color}>{statusMap[s]?.label}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: ActivityListItem) => (
        <Space>
          {record.status === 'ended' && (
            <Button
              type="primary"
              size="small"
              icon={<DollarOutlined />}
              onClick={() => openPreview(record)}
            >
              结算回收
            </Button>
          )}
          {record.status === 'settled' && (
            <Button
              type="default"
              size="small"
              icon={<HistoryOutlined />}
              onClick={() => openHistory(record)}
            >
              查看结算
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">结算管理</h2>
      </div>

      <Table<ActivityListItem>
        rowKey="id"
        columns={columns}
        dataSource={settleActivities}
        loading={isLoading}
        pagination={false}
      />

      <SettlementPreviewModal
        open={previewOpen}
        activity={selectedActivity}
        onClose={() => setPreviewOpen(false)}
        onSuccess={handleSuccess}
      />

      <SettlementHistoryModal
        open={historyOpen}
        activityId={selectedActivity?.id ?? null}
        activityName={selectedActivity?.name ?? ''}
        onClose={() => setHistoryOpen(false)}
        onSuccess={handleSuccess}
      />
    </div>
  )
}
