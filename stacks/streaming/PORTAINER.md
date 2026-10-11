# Portainer metadata — streaming

| Field | Value |
| --- | --- |
| **Purpose** | Serve video and music independently of acquisition and library-management updates. |
| **Services** | `jellyfin`, `navidrome` |
| **Recommended Portainer stack name** | `media-stack-streaming` |
| **Repository** | `dessun-lab/media-stack` |
| **Compose file path** | `stacks/streaming/compose.yml` |
| **Environment contract** | [stacks/streaming/.env.example](.env.example) |
| **Shared network** | Existing external `arrnet_managed` |
| **Network lifecycle** | Externally managed shared host infrastructure / Luis (USER-owned) |
| **Network prerequisite** | Must exist before deployment; [provisioning](../../docs/operations/networking.md) |
| **Operator** | Luis / USER-owned |

**Operational status:** cutover complete per Luis (update recorded 2026-10-11);
Portainer variables and the external network are populated/provisioned. Names
above remain repository recommendations, not a live export of Portainer settings.
Luis owns all subsequent Portainer operations, network lifecycle and rollback.
See the [stack README](README.md).
