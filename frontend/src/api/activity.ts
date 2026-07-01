import client from './client'
import type {
  Activity,
  ActivityListItem,
  ActivityDetail,
  ActivityCreateForm,
  ActivityUpdateForm,
  AddExecutionForm,
  ApiResponse,
} from '@/types'

/** 活动列表 */
export async function listActivities(): Promise<ActivityListItem[]> {
  const res = await client.get<ApiResponse<ActivityListItem[]>>('/activities')
  return res.data.data
}

/** 活动详情 */
export async function getActivity(id: number): Promise<ActivityDetail> {
  const res = await client.get<ApiResponse<ActivityDetail>>(`/activities/${id}`)
  return res.data.data
}

/** 新建活动 */
export async function createActivity(data: ActivityCreateForm): Promise<Activity> {
  const res = await client.post<ApiResponse<Activity>>('/activities', data)
  return res.data.data
}

/** 更新活动 */
export async function updateActivity(id: number, data: ActivityUpdateForm): Promise<Activity> {
  const res = await client.put<ApiResponse<Activity>>(`/activities/${id}`, data)
  return res.data.data
}

/** 录入执行次数 */
export async function addExecution(activityId: number, data: AddExecutionForm): Promise<void> {
  await client.post(`/activities/${activityId}/executions`, data)
}

/** 归档活动 */
export async function archiveActivity(id: number): Promise<void> {
  await client.put(`/activities/${id}/archive`)
}
