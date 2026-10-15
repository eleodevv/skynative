package tui

import (
	"os"
	"path/filepath"
	"sort"
)

// accionExplorador indica qué hace una entrada del explorador al pulsar Enter.
type accionExplorador int

const (
	accionSeleccionar accionExplorador = iota
	accionSubir
	accionEntrar
)

// entradaExplorador es una fila del explorador de carpetas.
type entradaExplorador struct {
	etiqueta string
	ruta     string
	accion   accionExplorador
}

// exploradorModelo es un explorador de carpetas integrado en la TUI. Evita
// depender del gestor de archivos del sistema para elegir el destino.
type exploradorModelo struct {
	directorio string
	entradas   []entradaExplorador
	indice     int
	err        error
}

// nuevoExplorador construye un explorador situado en el directorio indicado.
func nuevoExplorador(directorio string) exploradorModelo {
	if abs, err := filepath.Abs(directorio); err == nil {
		directorio = abs
	}
	e := exploradorModelo{directorio: directorio}
	e.recargar()
	return e
}

// recargar vuelve a leer el directorio actual.
func (e *exploradorModelo) recargar() {
	e.err = nil
	e.entradas = nil

	// Primera fila: usar la carpeta actual como destino.
	e.entradas = append(e.entradas, entradaExplorador{
		etiqueta: "« Usar esta carpeta »",
		ruta:     e.directorio,
		accion:   accionSeleccionar,
	})

	if padre := filepath.Dir(e.directorio); padre != e.directorio {
		e.entradas = append(e.entradas, entradaExplorador{
			etiqueta: "↑ Subir un nivel",
			ruta:     padre,
			accion:   accionSubir,
		})
	}

	info, err := os.ReadDir(e.directorio)
	if err != nil {
		e.err = err
		e.indice = 0
		return
	}

	var nombres []string
	for _, entrada := range info {
		if entrada.IsDir() {
			nombres = append(nombres, entrada.Name())
		}
	}
	sort.Strings(nombres)
	for _, nombre := range nombres {
		e.entradas = append(e.entradas, entradaExplorador{
			etiqueta: nombre,
			ruta:     filepath.Join(e.directorio, nombre),
			accion:   accionEntrar,
		})
	}

	if e.indice >= len(e.entradas) {
		e.indice = 0
	}
}

// mover desplaza el cursor.
func (e *exploradorModelo) mover(delta int) {
	if len(e.entradas) == 0 {
		return
	}
	e.indice += delta
	if e.indice < 0 {
		e.indice = 0
	}
	if e.indice >= len(e.entradas) {
		e.indice = len(e.entradas) - 1
	}
}

// subir cambia al directorio padre.
func (e *exploradorModelo) subir() {
	padre := filepath.Dir(e.directorio)
	if padre == e.directorio {
		return
	}
	e.directorio = padre
	e.indice = 0
	e.recargar()
}

// entrar cambia al directorio indicado.
func (e *exploradorModelo) entrar(directorio string) {
	e.directorio = directorio
	e.indice = 0
	e.recargar()
}

// seleccionado devuelve la entrada bajo el cursor.
func (e exploradorModelo) seleccionado() (entradaExplorador, bool) {
	if e.indice < 0 || e.indice >= len(e.entradas) {
		return entradaExplorador{}, false
	}
	return e.entradas[e.indice], true
}
