# Scheduling Module UX

## Purpose
Expose scheduler health, pending queues, DRA allocation, and GPU migration status. The view must highlight actionable issues and link to pods and nodes.

## Layout
- Header: "Scheduling" title and view tabs.
- Summary row: pending pods, failed schedules, DRA claims, conflicts.
- Main area organized as tabs or stacked sections:
  1) Scheduler health and pending queue
  2) DRA allocation view
  3) Migration and conflicts

## Scheduler Health and Pending Queue
- Pending list grouped by scheduler name.
- Filters: scheduler, namespace, reason, pool.
- Each row shows namespace, pod, scheduler, reason, age, and GPU request.
- Clicking a row opens the pod detail drawer.

Empty states:
- No pending pods: show "Scheduler healthy" and last update time.

## DRA Allocation View
- Visual chain: ResourceClaim -> ResourceSlice -> Device.
- Summary counters: unallocated, allocating, allocated.
- Each claim row shows namespace, name, resource class, driver, and allocation detail.
- Clicking a claim opens a side panel with related pods and devices.

## Migration and Conflicts
- UsedBy distribution: tensor-fusion vs nvidia-device-plugin.
- Conflict table shows node, pod, usedBy, and resource type.
- Conflict rows link to pod or node detail drawers.
- A warning banner appears if conflicts are detected.

## Visual Treatment
- Use color chips for scheduler names.
- Use warning color for failed scheduling reasons.
- Use a subtle bar chart for usedBy distribution.

## Primary Interactions
- Filter and group in pending queue.
- Jump from warning to detail.
- Open claim and device details from the DRA chain.

## Error and Degraded States
- If DRA resources are unavailable, show "DRA not available" with details and a link to settings.
- If events are unavailable, show pending pods without reason and a notice banner.

## Success Criteria
- Users can identify why pods are pending in under 10 seconds.
- Conflicts are visible without scrolling.
- DRA allocation state is understandable without reading docs.
