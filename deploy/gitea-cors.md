# Admin remoto: Gitea, proxy y CORS

Gitea **no necesita Pages**. El admin estático se publica con el sitio
(`dist/admin`) y Gitea solo expone la API de contenido.

## Modo recomendado: proxy same-origin

El navegador llama a `/admin/api/...` en el mismo origen del sitio. El proxy
reenvía el `Authorization` a Gitea/GitHub. El proxy ve el PAT en memoria, pero
no lo guarda ni lo configura.

### En local

Con `admin.mode: remote` y `admin.api_mode: proxy`, `godian serve` monta el
proxy Go en `/admin/api/`. No hay CORS:

```yaml
admin:
  mode: remote
  provider: gitea
  base_url: https://gitea.example.com
  owner: my-org
  repo: my-site
  branch: main
  api_mode: proxy
```

### En Cloudflare Pages

El repo incluye `functions/admin/api/[[path]].js`. Configura las variables de
entorno del proyecto Pages:

- `GIT_PROVIDER=gitea` (o `github`)
- `GIT_BASE_URL=https://gitea.example.com` (solo Gitea)
- `GIT_OWNER=my-org`
- `GIT_REPO=my-site`

Cloudflare Pages detecta `functions/` con la integración Git. Si haces subida
directa, incluye el directorio de functions al desplegar (o usa `wrangler pages
functions build`). El proxy no necesita secretos: reenvía el `Authorization` del
navegador y limita las rutas al repositorio configurado.

El admin queda en `https://tu-dominio/admin/`. No se enlaza desde el sitio y
lleva `noindex` + `Disallow: /admin/`.

### GitHub Pages / hosting sin Functions

GitHub Pages es estático puro: no ejecuta `functions/`. Tienes dos opciones:

1. `admin.api_mode: direct` y habilitar CORS en Gitea (abajo).
2. Usar el admin nativo de servidor (modo local) o un proxy en otro host.

## Modo directo: CORS en Gitea

Solo si no puedes usar el proxy. El navegador llamará a
`https://gitea.example.com/api/v1/...` desde el origen del sitio.

Habilita CORS en Gitea o, preferiblemente, en el reverse proxy. Ejemplo Nginx:

```nginx
location /api/ {
    if ($request_method = OPTIONS) {
        add_header Access-Control-Allow-Origin "https://site.example.com" always;
        add_header Access-Control-Allow-Methods "GET, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type, Accept" always;
        add_header Access-Control-Max-Age 86400 always;
        return 204;
    }

    add_header Access-Control-Allow-Origin "https://site.example.com" always;
    add_header Access-Control-Allow-Methods "GET, PUT, DELETE, OPTIONS" always;
    add_header Access-Control-Allow-Headers "Authorization, Content-Type, Accept" always;

    proxy_pass http://127.0.0.1:3000/api/;
    proxy_set_header Host $host;
}
```

Puntos importantes:

- `Access-Control-Allow-Headers` debe incluir **Authorization** explícitamente;
  `*` no cubre ese header.
- Añade también `http://localhost:4321` y `http://localhost:4323` si pruebas en
  local.
- En algunas versiones de Gitea `ACCESS_CONTROL_ALLOW_ORIGIN` no cubre bien los
  preflight con Authorization. El proxy Nginx o el Pages Function son más
  fiables.

## Token

- Gitea: token con permiso de escritura del repositorio.
- GitHub: token fine-grained con `Contents: Read and write`.
- Se guarda en `sessionStorage` de la pestaña; no va al repositorio ni a la URL.
- En `api_mode: proxy`, el navegador lo envía a `/admin/api` y el proxy lo ve en
  memoria para reenviarlo. No se persiste. Si quieres que no pase por el proxy,
  usa `direct` + CORS; si quieres que no esté en el navegador, usa el admin de
  servidor.
- El admin no tiene OAuth sin servidor. Para un equipo, usa el admin de servidor
  con autenticación.

## Flujo

1. El admin lee/escribe `content/` y `assets/uploads/` con la API.
2. Cada guardado es un commit en la rama configurada.
3. El workflow del sitio ejecuta `godian check` y `godian build`.
4. El CI publica `dist/`, incluido el admin si `mode: remote`.

El build/CI es la validación final del frontmatter; el navegador valida campos
requeridos y tipos del schema.
