---
title: browse
---

# browse

Browse registry contents.

## Subcommands

- [`repos`](#repos) - List repositories
- [`tags`](#tags) - List repository tags  
- [`manifest`](#manifest) - Show manifest
- [`search`](#search) - Search artifacts

## repos

List repositories in a registry.

### Synopsis

```
lazyoci browse repos <registry-url> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<registry-url>` | Registry URL | Required |

**Argument validation:** ExactArgs(1)

### Examples

```bash
lazyoci browse repos docker.io
lazyoci browse repos ghcr.io
```

## tags

List tags for a repository.

### Synopsis

```
lazyoci browse tags <registry/repo> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<registry/repo>` | Repository reference | Required |

**Argument validation:** ExactArgs(1)

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--limit` | `20` | Maximum number of tags |
| `--offset` | `0` | Starting offset |
| `--filter` | `""` | Tag filter pattern |

### Examples

```bash
lazyoci browse tags nginx
lazyoci browse tags --limit 50 nginx
lazyoci browse tags --filter "alpine" nginx
```

## manifest

Show artifact manifest.

### Synopsis

```
lazyoci browse manifest <registry/repo:tag> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<registry/repo:tag>` | Artifact reference | Required |

**Argument validation:** ExactArgs(1)

### Examples

```bash
lazyoci browse manifest nginx:latest
lazyoci browse manifest ghcr.io/owner/repo:v1.0.0
```

## search

Search artifacts in a registry.

### Synopsis

```
lazyoci browse search <registry> <query> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<registry>` | Registry URL | Required |
| `<query>` | Search query | Required |

**Argument validation:** ExactArgs(2)

### Examples

```bash
lazyoci browse search docker.io nginx
lazyoci browse search ghcr.io kubernetes
```