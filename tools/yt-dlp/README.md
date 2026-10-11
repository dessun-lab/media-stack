# yt-dlp

[Tool index](../../docs/tools/README.md) · [Dockerfile](Dockerfile)

Luis builds this tool manually. From the Media Stack server checkout root,
use the following derived recipe; no build is performed by restructuring.

```bash
cd tools/yt-dlp
docker build --tag ytdlp:latest - < Dockerfile
```

Base: `python:3.12-alpine`; unpinned pip `yt-dlp` plus FFmpeg/curl/bash;
UID/GID 1000; image entrypoint `yt-dlp`. This recipe has no `COPY`/`ADD`, so the
command uses an empty filesystem context. It preserves manual execution while
avoiding inclusion of server-only `cookies.txt` beside the recipe.
[Docker build contexts](https://docs.docker.com/build/concepts/context/).

Service `ytdlp` overrides the entrypoint with an idle shell command, shares the
main VPN, and binds its runtime music/downloads plus the cookie file. Do not
read cookies or turn building into a deployment operation.

Version evidence above describes the recipe/toolchain. See the
[central version inventory](../../docs/versions.md) for fleet observations and limits.

## Build and ownership limits

The Dockerfile is unchanged. Keep its context here, never at the Media Stack root.
Manual procedure/tagging is reproducible; mutable base tags and unpinned packages
prevent a claim of byte-identical builds. Record source revision and resulting
image identity when Luis next builds. No Compose `build:` or CI build is provided.
