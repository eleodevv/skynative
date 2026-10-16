# Makefile de Skynative.
#
# Objetivos principales:
#   make compilar-local              Compila el binario para la plataforma actual.
#   make probar                      Ejecuta las pruebas de Go.
#   make verificar                   gofmt + go vet + pruebas.
#   make compilar-todas-plataformas  Genera los binarios para todas las plataformas.
#   make limpiar                     Elimina los artefactos generados.

BINARIO   := skynative
PAQUETE   := github.com/skynative/skynative
DIST      := dist
VERSION   ?= 0.1.0
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo desconocido)
FECHA     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   := -s -w \
	-X $(PAQUETE)/interno/version.Version=$(VERSION) \
	-X $(PAQUETE)/interno/version.Commit=$(COMMIT) \
	-X $(PAQUETE)/interno/version.Fecha=$(FECHA)

.PHONY: ayuda compilar-local probar verificar compilar-todas-plataformas limpiar

ayuda:
	@echo "Objetivos disponibles:"
	@echo "  compilar-local              Compila el binario para la plataforma actual."
	@echo "  probar                      Ejecuta las pruebas de Go."
	@echo "  verificar                   gofmt + go vet + pruebas."
	@echo "  compilar-todas-plataformas  Genera los binarios multiplataforma y sus checksums."
	@echo "  limpiar                     Elimina el directorio $(DIST)/."

compilar-local:
	@mkdir -p $(DIST)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARIO) ./cmd/$(BINARIO)
	@echo "Binario generado en $(DIST)/$(BINARIO)"

probar:
	go test ./...

verificar:
	@echo "==> gofmt"
	@test -z "$$(gofmt -l .)" || { echo "Archivos sin formatear:"; gofmt -l .; exit 1; }
	@echo "==> go vet"
	go vet ./...
	@echo "==> pruebas"
	go test ./...
	@echo "Verificación completada."

compilar-todas-plataformas:
	VERSION="$(VERSION)" COMMIT="$(COMMIT)" FECHA="$(FECHA)" bash scripts/compilar-binarios.sh

limpiar:
	rm -rf $(DIST)
