---
title: Docker Load Conversion
sidebar_position: 6
---

# Docker Load Conversion

The `docker load` command expects a specific format that differs from the OCI Image Layout specification. This incompatibility reflects the independent evolution of Docker's tooling and OCI standards, creating a need for conversion between formats.

## Two Standards, One Purpose

Both OCI Image Layout and Docker's save format solve the same fundamental problem: **how to represent container images on disk**. However, they evolved independently and made different design decisions.

### OCI Image Layout
```
layout/
├── oci-layout           # Version marker
├── index.json          # Entry point referencing manifests
└── blobs/sha256/
    ├── <digest>         # Manifest files
    ├── <digest>         # Config files  
    └── <digest>         # Layer files
```

The OCI layout is **content-addressed** throughout—every file is named by its SHA256 digest, making the structure self-verifying and enabling efficient deduplication.

### Docker Save Format
```
save/
├── manifest.json       # Array of image descriptors
├── <sha>.json         # Config files (named by hash)
└── <sha>/
    └── layer.tar      # Layer files (directory per layer)
```

Docker's format is **partially content-addressed**—config files use hash names, but layers are stored in directories named by hash with fixed filenames inside.

## Why the Incompatibility Exists

### Historical Context

Docker's save format predates OCI by several years. When Docker created this format, there was no standard for container image storage, so they designed a format that met their specific needs:

- **Human-readable structure** for debugging
- **Streaming support** for `docker save | docker load`  
- **Backward compatibility** with existing Docker versions

OCI Image Layout was designed later as an **interoperability standard**, prioritizing:

- **Content addressing** for integrity verification
- **Tooling independence** for multi-vendor ecosystems
- **Efficient storage** for registry backends

### Design Philosophies

These different priorities led to incompatible designs. Docker prioritized **operational convenience** while OCI prioritized **architectural purity**.

## Conversion Challenges

Converting between formats involves several technical challenges:

### Layer Compression
OCI layouts often store compressed layers (gzipped tarballs), while Docker load expects uncompressed layer.tar files. This means:

- **Decompression required**: Gzipped layers must be decompressed during conversion
- **Temporary storage**: Decompressed layers need temporary disk space  
- **Performance impact**: Compression/decompression adds CPU overhead

### Directory Structure Mapping
The two formats organize files differently:
- OCI uses flat blob storage with hash-named files
- Docker uses nested directories with fixed filenames

Converting requires **restructuring** the entire layout, not just renaming files.

### Metadata Transformation  
The manifest formats differ:
- OCI index.json references manifests by digest
- Docker manifest.json includes repository tags and file paths

Conversion requires **semantic translation** between metadata representations.

## lazyoci's Conversion Strategy

lazyoci implements a **multi-strategy approach** to Docker load conversion:

### Preferred: skopeo
When available, lazyoci delegates to skopeo:
```bash
skopeo copy oci:<path> docker-daemon:<ref>
```

This approach leverages skopeo's mature, battle-tested conversion logic. skopeo handles all the edge cases, compression scenarios, and format variations that have been discovered over years of production use.

### Fallback: Manual Conversion
When skopeo isn't available, lazyoci implements manual conversion:

1. **Parse OCI layout**: Read index.json, resolve manifest and config
2. **Create Docker structure**: Build manifest.json with required metadata  
3. **Copy and decompress layers**: Extract layers to Docker's directory structure
4. **Handle temporary files**: Clean up decompressed layers after conversion

### Why This Dual Approach?

**Reliability**: skopeo has handled countless edge cases in production environments. Using it when available provides maximum compatibility.

**Independence**: Manual conversion ensures lazyoci works even in environments where skopeo isn't available or can't be installed.

**Performance**: skopeo is optimized for this exact use case and typically performs better than general-purpose conversion code.

## The Decompression Requirement

Docker's load format specifically requires uncompressed layer.tar files. This requirement reflects Docker's internal architecture:

- **Streaming design**: Docker can stream uncompressed layers directly to the filesystem
- **Memory efficiency**: No need to buffer entire compressed layers in memory
- **Simplicity**: Avoid compression format detection and error handling

However, this requirement creates challenges:

**Storage overhead**: Temporary decompressed layers can be significantly larger than compressed layers.

**I/O amplification**: Read compressed, write uncompressed, then read uncompressed again during load.

**Cleanup complexity**: Temporary files must be properly cleaned up even if conversion fails.

## Why Not Direct Support?

You might wonder: why doesn't Docker just support OCI Image Layout directly? Several factors explain this design choice:

### Backward Compatibility
Changing `docker load` to accept different formats would break existing scripts, CI/CD pipelines, and tooling that depend on the current behavior.

### Implementation Complexity  
Supporting multiple input formats would complicate Docker's codebase and increase the surface area for bugs.

### Standard Evolution
When OCI Image Layout was standardized, Docker already had an established format with years of production usage. The cost of migration outweighed the benefits.

## Alternative Approaches

Other tools handle this incompatibility differently:

**Format detection** where tools automatically detect input format and convert as needed. This provides convenience but adds complexity and potential failure modes.

**Universal formats** where tools define their own format that can represent both Docker and OCI layouts. This solves the conversion problem but creates a third format to maintain.

**Registry-based workflows** where images are pushed to a registry and pulled by Docker, avoiding local format conversion entirely. This works well for networked environments but doesn't help with offline scenarios.

## Implications for Tooling

The Docker load conversion requirement illustrates broader challenges in container tooling:

**Format proliferation**: Multiple standards for similar functionality create integration overhead.

**Tool coupling**: Dependencies on specific tools (like skopeo) create deployment and maintenance complexities.

**User experience**: Format incompatibilities create friction for users who just want to move images between tools.

Understanding these challenges helps explain why lazyoci prioritizes **integration over innovation**—working with existing formats rather than creating new ones, even when the existing formats have limitations.