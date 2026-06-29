import client from './client'
import type { Activity, ApiResponse, PaginatedData, PaginationParams } from '@/types'

export interface ActivityForm {
  name: string
  campusId: number
  startDate: string
  endDate: string
  contactIds: number[]
  remark?: string
}

/** 活动列表 */
export async function listActivities(params: PaginationParams): Promise<PaginatedData<Activity>> {
  const res = await client.get<ApiResponse<PaginatedData<Activity>>>('/activities', { params })
  return res.data.data
}

/** 活动详情 */
export async function getActivity(id: number): Promise<Activity> {
  const res = await client.get<ApiResponse<Activity>>(`/activities/${id}`)
  return res.data.data
}

/** 新建活动 */
export async function createActivity(data: ActivityForm): Promise<Activity> {
  const res = await client.post<ApiResponse<Activity>>('/activities', data)
  return res.data.data
}

/** 更新活动 */
export async function updateActivity(id: number, data: Partial<ActivityForm>): Promise<Activity> {
  const res = await client.put<ApiResponse<Activity>>(`/activities/${id}`, data)
  return res.data.data
}

/** 取消活动 */
export async function cancelActivity(id: number): Promise<void> {
  await client.put(`/activities/${id}/cancel`)
}

/** 记录执行 */
export async function recordExecution(activityId: number, date: string): Promise<void> {
  await client.post(`/activities/${activityId}/executions`, { date })
}
