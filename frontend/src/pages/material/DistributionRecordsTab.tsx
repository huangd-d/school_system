import { useState } from 'react'
import { Button, DatePicker, Input, Select, Space, Table } from 'antd'
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import { listAllDistributions } from '@/api/material'
import type { ActivityListItem, DistributionRecord, DistributionQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

interface Props {
  activities: ActivityListItem[]
  isHQAdmin: boolean
}

export default function DistributionRecordsTab({ activities }: Props) {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)

  // 筛选条件（内部状态，点击"查询"时才生效）
  const [filterDraft, setFilterDraft] = useState<{
    activity_id?: number
    keyword?: string
    dateRange?: [dayjs.Dayjs, dayjs.Dayjs] | null
  }>({})

  // 实际生效的筛选条件
  const [appliedFilter, setAppliedFilter] = useState<DistributionQuery>({})

  const queryKey = ['allDistributions', { page, page_size: pageSize, ...appliedFilter }]
  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => listAllDistributions({ page, page_size: pageSize, ...appliedFilter }),
  })

  const handleSearch = () => {
    const query: DistributionQuery = {}
    if (filterDraft.activity_id) {
      query.activity_id = filterDraft.activity_id
    }
    if (filterDraft.keyword) {
      query.keyword = filterDraft.keyword
    }
    if (filterDraft.dateRange && filterDraft.dateRange[0] && filterDraft.dateRange[1]) {
      query.start_date = filterDraft.dateRange[0].format('YYYY-MM-DD')
      query.end_date = filterDraft.dateRange[1].format('YYYY-MM-DD')
    }
    setPage(1)
    setAppliedFilter(query)
  }

  const handleReset = () => {
    setFilterDraft({})
    setAppliedFilter({})
    setPage(1)
  }

  const columns: ColumnsType<DistributionRecord> = [
    {
      title: '物资名称',
      dataIndex: 'material_name',
      key: 'material_name',
      width: 150,
    },
    {
      title: '活动名称',
      dataIndex: 'activity_name',
      key: 'activity_name',
      width: 180,
    },
    {
      title: '派发数量',
      dataIndex: 'quantity',
      key: 'quantity',
      width: 100,
    },
    {
      title: '派发原因',
      dataIndex: 'reason',
      key: 'reason',
      width: 200,
      render: (v: string) => v || '-',
    },
    {
      title: '操作人ID',
      dataIndex: 'operator_id',
      key: 'operator_id',
      width: 100,
    },
    {
      title: '派发时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
    },
  ]

  return (
    <div>
      {/* 筛选栏 */}
      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          allowClear
          placeholder="选择活动"
          value={filterDraft.activity_id}
          onChange={(v) => setFilterDraft((prev) => ({ ...prev, activity_id: v }))}
          options={activities.map((a) => ({ label: a.name, value: a.id }))}
          style={{ width: 180 }}
        />
        <Input
          allowClear
          placeholder="物资名称"
          value={filterDraft.keyword}
          onChange={(e) => setFilterDraft((prev) => ({ ...prev, keyword: e.target.value }))}
          style={{ width: 180 }}
        />
        <RangePicker
          value={filterDraft.dateRange ?? null}
          onChange={(dates) => setFilterDraft((prev) => ({ ...prev, dateRange: dates as [dayjs.Dayjs, dayjs.Dayjs] | null }))}
          placeholder={['开始日期', '结束日期']}
        />
        <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
          查询
        </Button>
        <Button icon={<ReloadOutlined />} onClick={handleReset}>
          重置
        </Button>
      </Space>

      {/* 数据表格 */}
      <Table<DistributionRecord>
        rowKey="id"
        columns={columns}
        dataSource={data?.list ?? []}
        loading={isLoading}
        pagination={{
          current: page,
          pageSize,
          total: data?.total ?? 0,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (p, ps) => {
            setPage(p)
            setPageSize(ps)
          },
        }}
      />
    </div>
  )
}
