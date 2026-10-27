import { Route, Routes } from "react-router";
import { AppLayout } from "../components/app-layout";
import { OverviewPage } from "../pages/overview";
import { PayloadInspectorPage } from "../pages/payload-inspector";
import { StageCanvasPage } from "../pages/stage-canvas";
import TextPayloadEditor from "@/pages/TextPayloadEditor";

export function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<OverviewPage />} />
        <Route path="payloads/:payloadId" element={<PayloadInspectorPage />} />
        <Route path="stages" element={<StageCanvasPage />} />
        <Route path="upload/content" element={<TextPayloadEditor />} />
        <Route path="*" element={<OverviewPage />} />
      </Route>
    </Routes>
  );
}
