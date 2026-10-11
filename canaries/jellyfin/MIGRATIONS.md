# Jellyfin migrations

[Scenario preparation and status](README.md)

## Completed: operator-described Jellyfin 11 → Jellyfin 12

**Outcome:** Luis reports successful validation and promotion; the former migration
environment is disabled and not currently deployed. The exact shutdown date and
individual test results are not recorded in inspected Git history.

Preserve two kinds of evidence: Luis calls the source/target “Jellyfin 11/12”; the
literal prior production image was `linuxserver/jellyfin:10.11.11`, and the target
was `lscr.io/linuxserver/jellyfin:12.1ubu2604-ls50`. Do not infer different upstream
major-version numbering from the operator shorthand or silently rewrite tags.

| Repository date (commit offset -06:00) | Commit | Supported event |
| --- | --- | --- |
| 2026-09-17 | `27af7ed` | Added canary `jellyfin12` using the target image, separate `runtime/jellyfin12` writable state, port 8097 and read-only media. |
| 2026-09-18 | `89a6016` | Commit records addition of plugins to the canary. |
| 2026-09-19 00:17 | `a0984b8` | Changed canary writable paths to `runtime/jellysun`; explains present overlap with production. |
| 2026-09-19 00:38 | `4c398c7` | “using jellyfin12 in prd”: production adopted target image and `runtime/jellysun` paths; prior block retained as comments. |
| 2026-10-03 | `20c4e96` | Production media mount changed back to writable; retained canary media stayed read-only. |
| 2026-10-11 | This restructuring | Moved retained definition byte-for-byte into `canaries/jellyfin/compose.yml`; stable ownership and history added, no activation. |

Git supports definition changes and promotion intent; Luis supplies the successful
operational outcome and inactive status. This record does not claim a tested
rollback procedure or currently isolated canary.

## Future entries

Append a section per migration with baseline/target literal image identities,
source/config revision, purpose, isolation contract, Luis's validation outcome,
promotion/rollback decision, completion and retirement/data disposition. Reuse
this stable scenario after closing the prior trial; do not add version directories.
