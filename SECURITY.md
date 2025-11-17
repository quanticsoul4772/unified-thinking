# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest | :x:               |

## Reporting a Vulnerability

We take the security of our software seriously. If you believe you have found a security vulnerability in the Unified Thinking Server, please report it to us as described below.

**Please do not report security vulnerabilities through public GitHub issues.**

### How to Report

Please report security vulnerabilities by emailing the project maintainers. You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

Please include the following information:

- Type of issue (e.g. buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit the issue

### What to Expect

- **Acknowledgment**: We'll acknowledge receipt of your vulnerability report
- **Assessment**: We'll assess the vulnerability and determine its severity
- **Fix Development**: We'll develop a fix if the vulnerability is confirmed
- **Disclosure**: We'll coordinate disclosure with you

### Preferred Languages

We prefer all communications to be in English.

## Security Best Practices for Users

When using the Unified Thinking Server:

1. **Keep Software Updated**: Always use the latest version
2. **Secure Configuration**:
   - Use strong authentication if exposed to network
   - Limit file system access appropriately
   - Use SQLite persistence only with proper file permissions
3. **Environment Variables**: Never commit sensitive environment variables
4. **Access Control**: Restrict access to the MCP server to authorized users only
5. **Audit Logs**: Enable debug logging in production for security auditing

## Security Features

The Unified Thinking Server includes several security features:

- Input validation on all MCP tool inputs
- SQL injection prevention in SQLite storage
- Resource limits to prevent DoS attacks
- Secure defaults for all configurations
- No execution of arbitrary code from user input

## Known Security Considerations

- The server processes user input for reasoning tasks - ensure proper sandboxing if exposed to untrusted input
- SQLite storage writes to disk - ensure proper file permissions
- Debug mode may log sensitive information - disable in production

## Security Acknowledgments

We would like to thank the following people for responsibly disclosing security issues:

*(This list will be updated as we receive and address security reports)*