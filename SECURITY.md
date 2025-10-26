# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The WirePusher team takes security bugs seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do NOT report security vulnerabilities through public GitLab issues.**

Instead, please report security vulnerabilities via email to:

**security@wirepusher.com**

### What to Include

To help us triage and fix the issue quickly, please include:

1. **Type of vulnerability** (e.g., authentication bypass, injection, etc.)
2. **Full paths** of source files related to the vulnerability
3. **Location** of the affected source code (tag/branch/commit or direct URL)
4. **Step-by-step instructions** to reproduce the issue
5. **Proof-of-concept or exploit code** (if possible)
6. **Impact** of the vulnerability (what an attacker could achieve)
7. **Any mitigating factors** or workarounds you've identified

### What to Expect

After you submit a report:

1. **Acknowledgment** - We'll acknowledge receipt within 48 hours
2. **Assessment** - We'll assess the vulnerability and determine severity
3. **Updates** - We'll provide regular updates (at least every 7 days)
4. **Fix Timeline** - We aim to release fixes for:
   - **Critical** vulnerabilities: Within 7 days
   - **High** vulnerabilities: Within 14 days
   - **Medium** vulnerabilities: Within 30 days
   - **Low** vulnerabilities: Next regular release

5. **Disclosure** - We'll coordinate with you on public disclosure timing
6. **Credit** - We'll credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

### For Users

When using the WirePusher Go SDK:

1. **Keep the SDK updated** to the latest version
2. **Never commit credentials** to version control
3. **Use environment variables** for sensitive configuration
4. **Validate input** before sending to the SDK
5. **Handle errors gracefully** without exposing sensitive information
6. **Use HTTPS** for all network communication
7. **Limit token scope** to minimum required permissions

### Credential Management

```go
// ❌ Bad - Hardcoded credentials
client := wirepusher.NewClient("wpt_abc123", "user123")

// ✅ Good - Environment variables
token := os.Getenv("WIREPUSHER_TOKEN")
userID := os.Getenv("WIREPUSHER_USER_ID")
client := wirepusher.NewClient(token, userID)
```

### Error Handling

```go
// ❌ Bad - Exposes sensitive information
err := client.Send(ctx, options)
if err != nil {
    log.Printf("Error: %+v", err) // May log tokens or user IDs
}

// ✅ Good - Safe error handling
err := client.Send(ctx, options)
if err != nil {
    switch e := err.(type) {
    case *wirepusher.ValidationError:
        log.Printf("Validation error: %s", e.Message)
    case *wirepusher.AuthError:
        log.Println("Authentication failed - check credentials")
    default:
        log.Println("Notification failed - see logs for details")
    }
}
```

### Input Validation

```go
// ❌ Bad - No validation
http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
    title := r.FormValue("title")
    message := r.FormValue("message")
    client.SendSimple(ctx, title, message)
})

// ✅ Good - Validate input
http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
    title := r.FormValue("title")
    message := r.FormValue("message")

    if title == "" || message == "" {
        http.Error(w, "Missing required fields", 400)
        return
    }

    if len(title) > 256 || len(message) > 4096 {
        http.Error(w, "Content too long", 400)
        return
    }

    if err := client.SendSimple(ctx, title, message); err != nil {
        http.Error(w, "Failed to send notification", 500)
        return
    }

    w.WriteHeader(200)
})
```

### Context Timeouts

```go
// ❌ Bad - No timeout
err := client.Send(context.Background(), options)

// ✅ Good - Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
err := client.Send(ctx, options)
```

## Known Security Considerations

### API Token Security

- Tokens are transmitted in API requests and should be kept confidential
- Tokens are stored in plaintext by the SDK (secure storage is the user's responsibility)
- Compromised tokens can be used to send notifications as your user
- Rotate tokens regularly as a security best practice

### Network Communication

- All communication with WirePusher API is over HTTPS
- The SDK uses the Go standard library `net/http` which respects system-level TLS/SSL settings
- Certificate validation is handled by the Go runtime
- Minimum TLS 1.2 is enforced by default

### Dependencies

This SDK has **zero runtime dependencies** to minimize supply chain risks:
- Uses only the Go standard library (`net/http`, `encoding/json`, `context`)
- No external dependencies that could introduce vulnerabilities
- Regular security audits via `go list -m all`

### Memory Safety

- Go's memory safety features prevent buffer overflows and memory corruption
- Garbage collection prevents use-after-free vulnerabilities
- Type safety prevents many common programming errors

## Vulnerability Disclosure Process

When we receive a security bug report:

1. **Confirm the vulnerability** and determine affected versions
2. **Develop and test a fix** for all supported versions
3. **Prepare security advisory** with:
   - Description of the vulnerability
   - Affected versions
   - Fixed versions
   - Workarounds (if any)
   - Credit to reporter
4. **Release patched versions**
5. **Publish security advisory** on GitLab
6. **Notify users** via:
   - GitLab security advisory
   - Project README update
   - pkg.go.dev documentation

## Security Audit History

| Date | Type | Findings | Status |
|------|------|----------|--------|
| TBD  | TBD  | TBD      | TBD    |

## Security Hall of Fame

We thank the following individuals for responsibly disclosing security vulnerabilities:

- (None yet)

## Resources

- [Go Security Best Practices](https://golang.org/doc/security/)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Go Vulnerability Database](https://pkg.go.dev/vuln/)

## Questions?

For security-related questions that aren't reporting vulnerabilities:

- Email: security@wirepusher.com
- General questions: support@wirepusher.com

Thank you for helping keep WirePusher and its users safe!
