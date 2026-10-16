// Package diagnostico comprueba que el sistema tiene todo lo necesario para
// generar y compilar aplicaciones Tauri v2.
package diagnostico

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

// Estado describe el resultado de una comprobación concreta.
type Estado struct {
	// Nombre identifica la comprobación.
	Nombre string
	// Detalle explica el resultado (versión, error o recomendación).
	Detalle string
	// OK indica si la comprobación fue satisfactoria.
	OK bool
	// Critico indica si su ausencia impide compilar aplicaciones.
	Critico bool
	// Ayuda es la acción recomendada cuando algo falta.
	Ayuda string
}

// Ejecutar realiza todas las comprobaciones y devuelve su resultado.
func Ejecutar(ctx context.Context) []Estado {
	estados := []Estado{{
		Nombre:  "Go",
		Detalle: runtime.Version(),
		OK:      true,
		Critico: true,
	}}

	if v, err := primeraVersion(ctx, "rustc", "--version"); err == nil {
		estados = append(estados, Estado{Nombre: "Rust (rustc)", Detalle: v, OK: true, Critico: true})
	} else {
		estados = append(estados, Estado{
			Nombre:  "Rust (rustc)",
			Detalle: "no encontrado",
			Critico: true,
			Ayuda:   "instala Rust desde https://rustup.rs",
		})
	}

	if v, err := primeraVersion(ctx, "cargo", "--version"); err == nil {
		estados = append(estados, Estado{Nombre: "Cargo", Detalle: v, OK: true, Critico: true})
	} else {
		estados = append(estados, Estado{
			Nombre:  "Cargo",
			Detalle: "no encontrado",
			Critico: true,
			Ayuda:   "instala Rust desde https://rustup.rs",
		})
	}

	estados = append(estados, comprobarTauriCLI(ctx))
	estados = append(estados, dependenciasEscritorio(ctx)...)
	return estados
}

// HayCriticosFallidos indica si alguna comprobación crítica no pasó.
func HayCriticosFallidos(estados []Estado) bool {
	for _, e := range estados {
		if e.Critico && !e.OK {
			return true
		}
	}
	return false
}

// comprobarTauriCLI busca la CLI de Tauri como subcomando de cargo o suelta.
func comprobarTauriCLI(ctx context.Context) Estado {
	if v, err := primeraVersion(ctx, "cargo", "tauri", "--version"); err == nil {
		return Estado{Nombre: "CLI de Tauri", Detalle: v, OK: true, Critico: true}
	}
	if v, err := primeraVersion(ctx, "tauri", "--version"); err == nil {
		return Estado{Nombre: "CLI de Tauri", Detalle: v, OK: true, Critico: true}
	}
	return Estado{
		Nombre:  "CLI de Tauri",
		Detalle: "no disponible",
		Critico: true,
		Ayuda:   "cargo install tauri-cli --version \"^2\" (o npm install -g @tauri-apps/cli)",
	}
}

// dependenciasEscritorio comprueba las bibliotecas nativas por plataforma.
func dependenciasEscritorio(ctx context.Context) []Estado {
	switch runtime.GOOS {
	case "linux":
		libs := []struct{ pkg, nombre string }{
			{"webkit2gtk-4.1", "WebKitGTK 4.1"},
			{"gtk+-3.0", "GTK 3"},
			{"libsoup-3.0", "libsoup 3"},
			{"javascriptcoregtk-4.1", "JavaScriptCoreGTK 4.1"},
		}
		var faltan []string
		for _, l := range libs {
			if !pkgConfigExiste(ctx, l.pkg) {
				faltan = append(faltan, l.nombre)
			}
		}
		if len(faltan) == 0 {
			return []Estado{{Nombre: "Dependencias de escritorio (Linux)", Detalle: "todas presentes", OK: true, Critico: true}}
		}
		return []Estado{{
			Nombre:  "Dependencias de escritorio (Linux)",
			Detalle: "faltan: " + strings.Join(faltan, ", "),
			Critico: true,
			Ayuda: "Debian/Ubuntu: sudo apt install libwebkit2gtk-4.1-dev libgtk-3-dev " +
				"libsoup-3.0-dev build-essential curl wget file libxdo-dev libssl-dev " +
				"libayatana-appindicator3-dev librsvg2-dev",
		}}
	case "darwin":
		if _, err := exec.CommandContext(ctx, "xcode-select", "-p").Output(); err == nil {
			return []Estado{{Nombre: "Herramientas de macOS", Detalle: "Xcode Command Line Tools instaladas", OK: true, Critico: true}}
		}
		return []Estado{{
			Nombre:  "Herramientas de macOS",
			Detalle: "faltan las Xcode Command Line Tools",
			Critico: true,
			Ayuda:   "xcode-select --install",
		}}
	case "windows":
		return []Estado{{
			Nombre:  "WebView2",
			Detalle: "se requiere Microsoft Edge WebView2 Runtime (incluido en Windows 10/11)",
			OK:      true,
			Critico: false,
			Ayuda:   "https://developer.microsoft.com/microsoft-edge/webview2/",
		}}
	default:
		return nil
	}
}

// primeraVersion devuelve la primera línea no vacía de la salida de un comando.
func primeraVersion(ctx context.Context, nombre string, args ...string) (string, error) {
	salida, err := exec.CommandContext(ctx, nombre, args...).Output()
	if err != nil {
		return "", err
	}
	for _, linea := range strings.Split(string(salida), "\n") {
		linea = strings.TrimSpace(linea)
		if linea != "" {
			return linea, nil
		}
	}
	return "", nil
}

// pkgConfigExiste comprueba si pkg-config conoce una biblioteca.
func pkgConfigExiste(ctx context.Context, pkg string) bool {
	if _, err := exec.LookPath("pkg-config"); err != nil {
		return false
	}
	return exec.CommandContext(ctx, "pkg-config", "--exists", pkg).Run() == nil
}
