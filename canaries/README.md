# Media Stack canaries

[Documentation map](../docs/README.md) · [Design decision](../docs/design/canaries.md)

A canary is a temporary migration/test environment with a defined purpose and
acceptance contract. Luis alone owns activation/deactivation, Portainer settings,
promotion, cutover and rollback. Nothing here activates automatically or belongs
to a production Compose aggregate.

| Scenario | Status | Evidence |
| --- | --- | --- |
| [Jellyfin](jellyfin/README.md) | **INACTIVE**, not currently deployed; prior migration complete | [Migration log](jellyfin/MIGRATIONS.md) |

Use stable service/concern directory names. Record successive baselines, targets,
results and retirement in MIGRATIONS.md instead of creating jellyfin13/jellyfin14
folders. Each independent scenario owns its Compose and preparation contract;
there is no monolithic canary stack or permanent duplicate production owner.

Before concurrent testing, isolate writable configuration/database/cache/transcode/
metadata; use unique names, ports and addresses. Shared media should be read-only
where compatible, with host/GPU contention accounted for. Existing inactive files
are not necessarily safe to run. Luis provisions trial inputs and secrets.

Promotion requires accepted tests, deliberate state handover and one writer.
Rollback requires a recoverable baseline and compatible state/image, not just a
Git revert. After success/abort, deactivate, record the outcome and decide data
retention before retirement. No runtime activation or state deletion occurs here.
