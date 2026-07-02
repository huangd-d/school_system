import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router'
import { Table, Button, Space, Tag, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listUsers, disableUser, enableUser } from '@/api/user'
import { listCampuses } from '@/api/campus'
import { useAuthStore } from '@/stores/authStore'
import type { User, Role } from '@/types'
import CreateUserModal from './CreateUserModal'
import EditUserModal from './EditUserModal'
import ResetPasswordModal from './ResetPasswordModal'

const roleMap: Record<Role, { label: string; color: string }> = {
  hq_admin: { label: '总部管理员', color: 'red' },
  campus_operator: { label: '校区操作员', color: 'blue' },
  activity_contact: { label: '活动联系人', color: 'green' },
}

export default function UserPage() {
  const navigate = useNavigate()
  const role = useAuthStore((s) => s.user?.role)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [pwdOpen, setPwdOpen] = useState(false)

  // 仅总部管理员可访问
  if (role && role !== 'hq_admin') {
    navigate('/dashboard', { replace: true })
    return null
  }
  const [selectedUser, setSelectedUser] = useState<User | null>(null)
  const queryClient = useQueryClient()

  // 后端返回平铺数组，无分页
  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: listUsers,
  })

  // 校区列表 — 用于 id→名称 映射，queryKey ['campuses'] 与 CampusPage 共享缓存
  const { data: campuses } = useQuery({
    queryKey: ['campuses'],
    queryFn: listCampuses,
  })

  const campusMap = useMemo(() => {
    const map = new Map<number, string>()
    campuses?.forEach(c => map.set(c.id, c.name))
    return map
  }, [campuses])

  const disableMutation = useMutation({
    mutationFn: disableUser,
    onSuccess: () => {
      message.success('状态已更新')
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: (e: Error) => message.error(e.message),
  })

  const enableMutation = useMutation({
    mutationFn: enableUser,
    onSuccess: () => {
      message.success('状态已更新')
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
    onError: (e: Error) => message.error(e.message),
  })

  const refreshList = () => {
    queryClient.invalidateQueries({ queryKey: ['users'] })
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username' },
    { title: '手机号', dataIndex: 'phone', width: 130 },
    {
      title: '角色',
      dataIndex: 'role',
      width: 120,
      render: (r: Role) => (
        <Tag color={roleMap[r]?.color}>{roleMap[r]?.label}</Tag>
      ),
    },
    {
      title: '校区',
      dataIndex: 'campus_id',
      width: 120,
      render: (id: number) => campusMap.get(id) ?? `校区 #${id}`,
    },
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
              setEditOpen(true)
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => {
              if (record.status === 'active') {
                disableMutation.mutate(record.id)
              } else {
                enableMutation.mutate(record.id)
              }
            }}
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

      <CreateUserModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSuccess={refreshList}
      />

      <EditUserModal
        open={editOpen}
        user={selectedUser}
        onClose={() => setEditOpen(false)}
        onSuccess={refreshList}
      />

      <ResetPasswordModal
        open={pwdOpen}
        user={selectedUser}
        onClose={() => setPwdOpen(false)}
        onSuccess={refreshList}
      />
    </div>
  )
}
