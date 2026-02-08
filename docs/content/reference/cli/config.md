---
title: config
---

# config

Manage lazyoci configuration.

## Subcommands

- [`get`](#get) - Get configuration value
- [`set`](#set) - Set configuration value
- [`list`](#list) - List all configuration
- [`path`](#path) - Show configuration file path

## get

Get a configuration value.

### Synopsis

```
lazyoci config get <key> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<key>` | Configuration key | Required |

**Argument validation:** ExactArgs(1)

### Valid Keys

| Key | Aliases | Description |
|-----|---------|-------------|
| `artifact-dir` | `artifactdir` | Artifact storage directory |
| `cache-dir` | `cachedir` | Cache directory |
| `default-registry` | `defaultregistry` | Default registry |

### Examples

```bash
lazyoci config get artifact-dir
lazyoci config get artifactdir
lazyoci config get cache-dir
lazyoci config get default-registry
```

## set

Set a configuration value.

### Synopsis

```
lazyoci config set <key> <value> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<key>` | Configuration key | Required |
| `<value>` | Configuration value | Required |

**Argument validation:** ExactArgs(2)

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--create` | `false` | Create configuration if not exists |

### Examples

```bash
lazyoci config set artifact-dir /path/to/artifacts
lazyoci config set default-registry ghcr.io
lazyoci config set --create cache-dir /tmp/cache
```

## list

List all configuration values.

### Synopsis

```
lazyoci config list [flags]
```

### Examples

```bash
lazyoci config list
```

## path

Show configuration file path.

### Synopsis

```
lazyoci config path [flags]
```

### Examples

```bash
lazyoci config path
```