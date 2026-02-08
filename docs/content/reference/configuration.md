---
title: Configuration File
---

# Configuration File

Complete configuration file reference for lazyoci.

## File Location

Configuration file path resolution:

1. `$XDG_CONFIG_HOME/lazyoci/config.yaml`
2. `~/.config/lazyoci/config.yaml` (fallback)

## File Permissions

| Resource | Permissions |
|----------|------------|
| Config directory | `0755` |
| Config file | `0600` |

## Schema

```yaml
registries:
  - name: string
    url: string
    username: string      # optional
    password: string      # optional  
    insecure: boolean     # optional
cacheDir: string
artifactDir: string
defaultRegistry: string
theme: string
mode: string
```

## Field Reference

### registries

Array of registry configurations.

**Type:** `[]Registry`

**Default:**
```yaml
registries:
  - name: "Docker Hub"
    url: "docker.io"
  - name: "Quay.io" 
    url: "quay.io"
  - name: "GitHub Packages"
    url: "ghcr.io"
```

#### Registry Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | `string` | Yes | Display name |
| `url` | `string` | Yes | Registry URL |
| `username` | `string` | No | Authentication username |
| `password` | `string` | No | Authentication password |
| `insecure` | `boolean` | No | Allow insecure connections |

### cacheDir

Cache directory path.

**Type:** `string`  
**Default:** `~/.cache/lazyoci`

### artifactDir

Artifact storage directory path.

**Type:** `string`  
**Default:** `""` (resolves to `~/.cache/lazyoci/artifacts`)

### defaultRegistry

Default registry for operations.

**Type:** `string`  
**Default:** `"docker.io"`

### theme

UI theme name.

**Type:** `string`  
**Default:** `""` (resolves to `"default"`)

**Valid values:** See [Themes](./themes) reference.

### mode

Color mode preference.

**Type:** `string`  
**Default:** `""` (resolves to `"auto"`)

**Valid values:**
- `auto` - Detect from terminal
- `dark` - Force dark mode
- `light` - Force light mode

## Artifact Directory Resolution

Priority order for artifact directory:

1. `--artifact-dir` CLI flag
2. `$LAZYOCI_ARTIFACT_DIR` environment variable  
3. `artifactDir` config field
4. `~/.cache/lazyoci/artifacts` (fallback)

## Example Configuration

```yaml
registries:
  - name: "Docker Hub"
    url: "docker.io"
  - name: "Private Registry"
    url: "registry.company.com"
    username: "user"
    password: "pass"
    insecure: false
  - name: "Local Registry"
    url: "localhost:5000"
    insecure: true

cacheDir: "/tmp/lazyoci-cache"
artifactDir: "/home/user/artifacts"
defaultRegistry: "registry.company.com"
theme: "catppuccin-mocha"
mode: "dark"
```