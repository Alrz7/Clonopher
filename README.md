<div align="center">
  <a href="https://github.com/Alrz7/Marble">
    <img width="2011" height="505" alt="clonopher-banner" src="https://github.com/user-attachments/assets/4a3cc52f-7deb-487c-9991-9b3638857e62" />
  </a>
</div>

**A context-aware file organizer powered by Go and Jev.**

> **Clonopher is a work in progress.** CLI/API behavior, configuration, and internal architecture may change.

Clonopher organizes files using their context, directory structure, and configurable classifiers — not just filenames or extensions. It is under active development.

## Architecture at a glance

```text
             CLI                          HTTP API
              \                            /
               +-------- Payloads --------+
                            |
                     Worker / Serve
                            |
              +-------------+-------------+
              |                           |
        Jev decision core           Directory context
              |                       Root → Directories
      Relevant classifiers          Per-Root Global Map
       Weighted decisions                  |
              +-------------+-------------+
                            |
                     Filesystem action
                            |
                  Continue or complete
```

- **Jev** selects relevant classifiers for each file and combines their weighted decisions.
- **Workers** process payloads through a staged directory-exploration loop.
- Each **Root** has a shared **Global Map**, allowing workers to reuse discovered Directory information.
- `.clonopher.yaml` helper files preserve directory descriptions between runs. Language models assist with description generation and structured decision requests; they are not the whole organizer.

## Project layout

```text
.
├── api/                  # HTTP server, routes, and handlers
├── app/
│   ├── config/           # Application configuration types and setup
│   ├── core/             # Payload processing and decision loop
│   ├── providers/        # Jev and language-model provider definitions
│   ├── appMain.go        # Application bootstrap
│   └── appConfig.go      # Cobra commands and CLI flags
├── internal/
│   ├── fileSystem/       # Filesystem helpers, directory and staging data
│   ├── loggy/            # Logging and error helpers
│   └── ...
└── ...                   # Entry point and other project files
```

## Configuration

Clonopher loads its default configuration from its application config directory. You can provide a different file with `-c` / `--config`.

The configuration uses these top-level sections:

- `app`: roots, default destination, upload directory, environment, run mode, and action policy.
- `api`: HTTP server port.
- `jev_providers` and `llm_providers`: provider maps and the active provider/model for each role.

Example of the application/API settings:

```yaml
app:
  roots:
    Downloads: /home/user/Downloads
    Documents: /home/user/Documents
  default_root: /home/user/Downloads
  default_upload_dir: /home/user/.local/share/clonopher/uploads
  env: development
  action_policy: notify

api:
  port: 6520

jev_providers:
  active_provider: your-jev-provider
  active_model: your-jev-model
  all:
    # Add provider entries using the providers.Provider configuration.

llm_providers:
  active_provider: your-llm-provider
  active_model: your-llm-model
  all:
    # Add provider entries using the providers.Provider configuration.
```

`roots` maps friendly root names to directory paths. `default_root` is a directory path used when `cli` is run without `--dst`. Provider-specific fields are defined by `app/providers`; they are not shown here because the provider type definition is not part of the source snapshot used for this README.

## CLI

The command structure is:

```bash
clonopher [global flags] <command> [command flags]
```

### Global flags

| Flag | Description |
| --- | --- |
| `-c, --config` | Path to a configuration file. |
| `-r, --roots` | Comma-separated root mappings: `Name:/path,Other:/other/path`. |
| `-e, --env` | Application environment. Default: `development`. |

### `cli`

Organize source files from the terminal:

```bash
clonopher -c ./config.yaml cli \
  --src "/home/user/Downloads/report.pdf,/home/user/Downloads/notes.md" \
  --info "quarterly report,project notes" \
  --dst Downloads \
  --policy notify
```

| Flag | Description |
| --- | --- |
| `-s, --src` | Comma-separated source paths. Parentheses can group multiple files: `src1,(src2a,src2b),src3`. |
| `-i, --info` | Comma-separated descriptions matching the source entries. Use `-` to skip a description. |
| `-d, --dst` | Destination root name from `app.roots`, or a directory path. Defaults to `app.default_root`. |
| `-p, --policy` | Action policy: `copy`, `move`, `notify`, `ask-copy`, or `ask-move`. Default: `notify`. |
| `-l, --loglevel` | CLI log level. |

### `api`

Start the HTTP API server:

```bash
clonopher -c ./config.yaml api --port 6520
```

| Flag | Description |
| --- | --- |
| `-P, --port` | API server port. Default: `6520`. |

## HTTP API

The current routes registered by the API server are:

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/` | Health/status response |
| `GET` | `/api/v1/payloads/stats` | Payload monitoring |
| `GET` | `/api/v1/logs` | Global logs |
| `GET` | `/api/v1/stagings` | List staged files |
| `GET` | `/api/v1/directories/list` | List directories |
| `POST` | `/api/v1/directories/query` | Query directory information |
| `POST` | `/api/v1/directories/info` | Retrieve directory info |
| `POST` | `/api/v1/payloads/new` | Process selected staged files |
| `POST` | `/api/v1/payloads/text` | Submit a text payload |
| `POST` | `/api/v1/payloads/file` | Upload a file to staging |
| `DELETE` | `/api/v1/payloads/file` | Delete a staged file |

### Typical API flow

1. Upload files with `POST /api/v1/payloads/file`.
2. Inspect staged files with `GET /api/v1/stagings`.
3. Submit selected staging IDs to `POST /api/v1/payloads/new`.
4. Submit plain text separately through `POST /api/v1/payloads/text`.
5. Monitor processing through the payload stats and logs endpoints.

Uploaded files are stored under `app.default_upload_dir`. Staging metadata is persisted in `StagingProfiles.yaml` inside the application config directory.

## Status

Clonopher is a work in progress. CLI/API behavior, configuration, and internal architecture may change.
