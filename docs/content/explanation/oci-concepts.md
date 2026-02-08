---
title: OCI Registry Concepts
sidebar_position: 2
---

# OCI Registry Concepts

Understanding Open Container Initiative (OCI) concepts is crucial for grasping how modern container tools work. These concepts form the foundation of container image distribution and explain why tools like lazyoci can work across different registries and platforms.

## The Registry Hierarchy

Container registries organize content in a hierarchical structure that mirrors how we think about software distribution:

- **Registries** are like package repositories (think npm registry or Maven Central)
- **Repositories** within registries are like individual packages (think `lodash` in npm)
- **Tags and digests** within repositories are like package versions (think `lodash@4.17.21`)

This hierarchy enables distributed storage where anyone can run a registry, and clients can discover and retrieve content from any compatible registry.

### Why Content-Addressing Matters

OCI uses content-addressing through cryptographic digests (SHA256 hashes). This means:

- Content is identified by its hash, not its location
- Identical content has identical identifiers across all registries
- Content cannot be tampered with without changing its identifier
- Clients can verify integrity without trusting the transport

This design enables features like registry mirrors, content deduplication, and secure distribution across untrusted networks.

## Manifests and Blobs

The OCI specification separates metadata (manifests) from data (blobs):

**Manifests** describe what an artifact contains:
- Configuration blob digest and media type
- Layer blob digests and media types  
- Platform information (architecture, OS)
- Annotations (arbitrary metadata)

**Blobs** contain the actual data:
- Configuration JSON (how to run the container)
- Layer tarballs (filesystem changes)
- Arbitrary data (for non-image artifacts)

This separation allows efficient operations:
- Download only necessary layers
- Share common layers across different images
- Inspect metadata without downloading content
- Support different artifact types beyond container images

## OCI Image Layout

The OCI Image Layout specification defines how to store OCI artifacts on disk:

```
layout/
├── oci-layout           # JSON file marking this as OCI layout
├── index.json          # Points to manifests
└── blobs/sha256/
    ├── <digest>         # Manifest files
    ├── <digest>         # Config files
    └── <digest>         # Layer files
```

This format enables:
- **Portability**: Move artifacts between systems without re-uploading
- **Verification**: Check integrity using filename digests
- **Efficiency**: Share storage between different tools
- **Interoperability**: Standard format for local artifact storage

## Docker Hub's Special Cases

Docker Hub predates OCI and has evolved over time, creating special handling requirements:

### Registry URL Mapping
- Users type `docker.io` but clients connect to `registry-1.docker.io`
- Authentication happens against `https://index.docker.io/v1/`
- This split reflects Docker Hub's evolution from a single service to distributed infrastructure

### Repository Naming
- Official images like `nginx` are actually `library/nginx`
- This `library/` prefix is added automatically for single-name repositories
- User and organization repositories use full paths like `user/app`

### Why These Differences Exist

Docker Hub's special cases reflect its history as the original container registry. When Docker was created, there was no OCI specification, so Docker Hub developed its own conventions. As the ecosystem standardized around OCI, Docker Hub maintained backward compatibility while adopting new standards.

Understanding these quirks helps explain why container tools need special handling for Docker Hub and why newer registries (like GitHub Container Registry or AWS ECR) follow more consistent patterns.

## Media Types and Evolution

Media types identify what kind of content you're working with:
- `application/vnd.docker.distribution.manifest.v2+json` - Docker v2 manifest
- `application/vnd.oci.image.manifest.v1+json` - OCI image manifest  
- `application/vnd.cncf.helm.chart.content.v1.tar+gzip` - Helm chart
- `application/vnd.cyclonedx+json` - CycloneDX SBOM

This extensible system allows the same registry infrastructure to store different types of artifacts while maintaining compatibility with existing tooling.

The evolution from Docker-specific media types to OCI standards reflects the container ecosystem's maturation from a single-vendor solution to an open, interoperable standard.