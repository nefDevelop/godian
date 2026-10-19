---
title: Ejemplo completo de frontmatter
published: 2026-03-01
updated: 2026-03-10
description: Portada, autor, dificultad, updated, enlaces y descargas en un solo post.
image: /demo-3.jpg
tags: [guide]
categories: [design]
author: [nef]
difficulty: intermediate
draft: false
links:
  website: https://go.dev
  github: https://github.com/nefdevelop/godian
downloads:
  - id: imagen
    label: Imagen de ejemplo
    url: /demo-3.jpg
---

Este post existe para mostrar en el tema principal todos los campos visuales que
godian renderiza: **portada**, **autor**, **dificultad**, **fecha de actualización**,
**enlaces** y **descargas**.

## Portada

La imagen de `image` se muestra como cabecera del post y como miniatura en las
tarjetas de los listados.

## Metadatos

- `author` enlaza al perfil del autor.
- `difficulty` pinta el badge de nivel.
- `updated` muestra la última actualización junto a la fecha de publicación.
- `links` añade la caja de enlaces externos.
- `downloads` añade la lista de recursos descargables.

> [!tip] Ejemplo vivo
> Todo lo anterior se genera a partir del frontmatter, sin tocar el tema.
