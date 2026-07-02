import { useEffect, useState, useMemo } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router'
import { Layout, Menu, Button, theme } from 'antd'
import {
  DashboardOutlined,
  BankOutlined,
  UserOutlined,
  ScheduleOutlined,
  ShopOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { useAppStore } from '@/stores/appStore'

const { Header, Sider, Content } = Layout

const allMenuItems = [
  { key: '/dashboard', icon: <DashboardOutlined />, label: '工作台', roles: ['hq_admin', 'campus_operator', 'activity_contact'] },
  { key: '/campuses', icon: <BankOutlined />, label: '校区管理', roles: ['hq_admin'] },
  { key: '/users', icon: <UserOutlined />, label: '账户管理', roles: ['hq_admin'] },
  { key: '/activities', icon: <ScheduleOutlined />, label: '活动管理', roles: ['hq_admin', 'campus_operator', 'activity_contact'] },
  { key: '/materials', icon: <ShopOutlined />, label: '物资管理', roles: ['hq_admin'] },
]

const roleLabels: Record<string, string> = {
  hq_admin: '总部管理员',
  campus_operator: '校区操作员',
  activity_contact: '活动联系人',
}

export default function AppLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, restore, logout } = useAuthStore()
  const { collapsed, toggleCollapsed } = useAppStore()
  const [ready, setReady] = useState(false)
  const { token: themeToken } = theme.useToken()

  useEffect(() => {
    restore()
    setReady(true)
  }, [restore])

  // 未登录跳转
  useEffect(() => {
    if (ready && !user) {
      navigate('/login', { replace: true })
    }
  }, [ready, user, navigate])

  if (!ready || !user) {
    return (
      <div className="h-screen flex items-center justify-center">
        <span className="text-gray-400">加载中...</span>
      </div>
    )
  }

  // 按角色过滤菜单
  const menuItems = useMemo(
    () => allMenuItems
      .filter(item => item.roles.includes(user.role))
      .map(({ key, icon, label }) => ({ key, icon, label })),
    [user.role],
  )

  // 未授权页面自动跳转
  const allowedKeys = useMemo(() => new Set(menuItems.map(m => m.key)), [menuItems])
  const currentKey = menuItems.find((item) =>
    location.pathname.startsWith(item.key),
  )?.key
  useEffect(() => {
    if (currentKey && !allowedKeys.has(currentKey)) {
      navigate('/dashboard', { replace: true })
    }
  }, [currentKey, allowedKeys, navigate])

  const selectedKey = currentKey || '/dashboard'

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  return (
    <Layout className="h-screen">
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        theme="light"
        className="border-r border-gray-200"
      >
        <div className="h-16 flex items-center justify-center border-b border-gray-200">
          <h1 className={`font-bold text-gray-800 whitespace-nowrap ${collapsed ? 'text-sm' : 'text-base'}`}>
            {collapsed ? '物资' : '教培物资管理系统'}
          </h1>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          className="border-r-0 mt-2"
        />
      </Sider>
      <Layout>
        <Header
          className="flex items-center justify-between px-4 border-b border-gray-200"
          style={{ background: themeToken.colorBgContainer }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={toggleCollapsed}
          />
          <div className="flex items-center gap-4">
            <span className="text-gray-600 text-sm">
              {user.username}
              <span className="ml-2 text-gray-400">
                {roleLabels[user.role] || user.role}
              </span>
            </span>
            <Button
              type="text"
              icon={<LogoutOutlined />}
              onClick={handleLogout}
              danger
            >
              退出
            </Button>
          </div>
        </Header>
        <Content className="m-4 p-6 bg-white rounded-lg overflow-auto">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
