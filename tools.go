//go:build tools

// Package tools fija dependencias opcionales que solo se usan con build tags
// (p. ej. AVIF con `-tags avif`), para que `go mod tidy` no las elimine.
package tools

import _ "github.com/gen2brain/avif"
