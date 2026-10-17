---
title: Galería de islas
description: Ejemplos en vivo de todos los componentes del tema por defecto.
order: 3
section: guias
tags: [guide]
---

Ejemplos **en vivo**: cada isla se carga de forma diferida solo en esta página.
El código de uso va en `code span` para que puedas copiarlo.

## counter

`{{island counter start=5}}`

{{island counter start=5}}

## tabs

`{{island tabs items="Uno,Dos,Tres"}}`

{{island tabs items="Uno,Dos,Tres"}}

## accordion

`{{island accordion items="Primero=Contenido uno;Segundo=Contenido dos"}}`

{{island accordion items="Primero=Contenido uno;Segundo=Contenido dos"}}

## copy

`{{island copy text="Hola mundo" label="Copiar"}}`

{{island copy text="Hola mundo" label="Copiar"}}

## spoiler

`{{island spoiler text="Este es el secreto." label="Revelar"}}`

{{island spoiler text="Este es el secreto." label="Revelar"}}

## progress

`{{island progress value=72 max=100 label="Progreso"}}`

{{island progress value=72 max=100 label="Progreso"}}

## theme-picker

`{{island theme-picker}}`

{{island theme-picker}}

## rating

`{{island rating value=3}}`

{{island rating value=3}}

## carousel

`{{island carousel images="/demo-1.png,/demo-2.png"}}`

{{island carousel images="/demo-1.png,/demo-2.png"}}

## video

`{{island video id="dQw4w9WgXcQ" title="Demo"}}`

{{island video id="dQw4w9WgXcQ" title="Demo"}}

## share

`{{island share}}`

{{island share}}

## clock

`{{island clock}}`

{{island clock}}

## countdown

`{{island countdown to="2030-01-01T00:00:00"}}`

{{island countdown to="2030-01-01T00:00:00"}}

## Partials estáticos

`{{partial badge text="Nuevo" color="#16a34a"}}` → {{partial badge text="Nuevo" color="#16a34a"}}

`{{partial kbd keys="Ctrl+K"}}` → {{partial kbd keys="Ctrl+K"}}

`{{partial cta text="Ir a Docs" url="/docs/"}}` → {{partial cta text="Ir a Docs" url="/docs/"}}

## Islas con contenido (bloque)

Los shortcodes de bloque envuelven Markdown: `{{#island …}}` … `{{/island}}`
(o `{{#partial …}}` … `{{/partial}}`).

`{{#partial note title="Aviso"}}` … `{{/partial}}`

{{#partial note title="Aviso"}}
Esto es **Markdown** dentro de un partial (con enlaces, listas, etc.).
{{/partial}}

`{{#island tabs}}` con secciones `##`:

{{#island tabs}}
## Primera
Contenido de la primera pestaña con **Markdown**.

## Segunda
Contenido de la segunda pestaña.
{{/island}}

`{{#island accordion}}` con secciones `##`:

{{#island accordion}}
## Sección uno
Texto de la primera sección.

## Sección dos
Texto de la segunda sección.
{{/island}}
