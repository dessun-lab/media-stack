# Canary convention decision

[Design overview](restructuring.md) · [Canary authority](../../canaries/README.md)

Approved 2026-10-11: one reusable, independently activated scenario per stable
service/concern directory. The Jellyfin migration artifact moves into that owner
with unchanged Compose semantics, preparation notes and MIGRATIONS history.

| Option | Decision |
| --- | --- |
| Bare Compose per service | Independent activation, but insufficient without isolation/status/promotion/rollback ownership. |
| One shared canary Compose | Reject: unrelated trials inherit shared activation/redeploy scope and risk a new monolith. |
| Concern-owned scenario with Compose, README and migration log | Selected. Can include tightly coupled dependencies when a trial needs them; versions belong in history, not permanent directory names. |

Current status, preparation constraints and historical evidence now live in
[canaries/jellyfin](../../canaries/jellyfin/README.md). This decision page does not
maintain a competing operational manual. No isolation redesign or activation is
part of the structural move; Luis owns all Portainer operations.
