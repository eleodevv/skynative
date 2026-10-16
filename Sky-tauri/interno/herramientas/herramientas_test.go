package herramientas

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

// TestPlanIncluyeRustYTauri comprueba que todos los planes instalan Rust y la
// CLI de Tauri.
func TestPlanIncluyeRustYTauri(t *testing.T) {
	for _, so := range []string{"linux", "darwin", "windows"} {
		pasos := Plan(so, "")
		if len(pasos) == 0 {
			t.Fatalf("%s: plan vacío", so)
		}
		var hayRust, hayTauri bool
		for _, p := range pasos {
			if strings.Contains(strings.ToLower(p.Descripcion), "rust") {
				hayRust = true
			}
			if strings.Contains(strings.Join(p.Args, " "), "tauri-cli") {
				hayTauri = true
			}
		}
		if !hayRust {
			t.Errorf("%s: el plan no instala Rust", so)
		}
		if !hayTauri {
			t.Errorf("%s: el plan no instala la CLI de Tauri", so)
		}
	}
}

// TestPlanLinuxApt incluye las dependencias nativas de Tauri.
func TestPlanLinuxApt(t *testing.T) {
	pasos := Plan("linux", "apt-get")
	var encontrado bool
	for _, p := range pasos {
		texto := p.Descripcion + " " + strings.Join(p.Args, " ")
		if strings.Contains(texto, "libwebkit2gtk-4.1-dev") {
			encontrado = true
		}
	}
	if !encontrado {
		t.Fatalf("no se encontró la instalación de dependencias: %+v", pasos)
	}
}

// TestPlanLinuxSinGestor marca el paso informativo cuando no hay gestor.
func TestPlanLinuxSinGestor(t *testing.T) {
	pasos := Plan("linux", "")
	if len(pasos) == 0 {
		t.Fatal("plan vacío")
	}
	if pasos[0].Nombre != "" {
		t.Fatalf("se esperaba un paso informativo, se obtuvo %q", pasos[0].Nombre)
	}
}

// TestPaquetesLinux cubre los tres gestores admitidos.
func TestPaquetesLinux(t *testing.T) {
	casos := map[string]string{
		"apt-get": "libwebkit2gtk-4.1-dev",
		"dnf":     "webkit2gtk4.1-devel",
		"pacman":  "webkit2gtk-4.1",
	}
	for gestor, esperado := range casos {
		args := paquetesLinux(gestor)
		if !strings.Contains(strings.Join(args, " "), esperado) {
			t.Errorf("%s: faltan paquetes (%v)", gestor, args)
		}
	}
	if paquetesLinux("desconocido") != nil {
		t.Error("un gestor desconocido no debe devolver paquetes")
	}
}

// TestInstalarEjecutaComando comprueba que el runner ejecuta comandos.
func TestInstalarEjecutaComando(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo no es un ejecutable en Windows")
	}
	var registro bytes.Buffer
	pasos := []Paso{{Descripcion: "saludo", Nombre: "echo", Args: []string{"herramientas-ok"}}}
	if err := Instalar(context.Background(), pasos, &registro); err != nil {
		t.Fatalf("Instalar devolvió error: %v", err)
	}
	if !strings.Contains(registro.String(), "herramientas-ok") {
		t.Fatalf("la salida no contiene el texto esperado: %q", registro.String())
	}
}

// TestInstalarPasoInformativo no debe fallar con pasos manuales.
func TestInstalarPasoInformativo(t *testing.T) {
	var registro bytes.Buffer
	pasos := []Paso{{Descripcion: "haz esto a mano"}}
	if err := Instalar(context.Background(), pasos, &registro); err != nil {
		t.Fatalf("Instalar devolvió error: %v", err)
	}
	if !strings.Contains(registro.String(), "haz esto a mano") {
		t.Fatalf("no se registró la descripción: %q", registro.String())
	}
}

// TestInstalarDescarga verifica la descarga con un servidor de prueba.
func TestInstalarDescarga(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("contenido-de-prueba"))
	}))
	defer servidor.Close()

	destino := t.TempDir() + "/descarga.txt"
	pasos := []Paso{{Descripcion: "descargar", Descargar: servidor.URL, Destino: destino}}
	var registro bytes.Buffer
	if err := Instalar(context.Background(), pasos, &registro); err != nil {
		t.Fatalf("Instalar devolvió error: %v", err)
	}
	contenido, err := os.ReadFile(destino)
	if err != nil {
		t.Fatalf("no se creó el fichero: %v", err)
	}
	if string(contenido) != "contenido-de-prueba" {
		t.Fatalf("contenido inesperado: %q", contenido)
	}
}
