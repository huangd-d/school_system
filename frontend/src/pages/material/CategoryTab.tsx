import { useState } from 'react'
import { Button, Space, Table, Popconfirm, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteCategory } from '@/api/material'
import type { MaterialCategory } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import CategoryFormModal from './CategoryFormModal'

interface Props {
  categories: MaterialCategory[]
  isLoading: boolean
  isHQAdmin: boolean
}

export default function CategoryTab({ categories, isLoading, isHQAdmin }: Props) {
  const [modalOpen, setModalOpen] = useState(false)
  const [selected, setSelected] = useState<MaterialCategory | null>(null)
  const queryClient = useQueryClient()

  const deleteMutation = useMutation({
    mutationFn: deleteCategory,
    onSuccess: () => {
      message.success('分类已删除')
      queryClient.invalidateQueries({ queryKey: ['materialCategories'] })
    },
    onError: (e: Error) => message.error(e.message),
  })

  const columns: ColumnsType<MaterialCategory> = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', width: 150 },
    { title: '备注', dataIndex: 'note', ellipsis: true },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
  ]

  if (isHQAdmin) {
    columns.push({
      title: '操作',
      key: 'action',
      width: 160,
      render: (_: unknown, record: MaterialCategory) => (
        <Space>
          <Button
            type="link"
            size="small"
            onClick={() => {
              setSelected(record)
              setModalOpen(true)
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description={`确定要删除分类「${record.name}」吗？`}
            onConfirm={() => deleteMutation.mutate(record.id)}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button type="link" size="small" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    })
  }

  return (
    <div>
      {isHQAdmin && (
        <div className="mb-3">
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setSelected(null)
              setModalOpen(true)
            }}
          >
            新建分类
          </Button>
        </div>
      )}
      <Table
        rowKey="id"
        dataSource={categories}
        columns={columns}
        loading={isLoading}
        size="small"
        pagination={false}
      />
      <CategoryFormModal
        open={modalOpen}
        category={selected}
        onClose={() => {
          setModalOpen(false)
          setSelected(null)
        }}
        onSuccess={() => {
          setModalOpen(false)
          setSelected(null)
          queryClient.invalidateQueries({ queryKey: ['materialCategories'] })
        }}
      />
    </div>
  )
}
