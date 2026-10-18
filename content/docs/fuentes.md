---
title: Fuentes
description: Autohospedar fuentes (varias familias) en el build y aplicarlas.
order: 7
section: guias
tags: [guide]
---

La clave `fonts:` en `site.yaml` descarga tipografías de Google Fonts **durante el
build** y las sirve desde `/assets/fonts`. Así no hay peticiones a terceros (ni
cookies) en tiempo de uso.

## Declarar fuentes (varias familias)

```yaml
fonts:
  - family: Inter
    weights: [400, 700]
    styles: [normal]      # normal | italic
    subsets: [latin]      # latin, latin-ext, cyrillic…
    display: swap         # swap (por defecto) | block | fallback | optional
  - family: JetBrains Mono
    weights: [400]
    subsets: [latin]
```

- `family`: nombre en Google Fonts.
- `weights`: pesos a descargar (por defecto `[400]`).
- `styles`: `normal` / `italic` (por defecto `[normal]`).
- `subsets`: subconjuntos a conservar (por defecto `[latin]`).
- `display`: estrategia de `font-display` (por defecto `swap`).

Puedes declarar **tantas familias como necesites** (una entrada por familia).

## Qué genera el build

- Los ficheros `.woff2` en `/assets/fonts/`.
- `/assets/css/fonts.css` con los `@font-face` (con `font-display` y `unicode-range`),
  enlazado en el `<head>` como el resto de CSS (con fingerprint y minificado).
- Si el build **no tiene red**, avisa por stderr y continúa con las fuentes del
  sistema (no rompe el sitio).

## Aplicar la fuente

Los `@font-face` solo *cargan* la fuente; para *usarla* indica la familia del cuerpo:

```yaml
site:
  font: Inter
```

El tema aplica `--font-sans: 'Inter', system-ui, sans-serif` en `<html>`. Si
prefieres controlarlo por CSS, el tema usa dos variables:

- `--font-sans`: tipografía del cuerpo (por defecto la pila del sistema).
- `--font-mono`: código y bloques monoespaciados.

```css
:root { --font-sans: 'Inter', system-ui, sans-serif; --font-mono: 'JetBrains Mono', monospace; }
```

## Licencias

Las fuentes de Google Fonts son, en su mayoría, **OFL** (permiten autohospedaje).
Consulta la licencia de cada familia si vas a redistribuirla.
