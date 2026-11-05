import {
  useEffect,
  useRef,
  useState,
  type DragEvent,
  type ChangeEvent,
} from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  Check,
  FileText,
  LoaderCircle,
  Paperclip,
  RefreshCw,
  Send,
  Upload,
  X,
} from "lucide-react";
import { useDirectoryList } from "../../lib/queries";
import "../../styles/payload-composer.css";

type Asset = {
  id: string;
  fileName: string;
  src: string;
  status: string;
};

type UploadItem = {
  key: string;
  file: File;
  state: "uploading" | "uploaded" | "failed";
  error?: string;
};

type TargetMode = "root" | "path";
type Notice = { kind: "success" | "error"; text: string } | null;

const API = "/api/v1";

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object"
    ? (value as Record<string, unknown>)
    : {};
}

function readString(value: unknown, ...keys: string[]): string {
  const source = asRecord(value);
  for (const key of keys) {
    if (typeof source[key] === "string") return source[key] as string;
  }
  return "";
}

async function apiRequest<T = unknown>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${API}${path}`, { ...init, headers });
  const raw = await response.text();
  let body: unknown = null;

  if (raw) {
    try {
      body = JSON.parse(raw);
    } catch {
      body = raw;
    }
  }

  if (!response.ok) {
    const message =
      readString(body, "error", "message", "Error", "Message") ||
      (typeof body === "string" ? body : response.statusText);
    throw new Error(`API ${response.status}: ${message}`);
  }

  return body as T;
}

async function getStagings(): Promise<Asset[]> {
  const response = asRecord(await apiRequest("/stagings"));
  const raw = response.stagings ?? response.Stagings ?? {};
  const entries = Array.isArray(raw)
    ? raw.map((asset, index) => [String(index), asset] as const)
    : Object.entries(asRecord(raw));

  return entries
    .map(([key, value]) => ({
      id: readString(value, "id", "ID") || key,
      fileName:
        readString(value, "fileName", "FileName", "file_name") ||
        "Untitled file",
      src: readString(value, "src", "Src"),
      status: readString(value, "status", "Status") || "idle",
    }))
    .filter((asset) => asset.id.length > 0);
}

async function uploadAsset(file: File): Promise<void> {
  const form = new FormData();
  form.append("file", file);
  form.append("filename", file.name);
  await apiRequest("/payloads/file", { method: "POST", body: form });
}

async function deleteAsset(id: string): Promise<void> {
  await apiRequest("/payloads/file", {
    method: "DELETE",
    body: JSON.stringify({ id }),
  });
}

function DestinationPicker({
  roots,
  rootsLoading,
  mode,
  value,
  path,
  onModeChange,
  onValueChange,
  onPathChange,
}: {
  roots: string[];
  rootsLoading: boolean;
  mode: TargetMode;
  value: string;
  path: string;
  onModeChange: (mode: TargetMode) => void;
  onValueChange: (value: string) => void;
  onPathChange: (path: string) => void;
}) {
  return (
    <div className="pc-target-field">
      <div className="pc-field-heading">
        <label>Destination</label>
        <div
          className="pc-target-switch"
          role="group"
          aria-label="Destination type"
        >
          <button
            type="button"
            className={mode === "root" ? "is-active" : ""}
            onClick={() => onModeChange("root")}
          >
            Root
          </button>
          <button
            type="button"
            className={mode === "path" ? "is-active" : ""}
            onClick={() => onModeChange("path")}
          >
            Custom path
          </button>
        </div>
      </div>

      {mode === "root" ? (
        <select
          className="pc-control"
          value={value}
          onChange={(event) => onValueChange(event.target.value)}
          disabled={rootsLoading || roots.length === 0}
        >
          {roots.length === 0 ? (
            <option value="">
              {rootsLoading ? "Loading roots…" : "No roots available"}
            </option>
          ) : (
            roots.map((root) => (
              <option key={root} value={root}>
                {root}
              </option>
            ))
          )}
        </select>
      ) : (
        <input
          className="pc-control"
          value={path}
          onChange={(event) => onPathChange(event.target.value)}
          placeholder="/home/user/Documents/Inbox"
          spellCheck={false}
        />
      )}
    </div>
  );
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function PayloadComposerPage() {
  const queryClient = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const directoryQuery = useDirectoryList();
  const stagingQuery = useQuery({
    queryKey: ["stagings"],
    queryFn: getStagings,
    staleTime: 2_000,
    refetchOnWindowFocus: true,
  });

  const roots = directoryQuery.data?.list ?? [];
  const assets = stagingQuery.data ?? [];

  const [mode, setMode] = useState<"files" | "text">("files");
  const [targetMode, setTargetMode] = useState<TargetMode>("root");
  const [selectedRoot, setSelectedRoot] = useState("");
  const [customPath, setCustomPath] = useState("");

  const [payloadName, setPayloadName] = useState("");
  const [fileDescription, setFileDescription] = useState("");
  const [textName, setTextName] = useState("");
  const [textExtension, setTextExtension] = useState<"md" | "txt">("md");
  const [textDescription, setTextDescription] = useState("");
  const [textContent, setTextContent] = useState("");

  const [uploadQueue, setUploadQueue] = useState<UploadItem[]>([]);
  const [uploading, setUploading] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);

  useEffect(() => {
    if (!selectedRoot && roots.length > 0) setSelectedRoot(roots[0]);
  }, [roots, selectedRoot]);

  const destination =
    targetMode === "root" ? selectedRoot.trim() : customPath.trim();

  const refreshStagings = () =>
    queryClient.invalidateQueries({ queryKey: ["stagings"] });

  const chooseFiles = (files: File[]) => {
    if (!files.length || uploading) return;

    const batch: UploadItem[] = files.map((file, index) => ({
      key: `${Date.now()}-${index}-${file.name}`,
      file,
      state: "uploading",
    }));

    setUploadQueue(batch);
    setNotice(null);
    setUploading(true);

    void (async () => {
      for (const item of batch) {
        try {
          await uploadAsset(item.file);
          setUploadQueue((current) =>
            current.map((queued) =>
              queued.key === item.key
                ? { ...queued, state: "uploaded" }
                : queued,
            ),
          );
        } catch (error) {
          setUploadQueue((current) =>
            current.map((queued) =>
              queued.key === item.key
                ? {
                    ...queued,
                    state: "failed",
                    error:
                      error instanceof Error ? error.message : "Upload failed",
                  }
                : queued,
            ),
          );
        }
      }

      await refreshStagings();
      setUploading(false);
    })();
  };

  const onFileInput = (event: ChangeEvent<HTMLInputElement>) => {
    chooseFiles(Array.from(event.target.files ?? []));
    event.target.value = "";
  };

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragging(false);
    chooseFiles(Array.from(event.dataTransfer.files));
  };

  const onDragOver = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    if (!uploading) setDragging(true);
  };

  const onDragLeave = (event: DragEvent<HTMLDivElement>) => {
    if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
      setDragging(false);
    }
  };

  const handleDelete = async (asset: Asset) => {
    if (deletingId) return;
    setDeletingId(asset.id);
    setNotice(null);
    try {
      await deleteAsset(asset.id);
      await refreshStagings();
    } catch (error) {
      setNotice({
        kind: "error",
        text:
          error instanceof Error ? error.message : "Could not delete asset.",
      });
    } finally {
      setDeletingId(null);
    }
  };

  const submitFiles = async () => {
    if (!assets.length || !destination || uploading || submitting) return;
    setSubmitting(true);
    setNotice(null);
    try {
      await apiRequest("/payloads/new", {
        method: "POST",
        body: JSON.stringify({
          name: payloadName.trim(),
          stagings: assets.map((asset) => asset.id),
          destination,
          description: fileDescription.trim(),
        }),
      });
      setNotice({
        kind: "success",
        text: "Payload accepted. Follow its progress in Payload activity.",
      });
      setPayloadName("");
      setFileDescription("");
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["stagings"] }),
        queryClient.invalidateQueries({ queryKey: ["payloads"] }),
      ]);
    } catch (error) {
      setNotice({
        kind: "error",
        text:
          error instanceof Error ? error.message : "Could not start payload.",
      });
    } finally {
      setSubmitting(false);
    }
  };

  const submitText = async () => {
    const baseName = textName.trim().replace(/\.(md|txt)$/i, "");
    if (!baseName || !textContent.trim() || !destination || submitting) return;

    setSubmitting(true);
    setNotice(null);
    try {
      await apiRequest("/payloads/text", {
        method: "POST",
        body: JSON.stringify({
          fileName: `${baseName}.${textExtension}`,
          head: destination,
          content: textContent,
          description: textDescription.trim(),
        }),
      });
      setNotice({
        kind: "success",
        text: "Text payload accepted. Follow its progress in Payload activity.",
      });
      setTextContent("");
      await queryClient.invalidateQueries({ queryKey: ["payloads"] });
    } catch (error) {
      setNotice({
        kind: "error",
        text:
          error instanceof Error
            ? error.message
            : "Could not create text payload.",
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <main
      className={`pc-page${dragging ? " pc-page-dragging" : ""}`}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      <input
        ref={fileInputRef}
        className="pc-hidden-input"
        type="file"
        multiple
        onChange={onFileInput}
      />

      <header className="pc-header">
        <div className="pc-heading-copy">
          <span className="pc-eyebrow">CLONOPHER / COMPOSER</span>
          <h1>New payload</h1>
          <p>Prepare files or compose a text artifact for processing.</p>
        </div>
        <div
          className="pc-mode-switch"
          role="tablist"
          aria-label="Payload type"
        >
          <button
            type="button"
            role="tab"
            aria-selected={mode === "files"}
            className={mode === "files" ? "is-active" : ""}
            onClick={() => setMode("files")}
          >
            <Paperclip size={15} /> Files
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={mode === "text"}
            className={mode === "text" ? "is-active" : ""}
            onClick={() => setMode("text")}
          >
            <FileText size={15} /> Text file
          </button>
        </div>
      </header>

      {mode === "files" ? (
        <section className="pc-workspace">
          <div className="pc-section-heading">
            <div>
              <h2>Assets</h2>
              <span>{assets.length} ready to be grouped into one payload</span>
            </div>
            <button
              type="button"
              className="pc-icon-button"
              onClick={() => void stagingQuery.refetch()}
              title="Refresh assets"
              aria-label="Refresh assets"
            >
              <RefreshCw
                size={15}
                className={stagingQuery.isFetching ? "pc-spin" : ""}
              />
            </button>
          </div>

          <div className="pc-assets-area">
            {stagingQuery.isLoading ? (
              <div className="pc-state-line">
                <LoaderCircle className="pc-spin" size={16} /> Loading staged
                assets…
              </div>
            ) : stagingQuery.isError ? (
              <div className="pc-state-line pc-state-error">
                <AlertCircle size={16} /> Could not load staged assets.
              </div>
            ) : assets.length === 0 ? (
              <div className="pc-empty-drop">
                <span className="pc-empty-icon">
                  <Upload size={21} />
                </span>
                <strong>Drop files to start a payload</strong>
                <span>or choose files from your computer</span>
                <button
                  type="button"
                  className="pc-secondary-button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={uploading}
                >
                  Browse files
                </button>
              </div>
            ) : (
              <div className="pc-asset-list">
                {assets.map((asset) => (
                  <div className="pc-asset-row" key={asset.id}>
                    <span className="pc-file-icon">
                      <FileText size={17} />
                    </span>
                    <span className="pc-asset-name" title={asset.fileName}>
                      {asset.fileName}
                    </span>
                    <span className="pc-asset-state">READY</span>
                    <button
                      type="button"
                      className="pc-remove-button"
                      onClick={() => void handleDelete(asset)}
                      disabled={
                        deletingId === asset.id ||
                        Boolean(deletingId) ||
                        submitting
                      }
                      aria-label={`Remove ${asset.fileName}`}
                      title="Remove asset"
                    >
                      {deletingId === asset.id ? (
                        <LoaderCircle size={15} className="pc-spin" />
                      ) : (
                        <X size={16} />
                      )}
                    </button>
                  </div>
                ))}
              </div>
            )}

            {uploadQueue.length > 0 && (
              <div className="pc-upload-queue">
                <div className="pc-queue-heading">
                  <span>Recent uploads</span>
                  <span>
                    {
                      uploadQueue.filter((item) => item.state !== "uploading")
                        .length
                    }
                    /{uploadQueue.length}
                  </span>
                </div>
                {uploadQueue.map((item) => (
                  <div className="pc-queue-row" key={item.key}>
                    <span className={`pc-queue-icon pc-queue-${item.state}`}>
                      {item.state === "uploading" ? (
                        <LoaderCircle size={14} className="pc-spin" />
                      ) : item.state === "uploaded" ? (
                        <Check size={14} />
                      ) : (
                        <AlertCircle size={14} />
                      )}
                    </span>
                    <span className="pc-queue-name" title={item.file.name}>
                      {item.file.name}
                    </span>
                    <span className="pc-queue-meta">
                      {item.state === "failed"
                        ? item.error
                        : item.state === "uploaded"
                          ? "Uploaded"
                          : formatSize(item.file.size)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="pc-compose-area">
            <div className="pc-compose-fields">
              <label className="pc-field">
                <span>
                  Payload name <small>OPTIONAL</small>
                </span>
                <input
                  className="pc-control"
                  value={payloadName}
                  onChange={(event) => setPayloadName(event.target.value)}
                  placeholder="Auto-generated if empty"
                />
              </label>
              <DestinationPicker
                roots={roots}
                rootsLoading={directoryQuery.isLoading}
                mode={targetMode}
                value={selectedRoot}
                path={customPath}
                onModeChange={setTargetMode}
                onValueChange={setSelectedRoot}
                onPathChange={setCustomPath}
              />
            </div>

            <div className="pc-chat-composer">
              <textarea
                value={fileDescription}
                onChange={(event) => setFileDescription(event.target.value)}
                placeholder="Describe these files… All assets above will share this description and become one payload."
                rows={2}
                disabled={submitting}
              />
              <div className="pc-composer-actions">
                <button
                  type="button"
                  className="pc-attach-button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={uploading || submitting}
                >
                  <Paperclip size={16} /> <span>Add files</span>
                </button>
                <span className="pc-composer-hint">
                  {assets.length} asset{assets.length === 1 ? "" : "s"} · one
                  payload
                </span>
                <button
                  type="button"
                  className="pc-send-button"
                  onClick={() => void submitFiles()}
                  disabled={
                    !assets.length ||
                    !destination ||
                    uploading ||
                    submitting ||
                    Boolean(deletingId)
                  }
                  title="Create payload"
                  aria-label="Create payload"
                >
                  {submitting ? (
                    <LoaderCircle size={17} className="pc-spin" />
                  ) : (
                    <Send size={17} />
                  )}
                </button>
              </div>
            </div>
            {uploading && (
              <p className="pc-helper-line">Uploading files one by one…</p>
            )}
            {notice && (
              <p className={`pc-notice pc-notice-${notice.kind}`}>
                {notice.text}
              </p>
            )}
          </div>
        </section>
      ) : (
        <section className="pc-workspace pc-text-workspace">
          <div className="pc-text-fields">
            <label className="pc-field pc-name-field">
              <span>File name</span>
              <input
                className="pc-control"
                value={textName}
                onChange={(event) => setTextName(event.target.value)}
                placeholder="notes"
              />
            </label>
            <label className="pc-field pc-extension-field">
              <span>Format</span>
              <select
                className="pc-control"
                value={textExtension}
                onChange={(event) =>
                  setTextExtension(event.target.value as "md" | "txt")
                }
              >
                <option value="md">.md</option>
                <option value="txt">.txt</option>
              </select>
            </label>
            <DestinationPicker
              roots={roots}
              rootsLoading={directoryQuery.isLoading}
              mode={targetMode}
              value={selectedRoot}
              path={customPath}
              onModeChange={setTargetMode}
              onValueChange={setSelectedRoot}
              onPathChange={setCustomPath}
            />
          </div>

          <label className="pc-field pc-description-field">
            <span>
              Description <small>OPTIONAL</small>
            </span>
            <input
              className="pc-control"
              value={textDescription}
              onChange={(event) => setTextDescription(event.target.value)}
              placeholder="A short description for this text payload"
            />
          </label>

          <label className="pc-field pc-content-field">
            <span>
              Content{" "}
              <small>{textContent.length.toLocaleString()} CHARACTERS</small>
            </span>
            <textarea
              className="pc-text-editor"
              value={textContent}
              onChange={(event) => setTextContent(event.target.value)}
              placeholder="Start writing…"
              spellCheck={false}
            />
          </label>

          <div className="pc-text-footer">
            {notice && (
              <p className={`pc-notice pc-notice-${notice.kind}`}>
                {notice.text}
              </p>
            )}
            <span className="pc-composer-hint">
              Will be created as{" "}
              {textName.trim().replace(/\.(md|txt)$/i, "") || "filename"}.
              {textExtension}
            </span>
            <button
              type="button"
              className="pc-submit-text"
              onClick={() => void submitText()}
              disabled={
                !textName.trim() ||
                !textContent.trim() ||
                !destination ||
                submitting
              }
            >
              {submitting ? (
                <LoaderCircle size={16} className="pc-spin" />
              ) : (
                <Send size={16} />
              )}
              Create text payload
            </button>
          </div>
        </section>
      )}

      {dragging && (
        <div className="pc-drop-overlay" aria-hidden="true">
          <div>
            <Upload size={26} />
            <strong>Drop to upload</strong>
            <span>Files will be added to this payload.</span>
          </div>
        </div>
      )}
    </main>
  );
}
