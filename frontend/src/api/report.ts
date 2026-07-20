import client from './client'
import type { ActivityReport, ApiResponse, CampusReport, CategoryReportItem, DateRangeReportItem } from '@/types'

/** 按活动维度报表 */
export async function getReportByActivity(activityId: number): Promise<ActivityReport> {
  const res = await client.get<ApiResponse<ActivityReport>>('/reports/by-activity', {
    params: { activity_id: activityId },
  })
  return res.data.data
}

/** 按日期范围维度报表 */
export async function getReportByDateRange(startDate: string, endDate: string): Promise<DateRangeReportItem[]> {
  const res = await client.get<ApiResponse<DateRangeReportItem[]>>('/reports/by-date-range', {
    params: { start_date: startDate, end_date: endDate },
  })
  return res.data.data
}

/** 按校区维度报表 */
export async function getReportByCampus(campusId: number, startDate: string, endDate: string): Promise<CampusReport> {
  const res = await client.get<ApiResponse<CampusReport>>('/reports/by-campus', {
    params: { campus_id: campusId, start_date: startDate, end_date: endDate },
  })
  return res.data.data
}

/** 按品类维度报表 */
export async function getReportByCategory(startDate: string, endDate: string): Promise<CategoryReportItem[]> {
  const res = await client.get<ApiResponse<CategoryReportItem[]>>('/reports/by-category', {
    params: { start_date: startDate, end_date: endDate },
  })
  return res.data.data
}
