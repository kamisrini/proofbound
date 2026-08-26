# P3 GitHub fixture check

Date: 2026-08-26. This is fixture-selection evidence, not P3 acceptance.

The public repository `github/docs` was queried read-only through GitHub's REST API:

- `GET /repos/github/docs/actions/runs?per_page=1` returned a non-empty workflow-run result
  (`id: 33012280156`, with head SHA, workflow name, status, conclusion, and timestamps).
- `GET /repos/github/docs/deployments?per_page=1` returned a non-empty deployment result
  (`id: 2236474450`, with environment, SHA, status endpoint, and timestamps).

The project repository `kamisrini/proofbound` was also checked and returned no workflow runs and
no deployments, so it is not the live P3 acceptance fixture. The selected fixture is intentionally
documented here rather than hard-coded as a connector default.
