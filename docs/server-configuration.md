# Server Configuration Guide

## Overview

The Freightliner API server now supports flexible address binding and external URL configuration for better deployment flexibility.

## Configuration Options

### Host Binding

The server can bind to different network interfaces:

#### Default (localhost)
```bash
freightliner serve --host localhost --port 8080
```
- Binds to: `localhost:8080`
- Access: Only from the local machine
- Security: Most secure, default setting

#### All Interfaces (0.0.0.0)
```bash
freightliner serve --host 0.0.0.0 --port 8080
```
- Binds to: `0.0.0.0:8080` (all network interfaces)
- Access: From any network interface on the machine
- Security: Less secure, use with firewall/security groups

#### Specific IP Address
```bash
freightliner serve --host 192.168.1.100 --port 8080
```
- Binds to: `192.168.1.100:8080`
- Access: Only through the specified IP
- Security: Medium, useful for multi-homed systems

### External URL

For deployment behind load balancers, reverse proxies, or with custom domains:

```bash
freightliner serve --host 0.0.0.0 --port 8080 --external-url https://api.example.com
```

This setting:
- Provides the correct URL in logs and responses
- Useful for documentation and API discovery
- Handles port forwarding scenarios

### CORS Configuration

Control Cross-Origin Resource Sharing:

#### Enable/Disable CORS
```bash
freightliner serve --enable-cors=true
```

#### Configure Allowed Origins
```bash
freightliner serve --allowed-origins "https://app.example.com,https://dashboard.example.com"
```

#### Wildcard Domains
```bash
freightliner serve --allowed-origins "*.example.com"
```

## Configuration Examples

### Development Setup
```bash
freightliner serve \
  --host localhost \
  --port 8080 \
  --enable-cors true \
  --allowed-origins "*"
```

### Production Setup (Docker)
```bash
freightliner serve \
  --host 0.0.0.0 \
  --port 8080 \
  --external-url https://api.production.com \
  --enable-cors true \
  --allowed-origins "https://app.production.com,https://dashboard.production.com" \
  --tls \
  --tls-cert /etc/ssl/certs/server.crt \
  --tls-key /etc/ssl/private/server.key \
  --api-key-auth \
  --api-key "${API_KEY}"
```

### Behind Nginx Reverse Proxy
```bash
freightliner serve \
  --host 127.0.0.1 \
  --port 8080 \
  --external-url https://api.example.com
```

With Nginx configuration:
```nginx
server {
    listen 443 ssl;
    server_name api.example.com;

    ssl_certificate /etc/ssl/certs/api.example.com.crt;
    ssl_certificate_key /etc/ssl/private/api.example.com.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Kubernetes Deployment
```yaml
apiVersion: v1
kind: Service
metadata:
  name: freightliner-api
spec:
  type: LoadBalancer
  ports:
  - port: 443
    targetPort: 8080
  selector:
    app: freightliner
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: freightliner
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: freightliner
        image: freightliner:latest
        args:
        - serve
        - --host=0.0.0.0
        - --port=8080
        - --external-url=https://api.k8s.example.com
        - --enable-cors=true
        - --allowed-origins=*.example.com
        env:
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: freightliner-secrets
              key: api-key
```

## Server Methods

The server instance now provides utility methods for URL construction:

### GetBaseURL()
Returns the base URL for external access:
- Uses `ExternalURL` if configured
- Otherwise constructs from `Host:Port`
- Handles TLS protocol detection
- Omits standard ports (80, 443)

```go
// Example: http://localhost:8080
baseURL := server.GetBaseURL()
```

### GetAPIBaseURL()
Returns the full API base URL including the API version path:

```go
// Example: http://localhost:8080/api/v1
apiBaseURL := server.GetAPIBaseURL()
```

## Security Considerations

### Host Binding Security

1. **localhost** (Default)
   - Most secure
   - Only local access
   - Recommended for development

2. **0.0.0.0** (All interfaces)
   - Exposed to network
   - Requires firewall/security groups
   - Use with TLS in production

3. **Specific IP**
   - Limited exposure
   - Good for multi-homed systems
   - Balance of security and accessibility

### CORS Security

1. **Avoid wildcards (*) in production**
   - Only use for development
   - Specify exact origins in production

2. **Use HTTPS for allowed origins**
   - Prevents man-in-the-middle attacks
   - Required for secure communication

3. **Limit allowed origins**
   - Only allow trusted domains
   - Review regularly

### TLS Configuration

Always enable TLS in production:
```bash
freightliner serve \
  --tls \
  --tls-cert /path/to/cert.pem \
  --tls-key /path/to/key.pem
```

## Environment Variables

Configuration can also be set via environment variables:

```bash
export FREIGHTLINER_HOST=0.0.0.0
export FREIGHTLINER_PORT=8080
export FREIGHTLINER_EXTERNAL_URL=https://api.example.com
export FREIGHTLINER_ENABLE_CORS=true
export FREIGHTLINER_ALLOWED_ORIGINS=*.example.com
```

## Troubleshooting

### Cannot bind to address
```
Error: listen tcp 192.168.1.100:8080: bind: cannot assign requested address
```
- Ensure the IP address exists on your system
- Check if port is already in use: `lsof -i :8080`

### CORS errors in browser
```
Access to fetch at 'http://localhost:8080/api/v1/replicate' from origin 'http://localhost:3000'
has been blocked by CORS policy
```
- Verify `--enable-cors=true` is set
- Check `--allowed-origins` includes your origin
- Inspect browser console for specific origin

### External URL not working
- Ensure firewall allows traffic
- Verify DNS resolution
- Check load balancer health checks
- Test with curl: `curl -v https://api.example.com/health`

## Monitoring

The server logs startup information:
```json
{
  "level": "info",
  "address": "0.0.0.0:8080",
  "external_url": "https://api.example.com",
  "tls": true,
  "cors": true,
  "msg": "Starting HTTP server"
}
```

Health check endpoint:
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy"
}
```
