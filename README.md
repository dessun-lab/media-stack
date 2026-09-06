# ARR Stack

Automated media server with VPN protection. All download traffic routes through Gluetun — if the VPN drops, all traffic stops.

Based on [Tom Spark's ARR Stack](https://github.com/loponai/arrstack).

## Components

The Compose stack includes Jellyfin, qBittorrent, Prowlarr, Radarr, Sonarr,
Lidarr, Jellyseerr, Navidrome, Gluetun and helper services for subtitles,
scraping and media processing. Services that download or scrape use
`network_mode: service:gluetun` or the dedicated scraper VPN container.

The subdirectories `ffsubsync/`, `makemkv/`, `nensaysubs/`, `vuetorrent/` and
`yt-dlp/` contain supporting build or configuration files for local images.

## Usage

Configure the required values in `.env`, then start the stack from this
directory:

```bash
docker compose up -d
docker compose ps
```

Review the Gluetun and Compose network settings before exposing services. The
stack uses persistent host paths declared in `docker-compose.yml`, so those
paths must exist on the deployment host.