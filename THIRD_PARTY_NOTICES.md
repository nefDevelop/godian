# Third-party notices

godian is MIT-licensed (see `LICENSE`). It depends on and vendors third-party
software. This file lists the main components and their licenses. Always check
the upstream project for the authoritative license text.

## Go modules

| Module | License | Project |
|---|---|---|
| `gopkg.in/yaml.v3` | MIT / Apache-2.0 (see `LICENSE` and `NOTICE`) | https://gopkg.in/yaml.v3 |
| `github.com/alecthomas/chroma/v2` | MIT | https://github.com/alecthomas/chroma |
| `github.com/gen2brain/avif` | MIT | https://github.com/gen2brain/avif |
| `github.com/gen2brain/webp` | MIT | https://github.com/gen2brain/webp |
| `github.com/tdewolff/minify/v2` | MIT | https://github.com/tdewolff/minify |
| `github.com/yuin/goldmark` | MIT | https://github.com/yuin/goldmark |
| `golang.org/x/image` | BSD-3-Clause | https://golang.org/x/image |
| `github.com/dlclark/regexp2/v2` | MIT | https://github.com/dlclark/regexp2 |
| `github.com/ebitengine/purego` | Apache-2.0 | https://github.com/ebitengine/purego |
| `github.com/tdewolff/parse/v2` | MIT | https://github.com/tdewolff/parse |
| `github.com/tetratelabs/wazero` | Apache-2.0 | https://github.com/tetratelabs/wazero |
| `golang.org/x/sys` | BSD-3-Clause | https://golang.org/x/sys |
| `golang.org/x/text` | BSD-3-Clause | https://golang.org/x/text |

The Go module cache and the upstream repositories contain the full license
texts.

## Default theme vendor

The default theme includes vendored browser assets:

| Component | License | Project |
|---|---|---|
| KaTeX | MIT | https://katex.org/ |
| Mermaid | MIT | https://mermaid.js.org/ |
| PhotoSwipe | MIT | https://photoswipe.com/ |

## Admin vendor

The static admin vendors:

| Component | License | Project |
|---|---|---|
| js-yaml | MIT | https://github.com/nodeca/js-yaml |
| marked | MIT | https://github.com/markedjs/marked |
| DOMPurify | Apache-2.0 OR MPL-2.0 | https://github.com/cure53/DOMPurify |

The corresponding license files are under
`internal/infrastructure/adminui/static/vendor/`.

## Notes

- The default theme and its vendored assets are distributed with godian; keep
  their notices when redistributing a theme.
- If you add a dependency, add it here and include its license text if the
  license requires it.
- This notice is informational and is not legal advice.
