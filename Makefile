BINARY := godian
PKG    := ./...
GO     ?= go

# Versión para release (se puede sobrescribir: make release VERSION=1.2.3).
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
# Flags del linker para release: -s/-w reducen el binario (~26%) y -X inyecta la versión.
LDFLAGS  = -s -w -X github.com/nefdevelop/godian/internal/presentation/cli.Version=$(VERSION)

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

.PHONY: all build release test test-unit vet cover staticcheck vuln gosec security lint fmt tidy run serve clean

all: build

build:
	$(GO) build -trimpath -o bin/$(BINARY) ./cmd/godian

# Binario de release: sin símbolos de depuración y con la versión inyectada.
release:
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o bin/$(BINARY) ./cmd/godian

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

serve:
	$(GO) run ./cmd/godian serve

clean:
	rm -rf bin
