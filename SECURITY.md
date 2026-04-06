# Security Policy

## Reporting Vulnerabilities

If you discover a security vulnerability, please open a GitHub issue with the `security` label, or contact the maintainer via their GitHub profile.

For sensitive issues, please reach out privately before opening a public issue.

## Supported Versions

Only the latest release receives security patches.

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |
| Older   | No        |

## keys.yaml

`qs` can store API keys in `~/.qs/keys.yaml` in **plaintext**. Users should:

- Ensure proper file permissions on `keys.yaml` (readable only by your user)
- **Never** commit `keys.yaml` to version control
- Consider using environment variables instead of `keys.yaml` for sensitive keys

The `.gitignore` in this repo excludes `keys.yaml`, but be careful with copies or backups.
