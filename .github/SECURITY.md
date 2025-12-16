# Security Policy

## Supported Versions

We actively support the latest version of this project. Security updates are applied to the main branch.

| Version | Supported          |
| ------- | ------------------ |
| Latest (main) | :white_check_mark: |
| Older versions | :x: |

## Reporting a Vulnerability

We take the security of this project seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do Not** Open a Public Issue
Please do not report security vulnerabilities through public GitHub issues.

### 2. Report Privately
Send a detailed report to: **me@justinoleary.com**

Include:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)

### 3. What to Expect
- **Response Time:** You'll receive an acknowledgment within 48 hours
- **Updates:** We'll keep you informed of our progress
- **Credit:** Security researchers will be credited (if desired) once the issue is resolved

## Security Best Practices

When using this project:

### Secrets Management
- Never commit `.pulumi-passphrase` files
- Use environment variables for sensitive data
- Rotate passphrases regularly
- Use different passphrases for dev/staging/prod

### Access Control
- Limit access to your Pulumi state files
- Use proper RBAC within your Kubernetes cluster
- Keep your macOS system and dependencies updated

### Network Security
- The Lima VM uses host networking by default
- Consider using firewall rules for production-like testing
- Be cautious when exposing services from the cluster

### Dependency Updates
Keep these components updated:
```bash
brew upgrade pulumi go lima kind kubectl
go get -u ./...
go mod tidy
```

### Audit Logging
- Review cluster logs regularly
- Monitor resource usage
- Check for unexpected pods or services

## Known Security Considerations

### Local Development Environment
This project creates a **local development environment**. It is not designed for production use or to host sensitive workloads.

### Lima VM Access
The Lima VM has access to your home directory by default. Be mindful of what files are accessible.

### Docker Socket
The Docker socket is exposed to enable Kind cluster creation. Ensure only trusted containers are run.

## Updates and Patches

Security updates will be released as soon as possible after discovery. Check for updates regularly:

```bash
cd pulumi-kind-cluster
git pull origin main
pulumi refresh
pulumi up
```

## Security Checklist

- [ ] Using latest version from main branch
- [ ] Passphrase file is in `.gitignore`
- [ ] Dependencies are up to date
- [ ] macOS firewall is enabled
- [ ] Only running trusted container images
- [ ] Regular security updates applied

Thank you for helping keep this project secure!
