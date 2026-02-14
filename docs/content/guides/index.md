---
title: How-to Guides
description: Task-oriented guides for accomplishing specific goals with lazyoci
sidebar_position: 1
---

# How-to Guides

These guides help you accomplish specific tasks with lazyoci. Each guide assumes you understand the basics and focuses on practical steps to achieve your goal.

## Authentication & Access

- [**Configure Authentication**](authentication.md) - Set up credentials for private registries
- [**Connect to Cloud Registries**](cloud-registries.md) - Authenticate with AWS ECR, GCR, Azure ACR, and DigitalOcean
- [**Work with Insecure Registries**](insecure-registries.md) - Handle HTTP-only and self-signed certificate registries

## Image Operations

- [**Pull Images to Docker**](pulling-to-docker.md) - Download artifacts and load them directly into Docker
- [**Pull Multi-platform Images**](multi-platform-images.md) - Target specific architectures and operating systems
- [**Configure Custom Storage**](custom-storage.md) - Change where artifacts are cached locally

## Building & Pushing

- [**Build Artifacts**](building-artifacts.md) - Build container images, Helm charts, and generic artifacts with `.lazy` configs
- [**Push to Registries**](pushing-to-registries.md) - Push built artifacts to Docker Hub, GHCR, ECR, and other OCI registries
- [**CI/CD with GitHub Actions**](ci-cd-github-actions.md) - Automate builds with the lazyoci GitHub Action and semver detection

## Troubleshooting

- [**Fix Authentication Failures**](troubleshooting-auth.md) - Diagnose and resolve credential issues
- [**Fix Connectivity Issues**](troubleshooting-connectivity.md) - Resolve network and registry access problems
- [**Fix Docker Load Failures**](troubleshooting-docker.md) - Solve problems loading artifacts into Docker

:::tip Quick Start
New to lazyoci? Start with [Configure Authentication](authentication.md) to set up access to your registries, then try [Pull Images to Docker](pulling-to-docker.md) for your first artifact download.
:::