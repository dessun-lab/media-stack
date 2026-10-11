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

Documentation only. Luis supplies values and owns all Portainer operations,
network preparation, cutover and rollback. See the [stack README](README.md).
