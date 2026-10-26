import { create } from 'zustand'

export type LogFilter = 'all' | 'error' | 'warn' | 'info'

interface AppStore {
  logsOpen: boolean
  logFilter: LogFilter
  logSearch: string
  selectedStagePath: string | null
  setLogsOpen: (open: boolean) => void
  toggleLogs: () => void
  setLogFilter: (filter: LogFilter) => void
  setLogSearch: (search: string) => void
  setSelectedStagePath: (path: string | null) => void
}

export const useAppStore = create<AppStore>((set) => ({
  logsOpen: false,
  logFilter: 'all',
  logSearch: '',
  selectedStagePath: null,
  setLogsOpen: (logsOpen) => set({ logsOpen }),
  toggleLogs: () => set((state) => ({ logsOpen: !state.logsOpen })),
  setLogFilter: (logFilter) => set({ logFilter }),
  setLogSearch: (logSearch) => set({ logSearch }),
  setSelectedStagePath: (selectedStagePath) => set({ selectedStagePath }),
}))
