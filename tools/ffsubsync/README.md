# FFsubsync

[Tool index](../../docs/tools/README.md) · [Dockerfile](Dockerfile)

Luis builds this tool manually. From the Media Stack server checkout root,
use the following derived recipe; no build is performed by restructuring.

```bash
cd tools/ffsubsync
docker build --tag ffsubsync:latest .
```

Base: `python:3.12-slim`; unpinned pip `ffsubsync` plus distro FFmpeg.
Entrypoint `ffsubsync`. Runtime consumer is Compose service `ffsubsync`, invoked
by the host subtitle workflow as a one-off task, with writable `/data/media`,
no network, and restart policy `"no"`.

Version evidence above describes the recipe/toolchain. See the
[central version inventory](../../docs/versions.md) for fleet observations and limits.

## Build and ownership limits

The Dockerfile is unchanged. Keep its context here, never at the Media Stack root.
Manual procedure/tagging is reproducible; mutable base tags and unpinned packages
prevent a claim of byte-identical builds. Record source revision and resulting
image identity when Luis next builds. No Compose `build:` or CI build is provided.
