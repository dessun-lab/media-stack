# Host subtitle workflow

[Documentation map](../README.md) · [Subtitle unit](../../stacks/subtitles/README.md)

[scripts/download-and-sync-subs.bash](../../scripts/download-and-sync-subs.bash)
is a host-operated workflow coordinating NensaySubs and FFsubsync. It remains in
`scripts/`, because it serves both tools and can align existing subtitles without
a NensaySubs download.

| Mode | Preserved behavior |
| --- | --- |
| Series slug | `docker exec -it nensaysubs /usr/local/bin/nensaysubs`, then optional alignment. |
| `--move-only [baseVideoDir]` | NensaySubs move/rename path, then exit. |
| `--sync-only` | Skip download and enter file selection/alignment. |
| Manual/batch alignment | `sync_single_pair` runs `compose run --rm ffsubsync`; writes `.synced.<extension>` outputs; batch skips existing synced results. |

Host browser/matching defaults remain `/data/media` and `/data/media/subs`.
Containers/host must agree on paths and write permissions; interactive modes need
a terminal. Download/move modes require the existing container named `nensaysubs`.

The only adaptation derives SCRIPT_DIR from BASH_SOURCE and explicitly selects
`../stacks/subtitles/compose.yml` for FFsubsync. Compose-file selection no longer
depends on caller CWD; environment provisioning still belongs to Luis. No new
project-name override, profile, job conversion, full-stack startup or argument
change was added. FFsubsync has no dependency that would start the VPN group.

There is no tracked launcher/scheduler for this script. Luis handles external
callers and configuration context at the new checkout. The script still performs
real downloads/container operations/media writes; static Bash checks do not run
it. Agents must not execute it or use its Compose call to read protected `.env`.
