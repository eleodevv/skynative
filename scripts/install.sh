#!/bin/sh
#
# Instalador de Skynative para Linux y macOS.
#
# Uso:
#   curl -fsSL https://github.com/eleodevv/skynative/releases/latest/download/install.sh | sh
#
# Variables opcionales:
#   SKYNATIVE_VERSION  versión a instalar (por defecto "latest"; acepta "v0.1.0")
#   SKYNATIVE_DIR      directorio de instalación (por defecto "$HOME/.local/bin")
#   SKYNATIVE_REPO     repositorio "usuario/proyecto" (por defecto "eleodevv/skynative")
#
# No ejecuta ningún binario antes de verificar su checksum SHA-256 y no requiere
# privilegios de administrador.
#
# Desinstalación:
#   curl -fsSL https://github.com/eleodevv/skynative/releases/latest/download/desinstalar.sh | sh

set -eu

REPO="${SKYNATIVE_REPO:-eleodevv/skynative}"
VERSION="${SKYNATIVE_VERSION:-latest}"
DESTINO="${SKYNATIVE_DIR:-$HOME/.local/bin}"

error() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}
info() {
  printf '%s\n' "$1"
}

# 1. Detectar el sistema operativo.
so="$(uname -s)"
case "$so" in
  Linux) goos="linux" ;;
  Darwin) goos="darwin" ;;
  *) error "sistema operativo no compatible: $so (solo se admiten Linux y macOS)" ;;
esac

# 2. Detectar la arquitectura.
maquina="$(uname -m)"
case "$maquina" in
  x86_64 | amd64) goarch="amd64" ;;
  arm64 | aarch64) goarch="arm64" ;;
  *) error "arquitectura no compatible: $maquina" ;;
esac

artefacto="skynative-${goos}-${goarch}"

# 3. Construir la URL de descarga.
if [ "$VERSION" = "latest" ]; then
  base="https://github.com/${REPO}/releases/latest/download"
else
  base="https://github.com/${REPO}/releases/download/${VERSION}"
fi
url="${base}/${artefacto}"

if command -v curl >/dev/null 2>&1; then
  descargar() { curl -fsSL "$1" -o "$2"; }
elif command -v wget >/dev/null 2>&1; then
  descargar() { wget -qO "$2" "$1"; }
else
  error "se requiere 'curl' o 'wget' para descargar los artefactos"
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

info "Descargando ${artefacto} (${VERSION})..."
descargar "$url" "$tmp/$artefacto" || error "no se pudo descargar $url"
descargar "${url}.sha256" "$tmp/${artefacto}.sha256" || error "no se pudo descargar el checksum de $artefacto"

# 4. Verificar el checksum antes de instalar o ejecutar.
esperado="$(awk '{print $1}' "$tmp/${artefacto}.sha256")"
if command -v sha256sum >/dev/null 2>&1; then
  obtenido="$(sha256sum "$tmp/$artefacto" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  obtenido="$(shasum -a 256 "$tmp/$artefacto" | awk '{print $1}')"
else
  error "se requiere 'sha256sum' o 'shasum' para verificar la integridad"
fi

if [ "$esperado" != "$obtenido" ]; then
  error "el checksum no coincide; se cancela la instalación"
fi
info "Checksum verificado."

# 5. Instalar en el directorio del usuario (sin privilegios elevados).
mkdir -p "$DESTINO"
if ! install -m 0755 "$tmp/$artefacto" "$DESTINO/skynative" 2>/dev/null; then
  cp "$tmp/$artefacto" "$DESTINO/skynative"
  chmod 0755 "$DESTINO/skynative"
fi
info "Instalado en $DESTINO/skynative"

# 6. Informar sobre el PATH.
case ":${PATH:-}:" in
  *":${DESTINO}:"*) ;;
  *)
    info ""
    info "Añade el directorio al PATH si aún no está. Por ejemplo:"
    info "  export PATH=\"${DESTINO}:\$PATH\""
    ;;
esac

info ""
info "Comprueba la instalación con: skynative --version"
