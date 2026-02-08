---
title: Configure Authentication
description: Set up credentials for accessing private registries
sidebar_position: 2
---

# Configure Authentication

Configure lazyoci to access private registries using Docker's credential system or explicit configuration.

## Use Docker Login (Recommended)

The simplest approach is using Docker's login system, which lazyoci automatically inherits.

### Log in to a registry

```bash
docker login registry.example.com
```

Enter your username and password when prompted. Docker stores these credentials and lazyoci will use them automatically.

### Verify authentication works

```bash
lazyoci registry test registry.example.com
```

You should see a success message indicating the registry is accessible.

## Configure Explicit Credentials

For registries where Docker login isn't suitable, configure credentials directly in lazyoci.

### Add registry with credentials

```bash
lazyoci registry add registry.example.com \
  --user your-username \
  --pass your-password
```

### Test the configuration

```bash
lazyoci registry test registry.example.com
```

## Use Credential Helpers

For enhanced security, configure Docker credential helpers that lazyoci will automatically use.

### Configure a credential helper

Edit `~/.docker/config.json`:

```json
{
  "credHelpers": {
    "registry.example.com": "secretservice"
  }
}
```

Or set a default helper for all registries:

```json
{
  "credsStore": "secretservice"
}
```

### Available credential helpers

Common credential helpers include:
- `docker-credential-secretservice` - Linux Secret Service API
- `docker-credential-osxkeychain` - macOS Keychain
- `docker-credential-wincred` - Windows Credential Manager
- `docker-credential-pass` - Unix pass utility

### Verify helper is working

```bash
lazyoci registry test registry.example.com
```

The authentication should work without prompting for credentials.

## Authentication Priority

lazyoci tries authentication methods in this order:

1. **Per-registry credential helpers** - `credHelpers` in Docker config
2. **Default credential helper** - `credsStore` in Docker config  
3. **Docker config auths** - Base64 encoded credentials from `docker login`
4. **Explicit credentials** - From `lazyoci registry add`
5. **Anonymous access** - For public registries

## Verify Your Setup

### Check Docker configuration

```bash
cat ~/.docker/config.json
```

Look for `auths`, `credHelpers`, or `credsStore` entries for your registry.

### Test registry access

```bash
lazyoci registry test your-registry.com
```

### Pull a test artifact

```bash
lazyoci pull your-registry.com/your-repo:tag --quiet
```

If authentication is working correctly, the pull should succeed without credential prompts.

:::tip Security Best Practice
Use credential helpers instead of storing passwords in plain text. They provide better security by integrating with your system's secure credential storage.
:::