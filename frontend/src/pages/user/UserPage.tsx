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
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listUsers, createUser, updateUser, disableUser, resetPassword } from '@/api/user'
import type { User, UserCreateForm, UserUpdateForm, Role } from '@/types'

const roleOptions: { label: string; value: Role }[] = [
  { label: '总部管理员', value: 'hq_admin' },
  { label: '校区操作员', value: 'campus_operator' },
  { label: '活动联系人', value: 'activity_contact' },
]

const roleMap: Record<Role, { label: string; color: string }> = {
  hq_admin: { label: '总部管理员', color: 'red' },
  campus_operator: { label: '校区操作员', color: 'blue' },
  activity_contact: { label: '活动联系人', color: 'green' },
}

export default function UserPage() {
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [pwdOpen, setPwdOpen] = useState(false)
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [pwdForm] = Form.useForm()
  const queryClient = useQueryClient()

  // 后端返回平铺数组，无分页
  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: listUsers,
  })

  const createMutation = useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      message.success('账户创建成功')
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setCreateOpen(false)
      createForm.resetFields()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UserUpdateForm }) =>
      updateUser(id, data),
    onSuccess: () => {
      message.success('账户更新成功')
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setEditOpen(false)
      editForm.resetFields()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const disableMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: () => {
      message.success('状态已更新')
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: (e: Error) => message.error(e.message),
  })

  const pwdMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: { password: string } }) =>
      resetPassword(id, data),
    onSuccess: () => {
      message.success('密码已重置')
      setPwdOpen(false)
      pwdForm.resetFields()
    },
    onError: (e: Error) => message.error(e.message),
  })

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      width: 120,
      render: (r: Role) => (
        <Tag color={roleMap[r]?.color}>{roleMap[r]?.label}</Tag>
      ),
    },
    { title: '校区ID', dataIndex: 'campus_id', width: 80 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (s: string) =>
        s === 'active' ? <Tag color="green">正常</Tag> : <Tag color="gray">禁用</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', width: 170 },
    {
      title: '操作',
      key: 'action',
      width: 240,
      render: (_: unknown, record: User) => (
        <Space>
          <Button
            type="link"
            size="small"
            onClick={() => {
              setSelectedUser(record)
              editForm.setFieldsValue({
                role: record.role,
                campus_id: record.campus_id,
              })
              setEditOpen(true)
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => disableMutation.mutate(record.id)}
          >
            {record.status === 'active' ? '禁用' : '启用'}
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => {
              setSelectedUser(record)
              setPwdOpen(true)
            }}
          >
            重置密码
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">账户管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
          新建账户
        </Button>
      </div>

      <Table<User>
        rowKey="id"
        columns={columns}
        dataSource={users ?? []}
        loading={isLoading}
        pagination={false}
      />

      {/* 新建账户弹窗 */}
      <Modal
        title="新建账户"
        open={createOpen}
        onOk={() => createForm.submit()}
        onCancel={() => { setCreateOpen(false); createForm.resetFields() }}
        confirmLoading={createMutation.isPending}
        destroyOnClose
      >
        <Form
          form={createForm}
          layout="vertical"
          className="mt-4"
          onFinish={(values: UserCreateForm) => createMutation.mutate(values)}
        >
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]} initialValue="campus_operator">
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="campus_id" label="校区 ID" rules={[{ required: true, message: '请输入校区 ID' }]}>
            <Input type="number" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑账户弹窗 — 仅可改 role 和 campus_id */}
      <Modal
        title={`编辑账户 - ${selectedUser?.username}`}
        open={editOpen}
        onOk={() => editForm.submit()}
        onCancel={() => { setEditOpen(false); editForm.resetFields() }}
        confirmLoading={updateMutation.isPending}
        destroyOnClose
      >
        <Form
          form={editForm}
          layout="vertical"
          className="mt-4"
          onFinish={(values: UserUpdateForm) => {
            if (selectedUser) {
              updateMutation.mutate({ id: selectedUser.id, data: values })
            }
          }}
        >
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="campus_id" label="校区 ID" rules={[{ required: true }]}>
            <Input type="number" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 重置密码弹窗 */}
      <Modal
        title={`重置密码 - ${selectedUser?.username}`}
        open={pwdOpen}
        onOk={() => pwdForm.submit()}
        onCancel={() => { setPwdOpen(false); pwdForm.resetFields() }}
        confirmLoading={pwdMutation.isPending}
        destroyOnClose
      >
        <Form
          form={pwdForm}
          layout="vertical"
          className="mt-4"
          onFinish={(values: { password: string }) =>
            selectedUser && pwdMutation.mutate({ id: selectedUser.id, data: values })
          }
        >
          <Form.Item name="password" label="新密码" rules={[{ required: true, message: '请输入新密码' }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
