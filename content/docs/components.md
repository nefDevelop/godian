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
| `rating` | `max`, `value` | `{{island rating value=3}}` |

Y el partial estático `cta`: `{{partial cta text="Docs" url="/docs/"}}`.

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
