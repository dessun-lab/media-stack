# Production-unit decisions

[Design overview](restructuring.md)

Approved 2026-10-11. Exact membership and service-specific knowledge belong to
[streaming](../../stacks/streaming/README.md),
[management](../../stacks/management/README.md),
[acquisition](../../stacks/acquisition/README.md) and
[subtitles](../../stacks/subtitles/README.md).

## Boundary rationale

Streaming groups Jellyfin/Navidrome serving availability independently of library
management and acquisition maintenance. Management owns requests/library automation;
Prowlarr stays in acquisition because its main-VPN namespace is a hard constraint.
Acquisition keeps all main-VPN consumers together and MakeMKV with the download
tree it processes. Subtitles owns the independent scraper VPN, its loopback solver
and NensaySubs, plus FFsubsync coordinated by the same host workflow.

Two coarse units retained unrelated redeploy scope. Six dependency/per-tool units
added ownership overhead without operational benefit. Four units give useful
operator choices while accepting some shared lifecycle within each purpose.
No per-player, per-manager, per-tool or recovery stack is introduced. deunhealth
and its sole qBittorrent integration label are explicitly removed by Luis.

## Preservation and limits

Keep VPN owners with all `network_mode: service:` consumers and `depends_on`
edges inside their unit. Preserve image/tag, names, commands, defaults, bind modes,
ports, IPs, UID/GID, devices/plugins, restart/health behavior and comments/style.
The allowed differences are enumerated in the [design](restructuring.md).

Shared media, network and host resources remain cross-unit coupling. The split
does not guarantee playback during storage/host failures or application outages.
Luis chooses independent Portainer update policies and performs runtime acceptance;
repository implementation requires no Portainer settings or inspection.
