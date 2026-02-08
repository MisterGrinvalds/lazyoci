---
title: Artifact Types
---

# Artifact Types

OCI artifact type classification and storage mapping.

## Artifact Type Constants

| Type Constant | Display Name | Short Name | Storage Directory |
|---------------|--------------|------------|------------------|
| `image` | Container Image | `image` | `oci/` |
| `helm` | Helm Chart | `helm` | `helm/` |
| `sbom` | SBOM | `sbom` | `sbom/` |
| `signature` | Signature | `sig` | `sig/` |
| `attestation` | Attestation | `att` | `att/` |
| `wasm` | WebAssembly | `wasm` | `wasm/` |
| `unknown` | Unknown | `?` | - |

## Type Detection

Artifact types are detected from OCI manifest media types and annotations.

### Container Images

**Display:** "Container Image"  
**Storage:** Stored in `oci/` subdirectory  
**Detection:** Standard OCI image media types

### Helm Charts

**Display:** "Helm Chart"  
**Storage:** Stored in `helm/` subdirectory  
**Detection:** Helm chart media types and annotations

### Software Bill of Materials (SBOM)

**Display:** "SBOM"  
**Storage:** Stored in `sbom/` subdirectory  
**Detection:** SBOM-specific media types

### Signatures

**Display:** "Signature"  
**Storage:** Stored in `sig/` subdirectory  
**Detection:** Cosign and other signature formats

### Attestations

**Display:** "Attestation"  
**Storage:** Stored in `att/` subdirectory  
**Detection:** In-toto and other attestation formats

### WebAssembly Modules

**Display:** "WebAssembly"  
**Storage:** Stored in `wasm/` subdirectory  
**Detection:** WASM media types

### Unknown Types

**Display:** "Unknown"  
**Short Name:** `?`  
**Storage:** Not stored separately  
**Usage:** Fallback for unrecognized artifact types