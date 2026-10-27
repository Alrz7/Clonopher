import { useEffect, useMemo, useRef } from "react";
import { Search, X } from "lucide-react";
import { useAppStore } from "../stores/useAppStore";
import type { PayloadStatus } from "../lib/types";
import { logClass, logLabel } from "../lib/format";
import { IconButton } from "./primitives";
import { useGlobalLogs } from "../lib/queries";

function logTime(value?: string): string {
  if (!value) return "--:--:--";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "--:--:--";
  return date.toLocaleTimeString([], {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function LogDrawer({ payloads }: { payloads: PayloadStatus[] }) {
  const {
    logsOpen,
    logFilter,
    logSearch,
    setLogsOpen,
    setLogFilter,
    setLogSearch,
  } = useAppStore();
  const searchRef = useRef<HTMLInputElement>(null);
  const globalLogsQuery = useGlobalLogs(logsOpen);

  useEffect(() => {
    if (!logsOpen) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") setLogsOpen(false);
      const target = event.target;
      const isTyping =
        target instanceof HTMLElement &&
        (target.isContentEditable ||
          ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName));
      if (
        event.key === "/" &&
        !isTyping &&
        !event.metaKey &&
        !event.ctrlKey &&
        !event.altKey
      ) {
        event.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    const focusTimer = window.setTimeout(() => searchRef.current?.focus(), 40);
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      window.clearTimeout(focusTimer);
    };
  }, [logsOpen, setLogsOpen]);

  const logs = useMemo(
    () =>
      [
        ...(globalLogsQuery.data ?? []).map((log) => ({
          ...log,
          payloadId: "global",
          payloadName: "runtime",
        })),

        ...payloads.flatMap((payload) =>
          payload.logs.map((log) => ({
            ...log,
            payloadId: payload.id,
            payloadName: payload.name,
          })),
        ),
      ]
        .filter(
          (log) =>
            logFilter === "all" ||
            log.type.toLowerCase() === logFilter ||
            (logFilter === "warn" && log.type.toLowerCase() === "warning"),
        )
        .filter((log) =>
          `${log.content} ${log.payloadName} ${log.type}`
            .toLowerCase()
            .includes(logSearch.toLowerCase()),
        )
        .sort(
          (a, b) =>
            new Date(b.timestamp ?? 0).getTime() -
            new Date(a.timestamp ?? 0).getTime(),
        ),
    [payloads, globalLogsQuery.data, logFilter, logSearch],
  );
  if (!logsOpen) return null;

  return (
    <div className="drawer-layer">
      <button
        className="drawer-scrim"
        aria-label="Close live logs"
        onClick={() => setLogsOpen(false)}
      />
      <aside
        className="log-drawer"
        role="dialog"
        aria-modal="true"
        aria-label="Live logs"
      >
        <div className="drawer-header">
          <div className="drawer-title-row">
            <span className="drawer-live-indicator" />
            <h2>Live logs</h2>
            <span className="drawer-event-count">{logs.length}</span>
          </div>
          <IconButton label="Close logs" onClick={() => setLogsOpen(false)}>
            <X size={17} />
          </IconButton>
        </div>
        <label className="log-search">
          <Search size={16} />
          <input
            ref={searchRef}
            value={logSearch}
            onChange={(event) => setLogSearch(event.target.value)}
            placeholder="Search logs or payloads…"
          />
          <kbd>/</kbd>
        </label>
        <div
          className="log-filters"
          role="tablist"
          aria-label="Filter log types"
        >
          {(["all", "error", "warn", "info"] as const).map((filter) => (
            <button
              key={filter}
              role="tab"
              aria-selected={logFilter === filter}
              className={`log-filter ${logFilter === filter ? "log-filter-active" : ""}`}
              onClick={() => setLogFilter(filter)}
            >
              {filter === "all"
                ? "All"
                : filter === "warn"
                  ? "Warnings"
                  : filter === "error"
                    ? "Errors"
                    : "Info"}
            </button>
          ))}
        </div>
        <div className="drawer-log-list" aria-live="polite">
          {logs.length === 0 ? (
            <div className="log-empty">
              <span className="log-empty-mark">∅</span>
              <strong>No log entries</strong>
              <p>New payload events will appear here.</p>
            </div>
          ) : (
            logs.map((log) => (
              <article
                className="console-log"
                key={`${log.payloadId}-${log.id}`}
              >
                <time>{logTime(log.timestamp)}</time>
                <span className={`console-log-type ${logClass(log.type)}`}>
                  {logLabel(log.type).toLowerCase()}
                </span>
                <span className="console-log-payload">[{log.payloadName}]</span>
                <p>{log.content}</p>
              </article>
            ))
          )}
        </div>
        <div className="drawer-bottom">
          <span>
            {logs.length} matching {logs.length === 1 ? "entry" : "entries"}
          </span>
          <button
            onClick={() => {
              setLogSearch("");
              setLogFilter("all");
            }}
          >
            Clear filters
          </button>
        </div>
      </aside>
    </div>
  );
}
