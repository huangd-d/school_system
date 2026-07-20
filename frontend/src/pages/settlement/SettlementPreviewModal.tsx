import { useEffect, useState } from 'react'
import { Modal, Table, Descriptions, Button, message, Spin } from 'antd'
import { DollarOutlined } from '@ant-design/icons'
import { useMutation } from '@tanstack/react-query'
import { previewSettlement, executeSettlement } from '@/api/settlement'
import type { SettlementPreviewItem, ActivityListItem } from '@/types'

interface Props {
  open: boolean
  activity: ActivityListItem | null
  onClose: () => void
  onSuccess: () => void
}

export default function SettlementPreviewModal({ open, activity, onClose, onSuccess }: Props) {
  const [loading, setLoading] = useState(false)
  const [previewData, setPreviewData] = useState<{
    items: SettlementPreviewItem[]
    total_returned_amount: number
    total_executed: number
    planned_executions: number
    activity_name: string
  } | null>(null)

  useEffect(() => {
    if (open && activity?.id) {
      setLoading(true)
      setPreviewData(null)
      previewSettlement(activity.id)
        .then(setPreviewData)
        .catch(e => message.error(e.message))
        .finally(() => setLoading(false))
    }
  }, [open, activity?.id])

  const executeMutation = useMutation({
    mutationFn: () => executeSettlement(activity!.id),
    onSuccess: () => {
      message.success('结算完成')
      onSuccess()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const columns = [
    { title: '物资名称', dataIndex: 'material_name' },
    { title: '配发数量', dataIndex: 'distributed_qty', width: 100 },
    { title: '已用量', dataIndex: 'used_qty', width: 80 },
    { title: '回收量', dataIndex: 'recovery_qty', width: 80 },
    {
      title: '单价',
      dataIndex: 'unit_price',
      width: 100,
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    {
      title: '扣减金额',
      dataIndex: 'cost_deduction',
      width: 120,
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
  ]

  return (
    <Modal
      title={`结算预览 — ${activity?.name ?? ''}`}
      open={open}
      onCancel={onClose}
      width={700}
      footer={[
        <Button key="cancel" onClick={onClose}>取消</Button>,
        <Button
          key="execute"
          type="primary"
          icon={<DollarOutlined />}
          loading={executeMutation.isPending}
          onClick={() => executeMutation.mutate()}
          disabled={!previewData || previewData.items.length === 0}
        >
          确认结算
        </Button>,
      ]}
    >
      {loading ? (
        <div className="flex justify-center py-8"><Spin tip="计算中..." /></div>
      ) : previewData ? (
        <>
          <Descriptions size="small" column={3} className="mb-4">
            <Descriptions.Item label="已执行次数">
              {previewData.total_executed} / {previewData.planned_executions}
            </Descriptions.Item>
            <Descriptions.Item label="回收总金额">
              ¥{previewData.total_returned_amount.toFixed(2)}
            </Descriptions.Item>
            <Descriptions.Item label="回收物资数">
              {previewData.items.filter(i => i.recovery_qty > 0).length} 种
            </Descriptions.Item>
          </Descriptions>

          <Table
            rowKey="stock_id"
            dataSource={previewData.items}
            columns={columns}
            pagination={false}
            size="small"
            summary={() =>
              previewData.total_returned_amount > 0 ? (
                <Table.Summary.Row>
                  <Table.Summary.Cell index={0} colSpan={5} align="right">
                    <strong>回收总金额：</strong>
                  </Table.Summary.Cell>
                  <Table.Summary.Cell index={1}>
                    <strong style={{ color: '#cf1322' }}>¥{previewData.total_returned_amount.toFixed(2)}</strong>
                  </Table.Summary.Cell>
                </Table.Summary.Row>
              ) : null
            }
          />
        </>
      ) : null}
    </Modal>
  )
}
