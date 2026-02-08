---
title: Work with Insecure Registries
description: Handle HTTP-only registries and self-signed certificates
sidebar_position: 4
---

# Work with Insecure Registries

Configure lazyoci to work with registries that use HTTP or have self-signed certificates, common in development and internal environments.

## HTTP-Only Registries

For registries running on HTTP instead of HTTPS.

### Add HTTP registry

```bash
lazyoci registry add http://localhost:5000 --insecure
```

The `--insecure` flag allows HTTP connections and skips TLS certificate verification.

### Test HTTP registry

```bash
lazyoci registry test http://localhost:5000
```

### Pull from HTTP registry

```bash
lazyoci pull localhost:5000/my-repo:latest
```

Note: You can omit the `http://` scheme when pulling - lazyoci will try HTTPS first, then HTTP.

## Self-Signed Certificates

For registries with self-signed or invalid TLS certificates.

### Add registry with insecure TLS

```bash
lazyoci registry add https://registry.internal.com --insecure
```

The `--insecure` flag skips certificate validation while still using HTTPS.

### Alternative: Configure Docker daemon

If using Docker integration, also configure Docker to trust the registry:

```bash
# Create daemon configuration
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "insecure-registries": [
    "localhost:5000",
    "registry.internal.com"
  ]
}
EOF

# Restart Docker daemon
sudo systemctl restart docker
```

### Test insecure HTTPS registry

```bash
lazyoci registry test https://registry.internal.com
```

## Local Development Setup

Set up a local registry for development and testing.

### Run local registry

```bash
# Start local HTTP registry
docker run -d -p 5000:5000 --name registry registry:2
```

### Configure lazyoci

```bash
# Add local registry
lazyoci registry add http://localhost:5000 --insecure

# Test connectivity  
lazyoci registry test http://localhost:5000
```

### Push test image

```bash
# Tag and push image to local registry
docker tag alpine:latest localhost:5000/alpine:test
docker push localhost:5000/alpine:test

# Pull with lazyoci
lazyoci pull localhost:5000/alpine:test
```

## Corporate Environment

Handle registries behind corporate firewalls or with internal CA certificates.

### Add registry with authentication

```bash
lazyoci registry add https://registry.corp.com \
  --insecure \
  --user your-username \
  --pass your-password
```

### Configure proxy (if needed)

Set environment variables for proxy access:

```bash
export HTTP_PROXY=http://proxy.corp.com:8080
export HTTPS_PROXY=http://proxy.corp.com:8080
export NO_PROXY=localhost,127.0.0.1

# Test registry access
lazyoci registry test https://registry.corp.com
```

## Add Custom CA Certificate

For registries with certificates signed by internal CAs.

### System-wide CA installation

**Ubuntu/Debian:**
```bash
# Copy CA certificate
sudo cp your-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Restart Docker daemon
sudo systemctl restart docker
```

**CentOS/RHEL:**
```bash
# Copy CA certificate
sudo cp your-ca.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust

# Restart Docker daemon  
sudo systemctl restart docker
```

**macOS:**
```bash
# Add to system keychain
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain your-ca.crt
```

### Test with custom CA

```bash
# Should now work without --insecure
lazyoci registry add https://registry.internal.com
lazyoci registry test https://registry.internal.com
```

## Troubleshoot Certificate Issues

### Check TLS connection manually

```bash
# Test TLS handshake
openssl s_client -connect registry.example.com:443 -servername registry.example.com
```

### Verify certificate details

```bash
# Check certificate information
echo | openssl s_client -connect registry.example.com:443 2>/dev/null | \
openssl x509 -noout -text
```

### Test with curl

```bash
# Test HTTPS without verification
curl -k https://registry.example.com/v2/

# Test with proper certificates
curl https://registry.example.com/v2/
```

If curl works with `-k` but fails without it, you have a certificate validation issue.

## Security Considerations

:::warning Security Risk
The `--insecure` flag disables TLS certificate verification, making connections vulnerable to man-in-the-middle attacks. Only use this for trusted internal networks or development environments.
:::

### Best practices

1. **Use HTTPS with valid certificates** in production
2. **Add custom CA certificates** instead of using `--insecure` when possible  
3. **Limit insecure registries** to specific internal networks
4. **Use authentication** even with insecure connections
5. **Monitor access logs** for unusual activity

### Verify your setup

```bash
# Test all configured registries
lazyoci registry test http://localhost:5000
lazyoci registry test https://registry.internal.com

# Pull test artifacts
lazyoci pull localhost:5000/test:latest
lazyoci pull registry.internal.com/app:latest --quiet
```

Both commands should complete successfully without certificate errors.