# Portainer metadata — acquisition

| Field | Value |
| --- | --- |
| **Purpose** | Acquire and process media while retaining the main VPN namespace and its consumers. |
| **Services** | `gluetun`, `qbittorrent`, `prowlarr`, `flaresolverr`, `ytdlp`, `themearr`, `makemkv` |
| **Recommended Portainer stack name** | `media-stack-acquisition` |
| **Repository** | `dessun-lab/media-stack` |
| **Compose file path** | `stacks/acquisition/compose.yml` |
| **Environment contract** | [stacks/acquisition/.env.example](.env.example) |
| **Shared network** | Existing external `arrnet_managed` |
| **Network lifecycle** | Externally managed shared host infrastructure / Luis (USER-owned) |
| **Network prerequisite** | Must exist before deployment; [provisioning](../../docs/operations/networking.md) |
| **Operator** | Luis / USER-owned |

**Operational status:** cutover complete per Luis (update recorded 2026-10-11);
Portainer variables and the external network are populated/provisioned. Names
above remain repository recommendations, not a live export of Portainer settings.
Luis owns all subsequent Portainer operations, network lifecycle and rollback.
See the [stack README](README.md).
