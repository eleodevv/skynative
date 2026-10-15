package compilador

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLocalizarSrcTauri(t *testing.T) {
	base := t.TempDir()
	srcTauri := filepath.Join(base, "src-tauri")
	if err := os.MkdirAll(srcTauri, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcTauri, "tauri.conf.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := localizarSrcTauri(base); err != nil {
		t.Errorf("no se encontró src-tauri desde la raíz: %v", err)
	}
	if _, err := localizarSrcTauri(srcTauri); err != nil {
		t.Errorf("no se reconoció el propio directorio src-tauri: %v", err)
	}
	if _, err := localizarSrcTauri(t.TempDir()); err == nil {
		t.Error("un directorio sin tauri.conf.json debería fallar")
	}
}

func TestComandoTauriForzado(t *testing.T) {
	nombre, args, err := comandoTauri("/ruta/falsa/tauri")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if nombre != "/ruta/falsa/tauri" || len(args) != 0 {
		t.Errorf("comando forzado inesperado: %q %v", nombre, args)
	}
}

func TestCompilarRutaInvalida(t *testing.T) {
	_, err := Compilar(context.Background(), t.TempDir(), Opciones{}, nil)
	if err == nil {
		t.Error("Compilar debería fallar sin tauri.conf.json")
	}
}

func TestBuscarArtefactos(t *testing.T) {
	srcTauri := t.TempDir()
	dirBundle := filepath.Join(srcTauri, "target", "release", "bundle")
	deb := filepath.Join(dirBundle, "deb")
	appimage := filepath.Join(dirBundle, "appimage")
	for _, d := range []string{deb, appimage} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(deb, "app.deb"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appimage, "app.AppImage"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Binario suelto en target/release.
	if err := os.WriteFile(filepath.Join(srcTauri, "target", "release", "app"), []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	artefactos, err := BuscarArtefactos(srcTauri)
	if err != nil {
		t.Fatalf("BuscarArtefactos devolvió error: %v", err)
	}
	if len(artefactos) != 3 {
		t.Errorf("se esperaban 3 artefactos, se obtuvieron %d: %v", len(artefactos), artefactos)
	}
	for _, esperado := range []string{".deb", ".AppImage", "release/app"} {
		encontrado := false
		for _, a := range artefactos {
			if strings.Contains(a, esperado) {
				encontrado = true
			}
		}
		if !encontrado {
			t.Errorf("no se encontró un artefacto que contenga %q", esperado)
		}
	}
}

// TestCompilarConComandoFalso verifica el flujo completo (ejecución del proceso
// y descubrimiento de artefactos) sin necesidad de Rust, usando un guion que
// simula la CLI de Tauri.
func TestCompilarConComandoFalso(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("el guion de prueba requiere sh")
	}

	base := t.TempDir()
	srcTauri := filepath.Join(base, "src-tauri")
	if err := os.MkdirAll(srcTauri, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcTauri, "tauri.conf.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	dirBundle := filepath.Join(srcTauri, "target", "release", "bundle", "deb")
	artefacto := filepath.Join(dirBundle, "falso.deb")
	guion := "#!/bin/sh\nmkdir -p '" + dirBundle + "'\n: > '" + artefacto + "'\necho compilación simulada\nexit 0\n"
	rutaGuion := filepath.Join(base, "tauri-falso.sh")
	if err := os.WriteFile(rutaGuion, []byte(guion), 0o755); err != nil {
		t.Fatal(err)
	}

	var registro strings.Builder
	res, err := Compilar(context.Background(), base, Opciones{ComandoTauri: rutaGuion}, &registro)
	if err != nil {
		t.Fatalf("Compilar devolvió error: %v", err)
	}
	if res.CodigoSalida != 0 {
		t.Errorf("código de salida inesperado: %d", res.CodigoSalida)
	}
	if !strings.Contains(registro.String(), "compilación simulada") {
		t.Error("la salida del proceso no se registró")
	}
	if len(res.Artefactos) != 1 || res.Artefactos[0] != artefacto {
		t.Errorf("artefactos inesperados: %v", res.Artefactos)
	}
}
