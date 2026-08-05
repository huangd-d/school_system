import client from './client'
import type { ApiResponse, Settlement, SettlementOverviewItem, SettlementPreview } from '@/types'

/** 结算预览（只读查询，GET） */
export async function previewSettlement(activityId: number): Promise<SettlementPreview> {
  const res = await client.get<ApiResponse<SettlementPreview>>(`/settlements/preview/${activityId}`)
  return res.data.data
}

/** 执行结算 */
export async function executeSettlement(activityId: number): Promise<Settlement> {
  const res = await client.post<ApiResponse<Settlement>>(`/settlements/execute/${activityId}`)
  return res.data.data
}

/** 撤销结算 */
export async function reverseSettlement(settlementId: number): Promise<void> {
  await client.post(`/settlements/reverse/${settlementId}`)
}

/** 结算管理概览（一次返回可结算/已结算活动表格数据） */
export async function getSettlementOverview(): Promise<SettlementOverviewItem[]> {
  const res = await client.get<ApiResponse<SettlementOverviewItem[]>>('/settlements/overview')
  return res.data.data ?? []
}
