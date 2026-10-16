// Package version expone la información de compilación de Skynative.
//
// Los tres valores siguientes se pueden sustituir en tiempo de enlace con
// variables de enlace de Go, por ejemplo:
//
//	go build -ldflags "-X github.com/skynative/skynative/interno/version.Version=0.1.0"
package version

import "fmt"

// Valores por defecto. El Makefile y los flujos de GitHub Actions los
// sobrescriben mediante -ldflags "-X ...".
var (
	// Version es la versión semántica de Skynative.
	Version = "0.4.1"
	// Commit es el hash corto del commit con el que se compiló el binario.
	Commit = "desconocido"
	// Fecha es la fecha de compilación en formato RFC3339 (UTC).
	Fecha = "desconocida"
)

// Metadatos del proyecto. Los flujos de GitHub Actions y los scripts de
// instalación usan estas mismas rutas.
var (
	// Repositorio es la URL del repositorio del proyecto.
	Repositorio = "https://github.com/eleodevv/skynative"
	// URLInstaladorLinux es el instalador para Linux y macOS.
	URLInstaladorLinux = Repositorio + "/releases/latest/download/install.sh"
	// URLInstaladorWindows es el instalador para Windows (PowerShell).
	URLInstaladorWindows = Repositorio + "/releases/latest/download/install.ps1"
)

// Cadena devuelve una representación legible de la versión del binario.
func Cadena() string {
	return fmt.Sprintf("%s (commit %s, compilada %s)", Version, Commit, Fecha)
}
