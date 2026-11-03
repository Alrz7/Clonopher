import { memo, useCallback, useEffect, useMemo, useState } from "react";
import type { MouseEvent as ReactMouseEvent } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  Handle,
  MiniMap,
  Position,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  type Edge,
  type Node,
  type NodeProps,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import {
  Check,
  ChevronDown,
  FileText,
  Folder,
  FolderTree,
  GitBranch,
  X,
} from "lucide-react";
import {
  useGlobalDirectory,
  useDirInfo,
  useDirectoryList,
} from "../lib/queries";
import type { SnapShot } from "../lib/types";
import { useAppStore } from "../stores/useAppStore";

interface DirectoryNodeData extends Record<string, unknown> {
  label: string;
  path: string;
  fileCount: number;
  childCount: number;
  brief?: string;
  isRoot?: boolean;
  version?: number;
}

type DirectoryFlowNode = Node<DirectoryNodeData, "directoryCard">;
interface DirectoryNodePosition {
  id: string;
  parentId?: string;
  depth: number;
  data: DirectoryNodeData;
}

function flattenDirectoryTree(root: SnapShot): {
  nodes: DirectoryFlowNode[];
  edges: Edge[];
} {
  const positioned: DirectoryNodePosition[] = [];
  const queue: Array<{
    name: string;
    directory: SnapShot;
    depth: number;
    parentId?: string;
  }> = [
    {
      name: root.path.split(/[\\/]/).filter(Boolean).at(-1) || "Root",
      directory: root,
      depth: 0,
    },
  ];

  while (queue.length && positioned.length < 250) {
    const item = queue.shift()!;
    const id = `${item.directory.path || item.name}::${positioned.length}`;
    positioned.push({
      id,
      parentId: item.parentId,
      depth: item.depth,
      data: {
        label: item.name,
        path: item.directory.path,
        fileCount: item.directory.existingFiles?.length ?? 0,
        childCount: Object.keys(item.directory.children ?? {}).length,
        brief: item.directory.info?.brief,
        version: item.directory.info?.version,
        isRoot: item.depth === 0,
      },
    });
    for (const [name, child] of Object.entries(item.directory.children ?? {})) {
      queue.push({
        name,
        directory: child,
        depth: item.depth + 1,
        parentId: id,
      });
    }
  }

  const byDepth = new Map<number, DirectoryNodePosition[]>();
  for (const item of positioned) {
    const level = byDepth.get(item.depth) ?? [];
    level.push(item);
    byDepth.set(item.depth, level);
  }

  const nodes: DirectoryFlowNode[] = positioned.map((item) => {
    const level = byDepth.get(item.depth) ?? [item];
    const index = level.findIndex((entry) => entry.id === item.id);
    return {
      id: item.id,
      type: "directoryCard",
      position: {
        x: (index - (level.length - 1) / 2) * 290,
        y: item.depth * 165,
      },
      data: item.data,
      selected: false,
    };
  });

  const edges: Edge[] = positioned
    .filter((item) => item.parentId)
    .map((item) => ({
      id: `edge-${item.parentId}-${item.id}`,
      source: item.parentId!,
      target: item.id,
      type: "smoothstep",
      style: { stroke: "#393b46", strokeWidth: 1.4 },
    }));

  return { nodes, edges };
}

const DirectoryCardNode = memo(function DirectoryCardNode({
  data,
  selected,
}: NodeProps<DirectoryFlowNode>) {
  return (
    <div
      className={`directory-node ${selected ? "directory-node-selected" : ""} ${data.isRoot ? "directory-node-root" : ""}`}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="directory-handle"
      />
      <div className="directory-node-title-row">
        <span className="directory-node-icon">
          <Folder size={16} strokeWidth={1.7} />
        </span>
        <span className="directory-node-name">{data.label}</span>
        {data.isRoot && <span className="directory-root-indicator">ROOT</span>}
      </div>
      <div className="directory-node-path" title={data.path}>
        {data.path}
      </div>
      <div className="directory-node-meta">
        <span>
          <FileText size={12} />
          {data.fileCount}
        </span>
        <span>
          <FolderTree size={12} />
          {data.childCount}
        </span>
        {typeof data.version === "number" && <span>v{data.version}</span>}
      </div>
      <Handle
        type="source"
        position={Position.Bottom}
        className="directory-handle"
      />
    </div>
  );
});

const nodeTypes = { directoryCard: DirectoryCardNode };

export function DirectoryCanvasPage() {
  return (
    <ReactFlowProvider>
      <DirectoryCanvasInner />
    </ReactFlowProvider>
  );
}

function DirectoryCanvasInner() {
  const directoryListQuery = useDirectoryList();
  const names = directoryListQuery.data?.list ?? [];
  const [selectedRoot, setSelectedRoot] = useState("");
  const selectedPath = selectedRoot || names[0] || "";
  const directoryQuery = useGlobalDirectory(selectedPath);
  const tree = directoryQuery.data?.selected;
  const flow = useMemo(
    () => (tree ? flattenDirectoryTree(tree) : { nodes: [], edges: [] }),
    [tree],
  );
  const [nodes, setNodes, onNodesChange] = useNodesState<DirectoryFlowNode>(
    flow.nodes,
  );
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(flow.edges);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const { setSelectedDirectoryPath } = useAppStore();

  useEffect(() => {
    setNodes(flow.nodes);
    setEdges(flow.edges);
    setSelectedNodeId(null);
  }, [flow, setNodes, setEdges]);

  const selectedNode = nodes.find((node) => node.id === selectedNodeId);
  const DirInfoQuery = useDirInfo(selectedNode?.data.path ?? "");
  const DirInfo = DirInfoQuery.data?.info;
  const allFiles = useMemo(() => flattenFiles(tree), [tree]);

  const selectNode = useCallback(
    (_event: ReactMouseEvent, node: DirectoryFlowNode) => {
      setSelectedNodeId(node.id);
      setSelectedDirectoryPath(node.data.path);
    },
    [setSelectedDirectoryPath],
  );

  const selectChild = (child: DirectoryFlowNode) => {
    setSelectedNodeId(child.id);
    setSelectedDirectoryPath(child.data.path);
  };

  const children = selectedNode
    ? edges
        .filter((edge) => edge.source === selectedNode.id)
        .map((edge) => nodes.find((node) => node.id === edge.target))
        .filter((node): node is DirectoryFlowNode => Boolean(node))
    : [];
  const files = selectedNode
    ? allFiles.filter((file) => file.dirPath === selectedNode.data.path)
    : [];

  return (
    <section className="directory-page" aria-label="Directory canvas">
      <div className="directory-toolbar">
        <label className="directory-root-select">
          <GitBranch size={16} />
          <span>Global directory</span>
          <span className="directory-select-control">
            <select
              value={selectedPath}
              onChange={(event) => {
                setSelectedRoot(event.target.value);
                setSelectedNodeId(null);
              }}
              disabled={!names.length}
            >
              {!names.length && (
                <option value="" disabled>
                  Loading directories…
                </option>
              )}
              {names.map((name) => (
                <option key={name} value={name}>
                  {name}
                </option>
              ))}
            </select>
            <ChevronDown size={14} />
          </span>
        </label>
      </div>

      {directoryQuery.data?.warning && (
        <div className="directory-data-warning">
          {directoryQuery.data.warning}
        </div>
      )}

      <div className="directory-map-area">
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
          defaultEdgeOptions={{ type: "smoothstep" }}
          proOptions={{ hideAttribution: true }}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={25}
            size={1}
            color="#2a2b33"
          />
          <Controls position="bottom-left" showInteractive={false} />
          <MiniMap
            position="bottom-right"
            pannable
            zoomable
            nodeColor={(node) => (node.selected ? "#8580bd" : "#454650")}
            maskColor="rgba(10, 10, 13, 0.76)"
          />
        </ReactFlow>
        {nodes.length === 0 && (
          <div className="directory-map-empty">
            <FolderTree size={25} />
            <span>
              {directoryQuery.isLoading || directoryListQuery.isLoading
                ? "Loading directories…"
                : "No directories to display"}
            </span>
          </div>
        )}
      </div>

      {selectedNode && (
        <>
          <button
            className="directory-inspector-scrim"
            onClick={() => setSelectedNodeId(null)}
            aria-label="Close directory details"
          />
          <aside
            className="directory-inspector-overlay"
            aria-label={`${selectedNode.data.label} directory details`}
          >
            <div className="directory-inspector-head">
              <span className="directory-inspector-icon">
                <FolderTree size={17} />
              </span>
              <div className="directory-inspector-title">
                <h2>{selectedNode.data.label}</h2>
                <span>{selectedNode.data.path}</span>
              </div>
              <button
                className="icon-close"
                onClick={() => setSelectedNodeId(null)}
                aria-label="Close directory details"
              >
                <X size={17} />
              </button>
            </div>
            <div className="directory-inspector-content">
              <div className="directory-summary-line">
                <span>
                  <FileText size={14} />
                  {selectedNode.data.fileCount} files
                </span>
                <span>
                  <FolderTree size={14} />
                  {selectedNode.data.childCount} sub-directories
                </span>
              </div>
              <section className="directory-detail-section">
                <h3>Description</h3>
                <p>
                  {DirInfo?.brief ||
                    selectedNode.data.brief ||
                    "No description available."}
                </p>
              </section>
              <section className="directory-detail-section">
                <h3>Details</h3>
                <dl className="directory-detail-list">
                  <div>
                    <dt>Directory ID</dt>
                    <dd>{DirInfo?.dirId || "—"}</dd>
                  </div>
                  <div>
                    <dt>Version</dt>
                    <dd>
                      {DirInfo?.version ?? selectedNode.data.version ?? "—"}
                    </dd>
                  </div>
                  <div>
                    <dt>Created</dt>
                    <dd>
                      {DirInfo?.created_at
                        ? DirInfo.created_at.slice(0, 16).replace("T", " ")
                        : "—"}
                    </dd>
                  </div>
                  <div>
                    <dt>Updated</dt>
                    <dd>
                      {DirInfo?.updated_at
                        ? DirInfo.updated_at.slice(0, 16).replace("T", " ")
                        : "—"}
                    </dd>
                  </div>
                  <div>
                    <dt>Creator</dt>
                    <dd>
                      {DirInfo?.creator
                        ? `${DirInfo.creator.provider} / ${DirInfo.creator.model}`
                        : "—"}
                    </dd>
                  </div>
                </dl>
              </section>
              <section className="directory-detail-section">
                <h3>
                  Child directories <span>{children.length}</span>
                </h3>
                {children.length ? (
                  <div className="directory-child-list">
                    {children.map((child) => (
                      <button
                        className="directory-child-row"
                        key={child.id}
                        onClick={() => selectChild(child)}
                      >
                        <FolderTree size={15} />
                        <span>
                          <strong>{child.data.label}</strong>
                          <small>{child.data.fileCount} files</small>
                        </span>
                        <Check
                          size={13}
                          className={
                            child.id === selectedNode.id ? "child-current" : ""
                          }
                        />
                      </button>
                    ))}
                  </div>
                ) : (
                  <p className="directory-detail-muted">
                    No nested directories.
                  </p>
                )}
              </section>
              <section className="directory-detail-section">
                <h3>
                  Files <span>{files.length}</span>
                </h3>
                {files.length ? (
                  <div className="directory-files-list">
                    {files.map((file) => (
                      <div key={`${file.dirPath}/${file.name}`}>
                        <FileText size={14} />
                        <span>{file.name}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="directory-detail-muted">
                    No files reported for this directory.
                  </p>
                )}
              </section>
              {DirInfoQuery.data?.warning && (
                <p className="directory-detail-muted">
                  {DirInfoQuery.data.warning}
                </p>
              )}
            </div>
          </aside>
        </>
      )}
    </section>
  );
}

function flattenFiles(
  tree?: SnapShot,
): Array<{ name: string; dirPath: string }> {
  if (!tree) return [];
  const result = (tree.existingFiles ?? []).map((name) => ({
    name,
    dirPath: tree.path,
  }));
  for (const child of Object.values(tree.children ?? {}))
    result.push(...flattenFiles(child));
  return result;
}
