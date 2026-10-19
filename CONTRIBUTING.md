# Contribuir a godian

Gracias por el interés. godian es un generador estático en Go con un tema
embebido y un admin opcional. Antes de cambios grandes, abre una issue para
acordar el alcance.

## Requisitos

- Go 1.26 o superior (se recomienda 1.27.2).
- `make`.
- No se necesita Node/npm para construir el sitio ni el admin. La copia Svelte
  antigua (`godian-cms/`) no forma parte del proyecto soportado.

## Puesta en marcha

```sh
make build
make test
./bin/godian check -config site.yaml
./bin/godian build -config site.yaml
make serve
```

- `make test` ejecuta la suite.
- `make test-race` ejecuta la suite con el detector de carreras.
- `make serve` levanta el dev server y el admin nativo en `/admin/` (modo
  local).
- El sitio de ejemplo vive en `site.yaml`, `content/`, `themes/` y `assets/`.

## Antes de abrir una PR

Ejecuta lo que aplique a tu cambio:

```sh
make fmt
make fmt-check
make vet
make test
make test-race
make staticcheck
make gosec
make vuln
./bin/godian check -config site.yaml
./bin/godian build -config site.yaml
```

`make ci` agrupa el chequeo de formato, vet, tests con race, staticcheck,
govulncheck, gosec, build y la generación del sitio. La CI ejecuta ese target.
`make test-race` activa `CGO_ENABLED=1` porque el detector de carreras lo
necesita; el build normal del sitio puede seguir siendo `CGO_ENABLED=0` con el
tag `nodynamic`.

Si un análisis de seguridad falla por un falso positivo, explica la exclusión
en la PR; no la silencies sin justificación.

## Estilo

- Go idiomático, nombres de código en inglés y comentarios en español.
- Funciones pequeñas, nombres descriptivos y errores con contexto.
- No añadas dependencias sin discutirlo antes. Si es imprescindible, justifica
  la licencia, el peso y la alternativa de la stdlib.
- Respeta el formato de `gofmt`; no reformatees archivos no relacionados.
- Los tests deben ser deterministas y no depender de red ni de rutas absolutas.

## Commits y PRs

- Conventional Commits, en español o inglés, con scope cuando aporte:
  - `feat(theme): ...`
  - `fix(admin): ...`
  - `docs: ...`
  - `test: ...`
  - `chore: ...`
- Una PR = un cambio coherente. Incluye tests y documentación si cambia el
  comportamiento público.
- No hagas `git push --force` sobre ramas compartidas.
- No commitees `bin/`, `dist/`, `godian-cms/`, `.commandcode/`, `.env`, logs,
  tokens ni PAT.

## Temas

Antes de tocar el tema por defecto, lee `themes/default/README.md`. El contrato
incluye plantillas obligatorias, `PageData`, `theme.yaml`, assets, chunks,
tokens de movimiento e islas. No rompas compatibilidad sin discutirlo; el
versionado del contrato de temas es una fase aparte.

Para un tema nuevo:

```sh
./bin/godian new-theme mi-tema
./bin/godian new-theme --minimal mi-tema-minimal
```

Apunta `build.theme_dir` a la carpeta y valida con `godian build` y
`godian check`.

## Seguridad

- No incluyas PAT, cookies, tokens ni datos privados en issues, PRs, tests o
  capturas.
- Si encuentras una vulnerabilidad, sigue `SECURITY.md`; no abras un issue
  público con detalles explotables.
- El admin remoto con PAT en el navegador está documentado en
  `deploy/admin-remoto.md`; para equipo se recomienda el admin de servidor.

## Conducta

Al participar aceptas `CODE_OF_CONDUCT.md`.
