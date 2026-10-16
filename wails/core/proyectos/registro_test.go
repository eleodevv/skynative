package proyectos

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistroVacio(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	lista, err := Listar()
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if len(lista) != 0 {
		t.Errorf("se esperaba un registro vacío, hay %d entradas", len(lista))
	}
}

func TestAgregarYListar(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())

	if err := Agregar(Proyecto{Nombre: "Zeta", Ruta: "/tmp/zeta"}); err != nil {
		t.Fatalf("Agregar devolvió error: %v", err)
	}
	if err := Agregar(Proyecto{Nombre: "Alfa", Ruta: "/tmp/alfa"}); err != nil {
		t.Fatalf("Agregar devolvió error: %v", err)
	}

	lista, err := Listar()
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if len(lista) != 2 {
		t.Fatalf("se esperaban 2 proyectos, hay %d", len(lista))
	}
	if lista[0].Nombre != "Alfa" || lista[1].Nombre != "Zeta" {
		t.Errorf("el orden no es alfabético: %v", lista)
	}
	if lista[0].Creado == "" {
		t.Error("la fecha de creación debería rellenarse automáticamente")
	}
}

func TestAgregarReemplazaPorRuta(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	ruta := filepath.Join(t.TempDir(), "proyecto")

	if err := Agregar(Proyecto{Nombre: "Primero", Ruta: ruta}); err != nil {
		t.Fatal(err)
	}
	if err := Agregar(Proyecto{Nombre: "Segundo", Ruta: ruta}); err != nil {
		t.Fatal(err)
	}

	lista, err := Listar()
	if err != nil {
		t.Fatal(err)
	}
	if len(lista) != 1 {
		t.Fatalf("se esperaba una sola entrada, hay %d", len(lista))
	}
	if lista[0].Nombre != "Segundo" {
		t.Errorf("no se reemplazó la entrada: %v", lista[0])
	}
}

func TestEliminar(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	ruta := filepath.Join(t.TempDir(), "proyecto")

	if err := Agregar(Proyecto{Nombre: "Uno", Ruta: "/tmp/uno"}); err != nil {
		t.Fatal(err)
	}
	if err := Agregar(Proyecto{Nombre: "Dos", Ruta: ruta}); err != nil {
		t.Fatal(err)
	}

	if err := Eliminar(ruta); err != nil {
		t.Fatalf("Eliminar devolvió error: %v", err)
	}
	lista, err := Listar()
	if err != nil {
		t.Fatal(err)
	}
	if len(lista) != 1 || lista[0].Ruta != "/tmp/uno" {
		t.Fatalf("no se eliminó la entrada correcta: %v", lista)
	}
	if err := Eliminar(ruta); err == nil {
		t.Error("eliminar una ruta inexistente debería devolver error")
	}
}

func TestRegistroDanado(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", dir)
	if err := os.WriteFile(filepath.Join(dir, "proyectos.json"), []byte("{no es json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Cargar(); err == nil {
		t.Error("un registro dañado debería devolver error")
	}
}
