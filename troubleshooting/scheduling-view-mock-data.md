# Scheduling View Uses Mock Data

## What is the problem?
- The GPU Scheduling view always shows mock data and never connects to a real cluster.
- Snapshot/Delta events from the backend are ignored because aggregation is never started.

## What is the root cause?
- `DEBUG_MODE` is hard-coded to `true` in the scheduling view.
- `clusterId` is forced to `"mock-cluster"` and all aggregation actions are skipped.
- As a result, the store never subscribes to real cluster updates.

## What is the solution?
- Replace the hard-coded flag with an environment or build-time feature flag.
- Bind `clusterId` to the active cluster selection store.
- Ensure `startAggregation`, `stopAggregation`, and `refreshSnapshot` run when a real cluster is active.
- Add a smoke test or manual checklist to confirm live data flows in the scheduling view.
