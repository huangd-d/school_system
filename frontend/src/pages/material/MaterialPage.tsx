import { useMemo } from 'react'
import { Tabs } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { listCategories } from '@/api/material'
import { listActivities } from '@/api/activity'
import { useAuthStore } from '@/stores/authStore'
import CategoryTab from './CategoryTab'
import PurchaseStockTab from './PurchaseStockTab'
import DistributeTab from './DistributeTab'
import DistributionRecordsTab from './DistributionRecordsTab'

export default function MaterialPage() {
  const user = useAuthStore((s) => s.user)
  const isHQAdmin = user?.role === 'hq_admin'

  const { data: categories = [], isLoading: catLoading } = useQuery({
    queryKey: ['materialCategories'],
    queryFn: listCategories,
  })

  const { data: activities = [] } = useQuery({
    queryKey: ['activities'],
    queryFn: listActivities,
  })

  const tabItems = useMemo(() => {
    return [
      {
        key: 'distribute',
        label: '物资派发',
        children: (
          <DistributeTab
            categories={categories}
            activities={activities}
            isHQAdmin={isHQAdmin}
          />
        ),
      },
      {
        key: 'purchase-stock',
        label: '采购与库存',
        children: (
          <PurchaseStockTab categories={categories} isHQAdmin={isHQAdmin} />
        ),
      },
      {
        key: 'categories',
        label: '物资分类',
        children: (
          <CategoryTab
            categories={categories}
            isLoading={catLoading}
            isHQAdmin={isHQAdmin}
          />
        ),
      },
      {
        key: 'distribution-records',
        label: '派发记录',
        children: (
          <DistributionRecordsTab
            activities={activities}
            isHQAdmin={isHQAdmin}
          />
        ),
      },
    ]
  }, [categories, activities, isHQAdmin, catLoading])

  return (
    <div>
      <h2 className="text-lg font-bold mb-4">物资管理</h2>
      <Tabs defaultActiveKey="distribute" items={tabItems} />
    </div>
  )
}
