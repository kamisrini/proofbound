# VD-p3-github-connector-2026-08-26: GitHub Actions and deployments first

**Status:** Accepted (2026-08-26)
**Date:** 2026-08-26

**Context:** P3 needs a second tenant's real delivery data and a deployed-where / tested-what
view. GitHub is the default candidate in the roadmap and is also the project's remote, making
its public API reproducible without inventing a second integration surface. The repository
`github/docs` was checked on 2026-08-26 and exposed both recent Actions workflow runs and public
deployment records.

**Decision:** The first external connector is a narrow, read-only GitHub connector. It accepts
one organization and an explicit repository allowlist, collects bounded recent workflow runs and
deployments, and records them as immutable, content-addressed ledger events. The initial client
uses the standard library (`net/http` and `encoding/json`); authentication is an injected bearer
token and is never persisted in event payloads or logs. Tests use an injectable API base URL and
HTTP fixtures. `github/docs` is the live smoke-test fixture, not a hard-coded production default.

**Boundary:** Pull requests, issues, comments, users, webhooks, and a generic external-event
framework are out of scope. The connector, projection, and report surface are accepted for this
narrow slice; broader GitHub pagination and additional entities remain future scope.

Fixture-selection evidence is recorded in
`docs/verification/p3-github-fixture-check.md`.

**Revisit when:** the first real-data sync is accepted, the selected organization/repository
allowlist changes, or GitHub API limits require a different transport or pagination policy.
