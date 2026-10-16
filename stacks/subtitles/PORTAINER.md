# Portainer metadata — subtitles

| Field | Value |
| --- | --- |
| **Purpose** | Bazarr library subtitles and the existing scraper VPN / manual subtitle workflow. |
| **Services** | `bazarr`, `gluetun_scrapers`, `flaresolverr_subs`, `nensaysubs`, `ffsubsync` |
| **Recommended Portainer stack name** | `media-stack-subtitles` |
| **Repository** | `dessun-lab/media-stack` |
| **Compose file path** | `stacks/subtitles/compose.yml` |
| **Environment contract** | [stacks/subtitles/.env.example](.env.example) |
| **Shared network** | Existing external `arrnet_managed` |
| **Network lifecycle** | Externally managed shared host infrastructure / Luis (USER-owned) |
| **Network prerequisite** | Must exist before deployment; [provisioning](../../docs/operations/networking.md) |
| **Operator** | Luis / USER-owned |

**Operational status:** cutover complete per Luis (update recorded 2026-10-11);
Portainer variables and the external network are populated/provisioned. Names
above remain repository recommendations, not a live export of Portainer settings.
Luis owns all subsequent Portainer operations, network lifecycle and rollback.
See the [stack README](README.md).

**Bazarr addition (2026-10-15):** pending deployment/setup, separate from completed
cutover. New inputs for this stack: `BAZARR_HOST`, `PUID`, `PGID`; existing `TZ`
also configures Bazarr. Luis supplies a specific reachable host interface address
and the media-owner UID/GID. Port `6767` is now published on `gluetun_scrapers`;
Bazarr shares its existing `172.18.0.10` address. Verify port availability and the
new config directory before deployment. Apply the Gluetun port/firewall changes
and recreate Bazarr with its VPN owner as necessary; this affects the shared
scraper namespace and requires Luis's deployment window. API/provider credentials
belong in Bazarr's persistent UI configuration, not Git or the blank template. See the
[Bazarr handoff and acceptance checks](../../docs/operations/bazarr.md).
