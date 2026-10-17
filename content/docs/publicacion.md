---
title: Publicación, feeds y datos
description: Feeds, robots, aliases, data/ y enlaces externos.
order: 4
section: guias
tags: [guide]
---

## Feeds

En cada build se generan tres feeds, enlazados automáticamente en el `<head>`:

- `/rss.xml` — RSS 2.0
- `/atom.xml` — Atom 1.0
- `/feed.json` — JSON Feed 1.1

## Publicación programada

Un post con `published` en el **futuro** no se publica hasta esa fecha: queda excluido
del sitio, de los feeds y del sitemap. El build lo cuenta como *programado*:

```
posts:     3 (2 publicados, 1 programados, 0 borradores)
```

Con `godian build -drafts` se incluyen también los posts futuros (y los borradores).

## robots.txt

Se genera `/robots.txt` permitiendo todo y apuntando al sitemap:

```
User-agent: *
Allow: /

Sitemap: https://tu-sitio.com/sitemap.xml
```

## Aliases / redirects

En el frontmatter de un post o doc, declara URLs antiguas:

```yaml
aliases: ["/blog/url-vieja/", "/blog/otra-url/"]
```

Se generan páginas de **redirección** (con `<meta http-equiv="refresh">`, `noindex` y
`canonical`) hacia la URL actual. Ideal para no romper enlaces al renombrar.

## `data/`

Carpeta de datos estructurados (config `build.data_dir`, por defecto `data/`).
Acepta **JSON, YAML y CSV**. La clave es la ruta del archivo sin extensión.

`data/links.yaml`:

```yaml
- { label: Documentación, url: /docs/ }
- { label: GitHub, url: https://github.com/tuusuario }
```

En una plantilla del tema:

```
{{ with index .Data "links" }}
  <ul>{{ range . }}<li><a href="{{ .url }}">{{ .label }}</a></li>{{ end }}</ul>
{{ end }}
```

CSV con cabecera se convierte en lista de mapas (`data/equipo.csv` → `index .Data "equipo"`).

## Enlaces externos

Los enlaces Markdown a URLs absolutas (`[texto](https://…)`) se renderizan con
`target="_blank" rel="noopener noreferrer"`. Los internos, normales.

## AVIF (opt-in)

Por defecto se generan variantes **WebP**. Compilando con el tag `avif` se generan
además variantes **AVIF** y se añade `<source type="image/avif">` al `<picture>`:

```sh
GOFLAGS="-tags=nodynamic,avif" go build -o bin/godian ./cmd/godian
./bin/godian build -config site.yaml
```

> Nota: el encode AVIF es **pesado** (libavif/aom vía wazero). Recomendado en escritorio/CI.
