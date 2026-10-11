# Server checkout synchronization

[Documentation map](../README.md) · [Path-change history](../design/implementation-handoff.md)

The [caller](../../.github/workflows/sync-repo.yml) runs on main push (no path
filter) and manual dispatch, selects the dispatch server or DEFAULT_SERVER_NAME,
and inherits secrets. It calls the existing reusable workflow; no workflow ran here.

## Verified input contract

`repository: ${{ github.event.repository.name }}` supplies the short repository
name (`media-stack`) for either supported trigger. The `github` context is allowed
at reusable-job `with` inputs; event repository data provides the name, rather
than the owner/name pair from `github.repository`.
[GitHub context reference](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts#github-context).

The narrowly inspected `dl-actions` workflow at `8e8cb5b` expects a string short
slug. It constructs both `github.com/dessun-lab/${REPOSITORY}.git` and
`${DEPLOY_PATH}/${REPOSITORY}`. Therefore the dynamic name correctly changes the
GitHub identity and intended server checkout to `/home/devops/infra/media-stack`
with Luis's infrastructure DEPLOY_PATH. No reusable-workflow change is needed.
The previous design to preserve the old checkout path is superseded.

## Separate consumer, unchanged mechanics

The reusable workflow selects direct/jump SSH, pulls main into an existing Git
checkout or clones if absent, lists filenames and removes temporary runner SSH
keys. Its clean action targets runner workspace, not the remote checkout. It
contains no Docker build, Compose operation or Portainer request. This source
snapshot uses mutable `@main`; no live variables, secrets or jobs were inspected.

The cutover is complete per Luis; the current checkout contract uses the new
prefix. Pull/clone is still not a data migration or backup: future recovery must
preserve server-only `navidrome/music`, `yt-dlp/cookies.txt` and external
script/build callers/config context. The operator report does not establish the
exact sync-job history or whether the old checkout was retained. Source recipes
now live under `tools/`; cookie subpaths intentionally do not. No remote checkout
was created, renamed or
removed by this task. Portainer independently consumes GitHub and remains Luis-owned.
