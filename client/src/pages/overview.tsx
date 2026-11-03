import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router";
import { FileText, Search, AlertCircle, RefreshCw } from "lucide-react";
import { usePayloadStats } from "../lib/queries";
import { timeAgo } from "../lib/format";
import { StatusBadge } from "../components/status-badge";
import type { PayloadState } from "../lib/types";

const activeStates: PayloadState[] = [
  "queued",
  "exploring",
  "classifying",
  "generating",
  "modifying",
  "waiting",
  "retrying",
  "inferring",
  "selecting_classifiers",
];
type Filter = "all" | "active" | "completed" | "failed";

export function OverviewPage() {
  const navigate = useNavigate();
  const query = usePayloadStats();
  const payloads = query.data?.payloads ?? [];
  const [filter, setFilter] = useState<Filter>("all");
  const [search, setSearch] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
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
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  const visiblePayloads = useMemo(
    () =>
      payloads.filter((payload) => {
        const matchesFilter =
          filter === "all" ||
          (filter === "active" && activeStates.includes(payload.state)) ||
          payload.state === filter;
        const haystack =
          `${payload.name} ${payload.id} ${payload.DirPath} ${payload.lastAction}`.toLowerCase();
        return matchesFilter && haystack.includes(search.trim().toLowerCase());
      }),
    [payloads, filter, search],
  );

  const filters: Array<{ value: Filter; label: string }> = [
    { value: "all", label: "All" },
    { value: "active", label: "Active" },
    { value: "completed", label: "Completed" },
    { value: "failed", label: "Failed" },
  ];

  return (
    <section className="activity-page">
      <div className="activity-header">
        <h1>Payload activity</h1>
        <span className="activity-count">
          {visiblePayloads.length}
          <span> / </span>
          {payloads.length}
        </span>
      </div>

      {query.data?.warning && (
        <div className="data-warning">
          <AlertCircle size={15} />
          <span>{query.data.warning}</span>
        </div>
      )}

      <div className="activity-toolbar">
        <div
          className="activity-filters"
          role="tablist"
          aria-label="Filter payload activity"
        >
          {filters.map((item) => (
            <button
              key={item.value}
              type="button"
              role="tab"
              aria-selected={filter === item.value}
              className={`activity-filter ${filter === item.value ? "activity-filter-active" : ""}`}
              onClick={() => setFilter(item.value)}
            >
              {item.label}
            </button>
          ))}
        </div>
        <label className="activity-search">
          <Search size={16} />
          <input
            ref={searchRef}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search payloads…"
          />
          {search ? (
            <button
              type="button"
              className="search-clear"
              onClick={() => setSearch("")}
              aria-label="Clear search"
            >
              ×
            </button>
          ) : (
            <kbd>/</kbd>
          )}
        </label>
        <button
          className="refresh-button"
          onClick={() => void query.refetch()}
          title="Refresh payloads"
          aria-label="Refresh payloads"
        >
          <RefreshCw
            size={15}
            className={query.isFetching ? "refresh-spinning" : ""}
          />
        </button>
      </div>

      <div className="payload-list" role="table" aria-label="Payload activity">
        <div className="payload-list-head" role="row">
          <span role="columnheader">Payload</span>
          <span role="columnheader">State</span>
          <span role="columnheader">Directory</span>
          <span role="columnheader">Last action</span>
          <span role="columnheader">Updated</span>
        </div>
        {visiblePayloads.map((payload) => (
          <button
            type="button"
            className="payload-row"
            role="row"
            key={payload.id}
            onClick={() =>
              navigate(`/payloads/${encodeURIComponent(payload.id)}`)
            }
          >
            <span className="payload-primary" role="cell">
              <span className="payload-file-icon">
                <FileText size={17} strokeWidth={1.7} />
              </span>
              <span className="payload-primary-copy">
                <strong title={payload.name}>{payload.name}</strong>
                <small title={payload.id}>{payload.id}</small>
              </span>
            </span>
            <span className="payload-state-cell" role="cell">
              <StatusBadge state={payload.state} />
            </span>
            <span
              className="payload-directory-cell"
              role="cell"
              title={payload.DirPath || "Unassigned"}
            >
              <span className="directory-path-mark"></span>
              {payload.DirPath || "Unassigned"}
            </span>
            <span
              className="payload-action-cell"
              role="cell"
              title={payload.lastAction || ""}
            >
              {payload.lastAction ||
                (payload.iteration ? `Iteration ${payload.iteration}` : "—")}
            </span>
            <span className="payload-updated-cell" role="cell">
              {timeAgo(payload.updatedAt)}
            </span>
          </button>
        ))}
        {visiblePayloads.length === 0 && (
          <div className="activity-empty">
            <Search size={20} strokeWidth={1.6} />
            <strong>
              {query.isLoading ? "Loading payloads…" : "No payloads found"}
            </strong>
            <span>
              {query.isLoading
                ? "Waiting for the runtime snapshot."
                : "Try another search or filter."}
            </span>
          </div>
        )}
      </div>
    </section>
  );
}
