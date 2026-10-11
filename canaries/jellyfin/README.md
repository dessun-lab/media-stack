# Reusable Jellyfin migration canary

**Status: INACTIVE — not currently deployed.** Luis confirms the prior Jellyfin
migration completed and the former canary was disabled. Production already uses
the migrated current image, `lscr.io/linuxserver/jellyfin:12.1ubu2604-ls50`.

[Compose](compose.yml) was moved unchanged from `jellyfin12/docker-compose.yml`.
[MIGRATIONS.md](MIGRATIONS.md) owns historical evidence; [canary rules](../README.md)
own the common lifecycle contract. The historical service/container name remains
`jellyfin12` to preserve semantics; the stable directory can support future trials.

## Prepare a future migration

For a future target such as Jellyfin 13, Luis must explicitly select the baseline,
target image, validation/acceptance criteria, promotion and rollback plan. Review
names/ports/addresses, GPU/plugins and trial inputs before activating the scenario.

**This retained definition is NOT currently safe for concurrent production use.**
Writable config/cache/transcode/metadata point to production `runtime/jellysun`.
Read-only media alone does not provide isolation. Prepare independent writable
state and a recoverable baseline before any concurrent trial. The current port
8097, canary network `172.19.0.0/16` and address `172.19.0.2` also require Luis's
conflict check; they were preserved, not runtime-validated. Restart policy `no`
is not an activation guard and does not itself prove inactive status.

Compose references PUID, PGID, TZ, JELLYFIN_LOCAL_HOST, GID_VIDEO and GID_RENDER.
Luis supplies the trial configuration; never read the protected root `.env`.
No isolation redesign, image upgrade, activation or runtime change accompanied
this move. An existing ignored local `sso-new-redacted.xml` was moved unchanged
with Luis's explicit permission; its contents were not inspected and it remains
ignored. It is not a versioned or validated migration input.
