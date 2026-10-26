# Clonopher UI refresh

This is an overlay patch for the existing `web/` project. It replaces only the layout and page presentation files, plus makes Vite resolve `.env*` files relative to `vite.config.ts` instead of the shell's current working directory.

## Apply on Fedora/Linux

From your Clonopher repository root:

```bash
cd ~/GoPrj/Clonopher
mkdir -p /tmp/clonopher-ui-backup
cp --parents \
  web/src/styles.css \
  web/src/components/app-layout.tsx \
  web/src/components/log-drawer.tsx \
  web/src/pages/overview.tsx \
  web/src/pages/stage-canvas.tsx \
  web/src/pages/payload-inspector.tsx \
  web/vite.config.ts \
  /tmp/clonopher-ui-backup/
unzip -o ~/Downloads/clonopher-ui-refresh.zip -d ~/GoPrj/Clonopher
cd web
pnpm typecheck
pnpm build
pnpm dev
```

No `pnpm install` is needed. The patch does not include `package.json`, `pnpm-lock.yaml`, `.env`, or `.env.local`, so installed dependencies and your backend port setting are preserved.

## What changes

- Removes the sidebar, account/avatar, connection labels, decorative status cards, and page footer.
- Keeps the overview focused on one wide Payload activity list, with useful search/filter controls.
- Gives Stage Canvas the available viewport; the selected global stage is chosen from a single compact dropdown.
- Opens Stage details as an overlay after clicking a node, leaving the map behind it.
- Reworks Live logs into a console-style overlay with compact severity filters and search.
- Simplifies the payload detail page to useful context, action history, and logs without metric-card grids.
- Uses carbon/gray surfaces and muted matte indigo, blue, green, and red accents.
- Makes Vite load environment files from the directory containing `vite.config.ts`, which should help keep `VITE_BACKEND_ORIGIN` consistent when launched from different working directories.

## Scope note

This overlay updates the frontend presentation and Vite configuration only. The full Go API handler/routes source was not included in the UI source bundle, so backend handlers were not changed. API routes and response normalization in `src/lib/api.ts` are preserved.
