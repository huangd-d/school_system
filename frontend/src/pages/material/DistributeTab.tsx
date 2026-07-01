import { useState, useMemo } from 'react'
import { Button, Input, Select, Space, Table } from 'antd'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { listStock, getStockDistributions } from '@/api/material'
import type { ActivityListItem, MaterialCategory, StockItem } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import DistributeFormModal from './DistributeFormModal'
import DistributionListModal from './DistributionListModal'

interface Props {
  categories: MaterialCategory[]
  activities: ActivityListItem[]
  isHQAdmin: boolean
}

export default function DistributeTab({ categories, activities, isHQAdmin }: Props) {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [filter, setFilter] = useState<{ category_id?: number; keyword?: string }>({})

  const [distributeOpen, setDistributeOpen] = useState(false)
  const [distListOpen, setDistListOpen] = useState(false)
  const [selectedStock, setSelectedStock] = useState<StockItem | null>(null)

  // 分类 id→名称 映射
  const categoryMap = useMemo(() => {
    const map = new Map<number, string>()
    categories.forEach((c) => map.set(c.id, c.name))
    return map
  }, [categories])

  // 库存列表
  const queryKey = ['stock', { page, page_size: 20, ...filter }]
  const { data: stockData, isLoading } = useQuery({
    queryKey,
    queryFn: () => listStock({ page, page_size: 20, ...filter }),
  })

  // 查看派发记录时拉取
  const { data: distributions, isLoading: distLoading } = useQuery({
    queryKey: ['distributions', selectedStock?.id],
    queryFn: () => getStockDistributions(selectedStock!.id),
    enabled: distListOpen && !!selectedStock,
  })

  const onDistributeSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ['stock'] })
  }

  const onAdjustSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ['distributions'] })
    queryClient.invalidateQueries({ queryKey: ['stock'] })
  }

  const columns: ColumnsType<StockItem> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '物资名称', dataIndex: 'material_name', width: 150 },
    {
      title: '分类',
      dataIndex: 'category_id',
      width: 80,
      render: (id: number) => categoryMap.get(id) || String(id),
    },
    { title: '总量', dataIndex: 'total_quantity', width: 70 },
    {
      title: '剩余',
      dataIndex: 'remaining_qty',
      width: 70,
      render: (v: number) => (
        <span className={v < 5 ? 'text-red-500 font-bold' : ''}>{v}</span>
      ),
    },
    {
      title: '单价',
      dataIndex: 'unit_price',
      width: 80,
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    {
      title: '来源',
      dataIndex: 'source',
      width: 90,
      render: (s: string) => (
        <span className={s === 'purchase' ? 'text-blue-600' : 'text-green-600'}>
          {s === 'purchase' ? '采购入库' : '结算回收'}
        </span>
      ),
    },
    { title: '更新时间', dataIndex: 'updated_at', width: 170 },
  ]

  if (isHQAdmin) {
    columns.push({
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: StockItem) => (
        <Space>
          <Button
            type="link"
            size="small"
            disabled={record.remaining_qty <= 0}
            onClick={() => {
              setSelectedStock(record)
              setDistributeOpen(true)
            }}
          >
            派发
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => {
              setSelectedStock(record)
              setDistListOpen(true)
            }}
          >
            查看派发
          </Button>
        </Space>
      ),
    })
  }

  const categoryOptions = categories.map((c) => ({ label: c.name, value: c.id }))

  return (
    <div>
      <div className="mb-3 flex gap-3">
        <Select
          allowClear
          placeholder="分类筛选"
          options={categoryOptions}
          value={filter.category_id}
          onChange={(v) => {
            setFilter((prev) => ({ ...prev, category_id: v }))
            setPage(1)
          }}
          className="w-40"
        />
        <Input.Search
          allowClear
          placeholder="搜索物资名称"
          value={filter.keyword}
          onChange={(e) =>
            setFilter((prev) => ({ ...prev, keyword: e.target.value }))
          }
          onSearch={(v) => {
            setFilter((prev) => ({ ...prev, keyword: v }))
            setPage(1)
          }}
          className="w-64"
        />
      </div>
      <Table
        rowKey="id"
        dataSource={stockData?.list ?? []}
        columns={columns}
        loading={isLoading}
        size="small"
        pagination={{
          current: page,
          pageSize: 20,
          total: stockData?.total ?? 0,
          showTotal: (total) => `共 ${total} 条`,
          onChange: setPage,
        }}
      />

      <DistributeFormModal
        open={distributeOpen}
        stock={selectedStock}
        activities={activities}
        onClose={() => {
          setDistributeOpen(false)
          setSelectedStock(null)
        }}
        onSuccess={onDistributeSuccess}
      />

      <DistributionListModal
        open={distListOpen}
        stock={selectedStock}
        distributions={distributions ?? []}
        isLoading={distLoading}
        onClose={() => {
          setDistListOpen(false)
          setSelectedStock(null)
        }}
        onSuccess={onAdjustSuccess}
      />
    </div>
  )
}
