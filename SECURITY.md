# Security Policy

## Supported Versions

Currently, only the latest release of ADX is actively supported for security updates.

| Version | Supported          |
| ------- | ------------------ |
| v0.1.x  | :white_check_mark: |
| < v0.1  | :x:                |

## Reporting a Vulnerability

Security is a top priority for us. If you discover a security vulnerability in this project, please DO NOT report it by creating a public GitHub issue. 

Instead, please report it privately through GitHub:

1. **Where**: Open a private report at
   [Security → Report a vulnerability](https://github.com/bakhod1r/devicex/security/advisories/new).
   The report is visible only to the maintainers until an advisory is published.
2. **Details**: Please provide as much information as possible, including:
   - A description of the vulnerability.
   - Steps to reproduce the issue.
   - The potential impact of the vulnerability.
   - (Optional) A proof-of-concept (PoC) or exploit code.

We will review the report and respond to you as soon as possible, usually within 48 hours.

## Triage and Resolution

Once a vulnerability has been reported and confirmed:
1. We will acknowledge receipt of your report.
2. We will work on a fix in a private branch.
3. We will provide you with a timeline for the fix and the planned release date.
4. Once the fix is released, we will publicly disclose the vulnerability and acknowledge your contribution (unless you prefer to remain anonymous).

## Out of Scope

The following types of reports are generally considered out of scope:
- Denial of Service (DoS) attacks that rely on exhausting system resources, as these should be mitigated at the infrastructure or web framework level.
- Issues related to third-party dependencies (please report those to the respective upstream projects).
