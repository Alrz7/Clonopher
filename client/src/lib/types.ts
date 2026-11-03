export type PayloadState =
  | "queued"
  | "exploring"
  | "classifying"
  | "generating"
  | "modifying"
  | "waiting"
  | "retrying"
  | "inferring"
  | "selecting_classifiers"
  | "uncertain"
  | "completed"
  | "failed"
  | "cancelled"
  | "unknown";

export type PayloadLogType = "error" | "warn" | "info" | (string & {});

export interface PayloadLog {
  Type?: PayloadLogType;
  Content?: string;
  type?: PayloadLogType;
  content?: string;
  timestamp?: string;
  createdAt?: string;
}

export interface PayloadStatus {
  id: string;
  name: string;
  state: PayloadState;
  iteration: number;
  lastAction: string;
  actions: string[];
  logs: NormalizedPayloadLog[];
  paths: string[];
  DirPath: string;
  startedAt?: string;
  updatedAt?: string;
  source?: "api" | "demo";
}

export interface NormalizedPayloadLog {
  id: string;
  type: PayloadLogType;
  content: string;
  timestamp?: string;
}

export interface DirInfo {
  dirId?: string;
  path?: string;
  created_at?: string;
  updated_at?: string;
  version?: number;
  brief?: string;
  creator?: { provider: string; model: string } | null;
  modifier?: { provider: string; model: string } | null;
}

export interface SnapShot {
  path: string;
  children: Record<string, SnapShot>;
  existingFiles: string[];
  info?: DirInfo | null;
}

export interface PayloadStatsResult {
  payloads: PayloadStatus[];
  source: "api" | "demo";
  warning?: string;
}

export interface DirectoryListResult {
  list: string[];
  source: "api" | "demo";
  warning?: string;
}

export interface SelectedDirectoryResult {
  selected: SnapShot;
  list: string[];
  source: "api" | "demo";
  warning?: string;
}

export interface DirInfoResult {
  info: DirInfo | null;
  source: "api" | "demo";
  warning?: string;
}

export interface ApiEnvelope {
  payloads?: unknown;
  list?: unknown;
  selected?: unknown;
  directory?: unknown;
  info?: unknown;
  [key: string]: unknown;
}
