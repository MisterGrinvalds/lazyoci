---
title: Credential Chain
sidebar_position: 4
---

# Credential Chain

Modern credential management involves trying multiple sources in a specific order until one succeeds or all fail. The ChainedStore pattern in lazyoci embodies this approach, reflecting how real-world authentication actually works in complex environments.

## Why Chain Multiple Stores?

In practice, credentials might be stored in several places:
- Docker Desktop's secure storage
- Cloud provider CLI configurations (AWS, GCP, Azure)
- Kubernetes service account tokens
- CI/CD system environment variables
- Shared credential files

Rather than requiring users to configure which store to use for which registry, the ChainedStore tries each store in order until it finds working credentials. This approach prioritizes **convenience** and **robustness** over perfect efficiency.

## The ChainedStore Pattern

The ChainedStore acts as a **composite** that presents a single interface while managing multiple underlying stores:

```go
type ChainedStore struct {
    stores []CredentialStore
}

func (c *ChainedStore) Get(registry string) (Credential, error) {
    for _, store := range c.stores {
        cred, err := store.Get(registry)
        if err == ErrCredentialsNotFound || err == ErrNotImplemented {
            continue  // Try next store
        }
        if err != nil {
            return nil, err  // Real error, stop trying
        }
        return cred, nil  // Success
    }
    return nil, ErrCredentialsNotFound
}
```

This pattern enables **graceful degradation**—if one credential store is unavailable or misconfigured, the system continues with other stores.

## Error Propagation Strategy

The ChainedStore distinguishes between two types of errors:

### Expected Failures
- `ErrCredentialsNotFound` - This store doesn't have credentials for this registry
- `ErrNotImplemented` - This store doesn't support this registry type

These errors are **expected** and indicate "try the next store" rather than "stop processing."

### Actual Errors
- Network failures connecting to credential services
- Permission errors accessing credential files
- Malformed credential data

These errors indicate real problems and stop the chain immediately.

### Why This Distinction Matters

This error handling strategy reflects real-world credential scenarios:

**Mixed environments** where some stores are configured for some registries but not others. For example, a developer might have Docker Desktop credentials for Docker Hub but AWS CLI credentials for Amazon ECR.

**Partial deployments** where some credential stores are available in some environments but not others. A CI/CD system might have environment variables but no access to local credential files.

**Graceful degradation** where the system continues working even if some credential sources are temporarily unavailable.

## A Real-World Scenario

Consider a Docker Desktop user accessing multiple registries:

1. **User configuration**:
   ```json
   {
     "auths": {
       "https://index.docker.io/v1/": {}
     },
     "credsStore": "desktop"
   }
   ```

2. **lazyoci's credential chain**:
   - DockerConfigStore (reads config.json)
   - DockerDesktopStore (via credential helper)
   - EnvironmentStore (checks DOCKER_PASSWORD, etc.)
   - DefaultStore (anonymous access)

3. **Lookup for Docker Hub**:
   - DockerConfigStore: finds empty `{}` entry, delegates to credsStore
   - DockerDesktopStore: successfully retrieves credentials
   - **Result**: Authenticated access

4. **Lookup for private registry**:
   - DockerConfigStore: ErrCredentialsNotFound (no entry in auths)
   - DockerDesktopStore: ErrCredentialsNotFound (not configured)
   - EnvironmentStore: finds PRIVATE_REGISTRY_TOKEN
   - **Result**: Authenticated access with environment variable

5. **Lookup for public registry**:
   - DockerConfigStore: ErrCredentialsNotFound
   - DockerDesktopStore: ErrCredentialsNotFound  
   - EnvironmentStore: ErrCredentialsNotFound
   - DefaultStore: returns anonymous credentials
   - **Result**: Anonymous access

## Design Benefits

### User Experience
Users don't need to understand which credential store will be used for which registry. The system "just works" with their existing configuration.

### Environment Flexibility  
The same code works across different deployment environments (developer workstations, CI/CD systems, production containers) without requiring environment-specific configuration.

### Future-Proofing
New credential stores can be added to the chain without breaking existing functionality. The pattern scales naturally as new authentication methods emerge.

## Design Trade-offs

### Performance vs Convenience
The chain might try multiple stores before finding credentials, introducing latency. However, credential lookup is typically much faster than network operations, so this trade-off favors convenience.

### Debugging Complexity
When authentication fails, users need to understand which stores were tried and why they failed. lazyoci addresses this through detailed error messages that explain the chain traversal.

### Security Considerations
The chain might expose credentials from unexpected sources. For example, environment variables intended for one tool might be picked up by another. This is mitigated by explicit store ordering and clear documentation.

## Alternative Approaches

Other tools handle credential management differently:

**Single-source systems** require users to explicitly configure which credential store to use. This provides predictability but reduces convenience.

**Registry-specific configuration** where users configure credentials per registry. This provides fine-grained control but increases configuration complexity.

**Credential proxy systems** where a single service manages all credentials. This centralizes management but introduces additional infrastructure requirements.

The ChainedStore pattern reflects lazyoci's philosophy of **working with existing systems** rather than imposing new requirements. It acknowledges that real environments are messy and credential management is distributed across multiple tools and systems.

## Implications for Integration

Understanding the credential chain helps explain why lazyoci integrates well with existing container workflows. Rather than requiring users to reconfigure their authentication setup, it leverages whatever configuration already exists and gracefully falls back when specific configurations are unavailable.

This approach makes lazyoci a **good citizen** in complex container environments where multiple tools and systems need to share access to the same registries with minimal configuration overlap.