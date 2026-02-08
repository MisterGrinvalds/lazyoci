---
title: Registry Compatibility
---

# Registry Compatibility

OCI registry compatibility and configuration notes.

## Default Registries

lazyoci includes these registries by default:

| Name | URL | Authentication |
|------|-----|----------------|
| Docker Hub | `docker.io` | Optional |
| Quay.io | `quay.io` | Optional |
| GitHub Packages | `ghcr.io` | Optional |

## Registry Types

### Docker Hub

**URL:** `docker.io`  
**Protocol:** HTTPS  
**Authentication:** Optional (public repositories), Required (private repositories)  
**Notes:** Default registry for unqualified image references

### GitHub Container Registry

**URL:** `ghcr.io`  
**Protocol:** HTTPS  
**Authentication:** Personal Access Token or GitHub credentials  
**Notes:** Supports OCI artifacts, private repositories require authentication

### Quay.io

**URL:** `quay.io`  
**Protocol:** HTTPS  
**Authentication:** Optional (public repositories), Required (private repositories)  
**Notes:** Full OCI artifact support

### Harbor

**Protocol:** HTTPS (typically)  
**Authentication:** Username/password or token-based  
**Notes:** Self-hosted, full OCI artifact support  
**Configuration:** Requires custom URL in registry configuration

### Local Registries

**Protocol:** HTTP (typically) or HTTPS  
**Authentication:** Variable  
**Configuration Requirements:**
- Set `insecure: true` for HTTP registries
- Custom URL configuration required
- May require custom certificates for HTTPS

## Authentication Methods

### Username/Password

Standard authentication for most registries.

```yaml
registries:
  - name: "Private Registry"
    url: "registry.example.com"
    username: "user"
    password: "password"
```

### Token-based

For registries supporting token authentication.

```yaml
registries:
  - name: "Token Registry"
    url: "registry.example.com"
    username: "token"
    password: "your-access-token"
```

### Insecure Connections

For local or development registries without TLS.

```yaml
registries:
  - name: "Local Registry"
    url: "localhost:5000"
    insecure: true
```

## OCI Artifact Support

| Registry | Container Images | Helm Charts | SBOM | Signatures | Attestations | WASM |
|----------|------------------|-------------|------|------------|--------------|------|
| Docker Hub | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| GitHub Container Registry | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Quay.io | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Harbor | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Distribution v2.7+ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Common Configuration

### Docker Hub with Authentication

```yaml
registries:
  - name: "Docker Hub"
    url: "docker.io"
    username: "your-username"
    password: "your-password-or-token"
```

### GitHub Container Registry

```yaml
registries:
  - name: "GitHub Packages"
    url: "ghcr.io"
    username: "your-github-username"
    password: "your-personal-access-token"
```

### Self-hosted Harbor

```yaml
registries:
  - name: "Company Harbor"
    url: "harbor.company.com"
    username: "your-username"
    password: "your-password"
```

### Local Development Registry

```yaml
registries:
  - name: "Local Dev"
    url: "localhost:5000"
    insecure: true
```