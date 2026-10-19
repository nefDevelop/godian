# Quickstart

godian es un generador de sitios estáticos en Go. No necesita Node ni npm para
construir el sitio: produce HTML/CSS/JS en `dist/`.

## Requisitos

- Go 1.26 o superior (recomendado 1.27.2).
- `make` (opcional, pero recomendado).
- Un repositorio Git para el contenido, si vas a desplegar.

## Instalar

Con una release publicada:

```sh
go install github.com/nefdevelop/godian/cmd/godian@latest
godian version
```

Desde el código:

```sh
git clone https://github.com/nefDevelop/godian
cd godian
make build
./bin/godian version
```

## Probar el sitio de ejemplo

El repositorio incluye un sitio de ejemplo (`site.yaml`, `content/`, `themes/`):

```sh
./bin/godian build -config site.yaml
./bin/godian serve -config site.yaml
```

Abre `http://localhost:4321/`. En modo local, el admin nativo está en
`http://localhost:4321/admin/` (solo loopback).

## Estructura

```text
site.yaml          configuración del sitio
content/           blog, docs, pages, authors, series
themes/default/    tema por defecto (plantillas, CSS, JS, islas)
assets/            imágenes y estáticos fuente
i18n/              traducciones de UI
data/              datos del sitio
deploy/            despliegue, admin remoto y CORS
```

## Crear contenido

```sh
./bin/godian new -title "Mi primer post" mi-primer-post
./bin/godian new -collection docs -title "Guía" guia
./bin/godian check -config site.yaml
./bin/godian build -config site.yaml
```

- `new` escribe en `content/<colección>/<slug>.md` con `draft: true`.
- `check` valida imágenes, enlaces, series, títulos y frontmatter.
- `build` genera `dist/`; `-drafts` incluye borradores.

## Temas

El tema por defecto se puede copiar o crear desde cero:

```sh
./bin/godian new-theme mi-tema
./bin/godian new-theme --minimal mi-tema-minimal
```

Apunta `build.theme_dir` a la carpeta y lee
[`themes/default/README.md`](themes/default/README.md) para el contrato de
plantillas, `PageData`, chunks, islas y tokens.

## Admin

- **Local**: `godian serve` monta el admin nativo en `/admin/` (loopback, sin
  token).
- **Remoto**: opcional. Se publica en `dist/admin` y edita el repositorio con
  un PAT a través de Gitea/GitHub. Requiere proxy same-origin o CORS; lee
  [`deploy/admin-remoto.md`](deploy/admin-remoto.md).
- El PAT no se guarda en el repositorio ni en `site.yaml`.

## Desplegar

`dist/` es estático. Consulta [`deploy/README.md`](deploy/README.md) para
GitHub Pages, Cloudflare Pages, Netlify y Vercel, incluidas las subrutas.

## Desarrollo

```sh
make test
make test-race
make vet
make staticcheck
make gosec
make vuln
./bin/godian check -config site.yaml
./bin/godian build -config site.yaml
```

`make ci` agrupa la validación que ejecuta la CI. Para contribuir, lee
[`CONTRIBUTING.md`](CONTRIBUTING.md).

## Licencia

MIT. Consulta [`LICENSE`](LICENSE) y
[`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).
