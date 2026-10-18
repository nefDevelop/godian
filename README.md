# godian

Generador de blog/documentación **estático, minimalista y zero-framework, escrito en Go**.

- **Binario único**: sin Node ni Rust en el flujo principal (build cgo-free).
- **Config-driven**: contenido, tema y salida con rutas configurables.
- **Temas e islas data-driven**: crea componentes (interactivos o estáticos) como archivos, sin tocar Go.
- **Obsidian-first**: wikilinks entre colecciones, **transclusión** de notas (`![[nota]]`) y **backlinks**.

Diseño y decisiones: `PLAN.md`. Estado por fases: `ROADMAP.md`.

---

## Requisitos

- Go **>= 1.26** (se recomienda **>= 1.27.2**, que corrige vulnerabilidades de la stdlib reportadas por `govulncheck`).

## Instalación y build

```sh
git clone <repo> && cd godian
make build          # compila a bin/godian (usa -tags nodynamic)
```

> El build usa `-tags=nodynamic` (libwebp transpilado a Go, sin cgo ni dependencias de runtime).
> El `Makefile` lo aplica vía `GOFLAGS`. `CGO_ENABLED=0` por defecto.
> **AVIF opt-in**: compila con `-tags=nodynamic,avif` para generar también variantes AVIF.

## Uso (CLI)

```sh
godian build [-config site.yaml] [-drafts]        # genera el sitio en dist/
godian serve [-config site.yaml] [-addr :4321]    # dev server + hot reload (SSE)
godian new [-config site.yaml] [-collection blog] [-title "Título"] <slug>
godian new-theme <dir>                            # copia el tema por defecto a <dir>
godian check [-config site.yaml]                  # valida el contenido
godian version
```

- `build`: genera el sitio. `-drafts` incluye borradores.
- `serve`: sirve `dist/` y **reconstruye al detectar cambios** en contenido/tema/assets/config; recarga el navegador.
- `new`: crea `content/<colección>/<slug>.md` con frontmatter listo (`draft: true`). *Las banderas van antes del slug.*
- `new-theme`: copia el tema por defecto (`themes/default/`) a `<dir>` para personalizarlo; apunta `build.theme_dir` a esa carpeta.
- `check`: valida el contenido (imágenes referenciadas/huérfanas, `alt` faltante, series con capítulos inexistentes, títulos vacíos); sale con código 1 si hay problemas.
- `version`: imprime la versión (sobrescribible con `-ldflags "-X .../cli.Version=x.y.z"`).

---

## Configuración (`site.yaml`)

```yaml
site:
  title: Mi sitio
  description: Blog y documentación
  url: https://ejemplo.com
  language: es          # en | es (i18n de la interfaz)
  color: "#2563eb"      # color de acento del tema por defecto

build:
  content_dir: content
  theme_dir: themes/default   # si no existe, se usa el tema embebido
  output_dir: dist
  assets_dir: assets          # estáticos fuente (imágenes, favicon…) servidos desde /
  data_dir: data              # datos estructurados (JSON/YAML/CSV)
  i18n_dir: i18n              # traducciones de la interfaz (i18n/<lang>.json)
  minify: true                # minifica HTML, CSS y JS (tdewolff/minify)
  fingerprint: true           # hash en el nombre de CSS/JS
  optimize_images: true       # variantes WebP + <picture>/srcset + LQIP
  clean: true                 # limpia dist/ antes de cada build
  image_presets:              # opcional; para {{image src=… preset=…}}
    thumb: { width: 480, height: 270, fit: fill, focal: center }
    wide:  { width: 1200, fit: fit }

# Colecciones (opcional; por defecto: blog, docs, authors, series, pages)
collections:
  - { name: blog, folder: blog }
  - { name: docs, folder: docs }

# Taxonomías configurables (opcional; por defecto tags y categories)
taxonomies:
  - name: tags
    label: Etiquetas
    url: /tags
    field: tags            # campo del frontmatter del post
    prefix: "#"
    hierarchical: true     # expande "a/b" en "a" y "a/b"
    strict: true           # avisa de términos usados pero no declarados
    render_empty: false    # genera páginas de términos sin posts
    terms:
      - { name: go, description: Posts sobre Go }
      - { name: javascript, aliases: [js] }
      - { name: markdown, weight: 10 }

# Comentarios (Giscus / GitHub Discussions)
giscus:
  repo: ""
  repo_id: ""
  category: "General"
  category_id: ""

# Navegación (opcional; si se omite, se usa la estándar)
nav:
  - { label: Inicio, url: / }
  - { label: Blog, url: /blog/ }

# Enlaces sociales y de apoyo (opcional)
social:
  - { label: GitHub, url: https://github.com/tuusuario, icon: github }
  - { label: RSS, url: /rss.xml, icon: rss }
support:
  - { label: Buy me a coffee, url: https://buymeacoffee.com, icon: coffee }

# Analítica (opcional): provider plausible | ga4
analytics:
  provider: ""
  id: ""

# Fuentes autohospedadas (opcional): se descargan en el build a /assets/fonts
fonts:
  - family: Inter
    weights: [400, 700]
    subsets: [latin]
```

---

## Fuentes

`fonts:` en `site.yaml` descarga tipografías de Google Fonts **en el build** y las
sirve desde `/assets/fonts` (sin peticiones a terceros en tiempo de uso):

```yaml
fonts:
  - family: Inter
    weights: [400, 700]
    styles: [normal]      # normal | italic
    subsets: [latin]      # latin, latin-ext, cyrillic…
    display: swap         # swap (por defecto) | block | fallback | optional
```

- Se genera `/assets/css/fonts.css` con los `@font-face` (con `font-display`) y se añade al `<head>`.
- Se pueden declarar **varias** fuentes (una entrada por familia).
- Para **aplicarla** al cuerpo: `site.font: Inter` (el tema usa `--font-sans`/`--font-mono`).
- Si el build no tiene red, **avisa y sigue** con las fuentes del sistema (no rompe).
- Licencias: la mayoría son OFL (autohospedaje permitido).

Más detalle en la documentación del sitio: **`/docs/fuentes/`**.

---

## Internacionalización (UI)

Los textos de la interfaz (nav, botones, etiquetas) se cargan desde
`i18n/<lang>.{json,yaml}` (config `build.i18n_dir`), **superpuestos** sobre los valores
por defecto embebidos:

```json
{ "nav.home": "Inicio", "ui.search": "Buscar", "ui.backlinks": "Enlaces entrantes" }
```

- El idioma activo es `site.language`.
- El archivo externo debe ser **completo**: si faltan claves, o una taxonomía nueva
  (`taxonomy.<name>`) no tiene texto, el build **avisa** por stderr.
- Para soportar otro idioma basta con crear `i18n/<lang>.json` (sin recompilar).
- Los textos de **JS** (copiar/`Copied!`, “sin resultados”) también se traducen vía `window.__i18n`.

## Contenido

Estructura por defecto dentro de `content/`:

```
content/
├── blog/     # posts (.md / .mdx)
├── docs/     # documentación (.md)
├── authors/  # perfiles (.yaml)
├── series/   # series (.yaml)
└── pages/    # páginas sueltas (.md / .yaml)
```

Los **slugs** se derivan de la ruta relativa del archivo (p. ej. `guia/intro.md` → `guia/intro`). Se detectan slugs duplicados.

### Frontmatter

**Blog** (`content/blog/*.md`):

```yaml
---
title: "Título"
published: 2026-01-15
updated: 2026-02-01        # opcional
description: "Resumen"
image: "/assets/portada.jpg"
tags: [go, markdown]
categories: [dev]
author: [tu-slug]
difficulty: beginner        # beginner | intermediate | advanced
draft: false
pinned: false               # fijado (home/orden)
series: "mi-serie"          # opcional
order: 1                    # orden dentro de la serie
aliases: ["/blog/url-vieja/"]   # URLs antiguas → redirección
links: { website: "https://…", github: "https://…" }
downloads:
  - { id: "zip", label: "Código", url: "https://…" }
---
```

**Docs** (`content/docs/*.md`):

```yaml
---
title: "Empezando"
description: "Guía rápida"
order: 1
section: guias
tags: [guide]
heroImage: "/assets/doc.jpg"
---
```

**Authors** (`content/authors/*.yaml`):

```yaml
name: "Tu Nombre"
avatar: "/assets/avatar.jpg"
bio: "Breve biografía"
website: "https://tusitio.com"
github: "https://github.com/tuusuario"
twitter: "https://twitter.com/tuusuario"
```

**Series** (`content/series/*.yaml`):

```yaml
title: "Mi serie"
description: "Descripción"
cover: "/assets/cover.jpg"
author: [tu-slug]
chapters: [primer-post, segundo-post]
```

**Pages** (`content/pages/*.md` o `.yaml`): se generan en `/<slug>/`.

```yaml
---
title: "Acerca de"
description: "Sobre este sitio"
stack: [Go, goldmark]
---
```

---

## Markdown extendido

Además de GFM (tablas, tachado, listas de tareas, autolinks):

| Sintaxis | Resultado |
|---|---|
| `[[destino]]`, `[[destino\|alias]]` | wikilink interno |
| `![[imagen.png]]` | embed (imagen) |
| `==texto==` | resaltado (`<mark>`) |
| `> [!note] Título` | callout Obsidian (`note`, `tip`, `warning`, `danger`, `important`); `[!note]-` plegable |
| `{.clase}` / `{attr}` (línea propia) | atributos al bloque anterior (incl. tablas) |
| `![alt](/img.jpg)` | figura con `<figcaption>` |
| ` ```go {3,6-7} ` | resaltado con números de línea y líneas marcadas |
| ` ```js collapsible title="utils.js" ` | bloque plegable con título |
| ` ```mermaid ` | diagrama (carga diferida) |
| `$…$` / `$$…$$` | KaTeX (carga diferida) |
| `{{youtube id}}` | shortcode integrado |
| HTML crudo | se preserva (`<details>`, `<iframe>`…) |

---

## Temas

Un **tema** es una carpeta con plantillas, assets y un manifiesto. El tema por defecto va **embebido en el binario**; si defines `build.theme_dir`, se **superpone archivo a archivo** (los que falten caen al tema por defecto).

```
themes/<nombre>/
├── theme.yaml            # manifiesto
├── templates/            # plantillas html/template
│   ├── base/páginas .html (home, blog, post, terms, taxonomy, authors, author, series, series-detail, archive, search, docs, docs-index, 404)
│   ├── partials/         # head, header, footer, card, toc, pagination, breadcrumb, docsnav, searchmodal
│   └── islands/          # HTML SSR de islas (opcional)
└── assets/
    ├── css/              # estilos
    ├── js/               # scripts
    └── islands/          # módulos JS de islas (+ css)
```

`theme.yaml`:

```yaml
name: default
version: 0.1.0
assets:
  css:
    - /assets/css/styles.css
    - /assets/css/chroma.css
  js:
    - /assets/js/theme.js
    - /assets/js/search.js
    - /assets/js/islands.js
```

- **CSS**: el generador no compila CSS; el tema aporta el suyo (Tailwind precompilado, CSS plano, etc.).
- **Overlay**: puedes sobrescribir un template/asset concreto sin copiar todo el tema.
- **Fingerprinting/minify** se aplican a los assets del tema.

---

## Componentes / Islas

Las **islas** son componentes data-driven: **no requieren Go ni framework**. Cualquiera puede crear las suyas como archivos del tema.

### Contrato

```
assets/islands/<nombre>.js      # interactivo: export function mount(el, props) { … }  (opcional)
assets/islands/<nombre>.css     # estilos (opcional)
templates/islands/<nombre>.html # HTML SSR (opcional; props como .clave)
templates/partials/<nombre>.html# componentes estáticos (sin JS)
```

- **Con SSR** → se renderiza HTML en build y el JS lo activa (progressive enhancement).
- **Sin SSR** → se emite `<div data-island="<nombre>" data-props="…">` y el JS pinta.
- **Sin JS** → es un componente **estático** (solo template → usa `{{partial}}`).

### Usarlas

Desde Markdown:

```
{{island counter start=3}}
{{island tabs items="Uno,Dos,Tres" load=visible}}
{{partial cta text="Ir a Docs" url="/docs/"}}
```

Desde plantillas del tema:

```
{{ island "tabs" (dict "items" "A,B") }}
{{ partial "cta" (dict "text" "Docs" "url" "/docs/") }}
```

- Primer token = **nombre**; el resto son **props `clave=valor`** (con comillas para espacios).
- `load` es reservado: `idle` (defecto) | `visible` | `eager`.
- Si una isla referenciada **no existe**, el build **falla**.

### Carga diferida

`assets/js/islands.js` lee `assets/islands/manifest.json` (generado en build), importa cada isla **una vez por nombre** y **solo si está en la página**, inyecta su CSS y llama `mount(el, props)`.

### Ejemplo de isla

```js
// assets/islands/counter.js
export function mount(el, props) {
  let n = Number(props.start || 0);
  const out = el.querySelector('[data-counter-out]');
  el.querySelector('[data-counter-inc]').onclick = () => (out.textContent = ++n);
  out.textContent = n;
}
```

```html
<!-- templates/islands/counter.html -->
<div class="counter">
  <button type="button" data-counter-inc>+</button>
  <output data-counter-out>{{ if .start }}{{ .start }}{{ else }}0{{ end }}</output>
</div>
```

### Islas incluidas en el tema por defecto

| Isla | Props | Ejemplo |
|---|---|---|
| `counter` | `start` | `{{island counter start=3}}` |
| `tabs` | `items` (coma) | `{{island tabs items="Uno,Dos"}}` |
| `accordion` | `items` (`Título=Cuerpo;…`) | `{{island accordion items="A=1;B=2"}}` |
| `copy` | `text`, `label` | `{{island copy text="hola"}}` |
| `spoiler` | `text`, `label` | `{{island spoiler text="secreto"}}` |
| `progress` | `value`, `max`, `label` | `{{island progress value=72 max=100}}` |
| `theme-picker` | `colors` (coma) | `{{island theme-picker}}` |
| `carousel` | `images` (coma) | `{{island carousel images="/a.jpg,/b.jpg"}}` |
| `gallery` | `images` (coma) | `{{island gallery images="/a.jpg,/b.jpg"}}` |
| `rating` | `max`, `value` | `{{island rating value=3}}` |
| `video` | `id`, `title` | `{{island video id="dQw4w9WgXcQ"}}` |
| `share` | `url`, `title` | `{{island share}}` |
| `clock` | — | `{{island clock}}` |
| `countdown` | `to`, `label` | `{{island countdown to="2030-01-01"}}` |

Partials estáticos: `cta`, `badge`, `kbd`.

> Guía completa: `/docs/components/` · Galería en vivo: `/docs/islands-gallery/`.

---

## Funcionalidades

- **Contenido**: blog, docs, autores, series, páginas; drafts; pinned; series con capítulos; slugs con subrutas; detección de duplicados.
- **Taxonomías configurables**: tags/categorías con descripción, aliases, peso, `strict`, `render_empty`, jerárquicas.
- **Markdown extendido** y compatibilidad Obsidian (ver tabla arriba).
- **Componentes/islas** data-driven, con **contenido hijo** (`{{#island …}}…{{/island}}` / `{{#partial …}}…{{/partial}}`) que envuelve Markdown.
- **Código**: Chroma (temas claro/oscuro), números de línea, líneas resaltadas, bloques plegables, botón copiar.
- **Imágenes**: variantes WebP + `<picture>`/`srcset` con `width`/`height`.
- **Navegación**: TOC (sticky), paginación, archivo, búsqueda (índice propio en Go + JS) con **modal Ctrl/⌘K**, **docs** con sidebar y breadcrumbs, y **menús anidados** (config).
- **Publicación programada**: los posts con `published` futura no se publican hasta su fecha (`-drafts` los incluye).
- **Artículo**: progreso de lectura, volver arriba, prev/next, relacionados, navegación de serie y compartir.
- **Home**: post destacado (pinned) y widgets (tags, categorías, stats).
- **UI**: dark mode con persistencia, View Transitions, Giscus, KaTeX/Mermaid/PhotoSwipe (carga diferida).
- **SEO/Meta**: canonical por página, favicon + manifest + `theme-color` generados, JSON-LD (WebSite/BlogPosting/**CollectionPage**/**Person**), **`rel=prev/next`** en paginación, **`sizes`** en `srcset`, OG/Twitter y **OG images nativas**, RSS/**Atom**/**JSON Feed**, sitemap y **`robots.txt`**.
- **URLs**: **aliases/redirects** por frontmatter (páginas de redirección).
- **Datos**: carpeta **`data/`** (JSON/YAML/CSV) accesible en plantillas (`{{ index .Data "…" }}`).
- **Enlaces**: los externos en Markdown llevan `target="_blank" rel="noopener noreferrer"`.
- **Imágenes**: WebP + `srcset`, **LQIP** (placeholder difuminado) y **presets** (`fit`/`fill`/`focal`/`blur`) vía `{{image}}`; **AVIF opt-in** (`-tags avif`).
- **Offline**: KaTeX, Mermaid y PhotoSwipe **vendorizados** (sin CDN).
- **i18n**: textos de la interfaz externalizados en `i18n/<lang>.{json,yaml}` (overlay sobre los defaults embebidos). El build **avisa** si falta el archivo o claves (p. ej. una taxonomía nueva `taxonomy.<name>`). Ver abajo.
- **Verificación**: el build valida referencias (assets y enlaces internos) y **falla** si hay roturas.
- **Assets**: fingerprinting y minificado propio de HTML/CSS.

---

## Estructura del proyecto

```
cmd/godian/                          entrada CLI
internal/domain/                     entidades y configuración
internal/application/                casos de uso (puertos) y view models
internal/i18n/                       traducciones es/en
internal/infrastructure/
  config/                            carga de site.yaml
  content/                           descubrimiento y parseo de colecciones
  frontmatter/                       frontmatter YAML
  markdown/                          pipeline goldmark + extensiones + shortcodes/islas
  render/                            motor html/template
  theme/                             tema embebido + overlay
  islands/                           registro de componentes/islas
  images/                            variantes WebP + <picture>
  og/                                OG images nativas
  feed/                              RSS + sitemap
  output/                            escritura de artefactos
  minify/                            minificado (HTML/CSS/JS, tdewolff/minify)
  verify/                            verificación de referencias
internal/presentation/
  cli/                               comandos
  server/                            dev server + hot reload
content/                             contenido del sitio (ejemplo)
themes/default/                      tema por defecto (embebido; plantilla editable)
tools/genchroma/                     generador del CSS de Chroma
deploy/                              ejemplos de despliegue
```

---

## Desarrollo

```sh
make build        # bin/godian (con símbolos, depurable)
make release      # bin/godian sin símbolos (-s -w) y con la versión inyectada
make run          # go run ./cmd/godian build
make serve        # go run ./cmd/godian serve (dev server en :4321)
make test         # tests (unit + integración end-to-end)
make cover        # cobertura
make vet          # go vet
make staticcheck  # análisis estático
make gosec        # gosec (reglas por diseño excluidas y documentadas)
make vuln         # govulncheck
make security     # vuln + gosec
make lint         # fmt + vet + staticcheck
make fmt          # gofmt
make tidy         # go mod tidy
make clean        # rm -rf bin
```

- Regenerar el CSS de resaltado (Chroma): `go run ./tools/genchroma`.
- `-race`: no soportado en android/arm64; disponible en CI (linux, `CGO_ENABLED=1`).

## Despliegue

Ver `deploy/README.md` (GitHub Pages, Netlify, Cloudflare Pages, Vercel).

## Licencia

MIT
