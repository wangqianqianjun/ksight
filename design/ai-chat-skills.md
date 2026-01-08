# AI Chat Side Panel + Skills Design

This doc defines the **AI Chat** right side panel and a **Skills** abstraction that wraps Wails/`window.k` capabilities so users can instruct the assistant to analyze Pod/Node status, run operations, and perform data processing safely.

## Goals

- Provide a VSCode-style **right side panel** for chat + histories.
- Make Kubernetes actions **tool-driven** via a typed, permissioned **Skill** layer.
- Reuse existing capabilities: cluster connections, informer events/caches, scheduler aggregation, and `window.k` SDK.
- Support **auditability**: every skill execution has a stable `executionId` and saved input/output logs.
- Keep it extensible: core + plugins can register skills.

## Non-Goals (for initial iterations)

- Shipping a specific LLM provider by default (design for pluggable providers).
- Fully automated write operations without user confirmation.
- Replacing the existing `frontend/src/lib/k8s-sdk.ts` surface area in one go.

## UX: Chat Side Panel

### Layout

- Right panel (collapsible, resizable, width persisted).
- Left area stays primary for resource views.

### Core UI elements

- **Conversation list** (local history): rename, delete, pin.
- **Message timeline**: user ↔ assistant.
- **Tool execution blocks** embedded in chat:
  - Skill name + args summary
  - Status: running/success/failed/canceled
  - Result preview + “open in table / open in diff / copy JSON”
- **Confirmation dialog** for skills requiring write/danger permissions.
- **Quick actions** (optional): `/pods`, `/nodes`, `/analyze pods`, `/analyze nodes` as deterministic routes before LLM integration.

### Persisted state

- `chat.panel.open`, `chat.panel.width`
- conversation metadata + message history
- skill execution history (linked by `executionId`)

## Architecture

```
Chat UI
  -> (Router: slash-commands or LLM)
    -> SkillRunner
      -> SkillRegistry (core + plugins)
        -> window.k / Wails bound methods
      -> ExecutionLogStore (~/.ksight)
  <- structured results + streaming updates
```

### Components (frontend)

- `ChatPanel`: container (sessions + timeline + input).
- `ChatStore`: messages, sessions, UI state.
- `SkillRegistry`: owns `Skill` definitions.
- `SkillRunner`: validates args, permission checks, runs skill, streams results.
- `ExecutionLogStore`: writes/reads execution logs and indexes.

### Components (backend / Wails)

- Existing: cluster mgmt, informer events, scheduler aggregation.
- New (later milestones): streaming exec/logs/cp, apply YAML, diff endpoints, and operation history persistence.

## Skill Model

### Skill definition

Each skill wraps one “unit of capability” and declares:

- `id`: stable identifier (`k8s.listPods`, `analysis.podHealthSummary`).
- `name`, `description`: for UI and LLM prompt/tool schema.
- `argsSchema`: Zod schema (frontend) with JSON-serializable args.
- `permission`:
  - `read` (default, no confirm)
  - `write` (confirm required)
  - `danger` (strong confirm + optional dry-run)
- `run(ctx, args)`: returns a `SkillResult` (structured) and optionally streams progress.

### Context passed to skills

- `clusterId` (required unless explicitly “offline skill”)
- access to `window.k` SDK (typed wrapper)
- access to cached event store (in-memory state fed by `resource:event`)
- cancellation token
- logger (`executionId`, duration, error)

### Skill results (structured)

- `type`:
  - `table` (rows + columns)
  - `json` (arbitrary object)
  - `markdown` (renderable summary)
  - `diff` (left/right payloads)
  - `stream` (incremental chunks)
- `artifacts`: optional links to open in other UI surfaces (drawer, command tab, resource page).

## Security & Safety

- Default mode is **read-only**: listing, summarizing, and analysis.
- Any `write/danger` skill requires user confirmation with:
  - target cluster + namespace + resource identifiers
  - mutation summary (patch/apply/delete)
  - optional dry-run preview
- Redaction:
  - Chat should not display secret data by default (reuse backend sensitive redaction rules).
  - Provide explicit “show sensitive” toggle where supported, and log that it was enabled.
- Audit logs:
  - Every execution writes args + result metadata + timestamps + errors to `~/.ksight`.

## Built-in Skills (MVP)

Read-only skills to support “pod/node 状态分析、操作和数据处理”的第一步：

- `k8s.listPods({ namespace?, labelSelector?, fieldSelector? })`
- `k8s.listNodes()`
- `analysis.podHealthSummary({ namespace?, groupBy? })`
  - counts by phase, restarts (if available), common reasons, top offenders
- `analysis.nodePressureSummary()`
  - Ready/NotReady, Unschedulable, pressure conditions

Planned (later):

- `k8s.getResource({ gvr, namespace, name, showSensitive? })`
- `k8s.applyYaml({ yaml, namespace? })` (write)
- `pod.logs({ namespace, pod, container?, since?, tail? })` (stream)
- `pod.exec({ namespace, pod, container?, command })` (danger)
- `pod.files.download/upload(...)` (danger)

## Integration with `window.k` / Wails

- `window.k` remains the canonical frontend SDK surface.
- Skills call `window.k` methods; the SDK handles:
  - choosing Wails adapter vs HTTP adapter
  - standardized errors (`isNotFound/isConflict/...`)
  - event subscription for cache updates

For streaming skills (logs/exec), prefer:

- Wails events for chunk delivery (e.g. `skill:stream:{executionId}`), or
- WebSocket channel (already exists) for stream messages, depending on which path becomes the primary runtime.

## Extensibility

- Plugins can register skills via a `registerSkills(registry)` hook.
- Skills become discoverable by:
  - Command Palette (Cmd+T) entries (manual invocation)
  - Chat tool selection (LLM tool schema or slash-commands)

## Milestones

- **M1**: ChatPanel shell (toggle/resize/persist) + message history storage.
- **M1.5**: SkillRegistry/Runner + 4 read-only MVP skills + execution logging.
- **M2+**: Connect skills to informer-driven in-memory state and richer resource browser.
- **M4+**: Write/streaming skills (apply/logs/exec/files) behind confirmation gates.

## Open Questions

- Confirmation UX: single global modal vs per-execution inline confirmation card?
- Primary runtime: rely on Wails events only, or standardize on WebSocket for tool streaming?
- Where to store skill execution logs/index: JSONL + index vs SQLite?

