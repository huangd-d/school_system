import { createBrowserRouter, Navigate } from 'react-router'
import AppLayout from '@/components/layout/AppLayout'
import LoginPage from '@/pages/auth/LoginPage'
import DashboardPage from '@/pages/dashboard/DashboardPage'
import CampusPage from '@/pages/campus/CampusPage'
import UserPage from '@/pages/user/UserPage'

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <DashboardPage /> },
      { path: 'campuses', element: <CampusPage /> },
      { path: 'users', element: <UserPage /> },
    ],
  },
])
