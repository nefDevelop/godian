---
title: Componentes (islas)
description: Crea componentes interactivos y estáticos sin framework.
order: 2
section: guias
tags: [guide]
---

godian permite crear **componentes** como archivos del tema, sin framework ni bundler.
Hay dos tipos: **islas** (interactivas, con JS) y **partials** (estáticas).

> Galería con ejemplos en vivo de todas las islas: **[Galería de islas](/docs/islands-gallery/)**.

## Contrato de una isla

Crea estos archivos en el tema (todos opcionales salvo el JS si es interactiva):

```
assets/islands/<nombre>.js      # export function mount(el, props) { … }
assets/islands/<nombre>.css     # estilos (opcional)
templates/islands/<nombre>.html # HTML SSR (opcional; props como .clave)
```

- **Con SSR** → se renderiza HTML en build y el JS lo activa.
- **Sin SSR** → se emite `<div data-island="<nombre>" data-props="…">` y el JS pinta.
- **Sin JS** → es un componente estático (solo template).

## Usarlo en el contenido

Isla interactiva (el nombre va primero y luego props `clave=valor`):

`{{island counter start=3}}`

Componente estático desde un partial (`templates/partials/<nombre>.html`):

`{{partial cta text="Ir a la documentación" url="/docs/"}}`

Ejemplo en vivo:

{{island counter start=5}}

## Con contenido (shortcodes de bloque)

Envuelven **Markdown** y se lo pasan al componente como `content`:

```
{{#partial callout title="Aviso"}}
Contenido **Markdown** dentro del componente.
{{/partial}}

{{#island tabs}}
## Primera
Contenido de la primera pestaña.

## Segunda
Contenido de la segunda.
{{/island}}
```

- El partial recibe `{{ .content }}` (HTML ya renderizado) además de sus props.
- Las islas `tabs` y `accordion`, si tienen contenido, usan los encabezados `##` como
  secciones (no hace falta `items`).
- Los bloques dentro de fences de código se ignoran (se muestran literales).

## Desde plantillas del tema

```
{{ island "tabs" (dict "items" "Uno,Dos") }}
{{ partial "cta" (dict "text" "Docs" "url" "/docs/") }}
```

## Islas incluidas en el tema por defecto

| Isla | Props | Ejemplo |
|---|---|---|
| `counter` | `start` | `{{island counter start=3}}` |
| `tabs` | `items` (coma) | `{{island tabs items="Uno,Dos"}}` |
| `accordion` | `items` (`Título=Cuerpo;…`) | `{{island accordion items="A=1;B=2"}}` |
| `copy` | `text`, `label` | `{{island copy text="hola" label="Copiar"}}` |
| `spoiler` | `text`, `label` | `{{island spoiler text="secreto"}}` |
| `progress` | `value`, `max`, `label` | `{{island progress value=72 max=100}}` |
| `theme-picker` | `colors` (coma) | `{{island theme-picker}}` |
| `carousel` | `images` (coma) | `{{island carousel images="/a.jpg,/b.jpg"}}` |
| `gallery` | `images` (coma) | `{{island gallery images="/a.jpg,/b.jpg"}}` |
| `rating` | `max`, `value` | `{{island rating value=3}}` |
| `video` | `id`, `title` | `{{island video id="dQw4w9WgXcQ" title="Demo"}}` |
| `share` | `url`, `title` (opcionales) | `{{island share}}` |
| `clock` | — | `{{island clock}}` |
| `countdown` | `to`, `label` | `{{island countdown to="2030-01-01T00:00:00"}}` |
| `modal` | `label`, `title`, `text`, `action`, `size` | `{{island modal label="Abrir" title="Título" text="Contenido" action="Aceptar" size="lg"}}` |
| `popover` | `label`, `text`, `mode` | `{{island popover label="Info" text="Detalle"}}` |
| `progress-ring` | `value`, `max`, `size`, `label` | `{{island progress-ring value=72 max=100 size=80}}` |
| `toast` | `label`, `message`, `type`, `duration` | `{{island toast message="Guardado" type="success" label="Mostrar"}}` |

| `chart` | `type`, `data`, `labels`, `title` | `{{island chart type="bar" data="10,20,30" labels="A,B,C"}}` |
| `audio` | `src`, `title` | `{{island audio src="/audio.mp3" title="Podcast"}}` |
| `lightbox` | `src`, `alt`, `caption` | `{{island lightbox src="/img.png" alt="Demo" caption="Fig 1"}}` |
| `count-up` | `value`, `duration`, `prefix`, `suffix` | `{{island count-up value=1000 duration=1200 suffix="+"}}` |
| `steps` | `items` | `{{island steps items="Paso 1,Paso 2,Paso 3"}}` |
| `timeline` | `items` (`fecha\|texto`) | `{{island timeline items="2026-01:Lanzamiento,2026-02:Tema"}}` |
| `qr-code` | `text`, `size` | `{{island qr-code text="https://ejemplo.com" size=160}}` |
| `kanban` | `columns`, `cards` | `{{island kanban columns="Todo,Doing,Done" cards="A,B,C"}}` |
| `calendar` | `month`, `year`, `events` | `{{island calendar month=10 year=2026 events="5:Reunión,12:Entrega"}}` |

## Partials estáticos incluidos

| Partial | Props | Ejemplo |
|---|---|---|
| `badge` | `text`, `color` | `{{partial badge text="Nuevo" color="#16a34a"}}` |
| `breadcrumb` | — | `{{partial breadcrumb}}` |
| `card` | — | (uso interno del tema) |
| `cta` | `text`, `url` | `{{partial cta text="Docs" url="/docs/"}}` |
| `docsnav` | — | (uso interno del tema) |
| `footer` | — | (uso interno del tema) |
| `head` | — | (uso interno del tema) |
| `header` | — | (uso interno del tema) |
| `kbd` | `keys` | `{{partial kbd keys="Ctrl+K"}}` |
| `pagination` | — | (uso interno del tema) |
| `search-modal` | — | (uso interno del tema) |
| `toc` | — | (uso interno del tema) |
| `stat` | `value`, `label`, `icon` | `{{partial stat value=42 label="posts"}}` |
| `avatar` | `src`, `alt`, `size` | `{{partial avatar src="/img.png" alt="Nef" size=48}}` |
| `chip` | `text`, `color`, `href` | `{{partial chip text="Go" color="#00add8"}}` |
| `figure` | `src`, `alt`, `caption` | `{{partial figure src="/img.png" alt="x" caption="Fig 1"}}` |
| `quote` | `text`, `author`, `cite` | `{{partial quote text="..." author="Nef" cite="https://..."}}` |
| `divider` | `label` | `{{partial divider label="Sección"}}` |
| `callout` | `type`, `title`, `text` | `{{partial callout type="warn" title="Aviso" text="..."}}` |
| `embed` | `src`, `title`, `ratio` | `{{partial embed src="https://..." title="Vídeo" ratio="16x9"}}` |
| `tooltip` | `text`, `label` | `{{partial tooltip text="Ayuda" label="?"}}` |

## Carga diferida

El runtime (`assets/js/islands.js`) lee `assets/islands/manifest.json` e importa
cada isla **solo si está en la página** y **una vez por nombre**. Puedes elegir la
estrategia con el prop `load`:

- `idle` (por defecto): cuando el navegador está libre.
- `visible`: al entrar en pantalla.
- `eager`: inmediatamente.

`{{island tabs items="A,B" load=visible}}`

## Ejemplo de módulo

```js
// assets/islands/counter.js
export function mount(el, props) {
  let n = Number(props.start || 0);
  const out = el.querySelector('[data-counter-out]');
  el.querySelector('[data-counter-inc]').onclick = () => (out.textContent = ++n);
  out.textContent = n;
}
```

Si una isla referenciada no existe, **el build falla** con un error claro.
