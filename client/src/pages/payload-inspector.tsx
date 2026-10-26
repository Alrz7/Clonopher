import { useMemo } from 'react'
import { useNavigate, useParams } from 'react-router'
import { ArrowLeft, ArrowUpRight, Copy, FileText, GitBranch, History, TerminalSquare } from 'lucide-react'
import { usePayloadStats } from '../lib/queries'
import { logClass, logLabel, timeAgo } from '../lib/format'
import { useAppStore } from '../stores/useAppStore'
import { StatusBadge } from '../components/status-badge'
import type { PayloadStatus } from '../lib/types'

export function PayloadInspectorPage() {
  const { payloadId: routePayloadId } = useParams()
  const navigate = useNavigate()
  const query = usePayloadStats()
  const { setLogsOpen, setLogFilter } = useAppStore()
  const payloadId = decodeURIComponent(routePayloadId ?? '')
  const payload = query.data?.payloads.find((item) => item.id === payloadId)
    ?? query.data?.payloads.find((item) => item.name === payloadId)
  const related = useMemo(() => query.data?.payloads.filter((item) => item.id !== payload?.id).slice(0, 4) ?? [], [query.data?.payloads, payload?.id])

  if (!payload) {
    return <section className="payload-inspector-page">
      <button className="back-link" onClick={() => navigate('/')}><ArrowLeft size={15} /> Payload activity</button>
      <div className="inspector-empty"><FileText size={25} /><strong>Payload not found</strong><span>This payload is not present in the current snapshot.</span></div>
    </section>
  }

  return (
    <section className="payload-inspector-page">
      <div className="inspector-topline">
        <button className="back-link" onClick={() => navigate('/')}><ArrowLeft size={15} /> Payload activity</button>
        <button className="inspector-logs-button" onClick={() => { setLogFilter('all'); setLogsOpen(true) }}><TerminalSquare size={15} /> Live logs</button>
      </div>
      <div className="inspector-titlebar">
        <span className="inspector-file-icon"><FileText size={22} strokeWidth={1.6} /></span>
        <div className="inspector-title-copy"><h1>{payload.name}</h1><span>{payload.id}</span></div>
        <StatusBadge state={payload.state} />
      </div>

      <div className="payload-inspector-grid">
        <div className="inspector-section-column">
          <section className="inspector-section">
            <h2>Processing context</h2>
            <div className="inspector-destination"><GitBranch size={17} /><div><span>Destination stage</span><strong>{payload.stagePath || 'Unassigned'}</strong></div></div>
            <dl className="inspector-fields">
              <div><dt>Last action</dt><dd>{payload.lastAction || '—'}</dd></div>
              <div><dt>Iteration</dt><dd>{payload.iteration}</dd></div>
              <div><dt>Started</dt><dd>{timeAgo(payload.startedAt)}</dd></div>
              <div><dt>Last updated</dt><dd>{timeAgo(payload.updatedAt)}</dd></div>
            </dl>
            <button className="copy-id-button" onClick={() => void navigator.clipboard?.writeText(payload.id)}><Copy size={13} /> Copy payload ID</button>
          </section>
          <section className="inspector-section">
            <div className="inspector-section-heading"><h2>Related payloads</h2><span>{related.length}</span></div>
            {related.length ? <div className="related-payload-list">{related.map((item) => (
              <button key={item.id} className="related-payload-row" onClick={() => navigate(`/payloads/${encodeURIComponent(item.id)}`)}>
                <FileText size={15} /><span><strong>{item.name}</strong><small>{item.stagePath || 'Unassigned'}</small></span><StatusBadge state={item.state} /><ArrowUpRight size={14} />
              </button>
            ))}</div> : <p className="inspector-muted">No other payloads in this snapshot.</p>}
          </section>
        </div>

        <div className="inspector-section-column">
          <section className="inspector-section">
            <div className="inspector-section-heading"><h2>Action history</h2><span>{payload.actions.length}</span></div>
            {payload.actions.length ? <ol className="action-history">{payload.actions.map((action, index) => (
              <li key={`${action}-${index}`}><span className="action-history-index">{String(index + 1).padStart(2, '0')}</span><span>{action}</span>{index === payload.actions.length - 1 && <small>Latest</small>}</li>
            ))}</ol> : <p className="inspector-muted"><History size={15} /> No actions recorded.</p>}
          </section>
          <section className="inspector-section">
            <div className="inspector-section-heading"><h2>Payload logs</h2><button className="inline-link" onClick={() => { setLogFilter('all'); setLogsOpen(true) }}>Open all logs <ArrowUpRight size={13} /></button></div>
            <PayloadLogList payload={payload} />
          </section>
        </div>
      </div>
    </section>
  )
}

function PayloadLogList({ payload }: { payload: PayloadStatus }) {
  if (!payload.logs.length) return <p className="inspector-muted">No diagnostic events recorded for this payload.</p>
  return <div className="payload-log-list">{[...payload.logs].reverse().slice(0, 12).map((log) => (
    <article key={log.id} className="payload-log-row"><span className={`console-log-type ${logClass(log.type)}`}>{logLabel(log.type).toLowerCase()}</span><div><p>{log.content}</p><time>{timeAgo(log.timestamp)}</time></div></article>
  ))}</div>
}
