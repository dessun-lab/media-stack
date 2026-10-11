# Portainer metadata — subtitles

| Field | Value |
| --- | --- |
| **Purpose** | Acquire, place and align subtitles using the dedicated scraper VPN and cross-tool workflow. |
| **Services** | `gluetun_scrapers`, `flaresolverr_subs`, `nensaysubs`, `ffsubsync` |
| **Recommended Portainer stack name** | `media-stack-subtitles` |
| **Repository** | `dessun-lab/media-stack` |
| **Compose file path** | `stacks/subtitles/compose.yml` |
| **Environment contract** | [stacks/subtitles/.env.example](.env.example) |
| **Shared network** | Existing external `arrnet_managed` |
| **Operator** | Luis / USER-owned |

Documentation only. Luis supplies values and owns all Portainer operations,
network preparation, cutover and rollback. See the [stack README](README.md).
