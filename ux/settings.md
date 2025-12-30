# Settings Module UX

## Purpose
Configure clusters, kubeconfig sources, and global preferences.

## Layout
- Sectioned settings with left navigation.
- Panels: Clusters, Kubeconfig, Appearance, Shortcuts, Plugins.

## Primary Interactions
- Add or remove kubeconfig files.
- Toggle watch on default kubeconfig folder.
- Switch theme (light) and density mode.
- View and reset keyboard shortcuts.

## Error States
- Invalid kubeconfig: show validation error with file path.
- Watcher failure: show retry and open logs.

## Success Criteria
- Users can add a cluster in under 2 minutes.
- Preferences are saved and reflected immediately.
