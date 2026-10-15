#!/usr/bin/env bash
#
# Compila los binarios de Skynative para todas las plataformas compatibles y
# genera los archivos de checksum SHA-256.
#
# Go puede compilar para otra plataforma desde cualquier sistema con CGO
# desactivado. Esto NO genera instaladores de Tauri: esos requieren el
# toolchain de Rust y las dependencias de cada sistema destino.
#
# Uso:
#   VERSION=0.1.0 COMMIT=abc123 bash scripts/compilar-binarios.sh

set -euo pipefail

PAQUETE="github.com/skynative/skynative"
DIST="${DIST:-dist}"
VERSION="${VERSION:-0.1.0}"
COMMIT="${COMMIT:-desconocido}"
FECHA="${FECHA:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

LDFLAGS="-s -w \
-X ${PAQUETE}/interno/version.Version=${VERSION} \
-X ${PAQUETE}/interno/version.Commit=${COMMIT} \
-X ${PAQUETE}/interno/version.Fecha=${FECHA}"

# Formato: "goos/goarch:nombre-del-artefacto"
PLATAFORMAS=(
  "linux/amd64:skynative-linux-amd64"
  "linux/arm64:skynative-linux-arm64"
  "darwin/amd64:skynative-darwin-amd64"
  "darwin/arm64:skynative-darwin-arm64"
  "windows/amd64:skynative-windows-amd64.exe"
  "windows/arm64:skynative-windows-arm64.exe"
)

mkdir -p "$DIST"

echo "Compilando Skynative ${VERSION} (commit ${COMMIT}, ${FECHA})"
for entrada in "${PLATAFORMAS[@]}"; do
  plataforma="${entrada%%:*}"
  nombre="${entrada##*:}"
  goos="${plataforma%%/*}"
  goarch="${plataforma##*/}"

  echo "  -> ${goos}/${goarch}  ${nombre}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath -ldflags "$LDFLAGS" -o "${DIST}/${nombre}" ./cmd/skynative
done

# Checksums individuales y combinados.
echo "Generando checksums SHA-256"
(
  cd "$DIST"
  rm -f SHA256SUMS ./*.sha256
  for archivo in skynative-*; do
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "$archivo" > "${archivo}.sha256"
    else
      shasum -a 256 "$archivo" > "${archivo}.sha256"
    fi
  done
  cat ./*.sha256 > SHA256SUMS
)

echo "Listo. Artefactos en ${DIST}/:"
ls -1 "$DIST"
