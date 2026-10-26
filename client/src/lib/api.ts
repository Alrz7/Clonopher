import { demoPayloads, demoStageNames, demoStageTree } from './demo-data'
import type {
  ApiEnvelope,
  NormalizedPayloadLog,
  PayloadLog,
  PayloadStatsResult,
  PayloadStatus,
  PayloadState,
  SelectedStageResult,
  StageInfo,
  StageInfoResult,
  StageListResult,
  StageSnapshot,
} from './types'

const API = '/api/v1'

function record(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' ? value as Record<string, unknown> : {}
}

function valueOf(source: Record<string, unknown>, ...keys: string[]): unknown {
  for (const key of keys) {
    if (source[key] !== undefined) return source[key]
  }
  return undefined
}

function stringValue(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value : fallback
}

function numberValue(value: unknown, fallback = 0): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback
}

function arrayValue(value: unknown): unknown[] {
  return Array.isArray(value) ? value : []
}

function stateValue(value: unknown): PayloadState {
  const state = stringValue(value, 'unknown').toLowerCase() as PayloadState
  const supported: PayloadState[] = [
    'queued', 'exploring', 'classifying', 'generating', 'modifying',
    'waiting', 'retrying', 'inferring', 'selecting_classifiers',
    'uncertain', 'completed', 'failed', 'cancelled', 'unknown',
  ]
  return supported.includes(state) ? state : 'unknown'
}

function normalizeLog(input: unknown, index: number): NormalizedPayloadLog {
  const log = record(input) as PayloadLog & Record<string, unknown>
  return {
    id: stringValue(valueOf(log, 'id', 'ID'), `log-${index}`),
    type: stringValue(valueOf(log, 'type', 'Type'), 'info').toLowerCase(),
    content: stringValue(valueOf(log, 'content', 'Content', 'message', 'Message'), String(input ?? '')),
    timestamp: stringValue(valueOf(log, 'timestamp', 'createdAt', 'Timestamp', 'CreatedAt')) || undefined,
  }
}

function normalizePayload(input: unknown, fallbackId: string, index: number): PayloadStatus {
  const p = record(input)
  const logs = arrayValue(valueOf(p, 'logs', 'Logs')).map(normalizeLog)
  const actions = arrayValue(valueOf(p, 'actions', 'Actions')).map((item) => String(item))
  const stage = record(valueOf(p, 'stage', 'Stage'))

  return {
    id: stringValue(valueOf(p, 'id', 'ID'), fallbackId || `payload-${index}`),
    name: stringValue(valueOf(p, 'name', 'Name', 'filename', 'Filename'), fallbackId || `Payload ${index + 1}`),
    state: stateValue(valueOf(p, 'state', 'State')),
    iteration: numberValue(valueOf(p, 'iteration', 'Iteration')),
    lastAction: stringValue(valueOf(p, 'lastAction', 'LastAction')),
    actions,
    logs,
    paths: arrayValue(valueOf(p, 'road')).map((item) => String(item)),
    stagePath: stringValue(valueOf(p, 'path'), stringValue(valueOf(stage, 'path', 'Path'), '—')),
    startedAt: stringValue(valueOf(p, 'startedAt', 'StartedAt')) || undefined,
    updatedAt: stringValue(valueOf(p, 'updatedAt', 'UpdatedAt')) || undefined,
    source: 'api',
  }
}

async function requestJson<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  })

  if (!response.ok) {
    const body = await response.text().catch(() => '')
    throw new Error(`API ${response.status} ${response.statusText}${body ? ` — ${body.slice(0, 160)}` : ''}`)
  }

  return response.json() as Promise<T>
}

function normalizePayloadCollection(raw: unknown): PayloadStatus[] {
  const payloadObject = record(raw)
  const candidates = valueOf(payloadObject, 'payloads', 'Payloads', 'stats', 'Stats') ?? raw

  if (Array.isArray(candidates)) {
    return candidates.map((item, index) => normalizePayload(item, '', index))
  }

  return Object.entries(record(candidates)).map(([id, item], index) => normalizePayload(item, id, index))
}

export async function getPayloadStats(): Promise<PayloadStatsResult> {
  try {
    const raw = await requestJson<ApiEnvelope>(`${API}/payloads/stats`)
    const payloads = normalizePayloadCollection(raw)
    return { payloads, source: 'api' }
  } catch (error) {
    return {
      payloads: demoPayloads.map((payload) => ({ ...payload, logs: [...payload.logs], actions: [...payload.actions], source: 'demo' })),
      source: 'demo',
      warning: error instanceof Error ? error.message : 'The API is unavailable.',
    }
  }
}

export async function getStageList(): Promise<StageListResult> {
  try {
    const raw = record(await requestJson<ApiEnvelope>(`${API}/stages/list`))
    const listRaw = valueOf(raw, 'list', 'List', 'stages', 'Stages')
    const list = arrayValue(listRaw).map((item) => {
      if (typeof item === 'string') return item
      const stage = record(item)
      return stringValue(valueOf(stage, 'path', 'Path', 'name', 'Name'))
    }).filter(Boolean)
    return { list, source: 'api' }
  } catch (error) {
    return {
      list: [...demoStageNames],
      source: 'demo',
      warning: error instanceof Error ? error.message : 'The Stage API is unavailable.',
    }
  }
}

function normalizeStage(input: unknown, fallbackPath: string): StageSnapshot {
  const stage = record(input)
  const roots = record(valueOf(stage, 'substages', 'Substages', 'roots', 'Roots'))
  const files = arrayValue(valueOf(stage, 'existingFiles', 'ExistingFiles', 'files', 'Files')).map(String)
  const normalizedRoots: Record<string, StageSnapshot> = {}

  for (const [name, child] of Object.entries(roots)) {
    normalizedRoots[name] = normalizeStage(child, `${fallbackPath}/${name}`)
  }

  const infoRaw = valueOf(stage, 'info', 'Info')
  return {
    path: stringValue(valueOf(stage, 'path', 'Path'), fallbackPath),
    substages: normalizedRoots,
    existingFiles: files,
    info: infoRaw && typeof infoRaw === 'object' ? infoRaw as StageSnapshot['info'] : undefined,
  }
}

export async function getGlobalStage(path: string): Promise<SelectedStageResult> {
  try {
    const raw = record(await requestJson<ApiEnvelope>(`${API}/stages/query`, {
      method: 'POST',
      body: JSON.stringify({ select: path }),
    }))
    const selected = valueOf(raw, 'selected', 'Selected', 'stage', 'Stage')
    const listRaw = valueOf(raw, 'list', 'List', 'stages', 'Stages')
    const list = arrayValue(listRaw).map((item) => typeof item === 'string'
      ? item
      : stringValue(valueOf(record(item), 'path', 'Path', 'name', 'Name'))).filter(Boolean)

    if (!selected || typeof selected !== 'object') throw new Error('API response did not contain a selected Stage snapshot.')

    return {
      selected: normalizeStage(selected, path),
      list: list.length ? list : [path],
      source: 'api',
    }
  } catch (error) {
    return {
      selected: findDemoStage(path) ?? demoStageTree,
      list: [...demoStageNames],
      source: 'demo',
      warning: error instanceof Error ? error.message : 'The Stage API is unavailable.',
    }
  }
}

function findDemoStage(pathOrName: string): StageSnapshot | undefined {
  if (!pathOrName || pathOrName.toLowerCase() === 'root') return demoStageTree

  const wanted = pathOrName.replaceAll('\\', '/').replace(/\/$/, '').toLowerCase()

  function visit(stage: StageSnapshot): StageSnapshot | undefined {
    for (const [name, child] of Object.entries(stage.substages ?? {})) {
      const childPath = child.path.replaceAll('\\', '/').replace(/\/$/, '').toLowerCase()
      if (name.toLowerCase() === wanted || childPath === wanted || childPath.endsWith(`/${wanted}`)) return child
      const nested = visit(child)
      if (nested) return nested
    }
    return undefined
  }

  return visit(demoStageTree)
}


function normalizeStageInfo(input: unknown): StageInfo | null {
  if (!input || typeof input !== 'object') return null
  const info = record(input)
  const creatorRaw = valueOf(info, 'creator', 'Creator')
  const modifierRaw = valueOf(info, 'modifier', 'Modifier')
  const normalizeProvider = (raw: unknown) => {
    if (!raw || typeof raw !== 'object') return null
    const provider = record(raw)
    return {
      provider: stringValue(valueOf(provider, 'provider', 'Provider')),
      model: stringValue(valueOf(provider, 'model', 'Model')),
    }
  }

  return {
    stage_id: stringValue(valueOf(info, 'stage_id', 'StageID', 'stageId', 'StageId')),
    path: stringValue(valueOf(info, 'path', 'Path')),
    created_at: stringValue(valueOf(info, 'created_at', 'CreatedAt', 'createdAt')),
    updated_at: stringValue(valueOf(info, 'updated_at', 'UpdatedAt', 'updatedAt')),
    version: numberValue(valueOf(info, 'version', 'Version')),
    brief: stringValue(valueOf(info, 'brief', 'Brief')),
    creator: normalizeProvider(creatorRaw),
    modifier: normalizeProvider(modifierRaw),
  }
}

function findDemoInfo(pathOrName: string): StageInfo | null {
  const stage = findDemoStage(pathOrName)
  return stage?.info ?? null
}

export async function getStageInfo(path: string): Promise<StageInfoResult> {
  try {
    const raw = record(await requestJson<ApiEnvelope>(`${API}/stages/info`, {
      method: 'POST',
      body: JSON.stringify({ path }),
    }))
    const info = valueOf(raw, 'info', 'Info')
    return { info: normalizeStageInfo(info), source: 'api' }
  } catch (error) {
    return {
      info: findDemoInfo(path),
      source: 'demo',
      warning: error instanceof Error ? error.message : 'The Stage info API is unavailable.',
    }
  }
}
