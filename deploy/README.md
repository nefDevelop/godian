# Despliegue

`godian` genera HTML estático en `dist/`. Cualquier hosting estático sirve.

## Comando de build

```sh
GOFLAGS=-tags=nodynamic CGO_ENABLED=0 go build -o bin/godian ./cmd/godian
./bin/godian build -config site.yaml
```

El tag `nodynamic` usa el libwebp transpilado a Go (sin cgo ni dependencias de runtime).
`CGO_ENABLED=0` garantiza un binario estático.

## Plataformas

| Plataforma | Archivo | Notas |
|---|---|---|
| GitHub Pages | `.github/workflows/pages.yml` | Ya en el repo; se dispara en `main` |
| Netlify | `deploy/netlify.toml` | Copiar a la raíz del repo |
| Cloudflare Pages | `deploy/wrangler.toml` | Copiar a la raíz y ajustar `name` |
| Vercel | `deploy/vercel.json` | Copiar a la raíz |

## Alternativa sin Go en el CI

Genera `dist/` en local (`make build && ./bin/godian build`) y publica la carpeta
`dist/` directamente (arrastra el resultado al hosting o commitéalo en una rama `gh-pages`).
