import { create } from 'zustand'

interface AppState {
  /** 侧边栏折叠 */
  collapsed: boolean
  toggleCollapsed: () => void
}

export const useAppStore = create<AppState>((set) => ({
  collapsed: false,
  toggleCollapsed: () => set((s) => ({ collapsed: !s.collapsed })),
}))
