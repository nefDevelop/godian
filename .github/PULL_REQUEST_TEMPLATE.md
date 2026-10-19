## Qué cambia

Describe el cambio y el problema que resuelve.

## Tipo

- [ ] Bug fix
- [ ] Feature
- [ ] Tema / contrato de temas
- [ ] Documentación
- [ ] Refactor / mantenimiento
- [ ] Seguridad

## Cómo probarlo

Comandos, configuración mínima y resultado esperado.

## Checklist

- [ ] `make fmt-check`
- [ ] `make vet`
- [ ] `make test` / `make test-race`
- [ ] `make staticcheck`
- [ ] `godian check -config site.yaml`
- [ ] `godian build -config site.yaml`
- [ ] Tests añadidos/actualizados si cambia el comportamiento
- [ ] Documentación actualizada si cambia la API, el tema o la configuración
- [ ] Sin PAT, tokens, `.env`, contenido privado ni logs sensibles
- [ ] No incluye `bin/`, `dist/`, `godian-cms/` ni `.commandcode/`

## Notas

Cambios incompatibles, migraciones o seguimiento pendiente.
