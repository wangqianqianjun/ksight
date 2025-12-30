# Command Tabs UX

## Purpose
Provide a persistent workspace for common operational actions without leaving the current view.

## Tabs
- Apply resource: YAML editor with Apply and Cancel.
- Scripts: TypeScript editor with Run and output console.
- Pod shell: terminal view with container selector.
- Pod logs: log view with filters and container switch.
- Pod files: file tree on the left, editor on the right.
- Diff: side by side diff editor with merge options.

## Primary Interactions
- Tabs can be opened from context menus and pinned.
- Tabs can be minimized to a compact bar.
- Running actions show progress and allow cancel.

## Error Handling
- Failed runs show inline errors with link to logs.
- File transfers show progress and retry.

## Success Criteria
- Users can open a shell or logs without leaving the current page.
- Tabs remain stable across navigation within the same cluster.
