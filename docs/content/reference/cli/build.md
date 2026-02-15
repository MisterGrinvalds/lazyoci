---
title: build
---

# build

Build and push OCI artifacts from a `.lazy` configuration file.

## Synopsis

```
lazyoci build [path] [flags]
```

## Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `[path]` | Path to `.lazy` file or directory containing one | Optional |

**Argument validation:** MaximumNArgs(1)

If no path is given, lazyoci looks for a `.lazy` file in the current directory.

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | `""` | Path to `.lazy` config file |
| `--tag` | `-t` | `""` | Set `{{ .Tag }}` template variable (fallback: `LAZYOCI_TAG` env var) |
| `--push` | | `true` | Push to registries after build |
| `--no-push` | | `false` | Build only, don't push |
| `--dry-run` | | `false` | Show what would be built/pushed |
| `--artifact` | `-a` | `""` | Build specific artifact by name, type, or index |
| `--platform` | | `[]` | Override platforms for image builds (repeatable) |
| `--quiet` | `-q` | `false` | Suppress progress output |
| `--insecure` | | `false` | Allow HTTP for push targets |

## Inherited Flags

| Flag | Short | Default | Values |
|------|-------|---------|--------|
| `--output` | `-o` | `text` | `text`, `json`, `yaml` |
| `--artifact-dir` | | `""` | Artifact storage directory |
| `--theme` | | `""` | Theme name |

## Path Resolution

The build command finds the `.lazy` config in this order:

1. `--file` / `-f` flag (highest priority)
2. Positional `[path]` argument
3. `.lazy` in the current directory (default)

If a directory is given, lazyoci looks for `.lazy` inside it.

## Template Variables

Tag values and registry URLs in the `.lazy` file support Go template syntax:

| Variable | Source | Example |
|----------|--------|---------|
| `{{ .Registry }}` | `LAZYOCI_REGISTRY` env var | `localhost:5050` |
| `{{ .Tag }}` | `--tag` flag or `LAZYOCI_TAG` env var | `v1.0.0` |
| `{{ .GitSHA }}` | `git rev-parse --short HEAD` | `e5bce6f` |
| `{{ .GitBranch }}` | `git branch --show-current` | `main` |
| `{{ .ChartVersion }}` | `Chart.yaml` version (helm only) | `0.1.0` |
| `{{ .Timestamp }}` | UTC build time | `20260209153000` |
| `{{ .Version }}` | Semver (v prefix stripped) | `1.2.3` |
| `{{ .VersionMajor }}` | Major component | `1` |
| `{{ .VersionMinor }}` | Minor component | `2` |
| `{{ .VersionPatch }}` | Patch component | `3` |
| `{{ .VersionPrerelease }}` | Prerelease identifier | `rc.1` |
| `{{ .VersionMajorMinor }}` | Major.Minor | `1.2` |
| `{{ .VersionRaw }}` | Raw git tag string | `v1.2.3-rc.1` |

### Version resolution priority

The `{{ .Version }}` family is resolved in this order:

1. `LAZYOCI_VERSION` environment variable (if set and valid semver)
2. `--tag` flag value (if valid semver)
3. `git describe --tags --abbrev=0` (nearest git tag)

If the source is a valid semver (e.g., `v1.2.3-rc.1`), all component fields are populated.
If not a valid semver, only `{{ .Version }}` and `{{ .VersionRaw }}` are set.

## Environment Variables

| Variable | Fallback for | Description |
|----------|--------------|-------------|
| `LAZYOCI_REGISTRY` | (none) | Base registry URL for `{{ .Registry }}` in `.lazy` configs |
| `LAZYOCI_TAG` | `--tag` flag | Sets `{{ .Tag }}` when `--tag` is not specified |
| `LAZYOCI_VERSION` | Git tag detection | Overrides `{{ .Version }}` and all semver components |

## Artifact Filtering

The `--artifact` flag filters which artifacts to build. It matches against:

1. Artifact `name` (exact match)
2. Artifact `type` (all matching artifacts)
3. Zero-based index (e.g., `0` for the first artifact)

## Examples

```bash
# Build all artifacts
lazyoci build --tag v1.0.0

# Version auto-detected from git tag (no --tag needed)
lazyoci build

# Dry run
lazyoci build --tag v1.0.0 --dry-run

# Build from specific config
lazyoci build --file deploy/.lazy --tag v1.0.0

# Build from directory containing .lazy
lazyoci build examples/helm --tag v1.0.0

# Build only, no push
lazyoci build --tag v1.0.0 --no-push

# Build specific artifact by name
lazyoci build --tag v1.0.0 --artifact myapp

# Build all helm artifacts
lazyoci build --tag v1.0.0 --artifact helm

# Override platforms for image builds
lazyoci build --tag v1.0.0 --platform linux/amd64 --platform linux/arm64

# JSON output for scripting
lazyoci build --tag v1.0.0 -o json

# Push to insecure (HTTP) registries
lazyoci build --tag v1.0.0 --insecure

# CI: use env vars instead of flags
LAZYOCI_TAG=v1.2.3 lazyoci build

# CI: explicit version override
LAZYOCI_VERSION=1.2.3 lazyoci build

# Build against local dev registry
LAZYOCI_REGISTRY=localhost:5050 lazyoci build --tag v1.0.0 --insecure

# Build against DigitalOcean registry
LAZYOCI_REGISTRY=registry.digitalocean.com/greenforests lazyoci build --tag v1.0.0
```

## Output

### Text output

```
Building from /path/to/.lazy...
Tag: v1.0.0
Building myapp (type: image)...
  Pushing ghcr.io/owner/myapp:v1.0.0...
  Pushed ghcr.io/owner/myapp:v1.0.0 (digest: sha256:abc...)

OK    myapp (image)
      ghcr.io/owner/myapp:v1.0.0 [pushed] sha256:abc...
      ghcr.io/owner/myapp:latest [pushed] sha256:abc...
```

### JSON output

```json
[
  {
    "name": "myapp",
    "type": "image",
    "targets": [
      {
        "reference": "ghcr.io/owner/myapp:v1.0.0",
        "digest": "sha256:abc...",
        "pushed": true
      }
    ]
  }
]
```

## See Also

- [`.lazy` Config Reference](../lazy-config) -- full schema for the `.lazy` configuration file
- [Building Artifacts Guide](/guides/building-artifacts) -- step-by-step guide for each artifact type
