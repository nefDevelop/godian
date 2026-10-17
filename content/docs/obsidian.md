---
title: Obsidian
description: Wikilinks entre colecciones, transclusión y backlinks.
order: 6
section: guias
tags: [guide]
---

godian es **Obsidian-first**: entiende wikilinks, embeds y backlinks de forma nativa.

## Wikilinks entre colecciones

`[[nota]]` resuelve a cualquier nota (blog o docs) por **slug o nombre de archivo**:

- `[[components]]` → `/docs/components/` (por basename)
- `[[markdown-demo|Ver demo]]` → con alias de texto
- `[[guias/guide]]` → por ruta completa

Si no se encuentra, cae al comportamiento por defecto (`/blog/<nota>/`).

## Transclusión de notas

`![[nota]]` **incluye el contenido** de otra nota (no solo imágenes):

```
![[getting-started]]
```

Si el objetivo no es una nota (p. ej. `![[demo-2.png]]`), se trata como **imagen**
(con `srcset` y LQIP).

## Backlinks

Cada post y doc muestra una sección **Backlinks** con las notas que la enlazan,
calculada a partir de los wikilinks del resto del contenido.
