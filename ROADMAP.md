# Roadmap — godian

> Generador de blog estático en Go. Ver `PLAN.md` para el diseño y las decisiones.

## ✅ Completado

### F0 — Base
- [x] Módulo Go (`github.com/nefdevelop/godian`)
- [x] Configuración del sitio (`site.yaml` → `domain.Config`)
- [x] Modelo de dominio: BlogPost, Doc, Author, Series, Page, Content
- [x] Carga de colecciones con frontmatter compatible con `astro-ri`
- [x] CLI: `godian build` (resumen de contenido)
- [x] Separación frontmatter/cuerpo + test unitario
- [x] Makefile (build/test/vet/fmt/run)

### F1 — Tema y render
- [x] Carga de tema (`theme.yaml` + `go:embed` del tema default + overlay en disco)
- [x] Motor de plantillas (`html/template`)
- [x] Pipeline markdown (goldmark + GFM)
- [x] Páginas: home, blog list, post, 404
- [x] RSS + sitemap (sin dependencias)
- [x] Copia de assets del tema

### F2 — Derivados
- [x] Tags (jerárquicos, con conteos) + páginas por tag
- [x] Categorías + páginas por categoría
- [x] Autores (índice + perfil) + páginas por autor
- [x] Series (índice + detalle con capítulos)
- [x] Archivo por año
- [x] Paginación (tamaño 6) en blog, tags, categorías y autores

### F2.5 — Taxonomías configurables
- [x] `taxonomies:` en `site.yaml` (name, label, url, field, prefix, hierarchical)
- [x] Términos declarados en config con `description`, `aliases`, `weight`
- [x] Merge híbrido: declarados + descubiertos en posts
- [x] `render_empty` (páginas sin posts) y `strict` (aviso de términos no declarados)
- [x] Render genérico de taxonomías (reemplaza tags/categorías hardcodeadas)

### F3 — Markdown extendido
- [x] Callouts estilo Obsidian (note/tip/warning/danger/important, plegables)
- [x] Wikilinks `[[destino|alias]]` y `==highlight==`
- [x] Embeds Obsidian `![[recurso]]`
- [x] Atributos `{.clase}` / `{attr}` aplicados al bloque anterior (tablas incluidas)
- [x] Figuras automáticas con `figcaption` desde el alt
- [x] HTML crudo preservado (`<details>`, `<iframe>`, `<video>`…)
- [x] Shortcodes propios (mecanismo + `{{youtube id}}`)

### F4 — Código e imágenes
- [x] Resaltado con Chroma (temas github/github-dark, clases)
- [x] Line numbers, título y bloques plegables + botón copiar
- [x] TOC por página (h2/h3) con seguimiento activo
- [x] Reading time en posts y tarjetas
- [x] Pipeline de imágenes: variantes WebP + `srcset` (`<picture>`), dimensiones y copia de estáticos

### F5 — Interactividad
- [x] Búsqueda: índice JSON en Go + página `/search/` + buscador en JS
- [x] Dark mode con toggle y persistencia (sin FOUC)
- [x] View Transitions (CSS `@view-transition`)
- [x] Giscus (comentarios configurables)
- [x] KaTeX, Mermaid y PhotoSwipe con carga diferida (solo si la página los usa)
- [x] Mermaid renderizado en cliente, KaTeX (client-side) y lightbox PhotoSwipe

### F6 — Meta
- [x] OG images (render nativo con fuentes Go, sin cgo)
- [x] JSON-LD (BlogPosting) + meta og/twitter
- [x] i18n (es/en) de la interfaz vía config `site.language`
- [x] Dev server (`godian serve`) con recarga automática (SSE + polling)

### F7 — Calidad
- [x] Verificación de referencias en build (assets/enlaces; falla si hay roturas)
- [x] Tests (frontmatter, application, markdown, images, i18n, verify)
- [x] Configs de deploy (GitHub Pages, Netlify, Cloudflare Pages, Vercel)
- [x] README/ROADMAP/PLAN actualizados

### F8 — Pulido
- [x] Contenido propio de ejemplo + prueba end-to-end
- [x] Comando `godian new [-collection] [-title] <slug>`
- [x] Fingerprinting de assets CSS/JS (`build.fingerprint`)
- [x] Minificado propio de HTML y CSS (`build.minify`)

### F9 — Estabilización
- [x] Limpieza de `dist/` en cada build (`build.clean`) — evita artefactos obsoletos
- [x] Test de integración end-to-end (build + verify sobre fixture)
- [x] Fuzz tests (frontmatter, minify)
- [x] Comando `godian version`
- [x] CI (gofmt + vet + build + test) y build reproducible (`-trimpath`)
- [x] Carga de contenido recursiva, slugs con subrutas y detección de duplicados
- [x] Mensajes de error con contexto (config y frontmatter)
- [x] Tests de casos límite (config ausente/inválida, contenido vacío, duplicados, ocultos) + `make cover`
- [x] `staticcheck` (0 hallazgos) en Makefile (`make lint`) y CI; API deprecada de goldmark sustituida
- [x] `gosec` (0 hallazgos; reglas por diseño excluidas y documentadas) y `govulncheck`
- [x] Correcciones gosec: cierre de ficheros (G104), `ReadHeaderTimeout` (G112), contexto de shutdown (G118)
- [x] Workflow de seguridad y CI con Go 1.27.2 (corrige vulnerabilidades de stdlib)
- [x] Tests del dev server (notify, signature, estáticos)
- [x] Dev server: `Content-Type` por extensión (CSS/JS se servían como `text/plain` y el navegador no aplicaba el tema)
- [x] Tema por defecto pulido (layout, grid, TOC lateral, tags enlazados, color de acento configurable)
- [x] Responsividad: menú móvil (hamburguesa), tablas con scroll horizontal, perfiles apilados, grid a 1 columna

### F10 — Tema por defecto (ampliación)
- [x] SEO/meta: canonical por página, favicon+manifest generados, `theme-color`, `WebSite` JSON-LD, OG por defecto
- [x] Accesibilidad: skip-to-content, `aria-current` en nav, anclas `#` en headings
- [x] Artículo UX: progreso de lectura, volver arriba, prev/next, relacionados, navegación de serie, compartir
- [x] Documentación: `/docs/` con sidebar, índice y breadcrumbs
- [x] Home: post destacado (pinned) + widgets (tags, categorías, stats)
- [x] Búsqueda modal con Ctrl/⌘K
- [x] Markdown: líneas resaltadas (`{3,6-7}`), callouts extra, tablas ordenables
- [x] Config: `nav`, `social`, `support` y `analytics` (opt-in)

### F11 — Componentes / islas
- [x] Contrato data-driven: `assets/islands/<n>.{js,css}` + `templates/islands/<n>.html` (SSR opcional)
- [x] Shortcodes `{{island n k=v}}` y `{{partial n k=v}}`; funcs de plantilla `island`/`partial`/`dict`
- [x] Runtime `islands.js` con manifest, carga diferida (`idle|visible|eager`) e import único por nombre
- [x] `manifest.json` generado + minify; error de build si la isla no existe
- [x] Ejemplos: `counter`, `tabs`, `accordion`, `copy`, `spoiler`, `progress`, `theme-picker`, `carousel`, `rating`, `video`, `share`, `clock`, `countdown`; partials `cta`, `badge`, `kbd`
- [x] Documentación: `/docs/components/` (guía) y `/docs/islands-gallery/` (galería en vivo)

### F12 — Publicación, feeds y datos (Tanda 1)
- [x] `robots.txt` generado (permite todo + Sitemap)
- [x] **Atom** (`/atom.xml`) y **JSON Feed** (`/feed.json`) además de RSS
- [x] **Aliases/redirects** por frontmatter (páginas de redirección con `noindex` + canonical)
- [x] **`data/`** (JSON/YAML/CSV) accesible en plantillas
- [x] **Render hook**: enlaces externos con `target="_blank" rel="noopener noreferrer"`
- [x] **AVIF opt-in** (build tag `avif`) + `<source type="image/avif">`
- [x] Documentación en `/docs/publicacion/` y ejemplos en `data/` y contenido

### F13 — Islas con contenido hijo (Tanda 2)
- [x] Shortcodes de **bloque** `{{#island nombre}}…{{/island}}` / `{{#partial …}}…{{/partial}}` que envuelven Markdown
- [x] Contenido renderizado pasado como `content` (template.HTML) al componente
- [x] `tabs`/`accordion` con **secciones** (`##`) como contenido hijo; partial `note`
- [x] Los bloques dentro de fences de código se ignoran
- [x] Documentación (`/docs/components/`) y ejemplos en `/docs/islands-gallery/`
- [x] Tests (markdown + islands + integración)

### F14 — Publicación programada y menús anidados
- [x] **Publicación programada**: posts con `published` futura excluidos (sitio/feeds/sitemap); `-drafts` los incluye
- [x] Conteo de "programados" en el resumen del build
- [x] **Menús anidados**: `nav` con `children` (submenús desplegables; sublista en móvil)
- [x] Docs (`/docs/publicacion/`), ejemplo en `site.yaml` y tests (integración + unit)

### F15 — Imágenes (LQIP + presets) y vendorizado
- [x] **LQIP**: placeholder diminuto difuminado (data URI) en el `<img>`
- [x] **Presets** de imagen (`fit`/`fill`/`focal`/`quality`/`blur`) en `build.image_presets`
- [x] Shortcode `{{image src=… preset=… alt=…}}`
- [x] **Vendorizado** de KaTeX, Mermaid y PhotoSwipe (sin CDN; `assets/vendor/`), carga diferida local
- [x] Docs (`/docs/imagenes/`), ejemplo (`assets/demo-1.jpg`) y tests
- [x] Imágenes de ejemplo de **dominio público (CC0)** y **audio** de ejemplo generado (con créditos en `assets/CREDITS.md`)

### F16 — Obsidian-first y `godian check`
- [x] Wikilinks **entre colecciones** (resueltos por slug/basename)
- [x] **Transclusión** de notas (`![[nota]]` incluye su contenido; si no es nota → imagen)
- [x] **Backlinks** por nota (sección "Backlinks")
- [x] Comando **`godian check`**: imágenes referenciadas/huérfanas, `alt` faltante, series con capítulos inexistentes, títulos vacíos
- [x] Docs (`/docs/obsidian/`), tests (linkindex + check) y ejemplos

### F17 — SEO / rendimiento
- [x] `rel="prev"`/`rel="next"` en listados paginados
- [x] `sizes` en el `srcset` de las imágenes
- [x] JSON-LD **CollectionPage** (taxonomías) y **Person** (autores)
- [x] Tests (integración: paginación + rel=next)

### F18 — Feeds por taxonomía e isla gallery
- [x] **Feed RSS por término** de taxonomía (`/tags/<t>/rss.xml`, `/categories/<c>/rss.xml`) + link en `<head>`
- [x] Isla **`gallery`** (rejilla + lightbox PhotoSwipe vendorizado)
- [x] Docs/gallery y tabla de islas; test de integración (feed por término)

### F19 — UI externalizada (i18n)
- [x] Textos de interfaz de las plantillas pasados a claves `t(...)` (skip, share, related, backlinks, featured, stats, links, all posts, previous, chapter, website, not found, no series…)
- [x] Textos por defecto en `internal/i18n` + **overlay** `i18n/<lang>.{json,yaml}` (config `i18n_dir`)
- [x] El build **avisa** si falta el archivo o claves (`taxonomy.<name>` incluidas)
- [x] Textos de JS (copiar/`Copied!`, “sin resultados”) vía `window.__i18n`
- [x] `i18n/en.json` y `i18n/es.json`; docs en README

### F20 — Islas documentadas, datos en uso y páginas renderizadas
- [x] Tabla de islas de `components.md` completa (`video`, `share`, `clock`, `countdown`)
- [x] `data/equipo.csv` se muestra en la home (widget "Equipo", clave `ui.team`)
- [x] Colección `pages/` renderizada en `/<slug>/` (template `page` con TOC y `stack`)
- [x] `about` enlazado en `nav` de `site.yaml` y test de integración para `pages/`
- [x] Resaltado de líneas: las líneas vacías no se resaltan (evita barras grises); color de `hl` distinto al de los números de línea; demo con bloque normal y bloque resaltado

### F21 — Nuevas islas y partials (calidad)
- [x] Islas nuevas: `modal`, `popover`, `progress-ring`, `toast`, `chart` (bar/pie/donut), `audio`, `lightbox`, `count-up`, `steps`, `timeline`, `qr-code`, `kanban`, `calendar`
- [x] Partials nuevos: `stat`, `avatar`, `chip`, `figure`, `quote`, `divider`, `embed`, `tooltip`; `callout` unifica los antiguos `alert`/`note`
- [x] Accesibilidad: `tabs` con roles ARIA y navegación por teclado; `carousel` con teclado + dots + `alt`; `chart` con `role="img"` y descripción; `progress-ring` y `rating` con ARIA; `modal` con foco gestionado y `aria-labelledby`; `qr-code` con `title` y aviso de límite; `calendar` construido por DOM (sin `innerHTML`)
- [x] `calendar` con `source="site"`: marca los días con posts/docs (genera `content-dates.json`) y enlaza al contenido
- [x] Unificación de avisos: `alert`/`note` eliminados; `callout` como componente único
- [x] `share` con i18n para "copiar enlace"
- [x] Nueva isla `timeline` (`code-tabs` cubierto por `tabs` con secciones `##`)

### F22 — Accesibilidad (WCAG / axe-core)
- [x] Landmarks con etiquetas únicas (nav principal, pie, serie, entradas, docs, paginación, TOC)
- [x] Anclas de encabezado con nombre accesible; `:focus-visible` global; `.sr-only`
- [x] Etiquetas en `progressbar`, `output` y task lists; `<title>` en los SVG (chart, qr-code, progress-ring)
- [x] Contraste AA: paleta de Chroma completa en oscuro (`genchroma`), color de acento como texto (`--accent-text`), badge/chip/tag/fecha
- [x] Objetivos táctiles ≥24px (rating, carousel, kanban)
- [x] Enlaces externos avisan de "se abre en ventana nueva"; `cite` en `quote`; kanban con alternativa de teclado (◀ ▶)
- [x] Escaneo **axe-core**: 0 violaciones (claro y oscuro) en home, post, docs, 404, tags y autores

### F23 — Rendimiento (Lighthouse)
- [x] Imágenes con dimensiones/`aspect-ratio` explícitos (miniatura de vídeo, `figure`, carousel, lightbox)
- [x] `font-display: swap` en las fuentes de KaTeX
- [x] Eliminado CSS muerto (componentes retirados) y transiciones no compuestas (menos "unused CSS" y sin animaciones no compuestas)
- [x] **Mermaid con carga diferida**: el bundle (~3,5 MB) se descarga solo cuando el primer diagrama entra en pantalla (`IntersectionObserver`)
- [x] Botones primarios (`.cta`, `.modal-trigger`, `.modal-action`) con acento oscurecido → contraste 6.75:1
- [x] **CSS de cada isla en su propio archivo** (`assets/islands/<n>.css`): se inyecta solo si la isla está en la página → `styles.css` de ~30,6 KB a 23,2 KB en todas las páginas
- [x] **gzip en el dev server** (HTML/CSS/JS/JSON/SVG): la home de ejemplo pasa de 30,4 KB a 7,3 KB
- [x] **Minificado de JS** con `tdewolff/minify` (también HTML y CSS)
- [x] `progress` con `transform` (animación compuesta) y dimensiones explícitas/CLS en imágenes de islas (`gallery`, `carousel`, `lightbox`)
- [x] **`chroma.css` solo en páginas con código** (`NeedsCode`) y **KaTeX diferido por viewport**
- [x] **CSS por bloques (`data-css` + `assets/css/chunks/`)**: `article-ux`, `docs`, `lists` se inyectan solo en las páginas que los usan (sincrónico, sin FOUC); `styles.css` base baja de ~30 a ~21 KB
- [x] **CSS por bloques** (arriba) y **borde visible en modo claro** (`--border: #d1d5db`)
- [ ] Mermaid: ESM moderno con chunks por tipo **descartado** (el bundle pasaba de 3,5 MB/1 fichero a 5,5 MB/104); se mantiene el IIFE con carga diferida. Backlog: build ligero por tipo de diagrama
- [x] Prop `poster` en la isla `video` y en el shortcode `{{youtube}}` (miniatura propia)
- [ ] Backlog: dividir el CSS base por página, build de Mermaid con carga diferida por tipo de diagrama

### F24 — Tema extraído y `new-theme`
- [x] Tema movido a `themes/default/` (plantilla editable) y embebido vía el paquete `themes` (`//go:embed`)
- [x] Comando `godian new-theme <dir>` (+ `theme.Scaffold`) para copiar y personalizar el tema
- [x] Overlay: un `theme_dir` en disco se superpone al tema embebido archivo a archivo
- [x] README y `tools/genchroma` actualizados al nuevo path

### F25 — Feeds/SEO y render hooks (integrados)
- [x] **Sitemap index**: `sitemap.xml` se parte en `sitemap-N.xml` + índice si supera 50 000 URLs (los sitios pequeños siguen con un único sitemap)
- [x] **Render hooks configurables** (`render:` en `site.yaml`): `external_links` (target/rel), `headings` (anclar + símbolo) e `images` (loading/decoding/figure)
- [x] Tests de feed (sitemap) y de render hooks (enlaces, encabezados, imágenes)

### F26 — Fuentes autohospedadas y facade de vídeo
- [x] `fonts:` en `site.yaml`: descarga y autohospeda fuentes de Google Fonts en el build (`/assets/fonts` + `@font-face` con `font-display`); varias familias, pesos, estilos y subconjuntos; fallback con aviso si no hay red
- [x] Shortcode `{{youtube}}` en **facade**: miniatura + botón; el iframe (`youtube-nocookie`) se carga al pulsar → sin cookies ni peticiones a terceros hasta entonces
- [x] Isla `video` con `youtube-nocookie`

## 🧭 Backlog (competencias y mejoras)

> Comparado con Hugo / Zola / Eleventy / Jekyll / Astro. Cada ítem con una nota de qué es.
> Orden orientativo; se irá depurando ítem a ítem.

### Datos y contenido
- [ ] **Menús por idioma**.

### Rendimiento / entrega
- [ ] **Incremental builds** (reconstruir solo lo cambiado).
- [ ] **HMR granular** (hot swap de CSS sin recargar).
- [ ] **PWA / service worker** (offline).
- [ ] Dividir el CSS base por página; build de Mermaid con carga diferida por tipo de diagrama.

### Multilingüe (grande)
- [ ] **Multilingüe real**: contenido por idioma, `hreflang` y selector. (Hoy solo i18n de UI.)

### Publicación
- [x] Tema `default` como plantilla reutilizable: vive en `themes/default/` (editable) y `godian new-theme` lo copia (F24).
- [ ] Publicarlo como plantilla independiente en un repositorio aparte.

### Descartado por diseño (documentado)
- [x] **Sass/SCSS, PostCSS, bundling JS** (Hugo Pipes): el generador no compila CSS/JS; el tema los aporta.
- [x] **MDX / hidratación de frameworks** (Astro): usamos islas vanilla sin hidratación.
- [x] **Sistema de módulos** (Hugo Modules): tema embebido + overlay.
- [x] **Compatibilidad con `astro-cms`**: descartada (su editor escribe `category`/`author` como strings y en otra carpeta; no encaja sin modificarlo).
- [x] **Page bundles**: no prioritario (los slugs con subrutas ya cubren el caso; no aporta al enfoque).
