import client from './client'
import type { ApiResponse, Settlement, SettlementPreview } from '@/types'

/** 结算预览 */
export async function previewSettlement(activityId: number): Promise<SettlementPreview> {
  const res = await client.post<ApiResponse<SettlementPreview>>(`/settlements/preview/${activityId}`)
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
