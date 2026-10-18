# Plan — `godian`: generador de blog estático en Go

> Objetivo: generador estático **minimalista, zero-framework, escrito en Go**, con **paridad funcional (~90-95%)** respecto al tema Astro **`astro-ri`**, sin perder funciones clave.

---

## 1. Objetivo y encaje

- `godian` es un **generador de sitio estático** (SSG) en Go puro.
- **Lee el mismo contenido Markdown** que hoy consume el tema `astro-ri` (compatibilidad de
  esquema de frontmatter como referencia de partida).
- Entrada: **contenido** + **tema (personalización)** + **config**; todos con rutas configurables.
- Salida: `dist/` estático, desplegable en Cloudflare Pages, Vercel, Netlify o GitHub Pages.
- **Build = binario único**, sin toolchain Node/Rust en el flujo principal.

---

## 2. ¿Por qué Node o Rust?

No es por el generador en sí, sino por piezas concretas del stack de `astro-ri`:

| Pieza | Lenguaje real | ¿Equivalente en Go? | Impacto |
|---|---|---|---|
| Tailwind CSS v4 | **Rust** (`@tailwindcss/oxide`) + puente **Node** | **No** | Alto |
| Astro / Vite | Node | Sí (no hacen falta) | — |
| Pagefind (indexador) | **Rust** + UI JS | Sí (índice propio en Go) | Medio |
| Expressive Code + plugins remark/rehype | JavaScript | Sí (`chroma` + goldmark) | — |
| Sharp / Satori (OG images) | Node + libvips / JS | Sí (render nativo Go) | Medio |

**Conclusión:** Node/Rust solo son necesarios si queremos **Tailwind v4 tal cual** y/o **Pagefind tal cual**.
Al separar el CSS en el **tema** (§4), esa decisión deja de ser del generador y pasa a ser del autor del tema.

---

## 3. Arquitectura (Clean Architecture + DDD)

Aplica por tratarse de un dominio nuevo con entidades y casos de uso propios.

```
cmd/godian/                 # entrypoint CLI
internal/domain/            # entidades: Post, Page, Author, Series, Tag, Category, Theme + reglas
internal/application/       # casos de uso: Build, RenderPage, Watch
internal/infrastructure/
  content/                  # descubrir y parsear colecciones (fs + frontmatter/yaml)
  markdown/                 # pipeline goldmark + extensiones propias
  theme/                    # cargar theme.yaml, resolver templates y assets
  render/                   # motor de plantillas (html/template)
  feed/                     # rss, sitemap, json-ld
  search/                   # generación de índice
  assets/                   # copia/minificado/fingerprint de css, js, imágenes
internal/presentation/
  cli/                      # comandos (build / serve / new)
  server/                   # dev server + hot reload
```

**Regla de dependencia:** `domain` no importa nada de `infrastructure` ni `presentation`.
Los casos de uso de `application` orquestan dominio + interfaces; la infraestructura implementa.

---

## 4. Build configurable y theming

El build es **config-driven** con tres raíces separadas: **contenido**, **personalización (tema)** y **objetivo**.

```
content/            # markdown (fuente de datos)         -> configurable
themes/default/     # personalización (templates+css+js) -> configurable
dist/               # objetivo                            -> configurable
```

### Config propuesta (`site.yaml`)

```yaml
site:
  title: ...
  url: https://...
build:
  content_dir: content
  theme_dir: themes/default
  output_dir: dist
  minify: true
  fingerprint: true
collections:            # colecciones de contenido
  - name: blog
    folder: blog
    ...
```

### Contrato del tema (carpeta de personalización)

```
themes/default/
├── theme.yaml          # manifiesto: nombre, versión, lista de css/js, template base
├── templates/          # .html (html/template)
│   ├── base.html
│   ├── post.html
│   └── ...
└── assets/
    ├── css/            # styles.css  (Tailwind precompilado, CSS plano, lo que sea)
    ├── js/             # katex.js, mermaid.js, photoswipe.js...
    └── fonts/
```

`theme.yaml` declara qué CSS/JS incluir y en qué orden; el generador arma los `<link>`/`<script>`
y copia los assets a `dist/`. **El generador no interpreta el CSS**: solo lo copia/minifica/fingerprintea.

### Flujo del build

```
config + content_dir ─┐
                      ├─► [generador Go] ─► dist/
theme_dir (templates+assets) ─┘
```

### Cómo resuelve el dilema de Tailwind

Como el generador no compila CSS, la decisión pasa al **autor del tema**:

- Tema con Tailwind → precompilar una vez a `assets/css/styles.css` (Node/Rust solo en ese paso, opcional).
- Tema pure-Go → CSS plano con custom properties.
- Tema con CDN → poner el `<link>` en `base.html`; el generador no se entera.

**Tema `default` embebido en el binario con `go:embed`**, para funcionar sin carpeta externa; `theme_dir` lo sobrescribe.

---

## 5. Pipeline de build

1. Cargar config.
2. Resolver tema (embebido o `theme_dir`), cargar `theme.yaml` y manifest de assets.
3. Descubrir y parsear colecciones (frontmatter + markdown).
4. Validar (equivalente a los schemas Zod de Astro).
5. Pipeline markdown (goldmark + extensiones).
6. Resolver derivados: tags, categorías, series, archive, paginación, similares.
7. Render con plantillas del tema.
8. Generar RSS + sitemap + JSON-LD.
9. Construir índice de búsqueda.
10. Copiar/minificar/fingerprint assets del tema → `dist/`.
11. **Verificar** el resultado (ver abajo) y fallar el build si algo no cuadra.

### Verificación en build (sin compilar JS)

Como no compilamos ni empaquetamos JS, añadimos una **fase de validación** al final del pipeline
(un *linter* de build, no un compilador). Detecta referencias rotas antes de desplegar:

| Comprobación | Qué valida | Cómo |
|---|---|---|
| Assets del tema | Cada CSS/JS declarado en `theme.yaml` existe | fs + manifest |
| Referencias en HTML | Todo `<link href>`/`<script src>` emitido existe en `dist/` | escaneo del HTML generado |
| Templates | Todo template invocado resuelve (tema o fallback `default`) | registro de templates |
| Shortcodes | Todo shortcode usado está registrado | parser propio en el pipeline |
| Enlaces internos / wikilinks | `[[slug]]` y enlaces relativos resuelven a una página | índice de rutas |
| Imágenes embebidas | `![[img.png]]` existe | fs |
| Frontmatter | Campos requeridos y tipos | schema (§5.4) |

**Alcance honesto:** Go **no puede** analizar lógica JS (llamadas a funciones, `addEventListener`, etc.)
de forma fiable. Por eso:

- Mantenemos el JS **mínimo y vendored** (librerías + scripts pequeños propios), con poca superficie de error.
- La verificación cubre **existencia y referencias** (assets/links/templates), que es donde están los fallos reales.
- Si se quisiera análisis de funciones JS, requeriría Node (ESLint/`tsc`), **solo en CI, opcional y no obligatorio**.

---

## 6. Dependencias Go (DECIDIDO: set minimalista)

Se adopta el **set minimalista**: se implementa a mano todo lo trivial y solo se añaden
librerías donde no hay equivalente razonable en stdlib.

| Función | Decisión |
|---|---|
| Markdown | `yuin/goldmark` (+ `extension` GFM) |
| YAML / frontmatter | `gopkg.in/yaml.v3` |
| Resaltado de código | `alecthomas/chroma/v2` *(no hay equivalente en stdlib)* |
| Imágenes (WebP, §11) | `github.com/gen2brain/webp` (`nodynamic`) |
| Resize / fuentes (imágenes y OG) | `golang.org/x/image` |
| Plantillas | stdlib `html/template` |
| Watch / reload | **propio** (polling de `fsnotify`-free) |
| Slugs | **propio** |
| RSS / sitemap / JSON-LD | **propio** (stdlib `encoding/xml`, `encoding/json`) |
| Minify | `tdewolff/minify` (HTML/CSS/JS) |
| OG images | **nativo** (`golang.org/x/image/font` + dibujo stdlib) |

**Descartadas:** `gorilla/feeds`, `fsnotify`, `gosimple/slug`,
`disintegration/imaging`, `fogleman/gg`, `a-h/templ`.

**Pendiente de aprobación explícita (AGENTS.md):** `goldmark`, `yaml.v3`, `chroma`,
`gen2brain/webp`, `golang.org/x/image`, `tdewolff/minify`.

---

## 7. Paridad funcional por área

| Área (en `astro-ri`) | Plan en Go | Dificultad |
|---|---|---|
| Posts, tags, categorías, autores, series, archive, paginación | lógica propia | Fácil |
| Callouts, wikilinks, `==highlight==`, figuras, tablas, atributos | extensiones goldmark | Fácil |
| Highlighting (line numbers, copy, colapsables) | `chroma` + JS | Fácil |
| Búsqueda | índice JSON en Go + buscador vanilla | Media |
| KaTeX / Mermaid / PhotoSwipe / Giscus | assets JS locales (vendored) | Fácil |
| TOC, reading progress, dark mode, View Transitions, Ctrl+K, botones compartir | JS vanilla | Fácil |
| RSS / sitemap / JSON-LD / SEO / OG meta | generación en Go | Fácil |
| Paleta generativa desde seed | OKLCH/HSL en Go | Fácil |
| **MDX** | sin compilador JS no hay componentes | **Riesgo** (hoy solo 1 `.mdx`) |
| **OG images** | render nativo Go, no Satori | **Riesgo medio** (fidelidad visual) |
| **Tailwind CSS v4** | no lo compila el generador; lo aporta el tema | **Resuelto por §4** |

---

## 8. Relación con Hugo

**Sí, convergemos en el núcleo de Hugo** (es un diseño probado y sensato):

- Binario único en Go.
- Modelo `content` + `theme`.
- Markdown con **goldmark** (el mismo motor que usa Hugo).
- Plantillas con **`html/template`**.
- Taxonomías, RSS, sitemap.

**Diferencias (nuestro posicionamiento):**

| Aspecto | Hugo | godian |
|---|---|---|
| Alcance | Plataforma general | Generador enfocado a `astro-ri` |
| CSS/SCSS | Embebe `dart-sass` | No compila CSS (lo aporta el tema) |
| Bundling JS | Embebe `esbuild` | No |
| Shortcodes | Sistema completo | Mínimo / shortcodes propios |
| Módulos | Sistema de módulos | No |
| Funciones de plantilla | Cientos | Set mínimo |
| Wikilinks / callouts Obsidian | No nativo | Nativo |
| Esquema de frontmatter | Propio | Compatible con `astro-ri` |
| Pagefind / Giscus / PhotoSwipe | No por defecto | Parte del tema `default` |

**Por qué no usar Hugo directamente:** esquema de frontmatter compatible con `astro-ri`,
markdown extendido integrado, cero toolchain externo y una base de código pequeña y mantenible.
**Riesgo:** si queremos features de Hugo, reinventarlas tiene coste; hay que ceñirse al alcance.

---

## 9. Compatibilidad con Obsidian

**Clave:** la compatibilidad con Obsidian es de **sintaxis Markdown**, no de JavaScript.
Se resuelve entera en el **pipeline de markdown en Go** (goldmark + extensiones propias), igual que
`astro-ri` ya hace con `remark-obsidian`. El navegador no ejecuta nada de Obsidian.

### Sintaxis a soportar (lectura)

| Sintaxis Obsidian | Ejemplo | Se convierte en |
|---|---|---|
| Wikilink | `[[nota]]`, `[[nota\|alias]]` | enlace interno (`/blog/nota`) |
| Wikilink con ancla | `[[nota#seccion]]` | enlace interno con `#seccion` |
| Embed de nota | `![[nota]]` | inclusión/transclusión del contenido |
| Embed de media | `![[img.png]]` | `<img>` (con lightbox opcional) |
| Callout | `> [!note] ...` | bloque HTML de callout |
| Highlight | `==texto==` | `<mark>` |
| Tags en línea | `#tag` | enlace a `/tags/tag` |
| Frontmatter | `tags:`, `aliases:` | metadatos de la colección |

### Qué implica para el JS

- **No hay JS de Obsidian**: no se ejecutan plugins, Dataview, `dataviewjs` ni scripts embebidos en las notas.
- El único JS implicado es compartido con el resto del sitio: lightbox de imágenes y buscador.
- **Cada extensión de Obsidian es un transformer en Go**, verificable en la fase de validación (§5).

### Alcance recomendado

- **Sí:** compatibilidad de **sintaxis de lectura** (tabla anterior) para poder escribir en un *vault* de Obsidian
  y publicarlo tal cual. *Es lo que ya hace `astro-ri`.*
- **No:** ejecutar plugins/comunidad de Obsidian ni su JS. Queda fuera de alcance (requeriría un runtime JS
  y rompería el objetivo "zero-framework"). Si algún día se quiere, sería un paso opt-in con Node, nunca el default.

---

## 10. Fases propuestas

- **F0** — CLI + config + modelo de dominio + carga de colecciones.
- **F1** — Tema (`embed` + `theme_dir`), plantillas + blog/index + RSS + sitemap.
- **F2** — Tags / categorías / autores / series / archive + paginación.
- **F3** — Markdown extendido (callouts, wikilinks, atributos, figuras, tablas) + compatibilidad Obsidian (§9).
- **F4** — Pipeline de imágenes (§11: resize + WebP + `srcset`) + Chroma + TOC + reading time.
- **F5** — Búsqueda + KaTeX + Mermaid + PhotoSwipe + dark mode + View Transitions.
- **F6** — OG images + JSON-LD + i18n + dev server / hot reload.
- **F7** — Verificación de build (§5) + `make test` (Go test) + configs de deploy + README/ROADMAP.

---

## 11. Estrategia de imágenes y entrega

### Hallazgo: `astro-ri` NO optimiza imágenes

Verificado en el código: `astro-ri` renderiza `<img src=... loading=... decoding=...>` con las imágenes
**tal cual** (JPEG/PNG crudos). No usa `astro:assets` `Image`/`Picture`, ni `getImage`, ni `srcset`,
ni genera WebP/AVIF. El único uso de `astro:assets` es `Font`.

> Sí, Astro+**Sharp** *puede* optimizar en build, pero **este tema no lo usa**. Por tanto,
> **la paridad no requiere ningún códec ni CDN.**

### Conclusión

- Paridad con `astro-ri` = `<img>` crudo, **pero la decisión del proyecto es ir más allá**:
  la optimización de imágenes es **parte del build actual**, no una mejora futura.
- **WebP se genera en el build** (redimensionar + variantes WebP + `srcset`), por defecto.
- **AVIF: no** por defecto (+5.4 MB y lento); solo por build tag `avif` si se necesita.
- Coste asumido en el default: incluir `gen2brain/webp` → **+~2.5 MB**.

### Medido (android/arm64, `CGO_ENABLED=0`, `-s -w`)

| Códec | Librería | Runtime | Delta |
|---|---|---|---|
| WebP lossless | `nativewebp` | Go puro | **+0.32 MB** |
| WebP lossy (default) | `gen2brain/webp` (`nodynamic`) | wasm2go | **+2.49 MB** |
| AVIF (opt-in) | `gen2brain/avif` (`nodynamic`) | wazero | **+5.37 MB** |
| WebP + AVIF | ambos | wasm2go + wazero | **+7.73 MB** |

Verificados en runtime sin cgo. `gen2brain/webp` usa **wasm2go** (transpile), `gen2brain/avif` usa **wazero**.

### Decisión (fijada)

1. **El pipeline de imágenes es parte del build**: redimensiona, genera variantes **WebP** y emite
   `<picture>`/`<img>` con `srcset` + `loading` + `decoding` + `width`/`height`.
2. **WebP por defecto**: `gen2brain/webp` con `nodynamic` (sin cgo, +2.5 MB). Cubre lossy y lossless.
3. **AVIF opt-in** por build tag `avif` (no por defecto).
4. **Arquitectura:** interfaz de imágenes en `domain`; implementación WebP en `infrastructure` (default);
   AVIF como implementación opcional tras build tag.
5. **cgo descartado**; `cwebp`/`wazero`/CDN quedan como alternativas opcionales, nunca obligatorias.

---

## 12. Decisiones

### Ya resueltas (por el modelo de theming)

- **CSS / Tailwind** → el generador **no compila CSS**; cada tema aporta su CSS. Tema `default` con CSS plano;
  Tailwind opcional precompilado en el tema (§4).
- **Herencia de tema** → **base + override por archivo**: si un template/asset falta en `theme_dir`,
  se usa el del tema `default` embebido (como Hugo). *(recomendado)*
- **Assets** → fingerprinteo **configurable** (`fingerprint: true`); activo en prod, apagado en dev. *(recomendado)*
- **Post-proceso CSS** → solo **copiar/minificar** (sin autoprefixer ni crítico inline). *(recomendado)*
- **Nº de temas** → un tema activo por config, con herencia al `default`. *(recomendado)*
- **Tema por defecto** → **`go:embed`** en el binario; `theme_dir` lo sobrescribe. *(recomendado)*
- **Manifiesto del tema** → **`theme.yaml` mínimo** (lista de assets + template base) **combinado con convención
  de carpetas** `templates/` y `assets/`. *(recomendado)*

### Resueltas por la estrategia de imágenes (§11)

- **Imágenes** → **parte del build**: resize + variantes **WebP** + `srcset` (no opt-in). *(fijado)*
- **WebP** → `gen2brain/webp` `nodynamic` por defecto (+2.5 MB, sin cgo). *(fijado)*
- **AVIF** → opt-in por build tag `avif` (no por defecto). *(fijado)*
- **cgo** → **descartado** para el core (rompe static/cross-compile). *(fijado)*

### Resueltas por el usuario (finales)

1. **Nombre y módulo** → `github.com/nefdevelop/godian`. *(fijado)*
2. **MDX** → **descartar** (convertir el único `.mdx` a `.md`) + **shortcodes propios**. *(fijado)*
3. **Búsqueda** → **índice propio en Go + JS** (sin Pagefind/Rust). *(fijado)*
4. **OG images** → **render nativo en Go**. *(fijado)*
5. **Plantillas** → `html/template`. *(fijado)*
6. **Dependencias** → **set minimalista** (§6). *(fijado)*
7. **i18n** → se deja para **F6**. *(fijado)*
8. **Obsidian** → solo **sintaxis de lectura**, sin plugins/JS. *(fijado)*
9. **Verificación en build** → se incluye el *linter* de referencias. *(fijado)*
10. **Análisis JS** → **sin análisis estático** (build 100% Go). *(fijado)*

> No quedan decisiones pendientes. Listo para ejecutar fases.

---

## 13. Definition of Done (por fase)

- Código compila (`go build ./...`) y pasa `go vet`.
- Tests de la fase escritos; se ejecutan vía `make test` **solo cuando se indique**.
- Documentación de la fase actualizada (`PLAN.md` / `ROADMAP.md`).
- Sin dependencias nuevas fuera de las aprobadas.
