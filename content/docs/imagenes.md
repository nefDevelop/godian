---
title: Imágenes
description: Optimización (WebP/AVIF), LQIP y presets.
order: 5
section: guias
tags: [guide]
---

## Optimización

- **Variantes WebP** + `srcset` automáticas (480/768/1024/1440 + completo).
- **AVIF** opt-in compilando con `-tags avif`.
- **LQIP**: se genera un *placeholder* diminuto y difuminado (data URI) que se pone
  como fondo del `<img>` mientras carga la imagen real (mejora la percepción de carga).
- Las imágenes de Markdown se resuelven solas: `![alt](/demo-1.jpg)`.

## Presets

Define presets en `site.yaml` (dentro de `build`):

```yaml
build:
  image_presets:
    thumb: { width: 480, height: 270, fit: fill, focal: center, quality: 78 }
    wide:  { width: 1200, fit: fit }
```

- `fit`: escala manteniendo proporción (contain).
- `fill`: recorta para llenar exactamente la caja (cover), con `focal`
  (`center|top|bottom|left|right`).
- `quality` (0-100) y `blur` (radio).

Y úsalos en Markdown con el shortcode `{{image}}`:

```
{{image src="/demo-1.jpg" preset="thumb" alt="Miniatura 16:9"}}
```

Si el preset no existe, cae al comportamiento normal (`srcset` + LQIP).

> Recuerda: con `assets_dir: assets`, un archivo `assets/demo-1.jpg` se sirve en `/demo-1.jpg`.
