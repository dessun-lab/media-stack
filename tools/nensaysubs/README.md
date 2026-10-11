# NensaySubs

[Tool index](../../docs/tools/README.md) · [Dockerfile](Dockerfile)

Luis builds this tool manually. From the Media Stack server checkout root,
use the following derived recipe; no build is performed by restructuring.

```bash
cd tools/nensaysubs
docker build --tag nensaysubs:latest .
```

Base: `golang:1.26-alpine`; [go.mod](go.mod) declares `go 1.26.3`.
Build uses CGO with Tesseract/Leptonica, FFmpeg, and archive tools. `COPY go.*` and
`COPY . .` require the tool's source directory as context. The recipe compiles
`/usr/local/bin/nensaysubs`; its default command idles for later invocation.

The [Go source](nensaysubs.go) implements series downloading,
ZIP/RAR extraction, OCR/challenge handling, subtitle placement/renaming, and
conversion. The host script executes this binary in service `nensaysubs` on the
scraper VPN. It has writable media, a `/runtime` bind, and `restart: "no"`;
the image workdir is `/app`, so the `/runtime` mount alone does not establish
where every program artifact is stored. The tracked native binary is a separate
legacy review item, not the `/usr/local/bin` artifact rebuilt by the recipe.

Version evidence above describes the recipe/toolchain. See the
[central version inventory](../../docs/versions.md) for fleet observations and limits.

## Build and ownership limits

The Dockerfile is unchanged. Keep its context here, never at the Media Stack root.
Manual procedure/tagging is reproducible; mutable base tags and unpinned packages
prevent a claim of byte-identical builds. Record source revision and resulting
image identity when Luis next builds. No Compose `build:` or CI build is provided.
