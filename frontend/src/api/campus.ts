import client from './client'
import type { ApiResponse, Campus, CampusCreateForm, CampusUpdateForm } from '@/types'

/** 校区列表 — 后端返回平铺数组，无分页 */
export async function listCampuses(): Promise<Campus[]> {
  const res = await client.get<ApiResponse<Campus[]>>('/campuses')
  return res.data.data
}

/** 新建校区 */
export async function createCampus(data: CampusCreateForm): Promise<Campus> {
  const res = await client.post<ApiResponse<Campus>>('/campuses', data)
  return res.data.data
}

/** 更新校区 — 仅名称可改 */
export async function updateCampus(id: number, data: CampusUpdateForm): Promise<Campus> {
  const res = await client.put<ApiResponse<Campus>>(`/campuses/${id}`, data)
  return res.data.data
}

/** 删除校区 */
export async function deleteCampus(id: number): Promise<void> {
  await client.delete(`/campuses/${id}`)
}
