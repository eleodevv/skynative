// Command skynative abre la interfaz de terminal que convierte una URL web en
// una aplicación de escritorio con Tauri v2.
package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/skynative/sky-tauri/interno/tui"
	"github.com/skynative/sky-tauri/interno/version"
)

// mensajeAyuda es el texto que se muestra con -h/--help.
const mensajeAyuda = `Skynative — convierte una URL en una aplicación de escritorio (Tauri v2).

Skynative es una interfaz de terminal. Ejecútalo sin argumentos para abrirla:

    skynative

Opciones:
  -h, --help      Muestra esta ayuda.
  -v, --version   Muestra la versión instalada.

Dentro de la interfaz podrás crear aplicaciones, elegir la carpeta de salida con
el explorador integrado, compilarlas y ver el diagnóstico de tu entorno.`

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Println(mensajeAyuda)
			return
		case "-v", "--version":
			fmt.Printf("skynative %s\n", version.Cadena())
			return
		}
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Fprintln(os.Stderr, "Skynative necesita una terminal interactiva para mostrar su interfaz.")
		fmt.Fprintln(os.Stderr, "Ejecútalo directamente en una terminal (no a través de una tubería).")
		os.Exit(1)
	}

	if err := tui.Ejecutar(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
