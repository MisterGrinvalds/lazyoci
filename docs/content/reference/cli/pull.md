---
title: pull
---

# pull

Pull OCI artifacts from a registry.

## Synopsis

```
lazyoci pull <reference> [flags]
```

## Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<reference>` | OCI artifact reference | Required |

**Argument validation:** ExactArgs(1)

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--dest` | `-d` | `""` | Destination directory |
| `--platform` | | `""` | Target platform |
| `--docker` | | `false` | Pull to Docker daemon |
| `--quiet` | `-q` | `false` | Suppress output |

## Inherited Flags

| Flag | Short | Default | Values |
|------|-------|---------|--------|
| `--output` | `-o` | `text` | `text`, `json`, `yaml` |
| `--artifact-dir` | | `""` | Artifact storage directory |
| `--theme` | | `""` | Theme name |

## Examples

```bash
lazyoci pull nginx:latest
lazyoci pull --dest ./output nginx:alpine
lazyoci pull --docker nginx:latest
lazyoci pull --platform linux/amd64 nginx:latest
lazyoci pull --quiet nginx:latest
```