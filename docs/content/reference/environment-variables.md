---
title: Environment Variables
---

# Environment Variables

Environment variables recognized by lazyoci.

## lazyoci Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `LAZYOCI_ARTIFACT_DIR` | Override artifact storage directory | `~/.cache/lazyoci/artifacts` |

## System Variables

| Variable | Description | Default | Usage |
|----------|-------------|---------|-------|
| `XDG_CONFIG_HOME` | Override config directory base | `~/.config` | Config file location |
| `DOCKER_CONFIG` | Override Docker config directory | `~/.docker` | Docker daemon integration |
| `COLORFGBG` | Terminal background hint | - | Theme auto-detection |

## Variable Details

### LAZYOCI_ARTIFACT_DIR

Overrides the directory where pulled artifacts are stored.

**Priority:** Higher than config file `artifactDir`, lower than `--artifact-dir` flag.

**Examples:**
```bash
export LAZYOCI_ARTIFACT_DIR="/tmp/artifacts"
lazyoci pull nginx:latest
```

### XDG_CONFIG_HOME

Standard XDG Base Directory specification variable.

**Usage:** Configuration file location becomes `$XDG_CONFIG_HOME/lazyoci/config.yaml`

**Examples:**
```bash
export XDG_CONFIG_HOME="/home/user/.config"
# Config file: /home/user/.config/lazyoci/config.yaml
```

### DOCKER_CONFIG

Docker configuration directory for daemon integration.

**Usage:** Used when pulling artifacts directly to Docker daemon with `--docker` flag.

### COLORFGBG

Terminal background color hint for automatic theme detection.

**Format:** `foreground;background`

**Background Values:**
- `0-6` - Dark background
- `7-15` - Light background

**Examples:**
```bash
COLORFGBG="15;0"  # Light foreground, dark background
COLORFGBG="0;15"  # Dark foreground, light background
```