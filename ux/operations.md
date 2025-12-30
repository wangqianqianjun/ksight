# Operations Module UX

## Purpose
Run operational scripts with parameters, track history, and review output.

## Layout
- Left sidebar: operations list.
- Main area: overview, parameters, run history, and output.

## Primary Interactions
- Select an operation to view details and run history.
- Run action opens parameter editor and target cluster selector.
- Output appears in a split view: editor on the left, terminal on the right.

## Run Flow
- Step 1: Confirm parameters.
- Step 2: Select target cluster(s).
- Step 3: Run with live output.
- Step 4: Result summary with link to logs.

## Error Handling
- Failed runs show a banner and link to detailed logs.
- Partial success shows per cluster status.

## Success Criteria
- Users can run an operation in under 30 seconds.
- Output is readable and searchable.
