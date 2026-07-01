import { useState, useMemo } from 'react'
import { Button, Input, Select, Space, Table } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { listPurchaseOrders, listStock } from '@/api/material'
import type { MaterialCategory, PurchaseOrder, StockItem } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import PurchaseFormModal from './PurchaseFormModal'

interface Props {
  categories: MaterialCategory[]
  isHQAdmin: boolean
}

export default function PurchaseStockTab({ categories, isHQAdmin }: Props) {
  const queryClient = useQueryClient()
  const [purchasePage, setPurchasePage] = useState(1)
  const [stockPage, setStockPage] = useState(1)
  const [stockFilter, setStockFilter] = useState<{ category_id?: number; keyword?: string }>({})

  const [purchaseModalOpen, setPurchaseModalOpen] = useState(false)

  // 分类 id→名称 映射
  const categoryMap = useMemo(() => {
    const map = new Map<number, string>()
    categories.forEach((c) => map.set(c.id, c.name))
    return map
  }, [categories])

  // 采购单列表（分页）
  const { data: purchaseData, isLoading: purchaseLoading } = useQuery({
    queryKey: ['purchases', purchasePage],
    queryFn: () => listPurchaseOrders({ page: purchasePage, page_size: 20 }),
  })

  // 库存列表（分页+筛选）
  const stockQueryKey = ['stock', { page: stockPage, page_size: 20, ...stockFilter }]
  const { data: stockData, isLoading: stockLoading } = useQuery({
    queryKey: stockQueryKey,
    queryFn: () =>
      listStock({ page: stockPage, page_size: 20, ...stockFilter }),
  })

  const onPurchaseSuccess = () => {
    queryClient.invalidateQueries({ queryKey: ['purchases'] })
    queryClient.invalidateQueries({ queryKey: ['stock'] })
  }

  const purchaseColumns: ColumnsType<PurchaseOrder> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '物资名称', dataIndex: 'material_name', width: 150 },
    {
      title: '分类',
      dataIndex: 'category_id',
      width: 80,
      render: (id: number) => categoryMap.get(id) || String(id),
    },
    { title: '数量', dataIndex: 'quantity', width: 80 },
    {
      title: '单价',
      dataIndex: 'unit_price',
      width: 80,
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    {
      title: '总金额',
      dataIndex: 'total_amount',
      width: 100,
      render: (v: number) => `¥${v.toFixed(2)}`,
    },
    { title: '备注', dataIndex: 'notes', ellipsis: true },
    { title: '采购人', dataIndex: 'purchaser_id', width: 80 },
    { title: '时间', dataIndex: 'created_at', width: 170 },
  ]

  const stockColumns: ColumnsType<StockItem> = [
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

  const categoryOptions = categories.map((c) => ({ label: c.name, value: c.id }))

  return (
    <div>
      {/* ===== 采购单 ===== */}
      <div className="mb-4">
        <div className="flex items-center justify-between mb-2">
          <h3 className="text-base font-semibold">采购记录</h3>
          {isHQAdmin && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setPurchaseModalOpen(true)}>
              新建采购
            </Button>
          )}
        </div>
        <Table
          rowKey="id"
          dataSource={purchaseData?.list ?? []}
          columns={purchaseColumns}
          loading={purchaseLoading}
          size="small"
          pagination={{
            current: purchasePage,
            pageSize: 20,
            total: purchaseData?.total ?? 0,
            showTotal: (total) => `共 ${total} 条`,
            onChange: setPurchasePage,
          }}
        />
      </div>

      {/* ===== 库存 ===== */}
      <div>
        <div className="mb-2">
          <h3 className="text-base font-semibold">库存列表</h3>
        </div>
        <div className="mb-3 flex gap-3">
          <Select
            allowClear
            placeholder="分类筛选"
            options={categoryOptions}
            value={stockFilter.category_id}
            onChange={(v) => {
              setStockFilter((prev) => ({ ...prev, category_id: v }))
              setStockPage(1)
            }}
            className="w-40"
          />
          <Input.Search
            allowClear
            placeholder="搜索物资名称"
            value={stockFilter.keyword}
            onChange={(e) =>
              setStockFilter((prev) => ({ ...prev, keyword: e.target.value }))
            }
            onSearch={(v) => {
              setStockFilter((prev) => ({ ...prev, keyword: v }))
              setStockPage(1)
            }}
            className="w-64"
          />
        </div>
        <Table
          rowKey="id"
          dataSource={stockData?.list ?? []}
          columns={stockColumns}
          loading={stockLoading}
          size="small"
          pagination={{
            current: stockPage,
            pageSize: 20,
            total: stockData?.total ?? 0,
            showTotal: (total) => `共 ${total} 条`,
            onChange: setStockPage,
          }}
        />
      </div>

      <PurchaseFormModal
        open={purchaseModalOpen}
        categories={categories}
        onClose={() => setPurchaseModalOpen(false)}
        onSuccess={onPurchaseSuccess}
      />
    </div>
  )
}
