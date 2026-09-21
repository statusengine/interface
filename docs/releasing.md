# Releasing

A release is a tag. Pushing `v1.2.3` builds the binaries, creates the
release and uploads them.

```bash
git tag -a v1.2.3 -m "v1.2.3"
git push origin v1.2.3
```

Everything the pipeline does lives in the repository rather than in the
workflow file, so any of it can be run by hand:

| Command | What it does |
|---|---|
| `tools/ci/checks.sh` | Everything CI checks: vet, gofmt, Go tests, prettier, frontend tests, translations, build |
| `tools/ci/checks.sh go` | Just the Go side, for a quicker loop |
| `make dist VERSION=v1.2.3` | The release archives and their checksums, in `dist/` |
| `tools/release/publish.sh v1.2.3 dist` | Creates the release and uploads the archives |
| `tools/release/publish-test.sh` | Proves `publish.sh` still speaks both forge APIs |

## What a release contains

One archive per platform - `linux/amd64` and `linux/arm64` - each with
the `seid` binary, `seid.example.yaml` and the README. The binary is
static (`CGO_ENABLED=0`), built with `-trimpath`, and carries its version:

```
$ ./seid version
v1.2.3
```

`SHA256SUMS` sits beside them:

```bash
sha256sum -c SHA256SUMS --ignore-missing
```

The release notes are the commit subjects since the previous tag. A tag
with `-rc`, `-beta` or `-alpha` in it is published as a prerelease.

## Prerequisites

This repository lives on Gitea, which runs these workflows itself -
Gitea Actions reads `.github/workflows/` and speaks GitHub's syntax.
The same files work unchanged on GitHub.

Before the first tag:

- [ ] **Actions enabled** for the repository (Settings → Actions), and
      at least one `act_runner` registered that offers the
      `ubuntu-latest` label. Without a runner the workflow queues
      forever rather than failing, which is easy to mistake for nothing
      happening.
- [ ] **A token the workflow can publish with.** Gitea gives the
      workflow a `GITHUB_TOKEN` automatically; if that one is not
      allowed to create releases on your instance, add a repository
      secret `RELEASE_TOKEN` holding a personal access token with
      `write:repository`. `publish.sh` prefers `RELEASE_TOKEN` and falls
      back to the automatic one.
- [ ] **The runner can reach github.com**, because `actions/checkout`,
      `actions/setup-go` and `actions/setup-node` are fetched from
      there. That is Gitea's default; an air-gapped runner needs
      `DEFAULT_ACTIONS_URL` pointed at a mirror.

There is deliberately no `actions/upload-artifact` step. Its v4 needs an
artifact API that Gitea's runner does not serve, and its v3 no longer
works on GitHub - it is the one step that cannot be written for both
forges at once. The archives are attached to the release, and a failed
publish is a re-run.

## Moving to GitHub

Nothing in the workflows is Gitea-specific. On GitHub:

- `secrets.GITHUB_TOKEN` is provided automatically and can create
  releases, so `RELEASE_TOKEN` becomes unnecessary - the fallback in
  `release.yml` already prefers whichever exists.
- `publish.sh` notices `api.github.com` and switches to GitHub's upload
  host and encoding on its own. `publish-test.sh` covers that path.
- The `permissions: contents: write` block in `release.yml` is what
  GitHub needs and Gitea ignores.

The one thing worth doing on the day of the move is pushing a
`v0.0.0-rc` tag first and deleting the release afterwards, the same way
the pipeline was tried out here.

## Trying it without publishing anything

`publish.sh` will show exactly what it would create:

```bash
make dist VERSION=v1.2.3
GITHUB_API_URL=https://git.example.org/api/v1 \
GITHUB_REPOSITORY=statusengine/statusengine4-interface \
GITHUB_SERVER_URL=https://git.example.org \
DRY_RUN=1 tools/release/publish.sh v1.2.3 dist
```

That prints the tag, whether it counts as a prerelease, the full notes
and the list of files - and sends nothing.

## When something goes wrong

- **The release exists but has no files.** The upload step failed after
  the release was created. Delete the release in the web interface and
  re-run the workflow, or build locally with `make dist VERSION=v1.2.3`
  and attach the archives by hand.
- **The workflow never starts.** No runner has the `ubuntu-latest`
  label, or Actions is off for the repository.
- **`pattern all:dist: no matching files found`.** The frontend was not
  built before the Go build. `make dist` does both in order;
  `internal/webui/dist/.gitkeep` is what lets the Go code compile before
  that.
- **The notes are empty.** The checkout was shallow. The workflow uses
  `fetch-depth: 0` for exactly this reason.

## The container image

`docker build` produces the same binary in a 29 MB Alpine image that
runs as an unprivileged user and binds `:8090`:

```bash
docker build --build-arg VERSION=v1.2.3 -t seid:v1.2.3 .
docker run --rm seid:v1.2.3 version
```

Publishing it is not part of the release workflow yet: it needs a
registry and credentials, and which registry - Gitea's own or another -
is a deployment decision rather than a build one. The Dockerfile is
built and smoke-tested by hand for now.

## Versioning

`v<major>.<minor>.<patch>`, and the tag is the only place the version is
written down: the Makefile and the Dockerfile both take it from the tag
and stamp it into the binary with `-ldflags`. There is no version
constant to forget to bump.
