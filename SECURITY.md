# Security Policy

## Supported versions

go-openfga is pre-1.0; only the latest tagged release receives security fixes.
Pin a tagged version and upgrade to pick up fixes.

| Version        | Supported          |
| -------------- | ------------------ |
| latest release | :white_check_mark: |
| older          | :x:                |

## Reporting a vulnerability

**Please do not open a public issue for security reports.**

Report privately via GitHub's private vulnerability reporting:

→ https://github.com/sergiught/go-openfga/security/advisories/new

We aim to acknowledge reports within **5 business days** and to provide a status
update within **10 business days**. Coordinated disclosure is appreciated — we'll
agree on a timeline before any public detail is shared.

## Scope

go-openfga is a **client library** for the OpenFGA HTTP API. The most relevant
classes of issue are:

- Mishandling of credentials passed to `WithAPIToken`, `WithClientCredentials`, or
  `WithPrivateKeyJWT` (e.g. tokens or signing keys leaking into logs, errors, or
  request URLs).
- Incorrect TLS or transport handling that could expose requests to interception.
- Request construction flaws (header injection, request smuggling) reachable from
  caller-supplied input.

Out of scope:

- Vulnerabilities in the OpenFGA **server** itself — report those to the
  [OpenFGA project](https://github.com/openfga/openfga).
- Issues that require a malicious OpenFGA server you already fully control; the client
  trusts the server it is configured to talk to.
- Vulnerabilities in third-party dependencies that are already public — though we do
  appreciate a heads-up so we can bump them.
