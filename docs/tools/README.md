# Tool index and external provenance

[Documentation map](../README.md)

Locally owned utilities keep source, recipe and build/use notes together. Docker
is a packaging mechanism, not a separate `images/` class. Luis intentionally
builds manually; no Compose `build:`, CI or automatic rebuild lifecycle is added.

| Tool | Owning instructions | Consumed image | Unit |
| --- | --- | --- | --- |
| Subtitle alignment | [FFsubsync](../../tools/ffsubsync/README.md) | `ffsubsync:latest` | subtitles |
| Media conversion | [MakeMKV](../../tools/makemkv/README.md) | `makemkv:latest` | acquisition |
| Subtitle acquisition/placement | [NensaySubs](../../tools/nensaysubs/README.md) | `nensaysubs:latest` | subtitles |
| Media downloading | [yt-dlp](../../tools/yt-dlp/README.md) | `ytdlp:latest` | acquisition |
| Theme retrieval | External ThemeARR source below | `themearr:latest` | acquisition |

HomeLab Inventory at `32a8758`, dated 2026-10-03, records ytdlp, nensaysubs and
themearr custom `latest` tags with version `N/A`; no FFsubsync/MakeMKV rows.
This is dated tag evidence, not proof of deployed digests or current build versions.
A manual build creates an image/tag; it does not publish or redeploy a container.

## ThemeARR external source

Source ownership is external: the accessible sibling `animethemes-cli` repository
at `aa949a4b7b4850a69ea29b054a0bd7e342809191` contains an AnimeThemes CLI/TUI and
its Dockerfile. Its Go module is `github.com/dessun-lab/animethemes-cli`; source
requires Go 1.26.2. No ThemeARR recipe/source exists in Media Stack's inspected
current tree or targeted service history.

The match is substantive: `cmd/cli/root.go` reads `ANIMETHEMES_API_URL` and
`ANIMETHEMES_OUTPUT_DIR`, while its Dockerfile exposes 7681 through `ttyd` and
runs `animtheme tui`. These match Media Stack's variables, output bind and
Gluetun's 7681 forwarding. History `1f91a42` introduced `themearr:latest` with
that forwarding. Treat `animethemes-cli` as the matching external source/recipe;
no source reference or digest proves the exact current image was built from this
revision. Inventory records the image tag with application version `N/A`.

The recipe uses `debian:bookworm-slim`, installs certificates/curl/FFmpeg, creates
user `app`, and copies **prebuilt** `ttyd/ttyd` and `dist/animtheme-linux-amd64`.
Both required inputs are tracked. It does not compile Go inside the image.
The distribution metadata records `0.0.0-SNAPSHOT-cc506a1` (2026-05-09); that is
packaging evidence, not the deployed application version or proof of binary parity
with current source. `.goreleaser.yaml` describes rebuilding distributions.

Derived manual recipe, **not an attestation of Luis's original build**:

```bash
# Start at Luis's separate animethemes-cli checkout, with its intended build inputs.
docker build --tag themearr:latest .
```

Here the context/Dockerfile are the external repository root. Luis selects the
intended prebuilt inputs and keeps secrets out of that build context. No external
build or source import was performed. A source refresh must also refresh the copied
binary before claiming the image includes it. Manual ownership stays with Luis;
there is no new CI/Compose build mechanism.

Runtime consumer `themearr` stays with Gluetun in the acquisition unit,
retaining its API expression and output at
`/home/devops/runtime/arrstack/themearr/AnimeThemes`. Do not create
`tools/themearr/` here: document the external dependency instead of duplicating
its source. If source ownership ever transfers, the same `tools/<tool>/` rule
would apply under separate authorization.
