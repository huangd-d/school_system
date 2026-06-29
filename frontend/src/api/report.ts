import client from './client'
import type { AmortizationSnapshot, ApiResponse, ReportParams } from '@/types'

/** 摊销快照列表 */
export async function listSnapshots(params: {
  page: number
  pageSize: number
  startDate?: string
  endDate?: string
}): Promise<{ list: AmortizationSnapshot[]; total: number }> {
  const res = await client.get<ApiResponse<{ list: AmortizationSnapshot[]; total: number }>>(
    '/reports/snapshots',
    { params },
  )
  return res.data.data
}

/** 按维度统计 */
export async function getReport(params: ReportParams): Promise<Record<string, number>> {
  const res = await client.get<ApiResponse<Record<string, number>>>('/reports/statistics', { params })
  return res.data.data
}

/** 重算摊销 */
export async function recalculateAmortization(startDate: string, endDate: string): Promise<void> {
  await client.post('/reports/recalculate', { startDate, endDate })
}
