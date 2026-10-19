# Admin remoto: guía de puesta en marcha

Esta guía describe cómo publicar el admin estático de godian y editarlo desde
el navegador. El admin **no necesita npm**: se genera en `dist/admin` y usa JS
vanilla + librerías vendorizadas.

Hay dos modos:

| Modo | Dónde vive | Uso |
|---|---|---|
| `local` (por defecto) | Admin nativo en `godian serve` | Desarrollo local, sin token |
| `remote` | `dist/admin` publicado con el sitio | Edición desde cualquier navegador |
| `off` | No se sirve | Desactivar |

> Gitea **no necesita Pages**. Gitea solo es el proveedor Git/API; el admin se
> publica con el sitio (Cloudflare Pages, Netlify, Vercel, GitHub Pages…).

---

## 1. Requisitos

- Un PAT (token) con permiso de escritura sobre el repositorio de contenido.
  - **Gitea**: `Settings → Applications → Generate token` con permiso de
    repositorio (lectura/escritura). Ejemplo:
    `https://ggit.734038.xyz/user/settings/applications`.
  - **GitHub**: token fine-grained, acceso solo a ese repositorio, permiso
    `Contents: Read and write`.
- Un hosting que sirva el sitio y, si quieres evitar CORS, que ejecute
  Functions (Cloudflare Pages, Netlify Functions, Vercel Functions).
- Acceso al panel del hosting para configurar variables de entorno y, si
  procede, Cloudflare Access.

**No pongas el token en `site.yaml`, en el repositorio ni en una variable
pública.** El token se introduce en el navegador y se guarda en
`sessionStorage` de esa pestaña.

---

## 2. Configurar `site.yaml`

Descomenta y ajusta el bloque `admin`:

```yaml
admin:
  mode: remote
  provider: gitea
  base_url: https://ggit.734038.xyz
  owner: pi
  repo: godian
  branch: main
  api_mode: proxy
```

- `provider`: `gitea` o `github`.
- `base_url`: URL de Gitea (sin `/api/v1`). Para GitHub se puede omitir.
- `owner`/`repo`/`branch`: repositorio y rama donde viven `content/` y
  `assets/`.
- `api_mode`:
  - `proxy` (**recomendado**): el navegador llama a `/admin/api` en el mismo
    origen; no hay CORS. Requiere el proxy local o un Pages Function.
  - `direct`: el navegador llama al API del proveedor. Solo si el proveedor
    envía CORS. **El Gitea actual no lo envía** (comprobado: el preflight no
    devuelve `Access-Control-Allow-Origin`), así que no uses `direct` con
    Gitea.
- `mode: local` o vacío mantiene el admin nativo de `godian serve`.
- `mode: off` lo desactiva.

El token **no** va aquí. `config.json` se genera en el build y solo contiene
provider, owner, repo, branch y rutas públicas.

---

## 3. Probar en local

Con `mode: remote` y `api_mode: proxy`:

```sh
make build
./bin/godian serve -config site.yaml -addr 127.0.0.1:4323
```

Abre `http://127.0.0.1:4323/admin/`.

- `godian serve` construye el sitio (incluido `dist/admin`).
- En modo remoto no monta el admin nativo; sirve el admin estático.
- Monta el proxy Go en `/admin/api`, que reenvía el `Authorization` al
  proveedor. No guarda el token.
- Introduce el PAT en la pantalla de conexión.

Prueba: listar, buscar/ordenar, abrir un post, editar, previsualizar, guardar,
subir una imagen, duplicar, renombrar y borrar. Cada guardado crea un commit en
la rama configurada.

Si quieres volver al admin nativo local, cambia a `mode: local` y reinicia.

---

## 4. Cloudflare Pages (recomendado para el proxy)

El repo incluye el proxy same-origin:

```text
functions/admin/api/[[path]].js
```

### 4.1. Variables de entorno

En el proyecto de Cloudflare Pages:

| Variable | Ejemplo | Obligatoria |
|---|---|---|
| `GIT_PROVIDER` | `gitea` | Sí |
| `GIT_BASE_URL` | `https://ggit.734038.xyz` | Sí para Gitea |
| `GIT_OWNER` | `pi` | Sí |
| `GIT_REPO` | `godian` | Sí |

El proxy no necesita un token propio: reenvía el `Authorization` del navegador
y solo permite `contents/` del repo configurado.

### 4.2. Desplegar

- Con integración Git de Cloudflare Pages, `functions/` se detecta
  automáticamente. El build debe generar `dist`:
  ```sh
  GOFLAGS=-tags=nodynamic CGO_ENABLED=0 go build -o bin/godian ./cmd/godian
  ./bin/godian build -config site.yaml
  ```
- Con subida directa/Wrangler, asegúrate de incluir las Functions en el
  despliegue (según tu versión de Wrangler; si no, usa integración Git o un
  Worker equivalente).
- El archivo `dist/_headers` se genera con `X-Frame-Options: DENY` para
  `/admin/*` (Cloudflare/Netlify lo leen; GitHub Pages lo ignora).

### 4.3. URLs

- Admin: `https://tu-dominio/admin/`
- API same-origin: `https://tu-dominio/admin/api/...`
- El build escribe `robots.txt` con `Disallow: <base>/admin/`.
- La página lleva `noindex,nofollow` y no se enlaza desde el sitio.

### 4.4. Proteger el admin

El admin es público; el token es la autenticación. Para reducirlo:

1. **Cloudflare Access / Zero Trust**:
   - Aplicación self-hosted sobre `tu-dominio/admin*`.
   - Política por email, Google, GitHub o IdP.
   - Sesión corta.
   - Opcional: solo tu IP.
2. PAT fine-grained, caducidad corta, solo `Contents: write`.
3. Revoca el PAT si sospechas XSS, pérdida del portátil o acceso indebido.
4. No compartas el PAT.

Con Access, la página queda además protegida antes de ejecutar el admin.

---

## 5. Otros hostings

### GitHub Pages (estático puro)

GitHub Pages **no ejecuta `functions/`**. Opciones:

1. `api_mode: direct` + CORS en Gitea. Configura el reverse proxy con
   `Access-Control-Allow-Origin`, `Access-Control-Allow-Headers: Authorization,
   Content-Type, Accept` y métodos `GET, PUT, DELETE, OPTIONS`. Ejemplo en
   [`gitea-cors.md`](gitea-cors.md).
2. Usar el **admin nativo de servidor** (A) y acceder por SSH/Tailscale/VPN.
3. Poner el proxy en otro host (Cloudflare Worker, VPS, etc.) y apuntar
   `api_base` a él.

El Gitea actual no envía CORS, así que la opción 1 requiere configurarlo.

### Netlify / Vercel

Puedes portar `functions/admin/api/[[path]].js` a una Function equivalente.
Mientras no exista, usa `api_mode: direct` + CORS o el admin de servidor.

### Admin de servidor (A)

Si prefieres no exponer el token al navegador:

- Mantén `mode: local`.
- Ejecuta `godian serve` en una máquina/VPS con acceso restringido.
- Accede por SSH/Tailscale/VPN o pon un reverse proxy con autenticación.
- El token Git y las credenciales quedan en el servidor.

---

## 6. Verificación posterior al despliegue

1. Abre `/admin/`; debe aparecer la pantalla de token.
2. Conecta con el PAT; comprueba `owner/repo` y rama.
3. Dashboard: lista, búsqueda y orden.
4. Edita un post de prueba; guarda; comprueba:
   - el commit en Gitea/GitHub,
   - el workflow/CI,
   - el deploy del sitio.
5. Sube una imagen; el archivo aparece en `assets/uploads` y, tras el build,
   en `/uploads`.
6. Duplica, renombra y borra un archivo de prueba.
7. Fuerza un conflicto: edita el mismo archivo desde Gitea y guarda desde el
   admin; debe avisar por `sha` y no sobrescribir.
8. Si usas `site.base_path`, comprueba `/base/admin/` y los assets.
9. Comprueba `robots.txt` y `_headers`.
10. Cierra sesión y verifica que el token desaparece de la pestaña.

---

## 7. Troubleshooting

| Síntoma | Causa probable | Solución |
|---|---|---|
| `/admin/` 404 | `admin.mode` no es `remote` o no se desplegó `dist/admin` | Activa el modo y rebuild/deploy |
| `/admin` redirige mal | Falta la barra final | Usa `/admin/`; el servidor redirige |
| `404 /admin/api/...` | Function no desplegada o variables de entorno ausentes | Despliega `functions/` y define `GIT_*` |
| Error CORS | `api_mode: direct` y Gitea sin CORS | Usa `api_mode: proxy` |
| `401/403` | PAT caducado, sin scope o repo privado | Crea token con escritura del repo |
| `404` al conectar | `owner`, `repo` o `base_url` incorrectos; repo privado sin token | Revisa la config y el PAT |
| `409` al guardar | El archivo cambió en remoto | Recarga el editor y vuelve a aplicar |
| Imagen 404 tras subir | El sitio aún no se ha reconstruido/desplegado | Espera al CI o ejecuta el build |
| Cambios no aparecen en el sitio | El admin commitea, pero el deploy no se ha ejecutado | Revisa el workflow del hosting |
| CSP bloquea algo | Script inline o CDN | Mantén todo same-origin/vendorizado; no añadas inline |
| El admin no hereda el tema | Falta `assets/css/admin.css` o no se copió | Revisa el tema y rebuild |
| `dist/admin` en local | `admin.mode: remote` | Es esperado; usa `local` para el nativo |

---

## 8. Seguridad y límites

- Un editor de confianza; no hay roles, multiusuario ni auditoría propia.
- El token está en el navegador: un XSS o una dependencia comprometida podría
  robarlo. CSP, vendorizado, `sessionStorage`, PAT mínimo y Access reducen el
  riesgo, no lo eliminan.
- El CI/build es la validación final; el navegador valida el schema, pero no
  ejecuta `godian check`.
- El preview de islas/shortcodes/wikilinks es aproximado; Markdown y KaTeX sí.
- Sin OAuth sin servidor. Para OAuth/roles, usa el admin de servidor (A).
- El proxy no guarda tokens, pero es público: cualquiera con un PAT válido
  puede usarlo dentro del repo permitido. Protégelo con Access/rate limit si
  lo necesitas.

## 9. Desactivar

1. `admin.mode: off` (o elimina la sección) y rebuild/deploy.
2. Revoca el PAT.
3. Elimina la aplicación de Cloudflare Access si ya no la usas.
4. Opcional: borra `dist/admin` del despliegue (el build limpio ya no lo genera).
