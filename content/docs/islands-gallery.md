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

`{{island carousel images="/demo-1.jpg,/demo-2.jpg"}}`

{{island carousel images="/demo-1.jpg,/demo-2.jpg"}}

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

`{{#partial callout title="Aviso"}}` … `{{/partial}}`

{{#partial callout title="Aviso"}}
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

## gallery

`{{island gallery images="/demo-1.jpg,/demo-2.jpg,/demo-3.jpg"}}`

{{island gallery images="/demo-1.jpg,/demo-2.jpg,/demo-3.jpg"}}

## modal

`{{island modal label="Abrir modal" title="Título" text="Contenido del modal" action="Aceptar" size="lg"}}`

{{island modal label="Abrir modal" title="Título" text="Contenido del modal" action="Aceptar" size="lg"}}

## popover

`{{island popover label="Info" text="Detalle del popover"}}`

{{island popover label="Info" text="Detalle del popover"}}

## progress-ring

`{{island progress-ring value=72 max=100 size=80 label="72%"}}`

{{island progress-ring value=72 max=100 size=80 label="72%"}}

## toast

`{{island toast message="Guardado correctamente" type="success" label="Mostrar toast"}}`

{{island toast message="Guardado correctamente" type="success" label="Mostrar toast"}}

## qr-code

`{{island qr-code text="https://godian" size=160}}`

{{island qr-code text="https://godian" size=160}}

## kanban

`{{island kanban columns="Todo,Doing,Done" cards="Tarea A,Tarea B,Tarea C"}}`

{{island kanban columns="Todo,Doing,Done" cards="Tarea A,Tarea B,Tarea C"}}

## calendar

`{{island calendar month=10 year=2026 events="5:Reunión,12:Entrega,20:Review"}}`

{{island calendar month=10 year=2026 events="5:Reunión,12:Entrega,20:Review"}}

Con `source="site"` marca los días con posts/docs (clic para ver el contenido):

`{{island calendar source="site" month=2 year=2026}}`

{{island calendar source="site" month=2 year=2026}}

## chart

`{{island chart type="bar" data="10,20,30,40" labels="A,B,C,D" title="Ventas"}}`

{{island chart type="bar" data="10,20,30,40" labels="A,B,C,D" title="Ventas"}}

`{{island chart type="pie" data="14,4" labels="Islas,Partials" title="Componentes"}}`

{{island chart type="pie" data="14,4" labels="Islas,Partials" title="Componentes"}}

## audio

`{{island audio src="/demo-audio.mp3" title="Audio de ejemplo"}}`

{{island audio src="/demo-audio.mp3" title="Audio de ejemplo"}}

## lightbox

`{{island lightbox src="/demo-1.jpg" alt="Demo" caption="Figura 1"}}`

{{island lightbox src="/demo-1.jpg" alt="Demo" caption="Figura 1"}}

## count-up

`{{island count-up value=1234 duration=1200 suffix="+"}}`

{{island count-up value=1234 duration=1200 suffix="+"}}

## steps

`{{island steps items="Paso 1,Paso 2,Paso 3"}}`

{{island steps items="Paso 1,Paso 2,Paso 3"}}

## timeline

`{{island timeline items="2026-01:Lanzamiento,2026-02:Tema nuevo,2026-03:Islas"}}`

{{island timeline items="2026-01:Lanzamiento,2026-02:Tema nuevo,2026-03:Islas"}}

## Partials estáticos (ampliación)

`{{partial callout type="warn" title="Aviso" text="Texto de aviso"}}` → {{partial callout type="warn" title="Aviso" text="Texto de aviso"}}

`{{partial stat value=42 label="posts"}}` → {{partial stat value=42 label="posts"}}

`{{partial avatar src="/demo-1.jpg" alt="Nef" size=48}}` → {{partial avatar src="/demo-1.jpg" alt="Nef" size=48}}

`{{partial chip text="Go" color="#00add8"}}` → {{partial chip text="Go" color="#00add8"}}

`{{partial figure src="/demo-1.jpg" alt="Demo" caption="Figura 1" width=1200 height=630}}` → {{partial figure src="/demo-1.jpg" alt="Demo" caption="Figura 1" width=1200 height=630}}

`{{partial quote text="Godian es minimalista." author="Nef"}}` → {{partial quote text="Godian es minimalista." author="Nef"}}

`{{partial divider label="Sección"}}` → {{partial divider label="Sección"}}

`{{partial callout type="warn" title="Aviso" text="Texto de aviso"}}` → {{partial callout type="warn" title="Aviso" text="Texto de aviso"}}
