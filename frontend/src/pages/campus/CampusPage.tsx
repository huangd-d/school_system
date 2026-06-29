import { useState } from 'react'
import {
  Table,
  Button,
  Space,
  Tag,
  Modal,
  Form,
  Input,
  Select,
  message,
  Popconfirm,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listCampuses,
  createCampus,
  updateCampus,
  deleteCampus,
} from '@/api/campus'
import type { Campus, CampusCreateForm, CampusUpdateForm, CampusType } from '@/types'

export default function CampusPage() {
  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Campus | null>(null)
  const [form] = Form.useForm()
  const queryClient = useQueryClient()

  // 后端返回平铺数组，无分页
  const { data: campuses, isLoading } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const createMutation = useMutation({
    mutationFn: createCampus,
    onSuccess: () => {
      message.success('校区创建成功')
      queryClient.invalidateQueries({ queryKey: ['campuses'] })
      closeModal()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: CampusUpdateForm }) =>
      updateCampus(id, data),
    onSuccess: () => {
      message.success('校区更新成功')
      queryClient.invalidateQueries({ queryKey: ['campuses'] })
      closeModal()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteCampus,
    onSuccess: () => {
      message.success('校区已删除')
      queryClient.invalidateQueries({ queryKey: ['campuses'] })
    },
    onError: (e: Error) => message.error(e.message),
  })

  const closeModal = () => {
    setModalOpen(false)
    setEditing(null)
    form.resetFields()
  }

  const openCreate = () => {
    setEditing(null)
    form.resetFields()
    setModalOpen(true)
  }

  const openEdit = (record: Campus) => {
    setEditing(record)
    form.setFieldsValue(record)
    setModalOpen(true)
  }

  const handleSubmit = () => {
    form.validateFields().then((values: CampusCreateForm) => {
      if (editing) {
        // 编辑时仅传 name
        updateMutation.mutate({ id: editing.id, data: { name: values.name } })
      } else {
        createMutation.mutate(values)
      }
    })
  }

  const typeLabel = (t: CampusType) => (t === 'hq' ? '总部' : '普通')

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

      <Modal
        title={editing ? '编辑校区' : '新建校区'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={closeModal}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        destroyOnClose
      >
        <Form form={form} layout="vertical" className="mt-4">
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请输入校区名称' }]}
          >
            <Input />
          </Form.Item>
          {/* 新建时可选择类型，编辑时仅显示不可改 */}
          {editing ? (
            <Form.Item label="类型">
              <Input value={typeLabel(editing.type)} disabled />
            </Form.Item>
          ) : (
            <Form.Item
              name="type"
              label="类型"
              rules={[{ required: true, message: '请选择类型' }]}
              initialValue="normal"
            >
              <Select
                options={[
                  { label: '总部', value: 'hq' },
                  { label: '普通校区', value: 'normal' },
                ]}
              />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  )
}
