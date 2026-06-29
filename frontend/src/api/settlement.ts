import client from './client'
import type { ApiResponse, PaginatedData, PaginationParams, Settlement } from '@/types'

/** 结算预览 */
export async function previewSettlement(activityId: number): Promise<Settlement> {
  const res = await client.get<ApiResponse<Settlement>>(`/settlements/preview/${activityId}`)
  return res.data.data
}

/** 执行结算 */
export async function executeSettlement(activityId: number): Promise<Settlement> {
  const res = await client.post<ApiResponse<Settlement>>(`/settlements/execute/${activityId}`)
  return res.data.data
}

/** 撤销结算 */
export async function reverseSettlement(settlementId: number): Promise<void> {
  await client.put(`/settlements/reverse/${settlementId}`)
}

/** 结算列表 */
export async function listSettlements(params: PaginationParams): Promise<PaginatedData<Settlement>> {
  const res = await client.get<ApiResponse<PaginatedData<Settlement>>>('/settlements', { params })
  return res.data.data
}
