# lazyoci

A terminal UI for browsing OCI container registries.

Browse Docker Hub, Quay.io, GitHub Packages, and custom registries to discover container images, Helm charts, SBOMs, signatures, and other OCI artifacts.

## Features

- **Interactive TUI** -- navigate registries, repositories, and artifacts with keyboard shortcuts
- **Multi-registry support** -- Docker Hub, Quay.io, GHCR, Harbor, DigitalOcean, and any OCI-compliant registry
- **Artifact type detection** -- automatically identifies images, Helm charts, SBOMs, signatures, attestations, and WASM modules
- **Pull artifacts** -- download to local OCI layout or load directly into Docker
- **Docker credential integration** -- reads credentials from Docker Desktop, credential helpers, and `~/.docker/config.json`
- **Search** -- find repositories across registries
- **CLI commands** -- scriptable interface with JSON/YAML output
- **Theming** -- 7 built-in color themes with dark/light mode support

## Installation

```bash
go install github.com/mistergrinvalds/lazyoci/cmd/lazyoci@latest
```

Or build from source:

```bash
git clone https://github.com/mistergrinvalds/lazyoci.git
cd lazyoci
make build
# binary is at ./bin/lazyoci
```

## Quick Start

```bash
# Launch the TUI
lazyoci

# Launch with a specific theme
lazyoci --theme catppuccin-mocha

# Pull an image and load it into Docker
lazyoci pull nginx:latest --docker

# Search Docker Hub
lazyoci browse search docker.io nginx
```

## TUI Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus between panels |
| `Shift+Tab` | Go back to registry list |
| `1` / `2` / `3` / `4` | Focus registry / search / artifacts / details |
| `/` | Focus search |
| `j` / `k` | Move down/up |
| `g` / `G` | Scroll to top/bottom (details panel) |
| `Enter` | Select item |

### Actions

| Key | Action |
|-----|--------|
| `p` | Pull artifact |
| `d` | Pull and load into Docker |
| `T` | Theme picker |
| `S` | Settings |
| `?` | Help |
| `q` | Quit |

## CLI Commands

### Browse Registries

```bash
# List repositories
lazyoci browse repos localhost:5050

# List tags with filtering
lazyoci browse tags docker.io/library/nginx --limit 10 --filter alpine

# Show manifest details
lazyoci browse manifest docker.io/library/nginx:latest

# Search for repositories
lazyoci browse search docker.io nginx
```

### Pull Artifacts

```bash
# Pull to local OCI layout
lazyoci pull nginx:latest

# Pull and load into Docker
lazyoci pull alpine:latest --docker

# Pull specific platform
lazyoci pull nginx:latest --platform linux/arm64

# Pull to custom directory
lazyoci pull nginx:latest --dest ~/my-artifacts

# JSON output
lazyoci pull nginx:latest -q -o json
```

### Manage Registries

```bash
# List configured registries
lazyoci registry list

# Add a registry (auth is resolved from Docker credentials automatically)
lazyoci registry add harbor.example.com

# Add with explicit credentials
lazyoci registry add private.io --user admin --pass secret

# Add an insecure (HTTP) registry
lazyoci registry add localhost:5050 --insecure

# Test connectivity
lazyoci registry test harbor.example.com

# Remove a registry
lazyoci registry remove harbor.example.com
```

### Configuration

```bash
# Show config file path
lazyoci config path

# List all configuration
lazyoci config list

# Get/set values
lazyoci config get artifact-dir
lazyoci config set artifact-dir ~/my-artifacts --create
```

## Authentication

lazyoci resolves credentials automatically through a chain of sources (highest priority first):

1. **Per-registry credential helpers** -- `credHelpers` in `~/.docker/config.json`
2. **Default credential helper** -- `credsStore` in `~/.docker/config.json` (e.g. Docker Desktop)
3. **Docker config auths** -- base64 credentials from `docker login`
4. **lazyoci config** -- explicit username/password in `~/.config/lazyoci/config.yaml`
5. **Anonymous** -- no credentials

If you've already run `docker login` for a registry, lazyoci will use those credentials with no extra configuration.

## Configuration

Config file: `~/.config/lazyoci/config.yaml` (or `$XDG_CONFIG_HOME/lazyoci/config.yaml`)

```yaml
registries:
  - name: Docker Hub
    url: docker.io
  - name: My Harbor
    url: harbor.example.com
  - name: Local Registry
    url: localhost:5050
    insecure: true

cacheDir: ~/.cache/lazyoci
artifactDir: ~/oci-artifacts
defaultRegistry: docker.io
theme: catppuccin-mocha
mode: auto  # auto, dark, or light
```

### Artifact Directory Priority

1. `--artifact-dir` CLI flag
2. `LAZYOCI_ARTIFACT_DIR` environment variable
3. `artifactDir` in config file
4. Default: `~/.cache/lazyoci/artifacts`

## Artifact Types

lazyoci detects and displays these OCI artifact types:

| Type | Badge | Storage Dir |
|------|-------|-------------|
| Container Image | `IMG` | `oci/` |
| Helm Chart | `HELM` | `helm/` |
| SBOM | `SBOM` | `sbom/` |
| Signature | `SIG` | `sig/` |
| Attestation | `ATT` | `att/` |
| WebAssembly | `WASM` | `wasm/` |

## Themes

7 built-in themes with dark/light variants:

`default` `catppuccin-mocha` `catppuccin-latte` `dracula` `tokyonight` `gruvbox` `solarized-dark`

Set via CLI flag (`--theme dracula`), config file, or at runtime (press `T` in the TUI).

## Registry Compatibility

| Registry | List Repos | List Tags | Search | Pull |
|----------|------------|-----------|--------|------|
| Docker Hub | via search | Yes | Yes | Yes |
| Quay.io | via search | Yes | Yes | Yes |
| GHCR | No | Yes | No | Yes |
| Harbor | Yes | Yes | Yes | Yes |
| DigitalOcean | Yes | Yes | No | Yes |
| Local (distribution) | Yes | Yes | No | Yes |

GHCR doesn't support the catalog API; use `browse tags` with known repository paths.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LAZYOCI_ARTIFACT_DIR` | Override artifact storage directory |
| `XDG_CONFIG_HOME` | Override config directory (default `~/.config`) |
| `DOCKER_CONFIG` | Override Docker config directory (default `~/.docker`) |
| `COLORFGBG` | Terminal background hint for dark/light mode detection |

## Development

```bash
# Build
make build

# Run tests
make test

# Run with race detection
make test-all

# Start local OCI registry for testing
make registry-up

# Push all test artifact types to local registry
make registry-push-all

# Stop local registry
make registry-down
```

## Documentation

Full documentation is available as a Docusaurus site:

```bash
# Install dependencies and start dev server
make docs-dev

# Build for production
make docs-build
```

The docs are organized using the [Diataxis framework](https://diataxis.fr/):

- **Tutorials** -- step-by-step learning guides
- **How-to Guides** -- solve specific problems
- **Reference** -- CLI, keybindings, config, and API details
- **Explanation** -- background on OCI concepts and architecture

## License

MIT
