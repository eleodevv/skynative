package instalador

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLimpiarSlug(t *testing.T) {
	casos := map[string]string{
		"Mi Aplicación": "mi-aplicaci-n",
		"whatsapp":      "whatsapp",
		"App.NET":       "app.net",
		"  ..raro//  ":  "raro",
	}
	for entrada, esperado := range casos {
		if obtenido := limpiarSlug(entrada); obtenido != esperado {
			t.Errorf("limpiarSlug(%q) = %q; se esperaba %q", entrada, obtenido, esperado)
		}
	}
}

func TestComandoEscritorioEscapaRutas(t *testing.T) {
	obtenido := comandoEscritorio(`/ruta/con "comillas" y $variable`)
	esperado := `"/ruta/con \"comillas\" y \$variable"`
	if obtenido != esperado {
		t.Errorf("comandoEscritorio = %q; se esperaba %q", obtenido, esperado)
	}
}

func TestContenidoEscritorio(t *testing.T) {
	contenido := contenidoEscritorio(Opciones{
		Nombre:     "Mi App",
		Slug:       "mi-app",
		Ejecutable: "/opt/mi app/bin",
	})
	for _, esperado := range []string{
		"[Desktop Entry]",
		"Name=Mi App",
		"Icon=mi-app",
		`Exec="/opt/mi app/bin"`,
		"Terminal=false",
		"Categories=Network;Utility;",
	} {
		if !strings.Contains(contenido, esperado) {
			t.Errorf("la entrada .desktop no contiene %q:\n%s", esperado, contenido)
		}
	}
}

func TestInstalarLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("la instalación XDG solo aplica a Linux")
	}
	base := t.TempDir()
	exe := filepath.Join(base, "bin", "miapp")
	if err := os.MkdirAll(filepath.Dir(exe), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	icono := filepath.Join(base, "128x128.png")
	if err := os.WriteFile(icono, []byte("PNG"), 0o644); err != nil {
		t.Fatal(err)
	}
	datos := t.TempDir()

	res, err := instalarLinux(Opciones{
		Nombre:          "Mi App",
		Slug:            "mi-app",
		Ejecutable:      exe,
		IconoOrigen:     icono,
		DirectorioDatos: datos,
	})
	if err != nil {
		t.Fatalf("instalarLinux devolvió error: %v", err)
	}

	info, err := os.Stat(res.Entrada)
	if err != nil {
		t.Fatalf("no se creó la entrada .desktop: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Error("la entrada .desktop debería ser ejecutable")
	}
	contenido, err := os.ReadFile(res.Entrada)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contenido), "Name=Mi App") {
		t.Errorf("la entrada no incluye el nombre:\n%s", contenido)
	}

	if _, err := os.Stat(res.Icono); err != nil {
		t.Errorf("no se copió el icono: %v", err)
	}
	esperadoIcono := filepath.Join(datos, "icons", "hicolor", "256x256", "apps", "mi-app.png")
	if res.Icono != esperadoIcono {
		t.Errorf("icono en %q; se esperaba %q", res.Icono, esperadoIcono)
	}
}

func TestInstalarLinuxSinEjecutable(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("la instalación XDG solo aplica a Linux")
	}
	if _, err := instalarLinux(Opciones{Slug: "vacio"}); err == nil {
		t.Error("debería fallar sin ejecutable")
	}
}
