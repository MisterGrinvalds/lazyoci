---
title: Fix Connectivity Issues
description: Resolve network and registry access problems
sidebar_position: 9
---

# Fix Connectivity Issues

Diagnose and resolve network connectivity problems when accessing container registries. Follow these steps to identify and fix connection issues.

## Identify Connectivity Issues

### Common error messages

- `connection refused`
- `no route to host`
- `timeout`
- `network is unreachable`
- `certificate verify failed`
- `name resolution failed`

### Test basic connectivity

```bash
lazyoci registry test registry.example.com
```

If this fails with network errors, you have a connectivity issue.

## Check Network Basics

### Test DNS resolution

```bash
# Check if hostname resolves
nslookup registry.example.com

# Alternative check
dig registry.example.com
```

If DNS fails, you have a name resolution problem.

### Test network reachability

```bash
# Check if host is reachable
ping registry.example.com

# Check specific port
telnet registry.example.com 443
```

If ping fails but you expect the host to respond to ping, there may be firewall rules blocking ICMP.

### Test HTTP/HTTPS access

```bash
# Test HTTPS connectivity
curl -I https://registry.example.com/v2/

# Test HTTP connectivity (for insecure registries)
curl -I http://registry.example.com/v2/
```

## Diagnose Registry Endpoints

### Test registry API endpoint

```bash
# Test v2 API endpoint
curl https://registry.example.com/v2/
```

Should return `{}` for a healthy registry.

### Check registry version

```bash
# Check registry version header
curl -I https://registry.example.com/v2/ | grep -i docker
```

Look for `Docker-Distribution-Api-Version` header.

### Test with different schemes

```bash
# Try both HTTP and HTTPS
lazyoci registry test http://registry.example.com
lazyoci registry test https://registry.example.com
```

## Fix DNS Issues

### Check DNS configuration

```bash
# Check current DNS servers
cat /etc/resolv.conf
```

### Use alternative DNS

```bash
# Temporarily use Google DNS
sudo tee /etc/resolv.conf > /dev/null <<EOF
nameserver 8.8.8.8
nameserver 8.8.4.4
EOF

# Test resolution again
nslookup registry.example.com
```

### Add to hosts file

For development or internal registries:

```bash
# Add registry to hosts file
echo "192.168.1.100 registry.internal.com" | sudo tee -a /etc/hosts

# Test resolution
ping registry.internal.com
```

## Configure Proxy Settings

### Check current proxy configuration

```bash
# Check environment variables
echo $HTTP_PROXY
echo $HTTPS_PROXY  
echo $NO_PROXY
```

### Set proxy for current session

```bash
export HTTP_PROXY=http://proxy.corp.com:8080
export HTTPS_PROXY=http://proxy.corp.com:8080
export NO_PROXY=localhost,127.0.0.1,.internal.com
```

### Make proxy settings permanent

```bash
# Add to shell profile
tee -a ~/.bashrc > /dev/null <<EOF
export HTTP_PROXY=http://proxy.corp.com:8080
export HTTPS_PROXY=http://proxy.corp.com:8080
export NO_PROXY=localhost,127.0.0.1,.internal.com
EOF

source ~/.bashrc
```

### Test with proxy

```bash
# Test registry access through proxy
lazyoci registry test registry.example.com
```

### Configure Docker daemon proxy

```bash
# Create Docker daemon proxy configuration
sudo mkdir -p /etc/systemd/system/docker.service.d
sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf > /dev/null <<EOF
[Service]
Environment="HTTP_PROXY=http://proxy.corp.com:8080"
Environment="HTTPS_PROXY=http://proxy.corp.com:8080"
Environment="NO_PROXY=localhost,127.0.0.1,.docker.internal"
EOF

# Restart Docker
sudo systemctl daemon-reload
sudo systemctl restart docker
```

## Handle Firewall Issues

### Check firewall status

**Linux (ufw):**
```bash
sudo ufw status
```

**Linux (iptables):**
```bash
sudo iptables -L
```

**macOS:**
```bash
sudo pfctl -s all
```

### Test specific ports

```bash
# Test common registry ports
telnet registry.example.com 443   # HTTPS
telnet registry.example.com 80    # HTTP
telnet registry.example.com 5000  # Common Docker registry port
```

### Allow registry access through firewall

**Linux (ufw):**
```bash
# Allow outbound HTTPS
sudo ufw allow out 443

# Allow specific registry
sudo ufw allow out to registry.example.com port 443
```

**Corporate environments:**
Contact your network administrator to allow access to the registry.

## Fix TLS/Certificate Issues

### Test TLS connection

```bash
# Test TLS handshake
openssl s_client -connect registry.example.com:443 -servername registry.example.com
```

### Check certificate validity

```bash
# Check certificate details
echo | openssl s_client -connect registry.example.com:443 2>/dev/null | openssl x509 -noout -dates
```

### Skip certificate verification

For testing only (not recommended for production):

```bash
# Test with insecure flag
lazyoci registry add https://registry.example.com --insecure
lazyoci registry test https://registry.example.com
```

### Add custom CA certificate

If using internal CA:

```bash
# Add custom CA (Ubuntu/Debian)
sudo cp internal-ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates

# Test after adding CA
lazyoci registry test https://registry.example.com
```

## Test Local Registry Connectivity

### Start local test registry

```bash
# Run local registry for testing
docker run -d -p 5000:5000 --name test-registry registry:2
```

### Test local registry

```bash
# Test HTTP local registry
lazyoci registry add http://localhost:5000 --insecure
lazyoci registry test http://localhost:5000
```

### Push test image to local registry

```bash
# Push test image
docker tag alpine:latest localhost:5000/test:latest
docker push localhost:5000/test:latest

# Pull with lazyoci
lazyoci pull localhost:5000/test:latest --quiet
```

## Debug Network Routes

### Check routing table

```bash
# Check routes
route -n  # Linux
netstat -rn  # macOS/FreeBSD
```

### Trace network path

```bash
# Trace route to registry
traceroute registry.example.com
```

### Check network interfaces

```bash
# Check active network interfaces
ip addr show  # Linux
ifconfig     # macOS/FreeBSD
```

## Troubleshoot VPN Issues

### Test with VPN off

```bash
# Disconnect VPN and test
lazyoci registry test registry.example.com
```

### Configure split tunneling

For corporate VPNs, configure split tunneling to route registry traffic appropriately.

### Check VPN DNS

```bash
# Check DNS when VPN is connected
nslookup registry.example.com
```

VPN may override DNS servers and cause resolution issues.

## Verify Connectivity Fix

### Test registry endpoint

```bash
# Test basic connectivity
curl -I https://registry.example.com/v2/
```

Should return HTTP 200 or appropriate auth response.

### Test with lazyoci

```bash
# Test registry connectivity
lazyoci registry test registry.example.com
```

### Test full pull operation

```bash
# Test complete workflow
lazyoci pull registry.example.com/public-repo:latest --quiet
```

### Test from different locations

```bash
# Test from different network locations
# - Direct connection
# - Through corporate proxy  
# - From different geographic location
# - With/without VPN
```

All tests should complete without network errors.

:::tip Network Debugging
Use `curl` to test basic HTTP/HTTPS connectivity before troubleshooting lazyoci-specific issues. If curl fails, the problem is at the network level, not with lazyoci.
:::