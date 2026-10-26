import type { PayloadState } from './types'

export function titleCase(value: string): string {
  return value.replaceAll('_', ' ').replace(/\b\w/g, (letter) => letter.toUpperCase())
}

export function timeAgo(value?: string): string {
  if (!value) return 'just now'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000))
  if (seconds < 60) return `${seconds}s ago`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  return `${Math.floor(hours / 24)}d ago`
}

export function stateClass(state: PayloadState): string {
  switch (state) {
    case 'completed': return 'status status-success'
    case 'failed': return 'status status-error'
    case 'cancelled': return 'status status-muted'
    case 'uncertain': return 'status status-warn'
    case 'queued': return 'status status-queued'
    default: return 'status status-running'
  }
}

export function logClass(type: string): string {
  switch (type.toLowerCase()) {
    case 'error': return 'log-type log-type-error'
    case 'warn': case 'warning': return 'log-type log-type-warn'
    case 'info': return 'log-type log-type-info'
    default: return 'log-type log-type-other'
  }
}

export function logLabel(type: string): string {
  const lower = type.toLowerCase()
  return lower === 'warning' ? 'WARN' : lower.toUpperCase() || 'LOG'
}
