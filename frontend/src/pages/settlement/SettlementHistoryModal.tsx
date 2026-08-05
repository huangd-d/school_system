import { useEffect, useState } from 'react'
import { Table, Tag, Button, Popconfirm, message, Descriptions } from 'antd'
import DraggableModal from '@/components/DraggableModal'
import { UndoOutlined } from '@ant-design/icons'
import { useMutation } from '@tanstack/react-query'
import { reverseSettlement } from '@/api/settlement'
import client from '@/api/client'
import type { ApiResponse, Settlement } from '@/types'

interface Props {
  open: boolean
  activityId: number | null
  activityName: string
  onClose: () => void
  onSuccess: () => void
}

export default function SettlementHistoryModal({ open, activityId, activityName, onClose, onSuccess }: Props) {
  const [settlements, setSettlements] = useState<Settlement[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open && activityId) {
      setLoading(true)
      // 获取该活动的结算记录（从结算列表直接查活动维度的数据）
      client.get<ApiResponse<Settlement[]>>(`/settlements/by-activity/${activityId}`)
        .then(res => setSettlements(res.data.data ?? []))
        .catch(() => setSettlements([]))
        .finally(() => setLoading(false))
    }
  }, [open, activityId])

  const reverseMutation = useMutation({
    mutationFn: reverseSettlement,
    onSuccess: () => {
      message.success('结算已撤销')
      onSuccess()
    },
    onError: (e: Error) => message.error(e.message),
  })

  return (
    <DraggableModal
      title={`「${activityName}」结算记录`}
      open={open}
      onCancel={onClose}
      footer={null}
      width={700}
    >
      <Table<Settlement>
        rowKey="id"
        dataSource={settlements}
        loading={loading}
        pagination={false}
        columns={[
          { title: '结算ID', dataIndex: 'id', width: 80 },
          {
            title: '状态',
            dataIndex: 'status',
            width: 100,
            render: (s: string) =>
              s === 'settled' ? <Tag color="success">已结算</Tag> : <Tag color="default">已回撤</Tag>,
          },
          {
            title: '回收总金额',
            dataIndex: 'total_returned_amount',
            width: 120,
            render: (v: number) => `¥${(v / 100).toFixed(2)}`,
          },
          { title: '结算时间', dataIndex: 'created_at', width: 170 },
          {
            title: '操作',
            key: 'action',
            width: 120,
            render: (_: unknown, record: Settlement) =>
              record.status === 'settled' ? (
                <Popconfirm
                  title="确定撤销该结算？"
                  onConfirm={() => reverseMutation.mutate(record.id)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button type="link" size="small" icon={<UndoOutlined />} danger>
                    撤销
                  </Button>
                </Popconfirm>
              ) : null,
          },
        ]}
        expandable={{
          expandedRowRender: (record: Settlement) => (
            <Descriptions size="small" column={2}>
              <Descriptions.Item label="操作人ID">{record.operator_id}</Descriptions.Item>
              <Descriptions.Item label="回收金额">¥{(record.total_returned_amount / 100).toFixed(2)}</Descriptions.Item>
            </Descriptions>
          ),
        }}
      />
    </DraggableModal>
  )
}
