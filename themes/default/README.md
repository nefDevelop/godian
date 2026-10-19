# Tema `default`

Guía para usar, sobrescribir o crear un tema de godian. El tema por defecto va
embebido en el binario; si `build.theme_dir` apunta a una carpeta, sus archivos
**sobrescriben** los del tema embebido uno a uno. No hace falta copiar todo.

## Estructura

```
themes/<nombre>/
├── theme.yaml            # manifiesto: nombre, versión y assets globales
├── templates/
│   ├── *.html            # plantillas de página (definen "home", "post", …)
│   ├── partials/         # head, header, footer, card, toc, …
│   └── islands/          # HTML SSR opcional de islas
├── icons/                # iconos propios (opcional; sobrescriben Lucide)
└── assets/
    ├── css/
    │   ├── styles.css    # CSS base cargado en todas las páginas
    │   ├── motion.css    # tokens de movimiento, utilidades y colapso
    │   └── chunks/       # CSS por bloque (lists, docs, article-ux…)
    ├── js/               # theme.js, search.js, islands.js, motion.js…
    └── islands/          # <nombre>.js / <nombre>.css
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

El generador **no compila CSS ni JS**: copia, minifica y fingerprinta lo que
declares. Los assets de `assets/` se publican bajo `/assets/`.

## Plantillas obligatorias

El core invoca estos nombres con `html/template`. Si falta uno, la página que lo
usa falla al renderizar.

| Plantilla | Uso |
|---|---|
| `home` | portada |
| `blog` | listado paginado de posts |
| `post` | post individual |
| `terms` | índice de una taxonomía |
| `taxonomy` | listado de posts de un término |
| `authors` | índice de autores |
| `author` | perfil de autor |
| `series` | índice de series |
| `series-detail` | detalle de una serie |
| `archive` | archivo por año |
| `search` | página de búsqueda |
| `docs` | página de documentación |
| `docs-index` | índice de documentación |
| `docs-tags` | índice de tags de documentación |
| `docs-tag` | listado de docs de un tag |
| `page` | página suelta (`content/pages`) |
| `404` | página no encontrada |

Partials que el core espera encontrar (puedes sobrescribirlos, no eliminarlos):

`head`, `header`, `footer`, `card`, `pagination`, `breadcrumb`, `docsnav`,
`toc`, `toc-float`, `searchmodal`, `callout`.

## Datos disponibles (`PageData`)

Todas las plantillas reciben `application.PageData`. Campos comunes:

`Site`, `PageURL`, `Canonical`, `NavActive`, `Title`, `Description`, `Body`,
`Nav`, `CSS`, `JS`, `Year`, `Data`, `Social`, `Support`, `Analytics`, `Giscus`,
`OGImage`, `FeedURL`, `JSONLD`, `SiteJSONLD`, `NeedsMath`, `NeedsCode`.

Campos por tipo de página:

| Plantilla | Campos principales |
|---|---|
| `home` | `Posts`, `Featured`, `HomeTags`, `HomeCats`, `Stats`, `Data` |
| `blog`, `taxonomy` | `Posts`, `Page`, `Total` |
| `terms` | `Terms` |
| `post` | `Date`, `Updated`, `ReadingTime`, `Tags`, `Categories`, `Image`, `AuthorsRef`, `Difficulty`, `Downloads`, `Links`, `Pinned`, `TOC`, `Prev`, `Next`, `Related`, `Backlinks`, `SeriesNav` |
| `docs` | `Body`, `TOC`, `DocsNav`, `Breadcrumb`, `HeroImage`, `PubDate`, `UpdatedDate`, `DocTags`, `Backlinks` |
| `docs-index` | `DocsNav`, `DocTags` |
| `docs-tags` | `Terms`, `DocsNav`, `Breadcrumb` |
| `docs-tag` | `Docs`, `Total`, `DocsNav`, `Breadcrumb` |
| `authors` | `Authors` (`[]AuthorCard`) |
| `author` | `Author` (`*AuthorProfile`), `Posts` |
| `series` | `SeriesList` |
| `series-detail` | `Series` (`*SeriesDetail`) |
| `archive` | `Years` |
| `page` | `Body`, `Stack`, `TOC`, `Backlinks` |

Tipos auxiliares:

- `AuthorsRef`: `Name`, `Slug`, `URL`, `Avatar`.
- `Downloads`: `ID`, `Label`, `URL`.
- `Links`: `Website`, `Github`.
- `Terms`: `Name`, `URL`, `Count`.
- `Docs` en `docs-tag`: `Title`, `URL`, `Description`, `Date`, `HeroImage`, `Tags`.
- `Page`: `Current`, `Last`, `BaseURL`, `PrevURL`, `NextURL`.

La fuente de verdad es `internal/application/view.go`.

## CSS por chunks

`assets/css/chunks/<nombre>.css` **no** se carga en todas las páginas. Se inyecta
solo si el HTML declara `data-css="<nombre>"`. Se admiten varios chunks:

```html
<div class="wrap post-layout" data-css="article-ux lists">
```

Dependencias del tema por defecto:

| Página | `data-css` |
|---|---|
| `home`, `blog`, `terms`, `taxonomy`, `archive`, `authors`, `author`, `series`, `series-detail`, `search` | `lists` |
| `post` | `article-ux lists` |
| `docs`, `docs-index`, `docs-tags`, `docs-tag` | `docs lists` |

`lists` contiene `.card*`, `.tags`, `.tag`, `.cat`, `.term*`, `.difficulty`,
`.pinned-badge` y los listados. Si una plantilla usa alguno de esos estilos,
debe declarar `lists` o mover esos estilos a su propio chunk.

## Islas y partials

Contrato:

```
assets/islands/<nombre>.js       export function mount(el, props) { … }
assets/islands/<nombre>.css      estilos (opcional)
templates/islands/<nombre>.html  HTML SSR (opcional)
templates/partials/<nombre>.html componente estático
```

- Con SSR, el HTML se renderiza en build y el JS lo activa.
- Sin SSR, se emite `<div data-island="<nombre>" data-props="…">`.
- `load` es reservado: `idle` (defecto), `visible`, `eager`.
- Si se referencia una isla inexistente, el build falla.

Desde Markdown: `{{island counter start=3}}`, `{{partial cta text="Docs" url="/docs/"}}`.
Desde plantillas: `{{ island "tabs" (dict "items" "A,B") }}`.

## Movimiento

La escala vive en `assets/css/motion.css` (se carga después de `styles.css`).
No inventes curvas por componente: usa estos tokens.

| Token | Valor | Uso |
|---|---|---|
| `--motion-instant` | `90ms` | feedback inmediato |
| `--motion-fast` | `140ms` | hover, foco, salida |
| `--motion-base` | `220ms` | entrada, apertura |
| `--motion-slow` | `360ms` | cambios amplios |
| `--motion-slower` | `520ms` | trazos/progreso |
| `--motion-ease-standard` | `cubic-bezier(.2,.8,.2,1)` | movimiento general |
| `--motion-ease-enter` | `cubic-bezier(0,0,.2,1)` | algo entra en escena |
| `--motion-ease-exit` | `cubic-bezier(.4,0,1,1)` | algo sale |
| `--motion-ease-emphasized` | `cubic-bezier(.2,0,0,1)` | movimientos amplios |
| `--motion-ease-spring` | `cubic-bezier(.34,1.56,.64,1)` | pop / rebote suave |

Hay alias de compatibilidad: `--motion-ease` y `--motion-ease-out`.

Puedes redefinir la escala en el `motion.css` de tu tema (o en un CSS cargado
después). No hace falta tocar Go:

```css
:root {
  --motion-base: 180ms;
  --motion-ease-standard: cubic-bezier(.4, 0, .2, 1);
  --motion-ease-spring: cubic-bezier(.2, 1.4, .4, 1);
}
```

Reglas:

- Respeta `prefers-reduced-motion` (regla global en `motion.css`).
- Anima `transform` y `opacity`; evita `width`/`height` salvo en colapsos.
- Hover/foco: `--motion-fast` + `--motion-ease-standard`.
- Entrada: `--motion-base` + `--motion-ease-enter`.
- Salida: `--motion-fast` + `--motion-ease-exit`.
- Pop/feedback: `--motion-fast` + `--motion-ease-spring`.
- Para JS, importa `assets/js/motion.js`:

```js
import {
  animate,
  prefersReducedMotion,
  motionValue,
  motionMs
} from '/assets/js/motion.js';
```

`motionValue(name, fallback)` lee un token CSS y `motionMs(name, fallback)`
lo convierte a milisegundos para la Web Animations API.

### Colapsables

- Utilidad CSS por grid: `.collapse` + `.collapse.is-open`. Requiere un wrapper:

```html
<div class="collapse is-open"><div>contenido</div></div>
```

- `<details>` nativos: hay una mejora progresiva en `motion.css` para
  `.prose details:not(.accordion-item)` y `details.smooth`, usando
  `interpolate-size` + `::details-content` donde el navegador lo soporte.
- La isla `accordion` y la isla `spoiler` animan la altura con la Web
  Animations API y los mismos tokens (`motionMs`/`motionValue`).
- Sin JS, `<details>` sigue funcionando de forma nativa.

- Las animaciones de una isla viven en su propio `assets/islands/<nombre>.css`.
- No ocultes contenido de forma permanente sin fallback sin JS.
- Cambio de tema/acento: `window.godianThemeTransition(fn)` añade
  `html.theme-transition` durante `--motion-base` y lo retira al terminar. Úsalo
  al cambiar clases de tema o la variable `--accent`.

## Checklists

Las task lists GFM se renderizan con `disabled` (solo visuales). El tema activa
solo las que llevan la clase `checklist-interactive`:

```md
- [ ] Tarea pendiente
- [x] Tarea completada

{.checklist-interactive}
```

`theme.js` quita `disabled`, asigna un `aria-label` con el texto del ítem y guarda
el estado en `localStorage` (clave por ruta y lista). Sin la clase, los checkboxes
siguen siendo visuales. El Markdown fuente no cambia.

## i18n

Las plantillas usan `{{ t "clave" }}`. Las claves salen de
`internal/i18n/i18n.go` y del `i18n/<lang>.json` del sitio. Hoy **no existe
i18n a nivel de tema**: si creas un tema reutilizable, reutiliza las claves
existentes o añade las nuevas al i18n del sitio.

## Checklist para un tema nuevo

1. `godian new-theme mi-tema` y apunta `build.theme_dir` a esa carpeta.
2. Mantén los nombres de plantillas obligatorios.
3. Declara en `theme.yaml` los CSS/JS globales.
4. Asigna el chunk correcto con `data-css` en cada plantilla.
5. Copia o reimplementa las islas que uses y su `assets/js/islands.js`.
6. Respeta `prefers-reduced-motion` y los estados de foco.
7. Ejecuta `godian build` y `godian check`; el build valida assets y enlaces.
