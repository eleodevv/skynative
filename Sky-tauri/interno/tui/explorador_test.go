package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExploradorListaSubdirectorios(t *testing.T) {
	base := t.TempDir()
	for _, nombre := range []string{"zeta", "alfa"} {
		if err := os.MkdirAll(filepath.Join(base, nombre), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Un archivo no debe aparecer en el explorador de carpetas.
	if err := os.WriteFile(filepath.Join(base, "nota.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	e := nuevoExplorador(base)
	if len(e.entradas) != 4 { // seleccionar + subir + 2 carpetas
		t.Fatalf("se esperaban 4 entradas, hay %d: %+v", len(e.entradas), e.entradas)
	}
	if e.entradas[0].accion != accionSeleccionar {
		t.Error("la primera entrada debe ser «usar esta carpeta»")
	}
	if e.entradas[1].accion != accionSubir {
		t.Error("la segunda entrada debe ser «subir»")
	}
	if e.entradas[2].etiqueta != "alfa" || e.entradas[3].etiqueta != "zeta" {
		t.Errorf("las carpetas no están ordenadas: %v", []string{e.entradas[2].etiqueta, e.entradas[3].etiqueta})
	}
}

func TestExploradorMoverYEntrar(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "proyectos")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	e := nuevoExplorador(base)
	e.mover(100) // no debe salirse del rango
	if e.indice != len(e.entradas)-1 {
		t.Errorf("el cursor no se limitó: %d", e.indice)
	}
	e.mover(-100)
	if e.indice != 0 {
		t.Errorf("el cursor no se limitó por abajo: %d", e.indice)
	}

	e.entrar(sub)
	if e.directorio != sub {
		t.Errorf("no se entró en el subdirectorio: %q", e.directorio)
	}
	if _, ok := e.seleccionado(); !ok {
		t.Error("debería haber una entrada seleccionada")
	}
}

func TestExploradorSubir(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "hijo")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	e := nuevoExplorador(sub)
	e.subir()
	if e.directorio != base {
		t.Errorf("no se subió al directorio padre: %q (esperado %q)", e.directorio, base)
	}
}
