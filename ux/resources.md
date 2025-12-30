# Resources Module UX

## Purpose
Explore any Kubernetes resource type with dynamic filters, columns, and actions.

## Layout
- Left sidebar: resource type tree with pinned kinds at top.
- Main area: table view with filters, group by, and column controls.

## Primary Interactions
- Select a kind from the tree to load its table view.
- Filters: name, namespace, and dynamic field filters.
- Actions: dynamic actions configured by settings.
- Columns: from CRD print columns or built-in map.

## Data Table
Behavior:
- Supports grouping when configured for a resource type.
- Actions appear on hover in the row.
- Clicking a row opens the detail drawer.

## Detail Drawer
- Metadata, spec, status sections.
- YAML edit and status edit actions are separated with warnings.

## Empty and Error States
- No resources found: suggest adjusting namespace or filters.
- Permission denied: show access denied state from global copy.

## Success Criteria
- Users can locate a resource in under 15 seconds.
- Actions are discoverable without opening the row.
