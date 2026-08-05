import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router'
import { Tabs, Select, DatePicker, Table, Card, Statistic, Spin, Empty } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { listActivities } from '@/api/activity'
import { listCampuses } from '@/api/campus'
import {
  getReportByActivity,
  getReportByDateRange,
  getReportByCampus,
  getReportByCategory,
} from '@/api/report'
import { useAuthStore } from '@/stores/authStore'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line, PieChart, Pie, Cell, Legend,
} from 'recharts'
import dayjs from 'dayjs'

const COLORS = ['#8884d8', '#82ca9d', '#ffc658', '#ff7300', '#0088fe', '#00c49f', '#ff8042', '#a4de6c']

export default function ReportPage() {
  const navigate = useNavigate()
  const role = useAuthStore((s) => s.user?.role)
  const [activeTab, setActiveTab] = useState('activity')
  const [dateRange, setDateRange] = useState<[string, string]>([
    dayjs().subtract(30, 'day').format('YYYY-MM-DD'),
    dayjs().format('YYYY-MM-DD'),
  ])
  const [selectedActivityId, setSelectedActivityId] = useState<number | undefined>()
  const [selectedCampusId, setSelectedCampusId] = useState<number | undefined>()

  // 仅 hq_admin 可访问
  if (role && role !== 'hq_admin') {
    navigate('/dashboard', { replace: true })
    return null
  }

  const { data: activities } = useQuery({
    queryKey: ['activities'],
    queryFn: listActivities,
    enabled: !!role && role === 'hq_admin',
  })

  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
    enabled: !!role && role === 'hq_admin',
  })

  // 按活动报表
  const activityReport = useQuery({
    queryKey: ['reportActivity', selectedActivityId],
    queryFn: () => getReportByActivity(selectedActivityId!),
    enabled: activeTab === 'activity' && !!selectedActivityId,
  })

  // 按日期报表
  const dateRangeReport = useQuery({
    queryKey: ['reportDateRange', dateRange],
    queryFn: () => getReportByDateRange(dateRange[0], dateRange[1]),
    enabled: activeTab === 'date_range' && !!dateRange[0],
  })

  // 按校区报表
  const campusReport = useQuery({
    queryKey: ['reportCampus', selectedCampusId, dateRange],
    queryFn: () => getReportByCampus(selectedCampusId ?? 0, dateRange[0], dateRange[1]),
    enabled: activeTab === 'campus' && !!dateRange[0],
  })

  // 按品类报表
  const categoryReport = useQuery({
    queryKey: ['reportCategory', dateRange],
    queryFn: () => getReportByCategory(dateRange[0], dateRange[1]),
    enabled: activeTab === 'category' && !!dateRange[0],
  })

  const chartData = useMemo(() => {
    if (activityReport.data) {
      return [
        { name: '总投资', value: activityReport.data.total_investment / 100 },
        { name: '总摊销', value: activityReport.data.total_amortization / 100 },
      ]
    }
    return []
  }, [activityReport.data])

  const pieData = useMemo(() => {
    return (categoryReport.data ?? []).map(c => ({
      name: c.category_name,
      value: c.total_amount / 100,
    }))
  }, [categoryReport.data])

  const tabItems = [
    {
      key: 'activity',
      label: '按活动',
      children: (
        <div>
          <div className="mb-4">
            <Select
              placeholder="选择活动"
              style={{ width: 300 }}
              showSearch
              optionFilterProp="label"
              value={selectedActivityId}
              onChange={setSelectedActivityId}
              options={activities?.map(a => ({ value: a.id, label: a.name })) ?? []}
            />
          </div>
          {activityReport.isLoading ? <Spin /> : activityReport.data ? (
            <div>
              <div className="grid grid-cols-3 gap-4 mb-6">
                <Card><Statistic title="总投资" value={activityReport.data.total_investment / 100} prefix="¥" precision={2} /></Card>
                <Card><Statistic title="总摊销" value={activityReport.data.total_amortization / 100} prefix="¥" precision={2} /></Card>
                <Card><Statistic title="执行进度" value={`${activityReport.data.total_executed} / ${activityReport.data.planned_executions}`} /></Card>
              </div>
              <Card title="投资 vs 摊销">
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="value" fill="#8884d8" name="金额" />
                  </BarChart>
                </ResponsiveContainer>
              </Card>
            </div>
          ) : <Empty description="选择活动查看报表" />}
        </div>
      ),
    },
    {
      key: 'date_range',
      label: '按日期',
      children: (
        <div>
          <DatePicker.RangePicker
            className="mb-4"
            value={[dayjs(dateRange[0]), dayjs(dateRange[1])]}
            onChange={(dates) => {
              if (dates?.[0] && dates?.[1]) {
                setDateRange([dates[0].format('YYYY-MM-DD'), dates[1].format('YYYY-MM-DD')])
              }
            }}
          />
          {dateRangeReport.isLoading ? <Spin /> : dateRangeReport.data ? (
            <Card title="每日摊销趋势">
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={(dateRangeReport.data ?? []).map(d => ({ ...d, daily_amount: d.daily_amount / 100 }))}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" />
                  <YAxis />
                  <Tooltip />
                  <Line type="monotone" dataKey="daily_amount" stroke="#8884d8" name="日摊销额" />
                </LineChart>
              </ResponsiveContainer>
            </Card>
          ) : <Empty />}
        </div>
      ),
    },
    {
      key: 'campus',
      label: '按校区',
      children: (
        <div>
          <div className="mb-4 flex gap-4">
            <Select
              placeholder="选择校区（全部）"
              style={{ width: 200 }}
              allowClear
              value={selectedCampusId}
              onChange={setSelectedCampusId}
              options={campuses?.map(c => ({ value: c.id, label: c.name })) ?? []}
            />
            <DatePicker.RangePicker
              value={[dayjs(dateRange[0]), dayjs(dateRange[1])]}
              onChange={(dates) => {
                if (dates?.[0] && dates?.[1]) {
                  setDateRange([dates[0].format('YYYY-MM-DD'), dates[1].format('YYYY-MM-DD')])
                }
              }}
            />
          </div>
          {campusReport.isLoading ? <Spin /> : campusReport.data ? (
            <div>
              <div className="grid grid-cols-3 gap-4 mb-6">
                <Card><Statistic title="校区" value={campusReport.data.campus_name} /></Card>
                <Card><Statistic title="活动数" value={campusReport.data.activity_count} /></Card>
                <Card><Statistic title="总投资" value={campusReport.data.total_investment / 100} prefix="¥" precision={2} /></Card>
              </div>
            </div>
          ) : <Empty />}
        </div>
      ),
    },
    {
      key: 'category',
      label: '按品类',
      children: (
        <div>
          <DatePicker.RangePicker
            className="mb-4"
            value={[dayjs(dateRange[0]), dayjs(dateRange[1])]}
            onChange={(dates) => {
              if (dates?.[0] && dates?.[1]) {
                setDateRange([dates[0].format('YYYY-MM-DD'), dates[1].format('YYYY-MM-DD')])
              }
            }}
          />
          {categoryReport.isLoading ? <Spin /> : categoryReport.data ? (
            <div className="grid grid-cols-2 gap-4">
              <Card title="品类分布">
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={100} label>
                      {pieData.map((_, i) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
                    </Pie>
                    <Tooltip />
                    <Legend />
                  </PieChart>
                </ResponsiveContainer>
              </Card>
              <Card title="品类明细">
                <Table
                  rowKey="category_id"
                  dataSource={categoryReport.data}
                  pagination={false}
                  size="small"
                  columns={[
                    { title: '品类', dataIndex: 'category_name' },
                    { title: '数量', dataIndex: 'total_quantity', width: 80 },
                    {
                      title: '金额',
                      dataIndex: 'total_amount',
                      width: 120,
                      render: (v: number) => `¥${(v / 100).toFixed(2)}`,
                    },
                  ]}
                />
              </Card>
            </div>
          ) : <Empty />}
        </div>
      ),
    },
  ]

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">成本报表</h2>
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
    </div>
  )
}
