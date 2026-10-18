// Package themes embebe el tema por defecto de godian.
//
// El tema vive como archivos reales en themes/default/ (plantilla reutilizable
// y editable) y se embebe en el binario para funcionar sin carpeta externa. Un
// theme_dir en disco lo sobrescribe archivo a archivo.
package themes

import (
	"embed"
	"io/fs"
)

//go:embed default
var embedded embed.FS

// Default devuelve la raíz del tema por defecto embebido, que contiene
// theme.yaml, templates/ y assets/.
func Default() fs.FS {
	sub, err := fs.Sub(embedded, "default")
	if err != nil {
		// El embed es estático; fs.Sub solo falla si falta "default".
		panic("themes: tema embebido inválido: " + err.Error())
	}
	return sub
}
