# Tema mínimo

Scaffold de ejemplo generado por `godian new-theme --minimal`.

- Define `home`, `blog`, `post` y `404`; el resto de plantillas cae al tema embebido.
- Usa los assets del tema por defecto (`styles.css`, `motion.css`, chunks, JS).
- `assets/css/custom.css` se carga al final para tus ajustes.
- `theme.yaml` declara dependencias de chunks (`article-ux` y `docs` → `lists`).

Para usarlo, apunta `build.theme_dir` a esta carpeta en `site.yaml`.
