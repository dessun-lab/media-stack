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

**Operational status:** cutover complete per Luis (2026-10-11); Bazarr deployed
and functioning per Luis (2026-10-15). Portainer variables and the external
network are populated/provisioned. Names above remain repository recommendations,
not a live export of Portainer settings. Luis owns Portainer and network operations.

Bazarr uses `BAZARR_HOST`, `PUID`, `PGID` and `TZ`. Gluetun publishes its UI on
port `6767` and owns the shared `172.18.0.10` address. Bazarr API/provider
credentials reside in its persistent UI configuration, outside Git. See the
[stack README](README.md) and [Bazarr operations](../../docs/operations/bazarr.md).
