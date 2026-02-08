---
title: Fix Authentication Failures
description: Diagnose and resolve credential issues with registries
sidebar_position: 8
---

# Fix Authentication Failures

Diagnose and resolve authentication problems when accessing private registries. Follow these steps systematically to identify and fix credential issues.

## Identify Authentication Issues

### Common error messages

- `unauthorized: authentication required`
- `forbidden: insufficient scope` 
- `401 Unauthorized`
- `403 Forbidden`
- `no basic auth credentials`

### Test registry connectivity

```bash
lazyoci registry test your-registry.com
```

If this fails with auth errors, you have a credential problem.

## Check Docker Login Status

Verify your Docker authentication configuration.

### Check logged-in registries

```bash
# View current Docker config
cat ~/.docker/config.json
```

Look for `auths` entries for your registry.

### Test Docker login

```bash
# Try logging in with Docker
docker login your-registry.com
```

Enter your credentials and verify the login succeeds.

### Verify Docker pull works

```bash
# Test with Docker directly
docker pull your-registry.com/your-repo:tag
```

If Docker can pull but lazyoci cannot, the issue is with credential inheritance.

## Check Credential Helpers

Diagnose issues with Docker credential helpers.

### View credential helper configuration

```bash
# Check for credential helpers in Docker config
cat ~/.docker/config.json | jq '.credHelpers, .credsStore'
```

### Test credential helper manually

```bash
# Test specific credential helper
echo "your-registry.com" | docker-credential-osxkeychain get
```

Replace `osxkeychain` with your configured helper. Should return JSON with credentials.

### Common credential helpers

**macOS Keychain:**
```bash
# Test macOS keychain helper
echo "your-registry.com" | docker-credential-osxkeychain get
```

**Linux Secret Service:**
```bash
# Test Linux secret service helper
echo "your-registry.com" | docker-credential-secretservice get
```

**Windows Credential Manager:**
```bash
# Test Windows credential manager
echo "your-registry.com" | docker-credential-wincred.exe get
```

### Fix credential helper issues

```bash
# Reinstall credential helper (macOS example)
brew reinstall docker-credential-helper

# Re-login after fixing helper
docker login your-registry.com
```

## Inspect Docker Config

Examine your Docker configuration for issues.

### Check config file location

```bash
# Check if custom Docker config location is set
echo $DOCKER_CONFIG
# Default is ~/.docker if unset
```

### Validate JSON format

```bash
# Check for JSON syntax errors
cat ~/.docker/config.json | jq .
```

Fix any JSON syntax errors before proceeding.

### Check registry key formats

Docker config stores auths with different key formats:

```bash
# Check all auth entries
cat ~/.docker/config.json | jq '.auths | keys[]'
```

Common formats:
- `registry.com`
- `https://registry.com`
- `https://registry.com/v1/`
- `https://registry.com/v2/`

### Fix registry key mismatch

If your registry appears with a different format than expected:

```bash
# Example: Fix registry key format
# Back up config first
cp ~/.docker/config.json ~/.docker/config.json.backup

# Edit to use correct format (replace with your registry)
cat ~/.docker/config.json | jq '.auths["registry.com"] = .auths["https://registry.com"] | del(.auths["https://registry.com"])' > /tmp/fixed-config.json
mv /tmp/fixed-config.json ~/.docker/config.json
```

## Test Registry Authentication

Verify specific registry authentication.

### Use registry test command

```bash
# Test specific registry
lazyoci registry test your-registry.com
```

### Test with curl

```bash
# Get auth token manually
auth=$(echo -n "username:password" | base64)
curl -H "Authorization: Basic $auth" https://your-registry.com/v2/
```

Should return `{}` for successful auth.

### Test Docker Hub authentication

Docker Hub uses special registry URLs:

```bash
# Test Docker Hub access
lazyoci registry test https://index.docker.io/v1/

# Alternative test
docker login
lazyoci pull your-username/your-repo:tag
```

## Fix Common Authentication Issues

### Issue: Credential helper not found

```bash
# Error: credential helper not found
# Fix: Install the helper
brew install docker-credential-helper  # macOS
```

### Issue: Invalid or expired credentials

```bash
# Fix: Re-login to refresh credentials
docker login your-registry.com
```

### Issue: Wrong registry URL

```bash
# Fix: Check registry URL format
# Use registry test to verify correct URL
lazyoci registry test https://registry.example.com
lazyoci registry test registry.example.com
```

### Issue: Multi-factor authentication

For registries requiring MFA:

```bash
# Use app-specific password or token
docker login your-registry.com
# Enter username: your-username
# Enter password: <app-specific-token>
```

### Issue: Corporate proxy or firewall

```bash
# Configure proxy for Docker
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf > /dev/null <<EOF
[Service]
Environment="HTTP_PROXY=http://proxy.corp.com:8080"
Environment="HTTPS_PROXY=http://proxy.corp.com:8080"
Environment="NO_PROXY=localhost,127.0.0.1"
EOF

# Restart Docker
sudo systemctl daemon-reload
sudo systemctl restart docker

# Re-login
docker login your-registry.com
```

## Add Explicit Credentials

When Docker integration fails, configure credentials directly.

### Add registry credentials

```bash
lazyoci registry add your-registry.com \
  --user your-username \
  --pass your-password
```

### Test explicit credentials

```bash
lazyoci registry test your-registry.com
```

### Store credentials securely

Instead of passwords, use tokens when available:

```bash
# Use access tokens instead of passwords
lazyoci registry add your-registry.com \
  --user your-username \
  --pass ghp_your_github_token  # GitHub example
```

## Debug Authentication Chain

Understand which authentication method lazyoci is using.

### Check authentication priority

lazyoci tries authentication in this order:
1. Per-registry credential helpers (`credHelpers`)
2. Default credential helper (`credsStore`)
3. Docker config auths (`auths`)
4. Explicit lazyoci credentials
5. Anonymous access

### Force specific authentication method

```bash
# Clear Docker auth to test explicit credentials
mv ~/.docker/config.json ~/.docker/config.json.backup

# Add explicit credentials
lazyoci registry add your-registry.com --user username --pass password

# Test
lazyoci registry test your-registry.com

# Restore Docker config
mv ~/.docker/config.json.backup ~/.docker/config.json
```

## Verify Authentication Fix

### Test full workflow

```bash
# Test registry access
lazyoci registry test your-registry.com

# Test artifact pull
lazyoci pull your-registry.com/your-repo:tag --quiet

# Test Docker integration
lazyoci pull your-registry.com/your-repo:tag --docker --quiet
```

### Check Docker images

```bash
# Verify image loaded successfully
docker images | grep your-repo
```

All commands should complete without authentication errors.

:::tip Quick Debug
Start with `lazyoci registry test <registry>` to isolate authentication issues. If this fails, the problem is with credentials, not with the pull operation itself.
:::