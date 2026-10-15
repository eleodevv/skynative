package diagnostico

import (
	"context"
	"testing"
)

func TestEjecutarIncluyeGo(t *testing.T) {
	estados := Ejecutar(context.Background())
	if len(estados) == 0 {
		t.Fatal("Ejecutar no devolvió ningún estado")
	}
	for _, e := range estados {
		if e.Nombre == "Go" {
			if !e.OK || e.Detalle == "" {
				t.Errorf("la comprobación de Go debería ser satisfactoria: %+v", e)
			}
			return
		}
	}
	t.Error("no se encontró la comprobación de Go")
}

func TestHayCriticosFallidos(t *testing.T) {
	if HayCriticosFallidos([]Estado{{Nombre: "A", OK: true, Critico: true}}) {
		t.Error("no debería haber críticos fallidos")
	}
	if !HayCriticosFallidos([]Estado{{Nombre: "A", OK: false, Critico: true}}) {
		t.Error("debería detectar un crítico fallido")
	}
	if HayCriticosFallidos([]Estado{{Nombre: "A", OK: false, Critico: false}}) {
		t.Error("un fallo no crítico no debería contar")
	}
}
