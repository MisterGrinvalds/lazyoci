---
title: Authentication Architecture
sidebar_position: 3
---

# Authentication Architecture

Docker's credential system reflects the evolution of container tooling from simple command-line tools to complex, integrated development environments. Understanding this evolution helps explain why modern Docker installations behave differently than older ones and how tools like lazyoci integrate with existing workflows.

## The Evolution of Docker Credentials

### Early Days: Simple Storage
Originally, Docker stored credentials directly in `~/.docker/config.json`:
```json
{
  "auths": {
    "https://index.docker.io/v1/": {
      "auth": "dXNlcjpwYXNz"  // base64 encoded "user:pass"
    }
  }
}
```

This approach was simple but had security limitations—credentials were stored in plain text (base64 is encoding, not encryption) in a file that could be accidentally shared or compromised.

### Credential Helpers: External Security
Docker evolved to support external credential helpers:
```json
{
  "auths": {},
  "credsStore": "desktop",
  "credHelpers": {
    "gcr.io": "gcr"
  }
}
```

Credential helpers are external programs that handle secure credential storage. When Docker needs credentials for a registry, it executes `docker-credential-<name> get`, pipes the registry URL to stdin, and parses JSON from stdout.

## Why the Current Design Exists

### Security Through Delegation
By delegating credential storage to external helpers, Docker can leverage platform-specific secure storage:
- **macOS**: Keychain
- **Windows**: Credential Manager  
- **Linux**: Various keyrings (gnome-keyring, KDE Wallet)

This delegation means credentials are protected by the operating system's security mechanisms rather than being stored in plain text files.

### Flexibility for Different Workflows
The system supports multiple patterns:
- **Global helper** (`credsStore`): One helper for all registries
- **Per-registry helpers** (`credHelpers`): Different helpers for different registries
- **Direct storage** (`auths`): Backward compatibility for simple cases

This flexibility accommodates different organizational policies and security requirements.

## Docker Desktop's Pattern

Modern Docker Desktop installations typically show:
```json
{
  "auths": {
    "https://index.docker.io/v1/": {}
  },
  "credsStore": "desktop"
}
```

The empty `auths` entries serve as **registry registration** rather than credential storage. They tell Docker "I have credentials for this registry" without storing the actual credentials. The real credentials live in the `desktop` credential helper.

### Why This Design?

This pattern provides several benefits:
- **Security**: Credentials stored in secure system storage
- **User Experience**: Single sign-on integration with Docker Desktop
- **Compatibility**: Tools can discover which registries have credentials without accessing the credentials themselves

## Registry URL Normalization

Different contexts use different registry URL formats:
- **Docker CLI**: `docker.io`
- **Authentication**: `https://index.docker.io/v1/`
- **API calls**: `registry-1.docker.io`

Docker's credential system normalizes these URLs for consistent lookup. This normalization explains why credential helpers receive standardized URLs regardless of how users specify them.

## lazyoci's Integration Strategy

lazyoci doesn't reinvent credential management—it integrates with Docker's existing system. This approach provides several advantages:

### Seamless User Experience
Users don't need to configure credentials separately for lazyoci. If their Docker setup works, lazyoci works.

### Security Consistency
lazyoci leverages the same secure credential storage that users have already configured, maintaining consistent security posture.

### Organizational Compatibility
In enterprise environments where credential management is centrally configured, lazyoci automatically inherits those policies.

## Design Trade-offs

This integration approach brings trade-offs:

### Benefits
- No additional credential configuration
- Consistent with user expectations
- Leverages proven security mechanisms
- Works with existing organizational policies

### Limitations
- Dependent on Docker's credential system
- Complex debugging when credential lookup fails
- Platform-specific behaviors (especially on different Linux distributions)

## Alternative Approaches

Other container tools take different approaches:

**Self-contained credential storage** (like some CI/CD tools) stores credentials independently, providing predictable behavior at the cost of additional configuration.

**Cloud-specific authentication** (like cloud provider CLIs) integrates with specific identity systems, providing seamless experience within one ecosystem but requiring different configuration for different clouds.

lazyoci's choice to integrate with Docker's system reflects a philosophy of **working with existing tools** rather than creating new patterns. This choice prioritizes user experience and compatibility over complete control of the authentication flow.