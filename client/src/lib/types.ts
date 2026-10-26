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
  stagePath: string;
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

export interface StageInfo {
  stage_id?: string;
  path?: string;
  created_at?: string;
  updated_at?: string;
  version?: number;
  brief?: string;
  creator?: { provider: string; model: string } | null;
  modifier?: { provider: string; model: string } | null;
}

export interface StageSnapshot {
  path: string;
  substages: Record<string, StageSnapshot>;
  existingFiles: string[];
  info?: StageInfo | null;
}

export interface PayloadStatsResult {
  payloads: PayloadStatus[];
  source: "api" | "demo";
  warning?: string;
}

export interface StageListResult {
  list: string[];
  source: "api" | "demo";
  warning?: string;
}

export interface SelectedStageResult {
  selected: StageSnapshot;
  list: string[];
  source: "api" | "demo";
  warning?: string;
}

export interface StageInfoResult {
  info: StageInfo | null;
  source: "api" | "demo";
  warning?: string;
}

export interface ApiEnvelope {
  payloads?: unknown;
  list?: unknown;
  selected?: unknown;
  stage?: unknown;
  info?: unknown;
  [key: string]: unknown;
}
