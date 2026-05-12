# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest  | ✅        |
| Older   | ❌        |

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Use [GitHub's private security advisory](https://github.com/oleg-koval/mac-dev-station/security/advisories/new) to report issues confidentially.

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact and affected versions

Expected response within 48 hours. Fixes tracked privately until a patch ships.

## Security Notes

- `mac-dev-station` runs shell commands on your local machine with your user permissions.
- Review the source before running on systems with sensitive data.
- The `install.sh` script verifies nothing beyond HTTPS transport — validate the checksum from the release page if supply-chain integrity matters to you.
- No telemetry, no network calls during setup phases beyond Homebrew.
