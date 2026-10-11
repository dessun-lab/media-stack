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
| **Operator** | Luis / USER-owned |

Documentation only. Luis supplies values and owns all Portainer operations,
network preparation, cutover and rollback. See the [stack README](README.md).
