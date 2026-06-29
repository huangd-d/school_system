import client from './client'
import type {
  ApiResponse,
  Inventory,
  MaterialCategory,
  PaginatedData,
  PaginationParams,
  PurchaseOrder,
} from '@/types'

// ===== 物资分类 =====

export async function listCategories(): Promise<MaterialCategory[]> {
  const res = await client.get<ApiResponse<MaterialCategory[]>>('/materials/categories')
  return res.data.data
}

export async function createCategory(data: { name: string; remark?: string }): Promise<MaterialCategory> {
  const res = await client.post<ApiResponse<MaterialCategory>>('/materials/categories', data)
  return res.data.data
}

// ===== 采购单 =====

export interface PurchaseForm {
  categoryId: number
  materialName: string
  quantity: number
  unit: string
  unitPrice: number
  purchaseDate: string
  remark?: string
}

export async function listPurchases(params: PaginationParams): Promise<PaginatedData<PurchaseOrder>> {
  const res = await client.get<ApiResponse<PaginatedData<PurchaseOrder>>>('/materials/purchases', { params })
  return res.data.data
}

export async function createPurchase(data: PurchaseForm): Promise<PurchaseOrder> {
  const res = await client.post<ApiResponse<PurchaseOrder>>('/materials/purchases', data)
  return res.data.data
}

// ===== 库存 =====

export async function listInventories(params: PaginationParams): Promise<PaginatedData<Inventory>> {
  const res = await client.get<ApiResponse<PaginatedData<Inventory>>>('/materials/inventories', { params })
  return res.data.data
}
