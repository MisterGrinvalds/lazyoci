---
title: Understanding lazyoci
sidebar_position: 1
---

# Understanding lazyoci

This section provides deeper understanding of the concepts, architectures, and design decisions that shape lazyoci. These explanations will help you understand **why** things work the way they do, rather than just **how** to use them.

## Key Areas of Understanding

### [OCI Concepts](./oci-concepts.md)
The foundational concepts of container registries, repositories, manifests, and blobs. Understanding these building blocks helps explain why container tools work the way they do and how content addressing enables distributed, reliable artifact storage.

### [Authentication Architecture](./authentication-architecture.md)
Docker's credential system represents years of evolution in container tooling. Understanding its design helps explain why modern tools like Docker Desktop behave differently than earlier versions, and how lazyoci integrates seamlessly with existing workflows.

### [Artifact Type Detection](./artifact-type-detection.md)
The container ecosystem has expanded far beyond Docker images. This explanation covers how lazyoci determines what kind of artifact it's examining and why certain detection strategies take precedence over others.

### [Credential Chain](./credential-chain.md)
Modern credential management involves trying multiple sources in a specific order. Understanding the ChainedStore pattern explains how lazyoci can work with diverse credential configurations while maintaining predictable behavior.

### [Docker Load Conversion](./docker-load-conversion.md)
Two different but related standards—OCI Image Layout and Docker's save format—solve similar problems in incompatible ways. This explanation covers why the conversion is necessary and how lazyoci handles the complexity.

## Perspective on Design

lazyoci embodies a particular philosophy: **leverage existing standards and tools** rather than reinventing them. This means understanding and working with Docker's credential system, OCI specifications, and established tools like skopeo. The result is a tool that integrates naturally into existing container workflows while providing new capabilities.

This approach brings trade-offs. It means learning more about how existing systems work, but it also means better compatibility and less vendor lock-in. Understanding these explanations will help you make informed decisions about when and how to use lazyoci in your container workflows.