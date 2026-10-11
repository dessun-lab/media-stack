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
| **Network lifecycle** | Externally managed shared host infrastructure / Luis (USER-owned) |
| **Network prerequisite** | Must exist before deployment; [provisioning](../../docs/operations/networking.md) |
| **Operator** | Luis / USER-owned |

**Operational status:** cutover complete per Luis (update recorded 2026-10-11);
Portainer variables and the external network are populated/provisioned. Names
above remain repository recommendations, not a live export of Portainer settings.
Luis owns all subsequent Portainer operations, network lifecycle and rollback.
See the [stack README](README.md).
