# Applications Module UX

## Purpose
Show pods grouped by application labels, with fast access to detail, logs, exec, and YAML edits. The view should feel like an operations console, not a developer note.

## Layout
- Header: "Applications" title, view tabs, view switcher, and primary actions.
- Summary row: compact pills for Apps, Pods, Namespaces, Unhealthy (only if > 0).
- Control row: filters on the left, group by and column controls on the right.
- Data table: grouped by app label and workload type by default.

## Primary Interactions
- Group by: app labels, env labels, workload type, node.
- Filters: app name, namespaces, resource types, labels, images, nodes.
- Row actions: detail, edit YAML, exec, logs, port forward.
- Group actions: expand or collapse, apply actions to group.
- Selection: row, group, and page level selection.

## Data Table
Key columns:
- Pod
- Workload
- Node
- Image
- Status
- Age
- Ready
- Containers

Behavior:
- Group headers show pod count and readiness summary.
- Large groups use virtualized lists.
- Clicking a row opens the detail drawer.

## Detail Drawer
Tabs:
- Compute: containers, image, command, resources.
- Configuration: configmaps, secrets, env.
- Networking: services, ports, port forward.
- Storage: volumes and claims.
- Events: recent events and restarts.

Top area:
- Owner chain with clickable workload references.
- Primary actions: edit YAML, logs, exec.

## Empty and Connection States
Use global state copy from ux/application-ux.md with module specific nouns.

Module specific copy:
- No applications found
  - Title: No applications found
  - Body: This cluster has no running workloads yet.
  - Primary: Create workload
  - Secondary: Open templates

- No results match filters
  - Title: No results match your filters
  - Body: Try adjusting filters or clear them.
  - Primary: Clear filters
  - Secondary: Edit filters

## Command Tabs Integration
From this module, the command tabs should open with context:
- Logs: selected pod and container.
- Shell: selected pod, default container.
- Diff: current YAML vs template.

## Success Criteria
- Users can find a workload in under 10 seconds.
- Users can open logs or exec within two clicks.
- Empty states guide the next action clearly.
