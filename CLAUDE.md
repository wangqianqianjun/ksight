# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Development
make dev                    # Run with hot-reload (Wails dev mode)
wails build                 # Production build

# Frontend (run from frontend/)
pnpm install                # Install dependencies
pnpm run lint:fix           # Lint and auto-fix
pnpm run type-check         # TypeScript validation

# Backend
make fmt                    # Format Go code (go fmt + goimports)
make lint                   # Run golangci-lint
```

## Testing

```bash
# All tests
make test                   # Unit + integration tests

# Specific tests
make test-unit              # Unit tests (./pkg/informer, ./pkg/service)
make test-integration       # Integration tests with real K8s API (envtest)
make test-informer          # InformerManager tests only
make test-service           # ClusterService tests only
make test-watch             # Watch mode for TDD

# Single test pattern (Ginkgo)
KUBEBUILDER_ASSETS=$(pwd)/bin/k8s go run github.com/onsi/ginkgo/v2/ginkgo -v --focus="test name pattern" ./pkg/test

# Coverage
make test-coverage          # Generate HTML coverage report
```

**Testing framework**: Ginkgo/Gomega with envtest (real K8s API server)

## Development Workflow Requirements

1. **Unit Tests Required**: Add unit tests for new/modified modules
2. **Web App Testing**: After development, test the running application (`make dev`) to verify functionality
3. **Fix Issues**: If problems are found during testing, fix them before considering the task complete

## Architecture

### Backend (Go + Wails)

```
app.go                      # Wails-exposed methods (frontend bindings)
pkg/
├── informer/manager.go     # InformerManager - K8s resource watching, dynamic informers
├── service/
│   ├── cluster.go          # ClusterService - cluster management, kubeconfig, events
│   └── scheduler/          # GPU scheduler aggregation (types, collector, aggregator, views)
├── client/                 # Kubernetes client utilities
└── test/                   # Integration tests with envtest
```

**Event Flow**: K8s API → InformerManager → ResourceEvent → ClusterService → Wails EventEmitter → Frontend

### Frontend (Vue 3 + TypeScript)

```
frontend/src/
├── core/
│   ├── layout/             # MainLayout, Sidebar
│   └── router/             # Vue Router with dynamic plugin loading
├── plugins/                # Feature plugins (each is independent route+component+store)
│   ├── applications/       # Pod listing by app label
│   ├── nodes/              # Node management
│   ├── resources/          # Resource browser
│   ├── templates/          # K8s YAML templates
│   ├── boards/             # Custom dashboards
│   ├── operations/         # Troubleshooting operations
│   └── scheduling/         # GPU scheduler visualization
└── shared/
    ├── components/         # shadcn-vue components
    ├── composables/        # Vue 3 composables
    ├── stores/             # Pinia state management
    └── utils/              # Helper functions
```

**Frontend SDK**: `window.k` with strong typing for K8s operations (k.pods.list().first().exec())

### Key Patterns

- **Layered Backend**: Wails bindings → Service layer → Informer layer → K8s client
- **Plugin Architecture**: Each plugin is self-contained with independent routes and state
- **Event-Driven UI**: Frontend subscribes to Wails events emitted from backend
- **SQLite Cache**: Local persistence for 100K+ resources with resourceVersion tracking

## Design Memory

### Core Concept
User-centered Kubernetes GUI focused on real-world operations, not resource types.

### Layout Structure
- **Top**: Chrome-style tabs for cluster connections + settings gear
- **Left**: VSCode-style icon sidebar (Applications, Operations, Nodes, Resources, Templates, Boards)
- **Main**: DataTable with filters, groupBy, views system
- **Right**: AI chatbox with k.operation() calls
- **Footer**: Ad-hoc command tabs (Add resource, K8S scripts, Pod shell/logs/files, Resource diff)

### Tech Stack
- **Backend**: Golang dynamic informers + Wails v2.11.0
- **Frontend**: Shadcn-Vue + Pinia + TailwindCSS + Vite
- **Cache**: SQLite with resourceVersion for consistency
- **Extensions**: Dynamic component loading, VSCode-style extension system

### Key Behaviors
- Views system: saved filters/groupBy/orderBy per menu
- Topology dialog: kubectl tree-like structure for any resource
- YAML editor: 3-part view (metadata/spec/status) with quick inputs
- Keyboard shortcuts: Cmd+T palette, S for shell, L for logs
- Resource diff: Monaco diff viewer with one-click merge
