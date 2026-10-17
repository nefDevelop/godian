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

## 🧭 Backlog (competencias y mejoras)

> Comparado con Hugo / Zola / Eleventy / Jekyll / Astro. Cada ítem con una nota de qué es.
> Orden orientativo; se irá depurando ítem a ítem.

### Feeds / SEO (pequeños, alto valor)
- [ ] **Sitemap index** (sitios grandes) y **feeds por taxonomía**.
- [ ] **Publicación programada**: los posts futuros no se publican hasta su fecha.

### Datos y contenido
- [ ] **Render hooks** adicionales configurables (imágenes, headings).
- [ ] **Menús anidados** y por idioma.
- [ ] **Page bundles / secciones anidadas** (recursos co-localizados con la página).

### Imágenes
- [ ] **LQIP/placeholder** difuminado y filtros (`fit` / `fill` / `blur`).

### Rendimiento / entrega
- [ ] **Incremental builds** (reconstruir solo lo cambiado).
- [ ] **HMR granular** (hot swap de CSS sin recargar).
- [ ] **PWA / service worker** (offline).
- [ ] **Minificado de JS** (o integrar un minificador seguro).
- [ ] **Vendorizar** KaTeX/Mermaid/PhotoSwipe para funcionar 100% offline.

### Multilingüe (grande)
- [ ] **Multilingüe real**: contenido por idioma, `hreflang` y selector. (Hoy solo i18n de UI.)

### Publicación
- [ ] Publicar el tema `default` como plantilla reutilizable.

### Descartado por diseño (documentado)
- [x] **Sass/SCSS, PostCSS, bundling JS** (Hugo Pipes): el generador no compila CSS/JS; el tema los aporta.
- [x] **MDX / hidratación de frameworks** (Astro): usamos islas vanilla sin hidratación.
- [x] **Sistema de módulos** (Hugo Modules): tema embebido + overlay.
- [x] **Compatibilidad con `astro-cms`**: descartada (su editor escribe `category`/`author` como strings y en otra carpeta; no encaja sin modificarlo).
