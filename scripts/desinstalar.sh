#!/bin/sh
#
# Desinstalador de Skynative para Linux y macOS.
#
# Uso:
#   curl -fsSL https://github.com/eleodevv/skynative/releases/latest/download/desinstalar.sh | sh
#
# Variables opcionales:
#   SKYNATIVE_DIR  directorio donde se instaló (por defecto "$HOME/.local/bin")
#
# Elimina el binario «skynative» del directorio indicado. No toca los proyectos
# ni las aplicaciones que hayas creado con la herramienta.

set -eu

DESTINO="${SKYNATIVE_DIR:-$HOME/.local/bin}"
BINARIO="$DESTINO/skynative"

info() {
  printf '%s\n' "$1"
}

if [ ! -e "$BINARIO" ]; then
  info "Skynative no está instalado en $BINARIO"
  info ""
  info "Si lo instalaste en otro directorio, indica la ruta con SKYNATIVE_DIR:"
  info "  SKYNATIVE_DIR=/otra/ruta sh desinstalar.sh"
  exit 0
fi

rm -f "$BINARIO"
info "Skynative eliminado de $BINARIO"

# Restos del PATH: solo informamos, no modificamos los perfiles del usuario.
case ":${PATH:-}:" in
  *":${DESTINO}:"*)
    info ""
    info "Nota: $DESTINO sigue en tu PATH. Puedes quitar la línea que lo añade"
    info "de tu ~/.bashrc, ~/.zshrc o ~/.profile si ya no lo necesitas."
    ;;
esac

info ""
info "Hecho. Los proyectos creados con Skynative permanecen intactos."
