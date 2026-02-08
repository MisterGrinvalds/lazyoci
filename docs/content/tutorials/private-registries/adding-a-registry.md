---
sidebar_position: 1
title: Adding a Registry
---

# Adding a Private Registry

Let's add your organization's private registry to lazyoci. We'll use authentication and verify the connection works.

## Add with Authentication

Use the `registry add` command with your credentials. Replace the URL and credentials with your actual registry details:

```bash
lazyoci registry add https://my-company-registry.com \
  --name "My Company" \
  --user your-username \
  --pass your-password
```

You'll see confirmation:
```
✓ Added registry https://my-company-registry.com as "My Company"
```

## For Self-Signed Certificates

If your registry uses self-signed certificates, add the `--insecure` flag:

```bash
lazyoci registry add https://internal-registry.local:5000 \
  --name "Internal Registry" \
  --user admin \
  --pass secret123 \
  --insecure
```

## Test the Connection

Verify lazyoci can connect to your registry:

```bash
lazyoci registry test https://my-company-registry.com
```

If successful, you'll see:
```
✓ Connection to https://my-company-registry.com successful
✓ Authentication verified
```

If there are issues, you'll see specific error messages to help troubleshoot.

## View All Registries

Check that your registry is now in the list:

```bash
lazyoci registry list
```

You'll see output like:
```
NAME                URL                              AUTH
Docker Hub          https://docker.io                No
Quay.io            https://quay.io                   No  
GitHub Packages    https://ghcr.io                   No
My Company         https://my-company-registry.com   Yes
```

## Remove a Registry (if needed)

If you made a mistake, you can remove and re-add:

```bash
lazyoci registry remove https://my-company-registry.com
```

## What You've Done

You've successfully:
- ✅ Added a private registry with authentication
- ✅ Tested the connection to verify credentials
- ✅ Viewed the registry in your list
- ✅ Learned about insecure connections for self-signed certs

Next, let's see how [Docker credentials work automatically](./docker-credentials.md).