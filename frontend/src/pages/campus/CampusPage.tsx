import { useState } from 'react'
import { Table, Button, Space, Tag, Popconfirm, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listCampuses, deleteCampus } from '@/api/campus'
import type { Campus, CampusType } from '@/types'
import CampusFormModal from './CampusFormModal'

export default function CampusPage() {
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Campus | null>(null)
  const queryClient = useQueryClient()

  // 后端返回平铺数组，无分页
  const { data: campuses, isLoading } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const deleteMutation = useMutation({
    mutationFn: deleteCampus,
    onSuccess: () => {
      message.success('校区已删除')
      queryClient.invalidateQueries({ queryKey: ['campuses'] })
    },
    onError: (e: Error) => message.error(e.message),
  })

  const refreshList = () => {
    queryClient.invalidateQueries({ queryKey: ['campuses'] })
  }

  const openCreate = () => {
    setEditing(null)
    setModalOpen(true)
  }

  const openEdit = (record: Campus) => {
    setEditing(record)
    setModalOpen(true)
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name' },
    {
      title: '类型',
      dataIndex: 'type',
      width: 100,
      render: (t: CampusType) =>
        t === 'hq' ? (
          <Tag color="blue">总部</Tag>
        ) : (
          <Tag>普通</Tag>
        ),
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      render: (_: unknown, record: Campus) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确定删除该校区？"
            onConfirm={() => deleteMutation.mutate(record.id)}
          >
            <Button type="link" size="small" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">校区管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新建校区
        </Button>
      </div>

      <Table<Campus>
        rowKey="id"
        columns={columns}
        dataSource={campuses ?? []}
        loading={isLoading}
        pagination={false}
      />

      <CampusFormModal
        open={modalOpen}
        campus={editing}
        onClose={() => setModalOpen(false)}
        onSuccess={refreshList}
      />
    </div>
  )
}
