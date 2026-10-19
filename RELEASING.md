# Releasing godian

Releases use [Semantic Versioning](https://semver.org/). Do not create a tag,
publish a release, or push a release commit without the maintainer's explicit
approval.

## Checklist

1. Ensure the working tree is clean and `main` is up to date.
2. Run the CI target:
   ```sh
   make ci
   ```
3. Update `CHANGELOG.md`:
   - Move items from `Unreleased` to the new version.
   - Add the release date.
   - Summarize breaking changes and migrations.
4. Review `README.md`, `QUICKSTART.md` and the deployment docs for stale
   commands or placeholders.
5. Create the tagged commit:
   ```sh
   git tag -a vX.Y.Z -m "godian vX.Y.Z"
   git push origin main
   git push origin vX.Y.Z
   ```
6. Build the release binary and checksum:
   ```sh
   make release
   make checksums
   ```
7. Build the cross-platform binaries (the site does not need cgo):
   ```sh
   for os in linux darwin; do
     for arch in amd64 arm64; do
       GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 GOFLAGS=-tags=nodynamic \
         go build -trimpath \
         -ldflags="-s -w -X github.com/nefdevelop/godian/internal/presentation/cli.Version=vX.Y.Z" \
         -o "bin/godian-${os}-${arch}" ./cmd/godian
     done
   done
   GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOFLAGS=-tags=nodynamic \
     go build -trimpath \
     -ldflags="-s -w -X github.com/nefdevelop/godian/internal/presentation/cli.Version=vX.Y.Z" \
     -o bin/godian-windows-amd64.exe ./cmd/godian
   ```
8. Generate checksums for every artifact:
   ```sh
   (cd bin && sha256sum godian-* > SHA256SUMS)
   ```
9. Publish the release with the binaries, `SHA256SUMS` and the changelog
   section.
10. Verify the published module:
    ```sh
    go install github.com/nefdevelop/godian/cmd/godian@vX.Y.Z
    godian version
    ```
11. Deploy the site and check `/`, `/admin/` (if enabled), feeds, sitemap and
    search.

## Versioning rules

- `MAJOR`: breaking changes in the CLI, config, content schema or theme
  contract.
- `MINOR`: backwards-compatible features.
- `PATCH`: backwards-compatible fixes.
- Pre-1.0 versions may change the theme/content contract between minor
  versions; document it in `CHANGELOG.md`.

## Rollback

- Delete the release/tag only if it has not been consumed.
- Re-publish the previous version.
- If a release contains a security fix, publish a new patch; do not rewrite the
  tag.
- Revoke any leaked tokens/credentials immediately.

## Admin and secrets

A release must not contain a PAT, `.env`, `config.json` with credentials, or
private content. Confirm with:

```sh
git grep -nE 'github_pat_|ghp_|glpat-|Bearer [A-Za-z0-9_-]{20,}'
```

The remote admin PAT is entered at runtime and stored in the browser; it is
never part of a release artifact.
