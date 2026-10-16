// Command skynative abre la interfaz de terminal que convierte una URL web en
// una aplicación de escritorio con Wails v2. Por ahora expone --help y
// --version; la TUI se traslada en el siguiente paso de la migración.
package main

import (
	"fmt"
	"os"

	"github.com/skynative/skynative-wails/core/version"
)

// mensajeAyuda es el texto que se muestra con -h/--help.
const mensajeAyuda = `Skynative — convierte una URL en una aplicación de escritorio (Wails v2).

Skynative es una interfaz de terminal. Ejecútalo sin argumentos para abrirla:

    skynative

Opciones:
  -h, --help      Muestra esta ayuda.
  -v, --version   Muestra la versión instalada.`

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Println(mensajeAyuda)
			return
		case "-v", "--version":
			fmt.Printf("skynative-wails %s\n", version.Cadena())
			return
		}
	}
	fmt.Println(mensajeAyuda)
}
