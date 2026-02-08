---
title: Connect to Cloud Registries
description: Authenticate with AWS ECR, Google Container Registry, Azure Container Registry, and DigitalOcean
sidebar_position: 3
---

# Connect to Cloud Registries

Connect lazyoci to major cloud container registries. Each provider has its own authentication method that works with Docker login.

## AWS Elastic Container Registry (ECR)

### Prerequisites

Install and configure the AWS CLI with appropriate permissions.

### Authenticate with ECR

```bash
# Get login token and authenticate
aws ecr get-login-password --region us-west-2 | \
docker login --username AWS --password-stdin 123456789012.dkr.ecr.us-west-2.amazonaws.com
```

Replace `us-west-2` with your region and `123456789012` with your AWS account ID.

### Verify access

```bash
lazyoci registry test 123456789012.dkr.ecr.us-west-2.amazonaws.com
```

### Pull ECR artifacts

```bash
lazyoci pull 123456789012.dkr.ecr.us-west-2.amazonaws.com/my-repo:latest
```

## Google Container Registry (GCR)

### Prerequisites

Install `gcloud` CLI and authenticate with your Google Cloud account.

### Authenticate with GCR

```bash
# Configure Docker to use gcloud as credential helper
gcloud auth configure-docker
```

This configures Docker to use `gcloud` for authentication with `gcr.io`, `us.gcr.io`, `eu.gcr.io`, and `asia.gcr.io`.

### Alternative: Direct token login

```bash
# Get access token and login
gcloud auth print-access-token | \
docker login -u oauth2accesstoken --password-stdin gcr.io
```

### Verify access

```bash
lazyoci registry test gcr.io
```

### Pull GCR artifacts

```bash
lazyoci pull gcr.io/my-project/my-repo:latest
```

## Google Artifact Registry

For the newer Artifact Registry service:

```bash
# Configure Docker authentication
gcloud auth configure-docker us-west1-docker.pkg.dev
```

Replace `us-west1` with your region.

## Azure Container Registry (ACR)

### Prerequisites

Install `az` CLI and authenticate with Azure.

### Authenticate with ACR

```bash
# Login to specific registry
az acr login --name myregistry
```

This logs you in to `myregistry.azurecr.io`.

### Alternative: Service principal login

```bash
# Using service principal
docker login myregistry.azurecr.io \
  --username <service-principal-id> \
  --password <service-principal-password>
```

### Verify access

```bash
lazyoci registry test myregistry.azurecr.io
```

### Pull ACR artifacts

```bash
lazyoci pull myregistry.azurecr.io/my-repo:latest
```

## DigitalOcean Container Registry

### Prerequisites

Install `doctl` CLI and authenticate with DigitalOcean.

### Authenticate with DOCR

```bash
# Login to registry
doctl registry login
```

This configures Docker to authenticate with `registry.digitalocean.com`.

### Alternative: Token-based login

```bash
# Using personal access token
echo $DIGITALOCEAN_ACCESS_TOKEN | \
docker login registry.digitalocean.com -u <your-email> --password-stdin
```

### Verify access

```bash
lazyoci registry test registry.digitalocean.com
```

### Pull DOCR artifacts

```bash
lazyoci pull registry.digitalocean.com/my-registry/my-repo:latest
```

## Token Refresh

Cloud registry tokens typically expire. Re-run the authentication commands when you encounter authentication errors.

### Automate token refresh

Consider setting up credential helpers for automatic token refresh:

**AWS ECR Helper:**
```bash
# Install ECR credential helper
go install github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cli/docker-credential-ecr-login@latest

# Configure in ~/.docker/config.json
{
  "credHelpers": {
    "123456789012.dkr.ecr.us-west-2.amazonaws.com": "ecr-login"
  }
}
```

**Google Cloud Helper:**
```bash
# Already configured by gcloud auth configure-docker
# Check ~/.docker/config.json for:
{
  "credHelpers": {
    "gcr.io": "gcloud",
    "us.gcr.io": "gcloud"
  }
}
```

## Verify Multiple Registries

Test access to all your configured registries:

```bash
# Test each registry
lazyoci registry test 123456789012.dkr.ecr.us-west-2.amazonaws.com
lazyoci registry test gcr.io
lazyoci registry test myregistry.azurecr.io
lazyoci registry test registry.digitalocean.com
```

All should return success messages if authentication is properly configured.

:::tip Automation Tip
Set up credential helpers for each cloud provider to avoid manual token refresh. They automatically handle token expiration and renewal.
:::