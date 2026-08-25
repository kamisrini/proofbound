# Sessions connector specification

The sessions connector observes Claude session metadata without ingesting message content.

## Interface

`New(Deps)` constructs a connector. `Sync(ctx, Appender)` scans the repository's encoded Claude
project directory, skips live files (mtime no older than ten minutes), parses JSONL metadata, and
appends one `session.observed` event per session file. A missing source directory is an empty,
successful sync.

## Payload

Each event payload is canonical JSON with exactly these fields:

`session_id`, `started_at`, `finished_at`, `message_count`, `tool_call_count`, `files_written_count`,
`parse_coverage`.

Timestamps may be null when unavailable. `parse_coverage` is the ratio of valid JSONL lines to all
non-empty lines. Files below 50% coverage emit no event and are counted as skipped.

## Invariants

| INV-1 | Source directory is derived from the absolute repository root by replacing `/` and `.` with `-` | sessions_test.go::TestProjectDir |
| INV-2 | Files newer than ten minutes are never ingested | sessions_test.go::TestSyncSkipsLiveFiles |
| INV-3 | Malformed lines are skipped and reduce parse coverage without aborting the file | sessions_test.go::TestSyncRecordsParseCoverage |
| INV-4 | A file below 50% parse coverage emits no event | sessions_test.go::TestSyncDropsLowCoverage |
| INV-5 | Session payload contains metadata only and no source-line content | sessions_test.go::TestSyncRecordsParseCoverage |
| INV-6 | Repeating a sync is idempotent through the ledger appender | sessions_test.go::TestSyncDeduplicates |

## Non-goals

- Never read or persist real session content in tests or event payloads.
- Never fail a whole sync because one line is malformed.
- Never ingest a live session file.
