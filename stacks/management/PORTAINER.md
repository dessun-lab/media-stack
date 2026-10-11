# Portainer metadata — management

| Field | Value |
| --- | --- |
| **Purpose** | Coordinate media libraries and requests independently of serving and VPN maintenance. |
| **Services** | `radarr`, `sonarr`, `lidarr`, `jellyseerr` |
| **Recommended Portainer stack name** | `media-stack-management` |
| **Repository** | `dessun-lab/media-stack` |
| **Compose file path** | `stacks/management/compose.yml` |
| **Environment contract** | [stacks/management/.env.example](.env.example) |
| **Shared network** | Existing external `arrnet_managed` |
| **Operator** | Luis / USER-owned |

Documentation only. Luis supplies values and owns all Portainer operations,
network preparation, cutover and rollback. See the [stack README](README.md).
