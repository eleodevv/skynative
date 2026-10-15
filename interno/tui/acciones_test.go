package tui

import (
	"strings"
	"testing"
)

func TestEscritorCanalDivideLineas(t *testing.T) {
	canal := make(chan string, 10)
	escritor := nuevoEscritorCanal(canal)

	// Escrituras fragmentadas a propósito.
	if _, err := escritor.Write([]byte("hola ")); err != nil {
		t.Fatal(err)
	}
	if _, err := escritor.Write([]byte("mundo\nsegunda\nter")); err != nil {
		t.Fatal(err)
	}
	escritor.cerrar()
	close(canal)

	var lineas []string
	for linea := range canal {
		lineas = append(lineas, linea)
	}
	esperado := []string{"hola mundo", "segunda", "ter"}
	if len(lineas) != len(esperado) {
		t.Fatalf("líneas = %v; se esperaban %v", lineas, esperado)
	}
	for i := range esperado {
		if lineas[i] != esperado[i] {
			t.Errorf("línea %d = %q; se esperaba %q", i, lineas[i], esperado[i])
		}
	}
}

func TestEscritorCanalNoBloqueaSiEstaLleno(t *testing.T) {
	canal := make(chan string, 1)
	escritor := nuevoEscritorCanal(canal)
	// Muchas líneas con un canal diminuto no deben bloquear.
	for i := 0; i < 100; i++ {
		if _, err := escritor.Write([]byte("linea\n")); err != nil {
			t.Fatal(err)
		}
	}
	if len(canal) == 0 {
		t.Error("debería haberse conservado al menos una línea")
	}
}

func TestAbrirEnExploradorRutaVaciaNoHaceNada(t *testing.T) {
	// No debe entrar en pánico ni lanzar procesos.
	abrirEnExplorador("")
	_ = strings.TrimSpace("")
}
