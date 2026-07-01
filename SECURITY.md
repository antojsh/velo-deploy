# Security Policy

## Supported versions

Velo Deploy follows [Semantic Versioning](https://semver.org/). Security patches are backported to the following releases:

| Version | Supported           |
| ------- | ------------------- |
| Latest  | ✅ Active           |
| `latest - 1` minor | ✅ Critical fixes only |
| Older   | ❌ End of life      |

We always recommend running the latest stable release. Older versions are not patched unless there is a critical, actively exploited vulnerability.

## Reporting a vulnerability

**Please do not open a public issue for security vulnerabilities.**

Send a detailed report to **[security@velo-deploy.dev](mailto:security@velo-deploy.dev)** with:

- A clear description of the issue and its impact
- Steps to reproduce or a proof-of-concept
- Affected version(s) and commit hash(es) if known
- Your name and how you would like to be credited (optional)

You can also use [GitHub's private vulnerability reporting](https://github.com/antojsh/velo-deploy/security/advisories/new) to disclose directly through the platform.

### What to expect

1. **Acknowledgement** within 48 hours of your report.
2. **Triage** within 5 business days — we will confirm the issue, assess severity, and propose a timeline.
3. **A CVE ID** assigned via GitHub Security Advisories for confirmed issues.
4. **A fix** in a private fork until the patch is ready.
5. **Coordinated disclosure** — we will agree on a release date with you, typically within 90 days of the report.
6. **Public disclosure** — the advisory, the fix, and credit (if you want it) are published on the same day.
7. **A new release** with the fix and a CHANGELOG entry.

### What we will not do

- We will not threaten legal action against researchers who follow this policy.
- We will not request a CVE be rescinded once it has been assigned.
- We will not delay disclosure past the agreed-upon date without a compelling reason and your consent.

## Severity ratings

We use [CVSS v3.1](https://www.first.org/cvss/calculator/3.1) to assess severity. The fix timeline depends on the score:

| CVSS | Severity | Fix timeline |
| --- | --- | --- |
| 9.0 – 10.0 | Critical | ≤ 7 days |
| 7.0 – 8.9  | High     | ≤ 30 days |
| 4.0 – 6.9  | Medium   | ≤ 90 days |
| 0.1 – 3.9  | Low      | Next minor release |

## Out-of-scope issues

The following are generally not considered security vulnerabilities and should be opened as regular bug reports:

- Denial-of-service attacks that require authenticated access
- Issues that only affect unsupported versions
- Issues in third-party dependencies that do not affect Velo Deploy directly
- Theoretical issues without a realistic attack scenario
- Self-XSS or social engineering

## Hall of fame

We thank the following researchers for responsible disclosure:

- _Your name here — be the first._

## Contact

- **Email**: [security@velo-deploy.dev](mailto:security@velo-deploy.dev)
- **GitHub**: [private vulnerability reporting](https://github.com/antojsh/velo-deploy/security/advisories/new)
- **PGP key**: [https://velo-deploy.dev/.well-known/pgp-key.asc](https://velo-deploy.dev/.well-known/pgp-key.asc)
