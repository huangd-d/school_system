import client from './client'
import type {
  AdjustDistributionForm,
  ApiResponse,
  CategoryCreateForm,
  DistributeForm,
  Distribution,
  DistributionQuery,
  DistributionRecord,
  MaterialCategory,
  PaginatedData,
  PaginationParams,
  PurchaseForm,
  PurchaseOrder,
  StockItem,
  StockQuery,
} from '@/types'

// ===== 物资分类 =====

export async function listCategories(): Promise<MaterialCategory[]> {
  const res = await client.get<ApiResponse<MaterialCategory[]>>('/materials/categories')
  return res.data.data
}

export async function createCategory(data: CategoryCreateForm): Promise<MaterialCategory> {
  const res = await client.post<ApiResponse<MaterialCategory>>('/materials/categories', data)
  return res.data.data
}

export async function updateCategory(id: number, data: CategoryCreateForm): Promise<MaterialCategory> {
  const res = await client.put<ApiResponse<MaterialCategory>>(`/materials/categories/${id}`, data)
  return res.data.data
}

export async function deleteCategory(id: number): Promise<void> {
  await client.delete(`/materials/categories/${id}`)
}

// ===== 采购 =====

export async function listPurchaseOrders(params: PaginationParams): Promise<PaginatedData<PurchaseOrder>> {
  const res = await client.get<ApiResponse<PaginatedData<PurchaseOrder>>>('/materials/purchases', { params })
  return res.data.data
}

export async function createPurchase(data: PurchaseForm): Promise<StockItem> {
  const res = await client.post<ApiResponse<StockItem>>('/materials/purchase', data)
  return res.data.data
}

// ===== 库存 =====

export async function listStock(params: StockQuery): Promise<PaginatedData<StockItem>> {
  const res = await client.get<ApiResponse<PaginatedData<StockItem>>>('/materials/stock', { params })
  return res.data.data
}

export async function getStock(id: number): Promise<StockItem> {
  const res = await client.get<ApiResponse<StockItem>>(`/materials/stock/${id}`)
  return res.data.data
}

export async function getStockDistributions(stockId: number): Promise<Distribution[]> {
  const res = await client.get<ApiResponse<Distribution[]>>(`/materials/stock/${stockId}/distributions`)
  return res.data.data
}

// ===== 派发 =====

export async function distribute(data: DistributeForm): Promise<Distribution> {
  const res = await client.post<ApiResponse<Distribution>>('/materials/distribute', data)
  return res.data.data
}

export async function adjustDistribution(id: number, data: AdjustDistributionForm): Promise<void> {
  await client.put(`/materials/distribute/${id}`, data)
}

// ===== 派发记录查询 =====

export async function listAllDistributions(params: DistributionQuery): Promise<PaginatedData<DistributionRecord>> {
  const res = await client.get<ApiResponse<PaginatedData<DistributionRecord>>>('/materials/distributions', { params })
  return res.data.data
}
