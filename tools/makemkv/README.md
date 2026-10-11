# MakeMKV

[Tool index](../../docs/tools/README.md) · [Dockerfile](Dockerfile)

Luis builds this tool manually. From the Media Stack server checkout root,
use the following derived recipe; no build is performed by restructuring.

```bash
cd tools/makemkv
docker build --build-arg MAKEMKV_VERSION=1.18.3 --tag makemkv:latest .
```

Base: `ubuntu:24.04`; source downloads/builds use default tool version `1.18.3`.
Includes FFmpeg and mkvtoolnix; entrypoint `makemkvcon`. Service `makemkv` has no
network, restart policy `"no"`, runtime config, and qBittorrent's downloads child
mounted at `/downloads`. No invocation script or device passthrough is declared;
do not infer an optical-drive operating procedure from the image name.

Version evidence above describes the recipe/toolchain. See the
[central version inventory](../../docs/versions.md) for fleet observations and limits.

## Build and ownership limits

The Dockerfile is unchanged. Keep its context here, never at the Media Stack root.
Manual procedure/tagging is reproducible; mutable base tags and unpinned packages
prevent a claim of byte-identical builds. Record source revision and resulting
image identity when Luis next builds. No Compose `build:` or CI build is provided.
