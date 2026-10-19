// Package themes embebe el tema por defecto y el scaffold mínimo de godian.
//
// Los temas viven como archivos reales en themes/default/ y themes/minimal/ y se
// embeben en el binario para funcionar sin carpeta externa. Un theme_dir en disco
// sobrescribe el tema por defecto archivo a archivo.
package themes

import (
	"embed"
	"io/fs"
)

//go:embed default minimal
var embedded embed.FS

// Default devuelve la raíz del tema por defecto embebido, que contiene
// theme.yaml, templates/ y assets/.
func Default() fs.FS {
	return sub("default")
}

// Minimal devuelve la raíz del scaffold mínimo embebido.
func Minimal() fs.FS {
	return sub("minimal")
}

func sub(name string) fs.FS {
	sub, err := fs.Sub(embedded, name)
	if err != nil {
		// El embed es estático; fs.Sub solo falla si falta el directorio.
		panic("themes: tema embebido inválido: " + err.Error())
	}
	return sub
}
