// import { demoDirectoryNames, demoDirectoryTree } from "./demo-data";
import type {
  ApiEnvelope,
  NormalizedPayloadLog,
  PayloadLog,
  PayloadStatsResult,
  PayloadStatus,
  PayloadState,
  SelectedDirectoryResult,
  DirInfo,
  DirInfoResult,
  DirectoryListResult,
  SnapShot,
} from "./types";

const API = "/api/v1";

function record(value: unknown): Record<string, unknown> {
  return value && typeof value === "object"
    ? (value as Record<string, unknown>)
    : {};
}

function valueOf(source: Record<string, unknown>, ...keys: string[]): unknown {
  for (const key of keys) {
    if (source[key] !== undefined) return source[key];
  }
  return undefined;
}

function stringValue(value: unknown, fallback = ""): string {
  return typeof value === "string" ? value : fallback;
}

function numberValue(value: unknown, fallback = 0): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function arrayValue(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function stateValue(value: unknown): PayloadState {
  const state = stringValue(value, "unknown").toLowerCase() as PayloadState;
  const supported: PayloadState[] = [
    "queued",
    "exploring",
    "classifying",
    "generating",
    "modifying",
    "waiting",
    "retrying",
    "inferring",
    "selecting_classifiers",
    "uncertain",
    "completed",
    "failed",
    "cancelled",
    "unknown",
  ];
  return supported.includes(state) ? state : "unknown";
}

function normalizeLog(input: unknown, index: number): NormalizedPayloadLog {
  const log = record(input) as PayloadLog & Record<string, unknown>;
  return {
    id: stringValue(valueOf(log, "id", "ID"), `log-${index}`),
    type: stringValue(valueOf(log, "type", "Type"), "info").toLowerCase(),
    content: stringValue(
      valueOf(log, "content", "Content", "message", "Message"),
      String(input ?? ""),
    ),
    timestamp:
      stringValue(
        valueOf(log, "timestamp", "createdAt", "Timestamp", "CreatedAt"),
      ) || undefined,
  };
}

function normalizePayload(
  input: unknown,
  fallbackId: string,
  index: number,
): PayloadStatus {
  const p = record(input);
  const logs = arrayValue(valueOf(p, "logs", "Logs")).map(normalizeLog);
  const actions = arrayValue(valueOf(p, "actions", "Actions")).map((item) =>
    String(item),
  );
  const directory = record(valueOf(p, "directory", "Directory"));

  return {
    id: stringValue(valueOf(p, "id", "ID"), fallbackId || `payload-${index}`),
    name: stringValue(
      valueOf(p, "name", "Name", "filename", "Filename"),
      fallbackId || `Payload ${index + 1}`,
    ),
    state: stateValue(valueOf(p, "state", "State")),
    iteration: numberValue(valueOf(p, "iteration", "Iteration")),
    lastAction: stringValue(valueOf(p, "lastAction", "LastAction")),
    actions,
    logs,
    paths: arrayValue(valueOf(p, "road")).map((item) => String(item)),
    DirPath: stringValue(
      valueOf(p, "path"),
      stringValue(valueOf(directory, "path", "Path"), "—"),
    ),
    startedAt: stringValue(valueOf(p, "startedAt", "StartedAt")) || undefined,
    updatedAt: stringValue(valueOf(p, "updatedAt", "UpdatedAt")) || undefined,
    source: "api",
  };
}

async function requestJson<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      Accept: "application/json",
      ...(init?.body ? { "Content-Type": "application/json" } : {}),
      ...init?.headers,
    },
  });

  if (!response.ok) {
    const body = await response.text().catch(() => "");
    throw new Error(
      `API ${response.status} ${response.statusText}${body ? ` — ${body.slice(0, 160)}` : ""}`,
    );
  }

  return response.json() as Promise<T>;
}

function normalizePayloadCollection(raw: unknown): PayloadStatus[] {
  const payloadObject = record(raw);
  const candidates =
    valueOf(payloadObject, "payloads", "Payloads", "stats", "Stats") ?? raw;

  if (Array.isArray(candidates)) {
    return candidates.map((item, index) => normalizePayload(item, "", index));
  }

  return Object.entries(record(candidates)).map(([id, item], index) =>
    normalizePayload(item, id, index),
  );
}

export async function getPayloadStats(): Promise<PayloadStatsResult> {
  // try {
  const raw = await requestJson<ApiEnvelope>(`${API}/payloads/stats`);
  const payloads = normalizePayloadCollection(raw);
  return { payloads, source: "api" };
  // } catch (error) {
  //   return {
  //     payloads: demoPayloads.map((payload) => ({ ...payload, logs: [...payload.logs], actions: [...payload.actions], source: 'demo' })),
  //     source: 'demo',
  //     warning: error instanceof Error ? error.message : 'The API is unavailable.',
  //   }
  // }
}

export async function getDirectoryList(): Promise<DirectoryListResult> {
  // try {
  const raw = record(await requestJson<ApiEnvelope>(`${API}/directories/list`));
  const listRaw = valueOf(raw, "list", "List", "directories", "Directories");
  const list = arrayValue(listRaw)
    .map((item) => {
      if (typeof item === "string") return item;
      const directory = record(item);
      return stringValue(valueOf(directory, "path", "Path", "name", "Name"));
    })
    .filter(Boolean);
  return { list, source: "api" };
  // } catch (error) {
  // return {
  //   list: [...demoDirectoryNames],
  //   source: "demo",
  //   warning:
  //     error instanceof Error
  //       ? error.message
  //       : "The Directory API is unavailable.",
  // };
  // }
}

function normalizeDirectory(input: unknown, fallbackPath: string): SnapShot {
  const directory = record(input);
  const roots = record(valueOf(directory, "children"));
  const files = arrayValue(
    valueOf(directory, "existingFiles", "ExistingFiles", "files", "Files"),
  ).map(String);
  const normalizedRoots: Record<string, SnapShot> = {};

  for (const [name, child] of Object.entries(roots)) {
    normalizedRoots[name] = normalizeDirectory(
      child,
      `${fallbackPath}/${name}`,
    );
  }

  const infoRaw = valueOf(directory, "info", "Info");
  return {
    path: stringValue(valueOf(directory, "path", "Path"), fallbackPath),
    children: normalizedRoots,
    existingFiles: files,
    info:
      infoRaw && typeof infoRaw === "object"
        ? (infoRaw as SnapShot["info"])
        : undefined,
  };
}

export async function getGlobalDirectory(
  path: string,
): Promise<SelectedDirectoryResult> {
  // try {
  const raw = record(
    await requestJson<ApiEnvelope>(`${API}/directories/query`, {
      method: "POST",
      body: JSON.stringify({ select: path }),
    }),
  );
  const selected = valueOf(
    raw,
    "selected",
    "Selected",
    "directory",
    "Directory",
  );
  const listRaw = valueOf(raw, "list", "List", "directories", "Directories");
  const list = arrayValue(listRaw)
    .map((item) =>
      typeof item === "string"
        ? item
        : stringValue(valueOf(record(item), "path", "Path", "name", "Name")),
    )
    .filter(Boolean);

  if (!selected || typeof selected !== "object")
    throw new Error(
      "API response did not contain a selected Directory snapshot.",
    );

  return {
    selected: normalizeDirectory(selected, path),
    list: list.length ? list : [path],
    source: "api",
  };
  // } catch (error) {
  //   return {
  //     selected: findDemoDirectory(path) ?? demoDirectoryTree,
  //     list: [...demoDirectoryNames],
  //     source: "demo",
  //     warning:
  //       error instanceof Error
  //         ? error.message
  //         : "The Directory API is unavailable.",
  //   };
  // }
}

// function findDemoDirectory(pathOrName: string): DirectorySnapshot | undefined {
//   if (!pathOrName || pathOrName.toLowerCase() === "root") return demoDirectoryTree;

//   const wanted = pathOrName
//     .replaceAll("\\", "/")
//     .replace(/\/$/, "")
//     .toLowerCase();

//   function visit(directory: DirectorySnapshot): DirectorySnapshot | undefined {
//     for (const [name, child] of Object.entries(directory.children ?? {})) {
//       const childPath = child.path
//         .replaceAll("\\", "/")
//         .replace(/\/$/, "")
//         .toLowerCase();
//       if (
//         name.toLowerCase() === wanted ||
//         childPath === wanted ||
//         childPath.endsWith(`/${wanted}`)
//       )
//         return child;
//       const nested = visit(child);
//       if (nested) return nested;
//     }
//     return undefined;
//   }

//   return visit(demoDirectoryTree);
// }

function normalizeDirInfo(input: unknown): DirInfo | null {
  if (!input || typeof input !== "object") return null;
  const info = record(input);
  const creatorRaw = valueOf(info, "creator", "Creator");
  const modifierRaw = valueOf(info, "modifier", "Modifier");
  const normalizeProvider = (raw: unknown) => {
    if (!raw || typeof raw !== "object") return null;
    const provider = record(raw);
    return {
      provider: stringValue(valueOf(provider, "provider", "Provider")),
      model: stringValue(valueOf(provider, "model", "Model")),
    };
  };

  return {
    dirId: stringValue(
      valueOf(info, "dirId", "DirectoryID", "dir_id", "DirectoryId"),
    ),
    path: stringValue(valueOf(info, "path", "Path")),
    created_at: stringValue(
      valueOf(info, "created_at", "CreatedAt", "createdAt"),
    ),
    updated_at: stringValue(
      valueOf(info, "updated_at", "UpdatedAt", "updatedAt"),
    ),
    version: numberValue(valueOf(info, "version", "Version")),
    brief: stringValue(valueOf(info, "brief", "Brief")),
    creator: normalizeProvider(creatorRaw),
    modifier: normalizeProvider(modifierRaw),
  };
}

// function findDemoInfo(pathOrName: string): DirInfo | null {
//   const directory = findDemoDirectory(pathOrName);
//   return directory?.info ?? null;
// }

export async function getDirInfo(path: string): Promise<DirInfoResult> {
  // try {
  const raw = record(
    await requestJson<ApiEnvelope>(`${API}/directories/info`, {
      method: "POST",
      body: JSON.stringify({ path }),
    }),
  );
  const info = valueOf(raw, "info", "Info");
  return { info: normalizeDirInfo(info), source: "api" };
  // } catch (error) {
  //   return {
  //     info: findDemoInfo(path),
  //     source: "demo",
  //     warning:
  //       error instanceof Error
  //         ? error.message
  //         : "The Directory info API is unavailable.",
  //   };
  // }
}

export async function getGlobalLogs(): Promise<NormalizedPayloadLog[]> {
  const raw = record(await requestJson<ApiEnvelope>(`${API}/logs`));

  return arrayValue(valueOf(raw, "logs", "Logs")).map((item, index) =>
    normalizeLog(item, index),
  );
}
