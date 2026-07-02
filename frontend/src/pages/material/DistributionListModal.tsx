import { useState } from 'react'
import { Button, Modal, Space, Table } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import type { Distribution, StockItem } from '@/types'
import AdjustDistributionModal from './AdjustDistributionModal'

interface Props {
  open: boolean
  stock: StockItem | null
  distributions: Distribution[]
  isLoading: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function DistributionListModal({
  open,
  stock,
  distributions,
  isLoading,
  onClose,
  onSuccess,
}: Props) {
  const [adjustOpen, setAdjustOpen] = useState(false)
  const [selectedDist, setSelectedDist] = useState<Distribution | null>(null)

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '活动ID', dataIndex: 'activity_id', width: 80 },
    { title: '数量', dataIndex: 'quantity', width: 80 },
    { title: '操作人ID', dataIndex: 'operator_id', width: 90 },
    { title: '原因', dataIndex: 'reason', ellipsis: true },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: unknown, record: Distribution) => (
        <Button
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => {
            setSelectedDist(record)
            setAdjustOpen(true)
          }}
        >
          调整
        </Button>
      ),
    },
  ]

  return (
    <>
      <Modal
        title="派发记录"
        open={open}
        onCancel={onClose}
        footer={null}
        width={800}
        destroyOnHidden
      >
        {stock && (
          <div className="mb-3 text-sm text-gray-500">
            物资：{stock.material_name} | 库存余量：{stock.remaining_qty} |
            单价：¥{stock.unit_price.toFixed(2)}
          </div>
        )}
        <Table
          rowKey="id"
          dataSource={distributions}
          columns={columns}
          loading={isLoading}
          size="small"
          pagination={false}
        />
      </Modal>

      <AdjustDistributionModal
        open={adjustOpen}
        distribution={selectedDist}
        stock={stock}
        onClose={() => {
          setAdjustOpen(false)
          setSelectedDist(null)
        }}
        onSuccess={() => {
          setAdjustOpen(false)
          setSelectedDist(null)
          onSuccess()
        }}
      />
    </>
  )
}
