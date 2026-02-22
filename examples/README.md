# lazyoci examples

Each subdirectory demonstrates a different lazyoci capability — building OCI artifacts, or mirroring upstream charts and images to a private registry.

## Prerequisites

All examples use `{{ .Registry }}` in their `.lazy` configs, so you must set `LAZYOCI_REGISTRY` before running them.

### Local development

Start the local development registry and set the env var:

```sh
make registry-up   # starts localhost:5050
export LAZYOCI_REGISTRY=localhost:5050
```

### DigitalOcean Container Registry

```sh
export LAZYOCI_REGISTRY=registry.digitalocean.com/greenforests
```

Authenticate first via `docker login registry.digitalocean.com` or configure Docker credential helpers.

## Examples

### [`image/`](image/) — Container image from Dockerfile

Builds a multi-architecture Go HTTP server image using `docker buildx`.

```sh
lazyoci build examples/image --tag v1.0.0
lazyoci build examples/image --tag v1.0.0 --dry-run   # preview only
```

### [`helm/`](helm/) — Helm chart as OCI artifact

Packages a Helm chart directory and pushes it with the proper OCI media types. The tag `{{ .ChartVersion }}` is resolved from `Chart.yaml`.

```sh
lazyoci build examples/helm --tag v1.0.0
```

### [`artifact/`](artifact/) — Generic OCI artifact

Pushes arbitrary files (a JSON config and an OPA policy) with custom media types. Useful for distributing configs, policies, WASM modules, or any non-image content.

```sh
lazyoci build examples/artifact --tag v1.0.0
```

### [`docker/`](docker/) — Push existing Docker daemon image

Exports an image that already exists in the local Docker daemon and pushes it to a registry. No Dockerfile or build step needed.

```sh
docker pull nginx:alpine                          # ensure the image exists locally
lazyoci build examples/docker --tag v1.0.0
```

### [`multi/`](multi/) — Multiple artifacts in one config

A single `.lazy` file that builds an image, packages its Helm chart, and publishes its OpenAPI spec — all in one command. Use `--artifact` to target a specific one.

```sh
lazyoci build examples/multi --tag v1.0.0                      # build all three
lazyoci build examples/multi --tag v1.0.0 --artifact api-chart  # helm chart only
lazyoci build examples/multi --tag v1.0.0 --dry-run -o json     # preview as JSON
```

### [`mirror/`](mirror/) — Mirror upstream charts & images

Mirrors Helm charts and their container images from upstream sources to a target OCI registry. Includes a demo local chart and examples of all three source types (repo, oci, local).

```sh
# Preview what would be mirrored
lazyoci mirror --config examples/mirror/mirror.yaml --all --dry-run

# Mirror all charts
lazyoci mirror --config examples/mirror/mirror.yaml --all --insecure

# Mirror a specific chart
lazyoci mirror --config examples/mirror/mirror.yaml --chart vault
```

## Quick test (local registry)

```sh
# Start registry, build, and verify
make registry-up
LAZYOCI_REGISTRY=localhost:5050 lazyoci build examples/artifact --tag v0.1.0 --insecure
curl -s http://localhost:5050/v2/_catalog | python3 -m json.tool
```

Or use the Makefile convenience targets:

```sh
make build-local    # dry-run against localhost:5050
make build-docr     # dry-run against DigitalOcean
```

## Template variables

All `.lazy` files support Go template syntax in both registry and tag fields:

| Variable | Source | Example |
|---|---|---|
| `{{ .Registry }}` | `LAZYOCI_REGISTRY` env var | `localhost:5050` |
| `{{ .Tag }}` | `--tag` flag or `LAZYOCI_TAG` env | `v1.0.0` |
| `{{ .Version }}` | Semver from git tag (auto-detected) | `1.2.3` |
| `{{ .VersionMajorMinor }}` | Major.Minor | `1.2` |
| `{{ .VersionMajor }}` | Major component | `1` |
| `{{ .VersionMinor }}` | Minor component | `2` |
| `{{ .VersionPatch }}` | Patch component | `3` |
| `{{ .VersionPrerelease }}` | Prerelease identifier | `rc.1` |
| `{{ .VersionRaw }}` | Raw git tag string | `v1.2.3-rc.1` |
| `{{ .GitSHA }}` | `git rev-parse --short HEAD` | `e5bce6f` |
| `{{ .GitBranch }}` | `git branch --show-current` | `main` |
| `{{ .ChartVersion }}` | `Chart.yaml` version field (helm only) | `0.1.0` |
| `{{ .Timestamp }}` | UTC build time | `20260209153000` |

Version is auto-detected from git tags. Override with `LAZYOCI_VERSION` env var or a semver `--tag` value.

## Common flags

```
--tag / -t        Set {{ .Tag }} (also populates {{ .Version }} if semver)
--dry-run         Preview without building or pushing
--no-push         Build locally but don't push
--artifact / -a   Filter to a single artifact by name, type, or index
--platform        Override platforms for image builds
--insecure        Allow HTTP registries
--quiet / -q      Suppress progress output
-o json           Structured JSON output
```

## Environment variables

```
LAZYOCI_REGISTRY  Base registry URL for {{ .Registry }} in .lazy configs
LAZYOCI_TAG       Fallback for --tag when not set on CLI
LAZYOCI_VERSION   Override version detection (skips git describe)
```
