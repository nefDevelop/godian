# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Base OSS: MIT license, security policy, contribution guide, code of conduct,
  CI targets, issue/PR templates, quickstart and third-party notices.

## [0.1.0] - 2026-10-19

### Added

- Static site generator in Go with a single binary and no Node/runtime
  dependency for the site.
- Theme overlay with embedded default theme, `theme.yaml`, CSS chunks, islands,
  motion tokens and `godian new-theme` / `--minimal`.
- Content collections (blog, docs, pages), authors, series, taxonomies, RSS,
  Atom, JSON Feed, sitemap, Open Graph images and search index.
- Markdown extensions: callouts, wikilinks, embeds, attributes, checklists,
  math and code highlighting.
- Image pipeline with WebP/AVIF variants, LQIP and responsive `srcset`
  (AVIF opt-in).
- Configurable base path for root or subpath deployments, with verification.
- Native local admin at `/admin` with schema-driven forms, preview, uploads
  and atomic writes.
- Optional static remote admin with Gitea/GitHub providers, PAT login,
  browser-side schema forms, Markdown toolbar, media library, file actions and
  commit-per-save.
- Same-origin `/admin/api` proxy for local development and Cloudflare Pages
  Functions, with path restrictions and no stored tokens.
- Deployment documentation for GitHub Pages, Cloudflare Pages, Netlify and
  Vercel.
- Unit, integration, base-path, admin, proxy and theme tests.

### Security

- The remote admin keeps the PAT in `sessionStorage`, uses CSP, sanitizes
  previews, restricts the proxy to the configured repository `contents/` and
  documents Access/Zero Trust for public deployments.
