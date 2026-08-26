# P3 GitHub live acceptance

Date: 2026-08-26. Fixture: public `github/docs`. Code under test: `d9f41c7`.
Go runtime: `go1.26.7 linux/amd64`. The repository worktree was clean; the run used the
disposable PostgreSQL service on `127.0.0.1:55433`.

Using the pushed CLI and a disposable PostgreSQL database:

```text
VERA_GITHUB_OWNER=github VERA_GITHUB_REPOS=docs go run ./cmd/vera sync github
listed=200 appended=200 existing=0
duration_seconds=7
```

The live `vera report github` command returned `github_deliveries count=106`. Its rows contained
real workflow names and conclusions, deployment environments, `freshness=fresh`, and ledger proof
values in the form `event_id/seq`. Unmatched records were rendered explicitly as `tested=missing`
or `deployed=missing`; no relationship was inferred across different commit SHAs.

The database contained 100 `github.workflow_run` events and 100 `github.deployment` events, with
GitHub ledger sequences spanning 6–205. A representative exact rendered row was:

```text
- github/docs commit=077df53b0a1887aa6e941814e384fba00bceae8b tested=missing deployed=preview-env-35829:observed freshness=fresh proof=01M10316XFCA3Z8V09JRGJ5PC9/130
```

The unchanged sync was replayed against the same database:

```text
listed=200 appended=0 existing=200
duration_seconds=4
```

This meets the P3 DoD: a deployed-where / tested-what view ran on real external data, the cold sync
was under ten minutes, and freshness was rendered on the report surface. The connector remains
bounded to the configured repository and its v1 collection limit; broader GitHub pagination and
additional entities are future scope.

After `vera rebuild`, `vera report github` again returned `github_deliveries count=106` with proof
and freshness fields. The projection package's rebuild snapshot integration test also confirms
incremental and rebuilt row sets are identical.
