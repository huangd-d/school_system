import { useState } from 'react'
import { useNavigate } from 'react-router'
import { Table, Button, Space, Tag } from 'antd'
import { DollarOutlined, HistoryOutlined } from '@ant-design/icons'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getSettlementOverview } from '@/api/settlement'
import { useAuthStore } from '@/stores/authStore'
import type { ActivityStatus, SettlementOverviewItem } from '@/types'
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
  const [selectedActivity, setSelectedActivity] = useState<SettlementOverviewItem | null>(null)

  // 结算管理概览：后端一次返回可结算/已结算活动及结算后成本
  const { data: overview, isLoading } = useQuery({
    queryKey: ['settlement-overview'],
    queryFn: getSettlementOverview,
    enabled: !!role && role === 'hq_admin',
  })

  const openPreview = (record: SettlementOverviewItem) => {
    setSelectedActivity(record)
    setPreviewOpen(true)
  }

  const openHistory = (record: SettlementOverviewItem) => {
    setSelectedActivity(record)
    setHistoryOpen(true)
  }

  const handleSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ['settlement-overview'] })
  }

  // 仅 hq_admin 可访问
  if (role && role !== 'hq_admin') {
    navigate('/dashboard', { replace: true })
    return null
  }

  const columns = [
    { title: 'ID', dataIndex: 'activity_id', width: 60 },
    { title: '活动名称', dataIndex: 'activity_name' },
    {
      title: '执行进度',
      key: 'progress',
      width: 150,
      render: (_: unknown, r: SettlementOverviewItem) =>
        `${r.total_executed ?? 0} / ${r.planned_executions}`,
    },
    {
      title: '结算后物资成本',
      key: 'settled_cost',
      width: 140,
      render: (_: unknown, r: SettlementOverviewItem) =>
        r.status === 'settled' ? `¥${(r.settled_cost / 100).toFixed(2)}` : '-',
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
      render: (_: unknown, record: SettlementOverviewItem) => (
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

      <Table<SettlementOverviewItem>
        rowKey="activity_id"
        columns={columns}
        dataSource={overview ?? []}
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
        activityId={selectedActivity?.activity_id ?? null}
        activityName={selectedActivity?.activity_name ?? ''}
        onClose={() => setHistoryOpen(false)}
        onSuccess={handleSuccess}
      />
    </div>
  )
}
