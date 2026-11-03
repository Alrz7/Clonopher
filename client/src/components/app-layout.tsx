import { useMemo } from "react";
import { NavLink, Outlet } from "react-router";
import { Boxes, GitBranch, TerminalSquare } from "lucide-react";
import { usePayloadStats } from "../lib/queries";
import { useAppStore } from "../stores/useAppStore";
import { LogDrawer } from "./log-drawer";

export function AppLayout() {
  const payloadQuery = usePayloadStats();
  const payloads = payloadQuery.data?.payloads ?? [];
  const { toggleLogs } = useAppStore();
  const logCount = useMemo(
    () => payloads.reduce((count, payload) => count + payload.logs.length, 0),
    [payloads],
  );

  return (
    <div className="app-shell">
      <header className="app-topbar">
        <NavLink
          to="/"
          className="app-brand"
          aria-label="Clonopher payload activity"
        >
          <span className="app-brand-mark">
            <Boxes size={17} strokeWidth={1.8} />
          </span>
          <span>clonopher</span>
        </NavLink>
        <nav className="primary-nav" aria-label="Main navigation">
          <NavLink
            to="/upload/content"
            end
            className={({ isActive }) =>
              `primary-nav-link ${isActive ? "primary-nav-link-active" : ""}`
            }
          >
            New Content
          </NavLink>
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `primary-nav-link ${isActive ? "primary-nav-link-active" : ""}`
            }
          >
            Payload activity
          </NavLink>
          <NavLink
            to="/directories"
            className={({ isActive }) =>
              `primary-nav-link ${isActive ? "primary-nav-link-active" : ""}`
            }
          >
            <GitBranch size={15} />
            Directories
          </NavLink>
        </nav>
      </header>

      <main className="main-shell">
        <Outlet />
      </main>

      <button
        className="floating-log-button"
        onClick={toggleLogs}
        aria-label="Open live logs"
        aria-haspopup="dialog"
      >
        <TerminalSquare size={17} strokeWidth={1.8} />
        <span>Live logs</span>
        {logCount > 0 && (
          <span className="floating-log-count">
            {Math.min(logCount, 99).toString().padStart(2, "0")}
          </span>
        )}
      </button>
      <LogDrawer payloads={payloads} />
    </div>
  );
}
