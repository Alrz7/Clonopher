import { useEffect, useState, type FormEvent } from "react";
import { LoaderCircle, Send } from "lucide-react";

export type GlobalHead = string;

export type TextPayloadUploadResponse = {
  message?: string;
  address?: string;
};

type JsonRecord = Record<string, unknown>;

type FileType = "md" | "txt" | "custom";
type EditorMode = "content" | "description";

type TextPayloadEditorProps = {
  /** Override the head-list endpoint if the backend route changes. */
  headsEndpoint?: string;
  /** Override the text-payload endpoint if the backend route changes. */
  uploadEndpoint?: string;
  onProcessed?: (result: TextPayloadUploadResponse) => void;
};

function asRecord(value: unknown): JsonRecord {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as JsonRecord)
    : {};
}

function firstDefined(source: JsonRecord, ...keys: string[]): unknown {
  for (const key of keys) {
    if (source[key] !== undefined && source[key] !== null) return source[key];
  }
  return undefined;
}

function messageFromErrorBody(body: JsonRecord, fallback: string): string {
  const error = firstDefined(body, "error", "Error");
  if (typeof error === "string" && error.trim()) return error;
  if (error && typeof error === "object") {
    const message = firstDefined(asRecord(error), "message", "Message");
    if (typeof message === "string" && message.trim()) return message;
  }
  const message = firstDefined(body, "message", "Message");
  return typeof message === "string" && message.trim() ? message : fallback;
}

/** Fetch available global Directory heads from the existing Directory-list API. */
export async function fetchGlobalHeads(
  endpoint = "/api/v1/directories/list",
): Promise<GlobalHead[]> {
  const response = await fetch(endpoint, {
    headers: { Accept: "application/json" },
  });
  const body = asRecord(await response.json().catch(() => ({})));

  if (!response.ok) {
    throw new Error(
      messageFromErrorBody(
        body,
        `Could not load global directories (${response.status}).`,
      ),
    );
  }

  const envelope = asRecord(firstDefined(body, "data", "Data"));
  const source = Object.keys(envelope).length ? envelope : body;
  const rawList = firstDefined(
    source,
    "list",
    "List",
    "directories",
    "Directories",
    "heads",
    "Heads",
  );
  if (!Array.isArray(rawList)) {
    throw new Error("The Directory-list API response does not contain a list.");
  }

  const heads = rawList
    .map((item): string => {
      if (typeof item === "string") return item.trim();
      const directory = asRecord(item);
      const path = firstDefined(
        directory,
        "path",
        "Path",
        "name",
        "Name",
        "head",
        "Head",
      );
      return typeof path === "string" ? path.trim() : "";
    })
    .filter(Boolean);

  return [...new Set(heads)];
}

/** Submit text content using the JSON shape expected by HandleUploadTextFile. */
export async function uploadTextPayload(
  input: {
    fileName: string;
    head: string;
    content: string;
    description: string;
  },
  endpoint = "/api/v1/payloads/text",
): Promise<TextPayloadUploadResponse> {
  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(input),
  });
  const body = asRecord(await response.json().catch(() => ({})));

  if (!response.ok) {
    throw new Error(
      messageFromErrorBody(
        body,
        `Payload processing failed (${response.status}).`,
      ),
    );
  }

  const envelope = asRecord(firstDefined(body, "data", "Data"));
  const source = Object.keys(envelope).length ? envelope : body;
  const message = firstDefined(source, "message", "Message");
  const address = firstDefined(source, "address", "Address");

  return {
    message: typeof message === "string" ? message : undefined,
    address: typeof address === "string" ? address : undefined,
  };
}

function normalizeExtension(extension: string): string {
  return extension
    .trim()
    .replace(/^\.+/, "")
    .replace(/[^a-zA-Z0-9_-]/g, "");
}

function makeFileName(name: string, extension: string): string {
  const base = name.trim().replace(/[\\/]/g, "-").replace(/\.+$/, "");
  const normalizedExtension = normalizeExtension(extension);
  if (!base || !normalizedExtension) return "";

  const withoutSelectedExtension = base
    .toLowerCase()
    .endsWith(`.${normalizedExtension.toLowerCase()}`)
    ? base.slice(0, -(normalizedExtension.length + 1))
    : base;

  return `${withoutSelectedExtension}.${normalizedExtension}`;
}

export default function TextPayloadEditor({
  headsEndpoint = "/api/v1/directories/list",
  uploadEndpoint = "/api/v1/payloads/text",
  onProcessed,
}: TextPayloadEditorProps) {
  const [heads, setHeads] = useState<GlobalHead[]>([]);
  const [head, setHead] = useState("");
  const [isLoadingHeads, setIsLoadingHeads] = useState(true);
  const [headsError, setHeadsError] = useState("");
  const [name, setName] = useState("");
  const [fileType, setFileType] = useState<FileType>("md");
  const [customExtension, setCustomExtension] = useState("log");
  const [content, setContent] = useState("");
  const [description, setDescription] = useState("");
  const [mode, setMode] = useState<EditorMode>("content");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    const controller = new AbortController();

    async function loadHeads() {
      setIsLoadingHeads(true);
      setHeadsError("");
      try {
        const result = await fetchGlobalHeads(headsEndpoint);
        if (controller.signal.aborted) return;
        setHeads(result);
        setHead((current) =>
          result.includes(current) ? current : (result[0] ?? ""),
        );
      } catch (caught) {
        if (controller.signal.aborted) return;
        setHeads([]);
        setHead("");
        setHeadsError(
          caught instanceof Error
            ? caught.message
            : "Could not load global directories.",
        );
      } finally {
        if (!controller.signal.aborted) setIsLoadingHeads(false);
      }
    }

    void loadHeads();
    return () => controller.abort();
  }, [headsEndpoint]);

  const extension =
    fileType === "custom" ? normalizeExtension(customExtension) : fileType;
  const finalFileName = makeFileName(name, extension);
  const canSubmit =
    Boolean(head && finalFileName && content.trim()) &&
    !isLoadingHeads &&
    !headsError &&
    !isSubmitting;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSuccess("");

    if (!heads.includes(head)) {
      setError("Select a global Directory first.");
      return;
    }
    if (!finalFileName) {
      setError("Enter a file name and a valid extension.");
      return;
    }
    if (!content.trim()) {
      setError("Add some content before processing.");
      setMode("content");
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await uploadTextPayload(
        {
          fileName: finalFileName,
          head,
          content,
          description: description.trim(),
        },
        uploadEndpoint,
      );
      setSuccess(result.address || result.message || "Payload processed.");
      onProcessed?.(result);
    } catch (caught) {
      setError(
        caught instanceof Error ? caught.message : "Payload processing failed.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form className="text-payload" onSubmit={handleSubmit}>
      <div className="text-payload__topline">
        <label className="text-payload__name-field">
          <span className="field-label">FILE NAME</span>
          <input
            className="text-payload__input text-payload__name-input"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="e.g. meeting-notes"
            autoComplete="off"
            spellCheck={false}
          />
        </label>

        <div className="text-payload__type-group">
          <label className="text-payload__type-field">
            <span className="field-label">TYPE</span>
            <select
              className="text-payload__input text-payload__type-select"
              value={fileType}
              onChange={(event) => setFileType(event.target.value as FileType)}
            >
              <option value="md">.md</option>
              <option value="txt">.txt</option>
              <option value="custom">Custom</option>
            </select>
          </label>

          {fileType === "custom" && (
            <label className="text-payload__custom-field">
              <span className="field-label">EXTENSION</span>
              <input
                className="text-payload__input"
                value={customExtension}
                onChange={(event) =>
                  setCustomExtension(event.target.value.replace(/^\.+/, ""))
                }
                placeholder="log"
                autoComplete="off"
                spellCheck={false}
              />
            </label>
          )}
        </div>

        <label className="text-payload__head-field">
          <span className="field-label">GLOBAL STAGE</span>
          <select
            className="text-payload__input text-payload__head-select"
            value={head}
            onChange={(event) => {
              setHead(event.target.value);
              setError("");
              setSuccess("");
            }}
            disabled={isLoadingHeads || heads.length === 0}
          >
            <option value="">
              {isLoadingHeads
                ? "Loading directories…"
                : headsError
                  ? "Directories unavailable"
                  : heads.length
                    ? "Select a directory"
                    : "No directories available"}
            </option>
            {heads.map((item) => (
              <option key={item} value={item}>
                {item}
              </option>
            ))}
          </select>
        </label>
      </div>

      {headsError && (
        <p
          className="text-payload__inline-message text-payload__inline-message--error"
          role="status"
        >
          Could not load directories: {headsError}
        </p>
      )}

      <div className="text-payload__editor-toolbar">
        <div
          className="text-payload__mode-switch"
          role="tablist"
          aria-label="Payload text mode"
        >
          <button
            type="button"
            className={`button button-sm text-payload__mode-button${mode === "content" ? " is-active" : ""}`}
            role="tab"
            aria-selected={mode === "content"}
            onClick={() => setMode("content")}
          >
            Content
          </button>
          <button
            type="button"
            className={`button button-sm text-payload__mode-button${mode === "description" ? " is-active" : ""}`}
            role="tab"
            aria-selected={mode === "description"}
            onClick={() => setMode("description")}
          >
            Description
          </button>
        </div>
        <span className="text-payload__editor-hint">
          {mode === "content"
            ? `${content.length.toLocaleString()} characters`
            : "Optional classification context"}
        </span>
      </div>

      <textarea
        className="text-payload__editor"
        value={mode === "content" ? content : description}
        onChange={(event) =>
          mode === "content"
            ? setContent(event.target.value)
            : setDescription(event.target.value)
        }
        placeholder={
          mode === "content"
            ? "Write or paste the text you want Clonopher to process…"
            : "Add context about this text to help Clonopher classify it…"
        }
        aria-label={
          mode === "content" ? "Payload content" : "Payload description"
        }
        spellCheck={mode === "content"}
        autoCapitalize="off"
        autoComplete="off"
        autoCorrect="off"
      />

      <div className="text-payload__submit-area">
        {(error || success) && (
          <p
            className={`text-payload__feedback${error ? " text-payload__feedback--error" : " text-payload__feedback--success"}`}
            role="status"
            aria-live="polite"
          >
            {error || success}
          </p>
        )}
        <button
          className="send-button text-payload__process"
          type="submit"
          disabled={!canSubmit}
        >
          {isSubmitting ? (
            <LoaderCircle className="text-payload__spin" size={16} />
          ) : (
            <Send size={15} />
          )}
          {isSubmitting ? "Processing…" : "Process"}
        </button>
      </div>
    </form>
  );
}
