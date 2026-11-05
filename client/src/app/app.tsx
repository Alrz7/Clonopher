import { Route, Routes } from "react-router";
import { AppLayout } from "../components/app-layout";
import { OverviewPage } from "../pages/overview";
import { PayloadInspectorPage } from "../pages/payload-inspector";
import { DirectoryCanvasPage } from "../pages/stage-canvas";
import { PayloadComposerPage } from "../features/payloads/payload-composer";

export function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<OverviewPage />} />
        <Route path="payloads/:payloadId" element={<PayloadInspectorPage />} />
        <Route path="directories" element={<DirectoryCanvasPage />} />
        <Route path="upload/content" element={<PayloadComposerPage />} />
        <Route path="*" element={<OverviewPage />} />
      </Route>
    </Routes>
  );
}
