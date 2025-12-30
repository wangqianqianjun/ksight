# KSight UX System and Interaction Spec

## Scope
This document defines the global UX system, interaction patterns, and visual style for KSight. It is the single source of truth for shared UX behavior across modules. Module specific behavior lives in dedicated files listed at the end.

## Experience Principles
- Operator first: minimize clicks for common operational tasks.
- State clarity: users always know what is connected, loading, and actionable.
- Density with control: compact by default, with easy toggles for comfort.
- Progressive disclosure: details appear in drawers and panels, not in the main list.
- Consistent command surfaces: command palette and command tabs behave the same across modules.

## Global Layout
### Title Bar
- Cluster tabs for active and pinned clusters, plus "Add" to connect.
- Right side: settings and cluster status indicator.
- Status indicator shows connected, connecting, error, or disconnected.

### Left Sidebar
- Narrow icon rail: Applications, Operations, Nodes, Scheduling, Resources, Templates, Boards.
- Active module highlighted; supports drag to reorder.
- Tooltips show module name and last update time.

### Main Area
- Always shows a real state or data view. No developer placeholder text.
- View header contains title, view tabs, and primary actions.

### Right Panel
- AI chat panel, collapsible and resizable.
- Contextual prompts based on module and selection.

### Footer Command Tabs
- Collapsible tabs: apply resource, scripts, logs, shell, files, diff.
- Tabs can be added with a plus button and reordered.

## Visual System
### Typography
- Primary: IBM Plex Sans
- Secondary: IBM Plex Mono (code, logs, YAML)
- Sizes: 12, 13, 14, 16, 20, 24
- Weights: 400, 500, 600
- Line height: 1.3 for body, 1.2 for headings

### Color Tokens
Neutral scale:
- N0 #FFFFFF
- N50 #F7F8FA
- N100 #EEF1F4
- N200 #E1E6EB
- N300 #CBD2D9
- N400 #9FA8B3
- N500 #6B7280
- N600 #4B5563
- N700 #374151
- N800 #1F2937
- N900 #0B0F14

Semantic colors:
- Primary #2F6FED
- Primary hover #1F5BD7
- Primary active #1B4EC0
- Success #18A06B
- Warning #D87A16
- Error #D64A3A
- Info #2F6FED

Surfaces:
- App background: N50
- Surface 1: N0
- Surface 2: N100
- Borders: N200

### Spacing and Sizing
- Space scale: 4, 8, 12, 16, 24, 32, 40, 48
- Radius: 4, 6, 8, 12
- Border width: 1

### Density Modes
- Compact (default): table row height 28, input height 28, button height 28
- Comfort: table row height 36, input height 32, button height 32

### Motion
- Durations: 120ms (fast), 180ms (base), 240ms (slow)
- Easing: cubic-bezier(0.2, 0, 0, 1)
- Use fade + slide for drawers and modals. Avoid long spinners.

### Iconography
- Use lucide icons only.
- Keep icons 16 or 20 px.

## Global Interaction Patterns
### Data Table Framework
- Filters on the left, group by and column controls on the right.
- Group rows are collapsible; group counts visible.
- Selection supports row, group, and page selection.
- Actions appear as icon buttons on hover with tooltips.

### Detail Drawer
- Right side drawer with adjustable width.
- Top action bar and tabs for sections.
- Drawer can be pinned while switching selection.

### Modals and Confirmations
- Destructive actions require confirmation with explicit wording.
- Multi step actions show progress and allow cancel.

### Notifications
- Toasts for success, warning, error.
- Long running actions show inline status near the source action.

### Command Palette
- Open with Cmd+T.
- Supports fuzzy search and module scoped actions.
- Recent commands are listed first.

## Global State Copy Table
Use the following copy as defaults. Modules may override nouns.

| State | Title | Body | Primary | Secondary |
| --- | --- | --- | --- | --- |
| Disconnected | No cluster connected | Connect to a cluster to view data. | Connect to cluster | Open settings |
| Connecting | Connecting to <cluster> | Validating kubeconfig and starting watchers. | Cancel | Open logs |
| Loading | Loading data | This may take a few seconds. | None | None |
| Empty | No data found | This cluster has no matching resources. | Create or deploy | Switch namespace |
| Filtered empty | No results match your filters | Adjust or clear filters to continue. | Clear filters | Edit filters |
| Error | Something went wrong | We could not load data. | Retry | Open logs |
| Permission | Access denied | Your account lacks permission to list this resource. | Retry | Open settings |
| Stale | Data out of date | Last updated <time>. | Refresh | None |

## Accessibility
- Keyboard navigation for tables and drawers.
- Visible focus ring for interactive elements.
- High contrast mode supported via settings.

## Data Freshness
- Each module shows last sync time when data is older than 30s.
- Deltas are applied in order; if a sequence gap is detected, request a snapshot.

## Module UX Files
- Applications: ux/applications.md
- Nodes: ux/nodes.md
- Scheduling: ux/scheduling.md
- Resources: ux/resources.md
- Templates: ux/templates.md
- Operations: ux/operations.md
- Boards: ux/boards.md
- Settings: ux/settings.md
- AI Chat Panel: ux/chat.md
- Command Tabs: ux/command-tabs.md
