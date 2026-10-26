import type { PayloadStatus, StageSnapshot } from "./types";

const ago = (minutes: number) =>
  new Date(Date.now() - minutes * 60_000).toISOString();

export const demoPayloads: PayloadStatus[] = [
  {
    id: "pld_8f13a2",
    name: "complex_gin_rate_limiter.md",
    state: "completed",
    iteration: 4,
    lastAction: "ACT_STAY_HERE",
    paths: [],
    stagePath: "Golang / HTTP / Middleware",
    startedAt: ago(14),
    updatedAt: ago(11),
    actions: [
      "ACT_EXPLORE",
      "ACT_CLASSIFY",
      "ACT_SELECT_STAGE",
      "ACT_STAY_HERE",
    ],
    logs: [
      {
        id: "l1",
        type: "info",
        content:
          "Analyzed requirements and identified Gin middleware patterns.",
        timestamp: ago(13),
      },
      {
        id: "l2",
        type: "info",
        content: "Selected stage: Golang / HTTP / Middleware.",
        timestamp: ago(12),
      },
      {
        id: "l3",
        type: "info",
        content: "Payload completed successfully.",
        timestamp: ago(11),
      },
    ],
  },
  {
    id: "pld_7ab4d9",
    name: "pythonDictionaryTutorial.md",
    state: "inferring",
    iteration: 3,
    lastAction: "ACT_INFER_STAGE",
    paths: [],
    stagePath: "Python / Tutorials",
    startedAt: ago(1),
    updatedAt: ago(0.3),
    actions: ["ACT_EXPLORE", "ACT_CLASSIFY", "ACT_INFER_STAGE"],
    logs: [
      {
        id: "l4",
        type: "info",
        content:
          "Matched tutorial structure, examples, and Python-specific terms.",
        timestamp: ago(1),
      },
      {
        id: "l5",
        type: "warn",
        content:
          "Candidate path has overlapping topic labels; checking existing stages.",
        timestamp: ago(0.4),
      },
    ],
  },
  {
    id: "pld_2c90ef",
    name: "tomatoSoupRecipe.md",
    state: "completed",
    iteration: 2,
    lastAction: "ACT_STAY_HERE",
    paths: [],
    stagePath: "Recipes / Soups",
    startedAt: ago(22),
    updatedAt: ago(20),
    actions: ["ACT_EXPLORE", "ACT_CLASSIFY", "ACT_STAY_HERE"],
    logs: [
      {
        id: "l6",
        type: "info",
        content: "Recognized recipe ingredients and preparation steps.",
        timestamp: ago(21),
      },
      {
        id: "l7",
        type: "info",
        content: "Existing Recipes stage is a suitable destination.",
        timestamp: ago(20),
      },
    ],
  },
  {
    id: "pld_4d711e",
    name: "MyVsCodeCustomizations.md",
    state: "failed",
    iteration: 5,
    lastAction: "ACT_INFER_STAGE",
    paths: [],
    stagePath: "Developer tools",
    startedAt: ago(6),
    updatedAt: ago(5),
    actions: [
      "ACT_EXPLORE",
      "ACT_CLASSIFY",
      "ACT_CREATE_STAGE",
      "ACT_RETRY",
      "ACT_INFER_STAGE",
    ],
    logs: [
      {
        id: "l8",
        type: "info",
        content:
          "Identified editor settings, extensions, and keyboard customizations.",
        timestamp: ago(6),
      },
      {
        id: "l9",
        type: "warn",
        content:
          "No existing stage matched all customizations with high confidence.",
        timestamp: ago(5.5),
      },
      {
        id: "l10",
        type: "error",
        content:
          "Stage inference exhausted retry budget without a stable target.",
        timestamp: ago(5),
      },
    ],
  },
  {
    id: "pld_5e2f8a",
    name: "flight-sim-controls.md",
    state: "queued",
    iteration: 0,
    lastAction: "",
    paths: [],
    stagePath: "—",
    startedAt: ago(0.1),
    updatedAt: ago(0.1),
    actions: [],
    logs: [
      {
        id: "l11",
        type: "info",
        content: "Waiting for an available worker.",
        timestamp: ago(0.1),
      },
    ],
  },
  {
    id: "pld_90c6b1",
    name: "sqliteMigrationNotes.md",
    state: "cancelled",
    iteration: 1,
    lastAction: "ACT_CANCELLED",
    paths: [],
    stagePath: "Go / Persistence / SQLite",
    startedAt: ago(38),
    updatedAt: ago(36),
    actions: ["ACT_EXPLORE", "ACT_CLASSIFY", "ACT_CANCELLED"],
    logs: [
      {
        id: "l12",
        type: "warn",
        content: "Processing was cancelled by the operator.",
        timestamp: ago(36),
      },
    ],
  },
];

export const demoStageNames = [
  "Root",
  "Golang",
  "Python",
  "Recipes",
  "Developer tools",
];

export const demoStageTree: StageSnapshot = {
  path: "/home/alireza/Documents/My_TESTS/Clonopher2",
  existingFiles: ["README.md", "inbox.md", "folder-rules.md"],
  info: {
    stage_id: "stage_root_01",
    path: "/home/alireza/Documents/My_TESTS/Clonopher2",
    version: 3,
    brief:
      "Top-level destination for AI-assisted file organization across personal projects, development notes, and references.",
    creator: { provider: "ollama", model: "qwen3.5:4b" },
    modifier: { provider: "ollama", model: "qwen3.5:4b" },
  },
  substages: {
    Golang: {
      path: "/home/alireza/Documents/My_TESTS/Clonopher2/Golang",
      existingFiles: ["complex_gin_rate_limiter.md", "go-concurrency.md"],
      info: {
        stage_id: "stage_go_02",
        path: "/Golang",
        version: 2,
        brief:
          "Go language, backend engineering, concurrency, HTTP services, and tooling.",
      },
      substages: {
        HTTP: {
          path: "/home/alireza/Documents/My_TESTS/Clonopher2/Golang/HTTP",
          existingFiles: ["gin-routing.md"],
          info: {
            stage_id: "stage_http_03",
            version: 1,
            brief: "HTTP servers, routers, middleware, and API behavior.",
          },
          substages: {
            Middleware: {
              path: "/home/alireza/Documents/My_TESTS/Clonopher2/Golang/HTTP/Middleware",
              existingFiles: ["rate-limiter.md"],
              info: {
                stage_id: "stage_middleware_04",
                version: 1,
                brief:
                  "Reusable middleware patterns and request processing policies.",
              },
              substages: {},
            },
          },
        },
        Persistence: {
          path: "/home/alireza/Documents/My_TESTS/Clonopher2/Golang/Persistence",
          existingFiles: ["sqlite-migrations.md", "pgx-notes.md"],
          info: {
            stage_id: "stage_persist_05",
            version: 1,
            brief:
              "Database drivers, schema changes, and persistence patterns.",
          },
          substages: {},
        },
      },
    },
    Python: {
      path: "/home/alireza/Documents/My_TESTS/Clonopher2/Python",
      existingFiles: ["pythonDictionaryTutorial.md"],
      info: {
        stage_id: "stage_py_06",
        version: 2,
        brief: "Python language notes and legacy scripting references.",
      },
      substages: {
        Tutorials: {
          path: "/home/alireza/Documents/My_TESTS/Clonopher2/Python/Tutorials",
          existingFiles: ["pythonDictionaryTutorial.md"],
          info: {
            stage_id: "stage_python_tutorials_07",
            version: 1,
            brief: "Guides and worked examples organized by language concept.",
          },
          substages: {},
        },
      },
    },
    Recipes: {
      path: "/home/alireza/Documents/My_TESTS/Clonopher2/Recipes",
      existingFiles: ["tomatoSoupRecipe.md", "lentilSoup.md"],
      info: {
        stage_id: "stage_recipe_08",
        version: 4,
        brief: "Cooking notes, ingredients, preparation steps, and meal ideas.",
      },
      substages: {
        Soups: {
          path: "/home/alireza/Documents/My_TESTS/Clonopher2/Recipes/Soups",
          existingFiles: ["tomatoSoupRecipe.md"],
          info: {
            stage_id: "stage_soup_09",
            version: 1,
            brief: "Soup recipes and variations.",
          },
          substages: {},
        },
      },
    },
    "Developer tools": {
      path: "/home/alireza/Documents/My_TESTS/Clonopher2/Developer tools",
      existingFiles: ["MyVsCodeCustomizations.md"],
      info: {
        stage_id: "stage_tools_10",
        version: 1,
        brief:
          "Editor setup, extensions, and developer workflow customization.",
      },
      substages: {},
    },
  },
};
