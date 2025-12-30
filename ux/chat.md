# AI Chat Panel UX

## Purpose
Provide contextual assistance and execute actions based on the current module and selection.

## Layout
- Right side panel with chat history tabs.
- Input area with quick actions and context summary.

## Primary Interactions
- Messages can trigger actions such as open logs, filter table, or run operations.
- Context header shows current cluster, namespace, and selection count.
- History tabs support rename and pin.

## Behavior
- If no context is available, show guidance on how to select resources.
- Long responses collapse with a "Show more" control.

## Success Criteria
- Users can trigger a common action from chat in under 3 steps.
