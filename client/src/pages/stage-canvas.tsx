import { memo, useCallback, useEffect, useMemo, useState } from 'react'
import type { MouseEvent as ReactMouseEvent } from 'react'
import {
  Background, BackgroundVariant, Controls, Handle, MiniMap, Position,
  ReactFlow, ReactFlowProvider, useEdgesState, useNodesState,
  type Edge, type Node, type NodeProps,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { Check, ChevronDown, FileText, Folder, FolderTree, GitBranch, X } from 'lucide-react'
import { useGlobalStage, useStageInfo, useStageList } from '../lib/queries'
import type { StageSnapshot } from '../lib/types'
import { useAppStore } from '../stores/useAppStore'

interface StageNodeData extends Record<string, unknown> {
  label: string
  path: string
  fileCount: number
  childCount: number
  brief?: string
  isRoot?: boolean
  version?: number
}

type StageFlowNode = Node<StageNodeData, 'stageCard'>
interface StageNodePosition { id: string; parentId?: string; depth: number; data: StageNodeData }

function flattenStageTree(root: StageSnapshot): { nodes: StageFlowNode[]; edges: Edge[] } {
  const positioned: StageNodePosition[] = []
  const queue: Array<{ name: string; stage: StageSnapshot; depth: number; parentId?: string }> = [
    { name: root.path.split(/[\\/]/).filter(Boolean).at(-1) || 'Root', stage: root, depth: 0 },
  ]

  while (queue.length && positioned.length < 250) {
    const item = queue.shift()!
    const id = `${item.stage.path || item.name}::${positioned.length}`
    positioned.push({
      id,
      parentId: item.parentId,
      depth: item.depth,
      data: {
        label: item.name,
        path: item.stage.path,
        fileCount: item.stage.existingFiles?.length ?? 0,
        childCount: Object.keys(item.stage.substages ?? {}).length,
        brief: item.stage.info?.brief,
        version: item.stage.info?.version,
        isRoot: item.depth === 0,
      },
    })
    for (const [name, child] of Object.entries(item.stage.substages ?? {})) {
      queue.push({ name, stage: child, depth: item.depth + 1, parentId: id })
    }
  }

  const byDepth = new Map<number, StageNodePosition[]>()
  for (const item of positioned) {
    const level = byDepth.get(item.depth) ?? []
    level.push(item)
    byDepth.set(item.depth, level)
  }

  const nodes: StageFlowNode[] = positioned.map((item) => {
    const level = byDepth.get(item.depth) ?? [item]
    const index = level.findIndex((entry) => entry.id === item.id)
    return {
      id: item.id,
      type: 'stageCard',
      position: { x: (index - (level.length - 1) / 2) * 290, y: item.depth * 165 },
      data: item.data,
      selected: false,
    }
  })

  const edges: Edge[] = positioned.filter((item) => item.parentId).map((item) => ({
    id: `edge-${item.parentId}-${item.id}`,
    source: item.parentId!,
    target: item.id,
    type: 'smoothstep',
    style: { stroke: '#393b46', strokeWidth: 1.4 },
  }))

  return { nodes, edges }
}

const StageCardNode = memo(function StageCardNode({ data, selected }: NodeProps<StageFlowNode>) {
  return (
    <div className={`stage-node ${selected ? 'stage-node-selected' : ''} ${data.isRoot ? 'stage-node-root' : ''}`}>
      <Handle type="target" position={Position.Top} className="stage-handle" />
      <div className="stage-node-title-row">
        <span className="stage-node-icon"><Folder size={16} strokeWidth={1.7} /></span>
        <span className="stage-node-name">{data.label}</span>
        {data.isRoot && <span className="stage-root-indicator">ROOT</span>}
      </div>
      <div className="stage-node-path" title={data.path}>{data.path}</div>
      <div className="stage-node-meta"><span><FileText size={12} />{data.fileCount}</span><span><FolderTree size={12} />{data.childCount}</span>{typeof data.version === 'number' && <span>v{data.version}</span>}</div>
      <Handle type="source" position={Position.Bottom} className="stage-handle" />
    </div>
  )
})

const nodeTypes = { stageCard: StageCardNode }

export function StageCanvasPage() {
  return <ReactFlowProvider><StageCanvasInner /></ReactFlowProvider>
}

function StageCanvasInner() {
  const stageListQuery = useStageList()
  const names = stageListQuery.data?.list ?? []
  const [selectedRoot, setSelectedRoot] = useState('')
  const selectedPath = selectedRoot || names[0] || ''
  const stageQuery = useGlobalStage(selectedPath)
  const tree = stageQuery.data?.selected
  const flow = useMemo(() => tree ? flattenStageTree(tree) : { nodes: [], edges: [] }, [tree])
  const [nodes, setNodes, onNodesChange] = useNodesState<StageFlowNode>(flow.nodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(flow.edges)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const { setSelectedStagePath } = useAppStore()

  useEffect(() => {
    setNodes(flow.nodes)
    setEdges(flow.edges)
    setSelectedNodeId(null)
  }, [flow, setNodes, setEdges])

  const selectedNode = nodes.find((node) => node.id === selectedNodeId)
  const stageInfoQuery = useStageInfo(selectedNode?.data.path ?? '')
  const stageInfo = stageInfoQuery.data?.info
  const allFiles = useMemo(() => flattenFiles(tree), [tree])

  const selectNode = useCallback((_event: ReactMouseEvent, node: StageFlowNode) => {
    setSelectedNodeId(node.id)
    setSelectedStagePath(node.data.path)
  }, [setSelectedStagePath])

  const selectChild = (child: StageFlowNode) => {
    setSelectedNodeId(child.id)
    setSelectedStagePath(child.data.path)
  }

  const children = selectedNode ? edges
    .filter((edge) => edge.source === selectedNode.id)
    .map((edge) => nodes.find((node) => node.id === edge.target))
    .filter((node): node is StageFlowNode => Boolean(node)) : []
  const files = selectedNode ? allFiles.filter((file) => file.stagePath === selectedNode.data.path) : []

  return (
    <section className="stage-page" aria-label="Stage canvas">
      <div className="stage-toolbar">
        <label className="stage-root-select">
          <GitBranch size={16} />
          <span>Global stage</span>
          <span className="stage-select-control">
            <select value={selectedPath} onChange={(event) => { setSelectedRoot(event.target.value); setSelectedNodeId(null) }} disabled={!names.length}>
              {!names.length && <option value="" disabled>Loading stages…</option>}
              {names.map((name) => <option key={name} value={name}>{name}</option>)}
            </select>
            <ChevronDown size={14} />
          </span>
        </label>
      </div>

      {stageQuery.data?.warning && <div className="stage-data-warning">{stageQuery.data.warning}</div>}

      <div className="stage-map-area">
        <ReactFlow
          key={selectedPath}
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={selectNode}
          onPaneClick={() => setSelectedNodeId(null)}
          fitView
          fitViewOptions={{ padding: 0.22, minZoom: 0.18, maxZoom: 0.9 }}
          minZoom={0.12}
          maxZoom={1.6}
          defaultEdgeOptions={{ type: 'smoothstep' }}
          proOptions={{ hideAttribution: true }}
        >
          <Background variant={BackgroundVariant.Dots} gap={25} size={1} color="#2a2b33" />
          <Controls position="bottom-left" showInteractive={false} />
          <MiniMap position="bottom-right" pannable zoomable nodeColor={(node) => node.selected ? '#8580bd' : '#454650'} maskColor="rgba(10, 10, 13, 0.76)" />
        </ReactFlow>
        {nodes.length === 0 && <div className="stage-map-empty"><FolderTree size={25} /><span>{stageQuery.isLoading || stageListQuery.isLoading ? 'Loading stages…' : 'No stages to display'}</span></div>}
      </div>

      {selectedNode && (
        <>
          <button className="stage-inspector-scrim" onClick={() => setSelectedNodeId(null)} aria-label="Close stage details" />
          <aside className="stage-inspector-overlay" aria-label={`${selectedNode.data.label} stage details`}>
            <div className="stage-inspector-head">
              <span className="stage-inspector-icon"><FolderTree size={17} /></span>
              <div className="stage-inspector-title"><h2>{selectedNode.data.label}</h2><span>{selectedNode.data.path}</span></div>
              <button className="icon-close" onClick={() => setSelectedNodeId(null)} aria-label="Close stage details"><X size={17} /></button>
            </div>
            <div className="stage-inspector-content">
              <div className="stage-summary-line"><span><FileText size={14} />{selectedNode.data.fileCount} files</span><span><FolderTree size={14} />{selectedNode.data.childCount} sub-stages</span></div>
              <section className="stage-detail-section">
                <h3>Description</h3>
                <p>{stageInfo?.brief || selectedNode.data.brief || 'No description available.'}</p>
              </section>
              <section className="stage-detail-section">
                <h3>Details</h3>
                <dl className="stage-detail-list">
                  <div><dt>Stage ID</dt><dd>{stageInfo?.stage_id || '—'}</dd></div>
                  <div><dt>Version</dt><dd>{stageInfo?.version ?? selectedNode.data.version ?? '—'}</dd></div>
                  <div><dt>Created</dt><dd>{stageInfo?.created_at ? stageInfo.created_at.slice(0, 16).replace('T', ' ') : '—'}</dd></div>
                  <div><dt>Updated</dt><dd>{stageInfo?.updated_at ? stageInfo.updated_at.slice(0, 16).replace('T', ' ') : '—'}</dd></div>
                  <div><dt>Creator</dt><dd>{stageInfo?.creator ? `${stageInfo.creator.provider} / ${stageInfo.creator.model}` : '—'}</dd></div>
                </dl>
              </section>
              <section className="stage-detail-section">
                <h3>Child stages <span>{children.length}</span></h3>
                {children.length ? <div className="stage-child-list">{children.map((child) => (
                  <button className="stage-child-row" key={child.id} onClick={() => selectChild(child)}>
                    <FolderTree size={15} /><span><strong>{child.data.label}</strong><small>{child.data.fileCount} files</small></span><Check size={13} className={child.id === selectedNode.id ? 'child-current' : ''} />
                  </button>
                ))}</div> : <p className="stage-detail-muted">No nested stages.</p>}
              </section>
              <section className="stage-detail-section">
                <h3>Files <span>{files.length}</span></h3>
                {files.length ? <div className="stage-files-list">{files.map((file) => <div key={`${file.stagePath}/${file.name}`}><FileText size={14} /><span>{file.name}</span></div>)}</div> : <p className="stage-detail-muted">No files reported for this stage.</p>}
              </section>
              {stageInfoQuery.data?.warning && <p className="stage-detail-muted">{stageInfoQuery.data.warning}</p>}
            </div>
          </aside>
        </>
      )}
    </section>
  )
}

function flattenFiles(tree?: StageSnapshot): Array<{ name: string; stagePath: string }> {
  if (!tree) return []
  const result = (tree.existingFiles ?? []).map((name) => ({ name, stagePath: tree.path }))
  for (const child of Object.values(tree.substages ?? {})) result.push(...flattenFiles(child))
  return result
}
