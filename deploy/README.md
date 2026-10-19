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
| GitHub Pages | `.github/workflows/pages.yml` | Manual; ajusta `site.url`/`base_path` a Pages antes de ejecutarlo |
| Netlify | `deploy/netlify.toml` | Copiar a la raíz del repo |
| Cloudflare Pages | `deploy/wrangler.toml` | Copiar a la raíz y ajustar `name` |
| Vercel | `deploy/vercel.json` | Copiar a la raíz |

## Subruta de despliegue

Si el sitio se publica en una subruta (por ejemplo, un **project page** de GitHub
Pages en `https://usuario.github.io/godian/`), indica la URL pública completa:

```yaml
site:
  url: https://usuario.github.io/godian
```

Godian deriva `base_path: /godian` y prefija enlaces, assets, feeds, sitemap,
búsqueda y manifests. También puedes forzarlo con `site.base_path: /godian`.
En raíz (dominio propio, Cloudflare, Vercel, Netlify) no hace falta.

## Admin remoto

Guía paso a paso: [`admin-remoto.md`](admin-remoto.md).

El admin nativo se sirve en local con `godian serve`. Para publicar el admin
estático con el sitio:

1. Activa `admin.mode: remote` en `site.yaml` (ver el ejemplo comentado).
2. Usa `admin.api_mode: proxy` para evitar CORS: `godian serve` monta el proxy
   en local y Cloudflare Pages usa `functions/admin/api/[[path]].js`.
3. Configura `GIT_PROVIDER`, `GIT_BASE_URL`, `GIT_OWNER` y `GIT_REPO` en Pages.
4. El admin queda en `/admin/` de `dist`, con `noindex` y `Disallow`.

Si el hosting no ejecuta Functions (GitHub Pages puro), necesitas CORS en Gitea
o usar el admin de servidor. Detalle y ejemplo Nginx:
[`gitea-cors.md`](gitea-cors.md).

## Alternativa sin Go en el CI

Genera `dist/` en local (`make build && ./bin/godian build`) y publica la carpeta
`dist/` directamente (arrastra el resultado al hosting o commitéalo en una rama `gh-pages`).
