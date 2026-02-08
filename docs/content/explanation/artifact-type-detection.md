---
title: Artifact Type Detection
sidebar_position: 5
---

# Artifact Type Detection

The container ecosystem has evolved far beyond Docker images. Today's registries store Helm charts, software bills of materials (SBOMs), security signatures, WebAssembly modules, and many other artifact types. Understanding how lazyoci identifies these different types reveals the complexity and evolution of the container artifact ecosystem.

## Beyond Container Images

When OCI registries were first designed, they primarily stored container images. The registry API and manifest format were optimized for this use case. However, the content-addressable storage and standardized API proved valuable for distributing other types of software artifacts.

This evolution created a challenge: **how do you determine what kind of artifact you're looking at?** Unlike traditional package managers where file extensions provide hints (.deb, .rpm, .jar), OCI artifacts are identified by media types embedded in their metadata.

## Media Type Inspection Strategy

lazyoci examines three sources of media type information, in this order of priority:

1. **Manifest media type** - What kind of manifest structure is this?
2. **Config media type** - What kind of configuration does this artifact have?
3. **Layer media types** - What kinds of data layers does this contain?

This multi-level inspection handles cases where artifacts might have generic manifest types but specific config or layer types that reveal their true nature.

### Why This Approach?

Different artifact types use media types differently:

**Traditional container images** use well-established Docker or OCI media types throughout their structure.

**Helm charts** often use OCI manifest types but have distinctive layer media types like `application/vnd.cncf.helm.chart.content.v1.tar+gzip`.

**Security signatures** might use generic manifest types but have config media types that indicate they're signatures rather than executable artifacts.

## The Priority Order

lazyoci uses this priority order for artifact type detection:

1. **Helm** - Package manager for Kubernetes
2. **SBOM** (SPDX/CycloneDX) - Software bills of materials  
3. **Signature** (Cosign/Notary) - Security signatures
4. **Attestation** (in-toto/DSSE) - Security attestations
5. **WebAssembly** - WASM modules
6. **Container Image** - Traditional Docker/OCI images
7. **Unknown** - Unrecognized artifact types

### Why This Priority?

This ordering reflects both **specificity** and **ecosystem importance**:

**Helm charts** are highly specific - if something has Helm media types, it's almost certainly a Helm chart, not something else.

**Security artifacts** (SBOMs, signatures, attestations) are often overlaid on other artifact types. Detecting them early prevents misclassification as the artifacts they're securing.

**WebAssembly** represents a growing but distinct use case with specific runtime requirements.

**Container images** are the default fallback because they were the original use case and many tools generate OCI-compatible artifacts that are "image-like" even if they're not traditional container images.

## Case-Insensitive Detection

All media type matching is case-insensitive because different tools and registries have historically used inconsistent casing. This pragmatic approach prioritizes compatibility over strict standards compliance.

The detection also uses substring matching rather than exact matching, allowing for variations in vendor prefixes and version suffixes while still identifying the core artifact type.

## Detail Capture

Beyond the primary type, lazyoci captures **sub-type details**:
- SBOM format: "spdx" vs "cyclonedx"
- Signature system: "cosign" vs "notary"  
- Attestation format: "in-toto" vs "dsse"

This additional detail helps users understand not just what type of artifact they're working with, but which specific toolchain or standard was used to create it.

## Evolution and Future-Proofing

The artifact detection system reflects the container ecosystem's rapid evolution:

### Historical Context
- **2013-2016**: Primarily Docker images
- **2017-2019**: Helm charts, security scanning
- **2020-2022**: SBOMs, supply chain security
- **2023-present**: WebAssembly, AI model distribution

### Future Considerations

The detection system is designed to be **extensible**. As new artifact types emerge (ML models, configuration templates, policy definitions), they can be added to the priority list without breaking existing functionality.

This extensibility matters because the container registry ecosystem continues to evolve. What started as a way to distribute container images has become a general-purpose artifact distribution system.

## Alternative Detection Strategies

Other tools take different approaches to artifact type detection:

**Registry-side tagging** where registries provide artifact type metadata. This requires registry support and doesn't work with existing artifacts.

**Filename-based detection** where artifact types are inferred from repository names or tags. This is fragile and doesn't work with programmatically generated names.

**User-specified types** where users tell tools what type of artifact they're working with. This is accurate but requires additional user knowledge and configuration.

lazyoci's media type inspection approach balances **automation** with **accuracy**. It works with existing artifacts without requiring additional metadata or user input, while providing enough detail to enable type-specific behaviors.

## Implications for Tool Design

Understanding artifact type detection helps explain why modern container tools behave differently based on what they're examining. A tool that works perfectly with container images might need different behaviors for Helm charts or security signatures.

This type-aware approach represents the maturation of the container ecosystem from a single-purpose image distribution system to a multi-purpose artifact ecosystem with diverse use cases and requirements.