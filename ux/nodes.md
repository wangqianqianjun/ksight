# Nodes Module UX

## Purpose
Provide node health, capacity, and GPU usage at a glance. Support operational actions such as cordon, drain, and node shell.

## Layout
- Header: "Nodes" title and view tabs.
- Summary row: total nodes, ready, not ready, GPU nodes.
- Control row: filters on the left, group by and column controls on the right.
- Data table with resource thumbnails and quick actions.

## Primary Interactions
- Group by: region and zone, machine type, architecture, tenant label.
- Filters: node name, IP, labels, resource or device type.
- Toggle: show or hide resource usage thumbnails (cpu, memory, storage, device).
- Actions: detail, cordon, drain, node shell, metrics, events.
- Simulate schedule: open dialog to test pod placement with custom input.

## Data Table
Key columns:
- Node
- Status
- CPU and memory usage
- GPU usage and usedBy breakdown
- Pod count
- Cost (if configured)

Behavior:
- Resource thumbnails display bars with tooltips.
- GPU summary shows usedBy distribution and device count.
- Clicking a row opens the detail drawer.

## Detail Drawer
Tabs:
- Overview: labels, taints, allocatable.
- Resources: per resource usage and pods by namespace.
- GPU: devices, usedBy, and running apps.
- Events: recent events.

## Empty and Error States
- Disconnected: use global copy.
- No nodes found: suggest checking namespace or cluster connectivity.

## Success Criteria
- Users can identify a hot node within 5 seconds.
- GPU usage is visible without opening details.
- Node actions are reachable within two clicks.
