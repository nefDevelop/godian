BINARY := godian
PKG    := ./...
GO     ?= go

STATICCHECK := honnef.co/go/tools/cmd/staticcheck@v0.8.2
GOVULN      := golang.org/x/vuln/cmd/govulncheck@v1.9.0
GOSEC       := github.com/securego/gosec/v2/cmd/gosec@v2.29.0

# Reglas excluidas (por diseño en un generador de estáticos):
#   G301/G306 permisos de ficheros web generados (0644/0755)
#   G304/G703 leer rutas configurables por el usuario (contenido/assets)
#   G203/G705 HTML de autor confiable (html.WithUnsafe)
#   G122     WalkDir sobre rutas controladas por el usuario
GOSEC_EXCLUDE := G301,G306,G304,G703,G203,G705,G122

# "nodynamic" usa el libwebp transpilado a Go (sin cgo ni dependencias de runtime).
export GOFLAGS := -tags=nodynamic

.PHONY: all build test test-unit vet cover staticcheck vuln gosec security lint fmt tidy run clean

all: build

build:
	$(GO) build -trimpath -o bin/$(BINARY) ./cmd/godian

test:
	$(GO) test $(PKG)

test-unit:
	$(GO) test ./internal/...

vet:
	$(GO) vet $(PKG)

cover:
	$(GO) test -coverpkg=./internal/... -cover ./internal/...

staticcheck:
	$(GO) run $(STATICCHECK) $(PKG)

vuln:
	$(GO) run $(GOVULN) $(PKG)

gosec:
	$(GO) run $(GOSEC) -quiet -exclude=$(GOSEC_EXCLUDE) $(PKG)

security: vuln gosec

lint: fmt vet staticcheck

fmt:
	$(GO) fmt $(PKG)

tidy:
	$(GO) mod tidy

run:
	$(GO) run ./cmd/godian build

clean:
	rm -rf bin
