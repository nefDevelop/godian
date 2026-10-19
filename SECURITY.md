# Security policy

## Supported versions

godian is currently `0.x`. Security fixes are applied to the latest published
version and to `main`. If you use an older release, upgrade before reporting a
problem that may already be fixed.

## Reporting a vulnerability

Please do **not** open a public issue with exploit details, tokens, private
content, or personal data.

Use a private advisory in the repository hosting (GitHub Security Advisories,
Gitea private issue, or the maintainer's private contact). Include:

- Affected version/commit and component (generator, theme, admin, proxy).
- Reproduction steps and expected vs. actual behavior.
- Impact and attack prerequisites.
- Minimal proof of concept without real credentials.
- Any suggested fix.

We will acknowledge the report, coordinate a fix, and credit you if you want.
No response-time SLA is promised.

## Scope

In scope:

- The Go generator and its build pipeline.
- The default theme, templates, islands, and vendored browser assets.
- The local native admin (`godian serve`).
- The optional remote static admin and `/admin/api` proxy/Pages Function.
- Authentication/token handling in the admin.
- Path traversal, XSS, CSRF, SSRF, and unsafe file handling.
- Supply-chain issues in dependencies or vendored assets.

Out of scope:

- Content published by the site owner.
- Misconfiguration of a third-party host, CDN, or reverse proxy.
- Issues that require a compromised browser/extension or a stolen PAT.
- Denial of service against a private local development server.

## Security model

- The **local native admin** is loopback-only, uses a CSRF token, and writes
  directly to the local checkout. It is a development tool.
- The **remote static admin** is a public page protected by a PAT. The PAT is
  entered by the user, stored in `sessionStorage`, and never committed to the
  repository.
- In `api_mode: proxy`, the browser sends the PAT to `/admin/api`; the proxy
  forwards it to the Git provider. The proxy does not persist or configure the
  PAT, but it sees it in memory. Use Cloudflare Access/Zero Trust or a private
  proxy if the admin is public.
- In `api_mode: direct`, the browser calls the provider directly. The provider
  must allow CORS. Some Gitea deployments do not send CORS headers; check the
  preflight response before using this mode.
- The remote admin has no roles, multi-user accounts, or audit log of its own;
  Git commits are the audit trail. For team use, use a server-side admin with
  SSO/RBAC.
- The static admin uses CSP, no third-party runtime CDNs, sanitized Markdown
  previews, a restricted proxy path (`contents/` of the configured repo), and
  `noindex`/`robots.txt` disallow rules. `_headers` enables frame protection on
  hosts that support it; GitHub Pages ignores it.
- All tokens should be fine-grained, scoped to the content repository, and
  revocable. Revoke a token immediately if you suspect exposure.

## Dependencies and vendored assets

Dependencies are listed in `go.mod`; vendored browser assets and their licenses
are described in `THIRD_PARTY_NOTICES.md`. Report vulnerabilities in a
dependency upstream and, if godian needs a mitigation, open a private advisory.

## Safe defaults

- Do not put tokens in `site.yaml`, `config.json`, `.env` files that are
  committed, docs, tests, or issue reports.
- Run `git grep` for secrets before publishing a fork.
- Keep the admin behind authentication at the edge when it is public.
- Run `godian check` and `godian build` in CI; a commit is not validated until
  the build passes.
