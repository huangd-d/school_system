import { RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider, QueryCache, MutationCache } from '@tanstack/react-query'
import { ConfigProvider, App as AntApp, message } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { router } from '@/router'

// 判断是否为认证类错误 — 拦截器已处理，全局处理器应跳过避免重复提示
function isAuthError(err: Error): boolean {
  return err.message?.includes('未登录') || err.message?.includes('登录已过期')
}

// 全局 QueryClient — 统一错误兜底
const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error: Error) => {
      // 认证错误由拦截器处理（跳转登录），此处静默
      if (isAuthError(error)) return
      // 查询失败 → toast 提示，解决「查询静默失败、表格空数据」的问题
      message.error(error.message || '数据加载失败')
    },
  }),
  mutationCache: new MutationCache({
    onError: (error: Error) => {
      // 认证错误由拦截器处理
      if (isAuthError(error)) return
      // 组件级 onError 已弹 toast，此处只打日志避免重复
      // 若某 mutation 未注册 onError，这里至少能在控制台看到
      console.error('[Mutation Error]', error.message || '操作失败')
    },
  }),
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

export default function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <AntApp>
        <QueryClientProvider client={queryClient}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </AntApp>
    </ConfigProvider>
  )
}
