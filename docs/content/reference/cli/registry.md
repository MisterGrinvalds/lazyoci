---
title: registry
---

# registry

Manage OCI registries.

## Subcommands

- [`list`](#list) - List configured registries
- [`add`](#add) - Add registry
- [`remove`](#remove) - Remove registry
- [`test`](#test) - Test registry connection

## list

List all configured registries.

### Synopsis

```
lazyoci registry list [flags]
```

### Examples

```bash
lazyoci registry list
```

## add

Add a new registry.

### Synopsis

```
lazyoci registry add <url> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<url>` | Registry URL | Required |

**Argument validation:** ExactArgs(1)

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--name` | `-n` | `""` | Registry name |
| `--user` | `-u` | `""` | Username |
| `--pass` | `-p` | `""` | Password |
| `--insecure` | | `false` | Allow insecure connections |

### Examples

```bash
lazyoci registry add myregistry.com
lazyoci registry add --name "My Registry" myregistry.com
lazyoci registry add --user admin --pass secret myregistry.com
lazyoci registry add --insecure http://localhost:5000
```

## remove

Remove a registry.

### Synopsis

```
lazyoci registry remove <url> [flags]
```

**Aliases:** `rm`, `delete`

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<url>` | Registry URL | Required |

**Argument validation:** ExactArgs(1)

### Examples

```bash
lazyoci registry remove myregistry.com
lazyoci registry rm myregistry.com
lazyoci registry delete myregistry.com
```

## test

Test registry connection.

### Synopsis

```
lazyoci registry test <url> [flags]
```

### Arguments

| Argument | Description | Type |
|----------|-------------|------|
| `<url>` | Registry URL | Required |

**Argument validation:** ExactArgs(1)

### Examples

```bash
lazyoci registry test docker.io
lazyoci registry test myregistry.com
```