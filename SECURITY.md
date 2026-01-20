# Security Policy

## Overview

Stargazer is designed with security as a core principle. As a Kubernetes troubleshooting tool that connects to production clusters, we implement multiple layers of security to protect your infrastructure, credentials, and data.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

We recommend always using the latest stable release for the most up-to-date security features and patches.

## Reporting Security Vulnerabilities

We take security vulnerabilities seriously. If you discover a security issue, please follow responsible disclosure practices:

### How to Report

1. **DO NOT** create a public GitHub issue for security vulnerabilities
2. Email security details to: [security contact email - to be configured]
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

### What to Expect

- **Acknowledgment**: Within 48 hours of your report
- **Initial Assessment**: Within 1 week
- **Status Updates**: Regular updates on progress
- **Disclosure Timeline**: We aim to patch critical vulnerabilities within 30 days
- **Credit**: Security researchers will be credited (unless they prefer to remain anonymous)

## Security Best Practices for Users

### Kubernetes Access

1. **Use Read-Only Service Accounts**: Stargazer only requires read permissions
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: stargazer-readonly
   rules:
   - apiGroups: ["*"]
     resources: ["*"]
     verbs: ["get", "list", "watch"]
   ```

2. **Namespace Isolation**: Limit access to specific namespaces when possible
3. **Regular Credential Rotation**: Rotate kubeconfig credentials periodically
4. **Audit Access**: Review kubeconfig access and permissions regularly

### API Keys

1. **Minimum Permissions**: Use API keys with minimal required permissions
2. **Separate Keys**: Use different API keys for development and production
3. **Key Rotation**: Rotate LLM provider API keys regularly
4. **Environment Variables**: Consider using environment variables instead of storing keys in config files

### Network Security

1. **Local-Only Access**: By default, Stargazer binds to `127.0.0.1` (localhost only)
2. **Firewall Rules**: Ensure port 8000 is not exposed to external networks
3. **VPN/Bastion**: Access remote clusters through secure VPN or bastion hosts
4. **TLS in Production**: When exposing the API, use TLS/HTTPS (reverse proxy recommended)

## Data Storage and Security

### Storage Location

All Stargazer data is stored locally in:
```
~/.stargazer/
├── config.yaml           # Configuration (includes API keys)
├── storage/              # Scan results and historical data
│   └── scan-*.json      # Individual scan results
```

### File Permissions

Stargazer implements strict file permissions to protect sensitive data:

- **Config file**: `0600` (rw-------)
  - Only the owner can read/write
  - Contains sensitive API keys and credentials

- **Storage directory**: `0755` (rwxr-xr-x)
  - Owner has full access
  - Others can read/execute (scan results are non-sensitive)

- **Scan results**: `0644` (rw-r--r--)
  - Owner can read/write
  - Others can read

### Data Retention

Configure automatic cleanup of old scan results:

```yaml
# ~/.stargazer/config.yaml
storage:
  retain_days: 30          # Delete scans older than 30 days
  max_scan_results: 100    # Keep maximum 100 scan results
```

### What Data is Stored

**Configuration (`config.yaml`):**
- Kubeconfig path and context
- LLM provider API keys (plaintext - protect this file!)
- API server settings
- Storage preferences

**Scan Results (`storage/scan-*.json`):**
- Kubernetes resource metadata (names, namespaces, labels)
- Issue descriptions and recommendations
- Event logs and timestamps
- **Does NOT include**: Secrets, ConfigMap data, or sensitive workload information

## Kubeconfig Security

### Auto-Discovery

Stargazer discovers kubeconfig in this order:
1. Explicitly configured path in `~/.stargazer/config.yaml`
2. `$KUBECONFIG` environment variable
3. Default location: `~/.kube/config`

### Best Practices

1. **File Permissions**: Ensure kubeconfig has `0600` permissions
   ```bash
   chmod 600 ~/.kube/config
   ```

2. **Sensitive Data**: Never commit kubeconfig to version control

3. **Context Isolation**: Use different contexts for different environments
   ```bash
   kubectl config use-context production
   stargazer scan  # Uses production context
   ```

4. **Temporary Access**: For sensitive clusters, use short-lived tokens
   ```bash
   # Example: AWS EKS with temporary credentials
   aws eks get-token --cluster-name my-cluster
   ```

## API Key Handling

### Storage

API keys are stored in `~/.stargazer/config.yaml`:

```yaml
llm:
  default_provider: openai
  providers:
    openai:
      enabled: true
      api_key: "sk-..."        # Stored in plaintext
      model: "gpt-4o-mini"
      encrypted: false         # Future: encryption support
```

**Important**: The config file has `0600` permissions, but API keys are currently stored in plaintext. Protect this file carefully.

### Best Practices

1. **Dedicated Keys**: Create separate API keys for Stargazer
2. **Usage Limits**: Set spending limits on LLM provider accounts
3. **Monitor Usage**: Regularly review API usage for anomalies
4. **Revoke Unused Keys**: Disable providers you're not actively using

### Future Enhancements

We plan to implement:
- Encrypted API key storage
- OS keychain integration (macOS Keychain, Windows Credential Manager)
- Environment variable override: `STARGAZER_OPENAI_API_KEY`

## Network Security

### Local-Only by Default

The API server binds to `127.0.0.1` (localhost only):

```go
// app.go - Line 77
listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
```

This prevents external access by default.

### Port Configuration

Default port: `8000` (auto-increments to `8001-8010` if busy)

Configure in `~/.stargazer/config.yaml`:
```yaml
api:
  port: 8000
  enable_cors: true
  rate_limit_rps: 100
```

### Exposing to Network (Not Recommended)

If you must expose Stargazer to a network:

1. **Use a Reverse Proxy**: Never expose the API directly
   ```nginx
   # Example: nginx with TLS
   server {
     listen 443 ssl;
     server_name stargazer.internal.company.com;

     ssl_certificate /path/to/cert.pem;
     ssl_certificate_key /path/to/key.pem;

     location / {
       proxy_pass http://127.0.0.1:8000;
     }
   }
   ```

2. **Enable Authentication**: Add authentication at the reverse proxy level
3. **Firewall Rules**: Restrict access to specific IP ranges
4. **VPN Required**: Require VPN connection to access

## Security Features

### Rate Limiting

Per-IP rate limiting protects against abuse:

```yaml
# ~/.stargazer/config.yaml
api:
  rate_limit_rps: 100  # 100 requests per second per IP
```

Implementation details:
- Token bucket algorithm with burst allowance (2x rate limit)
- Automatic cleanup of inactive clients (10-minute timeout)
- Health check endpoint exempt from rate limiting
- Expensive endpoints (topology, recommendations) have stricter limits (5-10 req/min)

### CORS (Cross-Origin Resource Sharing)

Enabled by default for development:

```yaml
api:
  enable_cors: true
```

CORS headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`

**Production**: Disable CORS or restrict to specific origins using a reverse proxy.

### Security Headers

All responses include security headers:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Input Validation

- Namespace and resource names validated against Kubernetes naming rules
- Path traversal protection in file operations
- JSON schema validation for API requests
- SQL injection N/A (no database)

### Read-Only Operations

Stargazer **never** modifies cluster resources. All Kubernetes operations use read-only verbs:
- `get`
- `list`
- `watch`

This is enforced by:
1. RBAC permissions (user-configured)
2. Client code design (no create/update/delete operations)

## Authentication and Authorization

### Current State

- **No Built-In Authentication**: Stargazer runs locally and trusts the local user
- **Kubernetes Authentication**: Delegated to kubeconfig (certificates, tokens, etc.)
- **LLM Provider Authentication**: API keys stored in config

### Future Enhancements

Planned security features:
- [ ] Optional API authentication (bearer tokens)
- [ ] OIDC/SSO integration for enterprise deployments
- [ ] Encrypted credential storage
- [ ] Multi-user support with RBAC
- [ ] Audit logging

## Compliance and Privacy

### Data Privacy

- **No External Data Transmission**: Except to configured LLM providers
- **No Telemetry**: Stargazer does not send usage data or analytics
- **Local Processing**: All data processing happens locally
- **No Cloud Dependencies**: Fully self-contained (except optional LLM features)

### LLM Provider Considerations

When using AI features, cluster data is sent to third-party LLM providers:

**Data Sent:**
- Resource names and namespaces
- Issue descriptions
- Event messages
- Logs (when using troubleshooting features)

**Best Practices:**
- Review LLM provider privacy policies
- Use self-hosted models (Ollama) for sensitive data
- Disable LLM features for highly regulated environments
- Sanitize logs before AI analysis

### Audit Trail

Stargazer logs all API requests (except health checks):

```
[API] GET /api/cluster/health 200 2.5ms
[API] POST /api/troubleshoot 200 1.2s
```

For compliance, consider:
- Redirecting logs to SIEM systems
- Enabling verbose logging for audit purposes
- Implementing request ID tracking

## Incident Response

### Security Incident Checklist

If you suspect a security breach:

1. **Isolate**: Disconnect affected systems from the network
2. **Revoke Credentials**:
   - Rotate kubeconfig credentials
   - Regenerate LLM API keys
   - Update service account tokens
3. **Audit**:
   - Review `~/.stargazer/storage/` for suspicious scans
   - Check Kubernetes audit logs for unusual activity
   - Examine API server logs
4. **Report**: Contact our security team (see Reporting section)
5. **Update**: Upgrade to the latest version
6. **Review**: Update security policies and access controls

## Security Checklist

Use this checklist when deploying Stargazer:

- [ ] Kubeconfig has minimal required permissions (read-only)
- [ ] Config file (`~/.stargazer/config.yaml`) has `0600` permissions
- [ ] API server binds to `127.0.0.1` only
- [ ] Rate limiting is enabled
- [ ] LLM API keys are rotated regularly
- [ ] Scan results are cleaned up automatically (retention policy)
- [ ] CORS is disabled or restricted in production
- [ ] Running latest stable version
- [ ] Security updates are monitored
- [ ] Incident response plan is documented

## Additional Resources

- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

## Contact

For security questions or concerns:
- GitHub Issues: https://github.com/maplecitymadman/stargazer/issues (for non-sensitive topics)
- Security Email: [To be configured]

---

**Last Updated**: 2026-01-20
**Version**: 1.0
