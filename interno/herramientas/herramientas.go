// Package herramientas instala las dependencias externas que Skynative
// necesita para generar y compilar aplicaciones Tauri v2: Rust (rustc/cargo),
// la CLI de Tauri y, en Linux, las bibliotecas nativas del sistema.
//
// Los comandos se ejecutan con exec.CommandContext y argumentos separados:
// nunca se construye una cadena que se interprete en un shell. La descarga del
// instalador de rustup se realiza con net/http y se ejecuta el fichero
// descargado, sin canalizaciones hacia un intérprete.
package herramientas

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Paso describe una acción del plan de instalación.
//
// Si Descargar no está vacío, se descarga esa URL en Destino. En caso
// contrario, si Nombre no está vacío se ejecuta ese comando con Args. Si ambos
// están vacíos, el paso es informativo (instrucciones manuales).
type Paso struct {
	// Descripcion explica qué hace el paso y se muestra en el registro.
	Descripcion string
	// Descargar es la URL a descargar (opcional).
	Descargar string
	// Destino es la ruta del fichero descargado.
	Destino string
	// Nombre es el ejecutable a lanzar.
	Nombre string
	// Args son los argumentos del ejecutable.
	Args []string
	// Opcional indica que un fallo no debe abortar el resto del plan.
	Opcional bool
}

// PlanParaSistema construye el plan para el sistema operativo actual.
func PlanParaSistema() []Paso {
	return Plan(runtime.GOOS, DetectarGestor())
}

// Plan construye el plan de instalación para un sistema operativo y un gestor
// de paquetes concretos (por ejemplo "apt-get", "dnf", "pacman" o "").
func Plan(so, gestor string) []Paso {
	switch so {
	case "linux":
		var pasos []Paso
		if p := pasoPaquetesLinux(gestor); p != nil {
			pasos = append(pasos, *p)
		}
		pasos = append(pasos, pasosRustUnix()...)
		pasos = append(pasos, pasoCargoTauri(so))
		return pasos
	case "darwin":
		pasos := []Paso{{
			Descripcion: "Instalando las Command Line Tools de Xcode (si faltan)",
			Nombre:      "xcode-select",
			Args:        []string{"--install"},
			Opcional:    true,
		}}
		pasos = append(pasos, pasosRustUnix()...)
		pasos = append(pasos, pasoCargoTauri(so))
		return pasos
	case "windows":
		return []Paso{
			{
				Descripcion: "Instalando Rust (rustup) con winget",
				Nombre:      "winget",
				Args: []string{
					"install", "--id", "Rustlang.Rustup", "-e",
					"--accept-source-agreements", "--accept-package-agreements",
				},
			},
			{
				Descripcion: "Instalando las herramientas de compilación de Visual Studio (MSVC)",
				Nombre:      "winget",
				Args: []string{
					"install", "--id", "Microsoft.VisualStudio.2022.BuildTools", "-e",
					"--accept-source-agreements", "--accept-package-agreements",
				},
				Opcional: true,
			},
			pasoCargoTauri(so),
		}
	}
	return nil
}

// DetectarGestor devuelve el gestor de paquetes disponible en el sistema.
func DetectarGestor() string {
	for _, g := range []string{"apt-get", "dnf", "pacman"} {
		if _, err := exec.LookPath(g); err == nil {
			return g
		}
	}
	return ""
}

// pasosRustUnix descarga y ejecuta el instalador de rustup.
func pasosRustUnix() []Paso {
	guion := filepath.Join(os.TempDir(), "skynative-rustup-init.sh")
	return []Paso{
		{Descripcion: "Descargando el instalador de Rust (rustup)", Descargar: "https://sh.rustup.rs", Destino: guion},
		{Descripcion: "Instalando Rust", Nombre: "sh", Args: []string{guion, "-y"}},
	}
}

// pasoCargoTauri instala la CLI de Tauri con cargo.
func pasoCargoTauri(so string) Paso {
	return Paso{
		Descripcion: "Instalando la CLI de Tauri (cargo install tauri-cli)",
		Nombre:      cargoBin(so),
		Args:        []string{"install", "tauri-cli", "--version", "^2"},
	}
}

// cargoBin resuelve la ruta de cargo, teniendo en cuenta que tras instalar
// rustup puede no estar todavía en el PATH del proceso actual.
func cargoBin(so string) string {
	if p, err := exec.LookPath("cargo"); err == nil {
		return p
	}
	if so == "windows" {
		if perfil := os.Getenv("USERPROFILE"); perfil != "" {
			return filepath.Join(perfil, ".cargo", "bin", "cargo.exe")
		}
		return "cargo"
	}
	if inicio, err := os.UserHomeDir(); err == nil {
		return filepath.Join(inicio, ".cargo", "bin", "cargo")
	}
	return "cargo"
}

// pasoPaquetesLinux prepara la instalación de las bibliotecas nativas.
func pasoPaquetesLinux(gestor string) *Paso {
	if gestor == "" {
		return &Paso{
			Descripcion: "No se detectó un gestor de paquetes: instala manualmente " +
				"las dependencias de Tauri para Linux (webkit2gtk-4.1, gtk3, libsoup-3.0, librsvg).",
		}
	}
	args := paquetesLinux(gestor)
	if len(args) == 0 {
		return nil
	}
	nombre := gestor
	if !esRoot() {
		if !sudoSinPassword() {
			return &Paso{
				Descripcion: "Dependencias del sistema pendientes (requieren sudo). " +
					"Ejecuta manualmente:  sudo " + gestor + " " + strings.Join(args, " "),
			}
		}
		nombre = "sudo"
		args = append([]string{gestor}, args...)
	}
	return &Paso{
		Descripcion: "Instalando dependencias del sistema (requiere sudo)",
		Nombre:      nombre,
		Args:        args,
	}
}

// paquetesLinux devuelve los argumentos para instalar las dependencias.
func paquetesLinux(gestor string) []string {
	switch gestor {
	case "apt-get", "apt":
		return []string{
			"install", "-y",
			"libwebkit2gtk-4.1-dev", "libgtk-3-dev", "libsoup-3.0-dev",
			"build-essential", "curl", "wget", "file", "libxdo-dev",
			"libssl-dev", "libayatana-appindicator3-dev", "librsvg2-dev",
		}
	case "dnf":
		return []string{
			"install", "-y",
			"webkit2gtk4.1-devel", "gtk3-devel", "libsoup3-devel",
			"openssl-devel", "curl", "wget", "file", "libxdo-devel", "librsvg2-devel",
		}
	case "pacman":
		return []string{
			"-S", "--noconfirm", "--needed",
			"webkit2gtk-4.1", "gtk3", "libsoup3", "base-devel",
			"curl", "wget", "file", "xdotool", "openssl", "librsvg",
		}
	}
	return nil
}

// esRoot indica si el proceso se ejecuta como root (solo en Unix).
func esRoot() bool {
	return os.Geteuid() == 0
}

// sudoSinPassword indica si sudo puede elevar privilegios sin pedir contraseña.
func sudoSinPassword() bool {
	if _, err := exec.LookPath("sudo"); err != nil {
		return false
	}
	return exec.Command("sudo", "-n", "true").Run() == nil
}

// Instalar ejecuta los pasos en orden escribiendo el progreso en log.
func Instalar(ctx context.Context, pasos []Paso, log io.Writer) error {
	if log == nil {
		log = io.Discard
	}
	for _, p := range pasos {
		fmt.Fprintln(log, "→ "+p.Descripcion)
		switch {
		case p.Descargar != "":
			if err := descargar(ctx, p.Descargar, p.Destino); err != nil {
				if p.Opcional {
					fmt.Fprintf(log, "  aviso: %v\n", err)
					continue
				}
				return fmt.Errorf("%s: %w", p.Descripcion, err)
			}
		case p.Nombre == "":
			fmt.Fprintln(log, "  (acción manual requerida)")
		default:
			cmd := exec.CommandContext(ctx, p.Nombre, p.Args...)
			cmd.Stdout = log
			cmd.Stderr = log
			if err := cmd.Run(); err != nil {
				if p.Opcional {
					fmt.Fprintf(log, "  aviso: %v\n", err)
					continue
				}
				return fmt.Errorf("%s: %w", p.Descripcion, err)
			}
		}
	}
	return nil
}

// descargar guarda el contenido de una URL en un fichero.
func descargar(ctx context.Context, url, destino string) error {
	peticion, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	respuesta, err := http.DefaultClient.Do(peticion)
	if err != nil {
		return err
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusOK {
		return fmt.Errorf("respuesta inesperada: %s", respuesta.Status)
	}
	archivo, err := os.Create(destino)
	if err != nil {
		return err
	}
	defer archivo.Close()
	_, err = io.Copy(archivo, respuesta.Body)
	return err
}
