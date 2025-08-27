# Security Policy

We take the security of `oapi-codegen` seriously and appreciate coordinated disclosures.

## Supported versions

We generally support the latest minor release of the current major version (v2.x). Older versions may receive fixes on a best-effort basis. Security fixes are released as soon as practical.

## Reporting a vulnerability

- Please report suspected vulnerabilities privately via GitHub Security Advisories ("Report a vulnerability" on the repo) or by opening a security advisory draft addressed to the maintainers.
- If you cannot use GitHub Security Advisories, please open a minimal issue with no sensitive details and ask to initiate a private security report; a maintainer will follow up.
- Do not disclose details publicly until a fix is released and a coordinated disclosure date is agreed.

When reporting, include:

- A description of the issue and potential impact
- A minimal reproduction, if possible
- Affected versions / commit SHAs
- Any suggested mitigations or patches

## Response process

1. Triage to confirm impact and scope.
2. Coordinate privately to develop and review a fix.
3. Release patched versions and publish an advisory with CVE details where applicable.
4. Credit reporters who wish to be acknowledged.

## Verification and hardening

- We run CI across supported Go versions, linting, and generated-file checks.
- We welcome PRs that add additional security-related tests or hardening.
