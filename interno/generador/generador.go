// Package generador crea proyectos Tauri v2 completos a partir de una
// configuración validada.
//
// El paquete no ejecuta ninguna herramienta externa: solo escribe archivos de
// forma reproducible. La configuración de Tauri se construye como estructuras
// de Go y se serializa con encoding/json, de modo que los valores siempre
// quedan correctamente escapados.
package generador

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/skynative/skynative/interno/configuracion"
	"github.com/skynative/skynative/plantillas"
)

// Resultado describe el proyecto generado.
type Resultado struct {
	// Directorio es la ruta raíz del proyecto creado.
	Directorio string
	// Identificador es el identificador de paquete aplicado.
	Identificador string
	// NombrePaquete es el nombre del crate de Cargo.
	NombrePaquete string
	// RutaTauriConf es la ruta del archivo tauri.conf.json generado.
	RutaTauriConf string
	// Archivos lista todos los archivos escritos.
	Archivos []string
}

// datosPlantilla agrupa los valores que se insertan en las plantillas de texto.
type datosPlantilla struct {
	Nombre        string
	URL           string
	Identificador string
	NombrePaquete string
	Version       string
	Ancho         int
	Alto          int
}

// Generar crea el proyecto Tauri v2 en disco y devuelve el resultado.
//
// El proyecto se crea en <DirectorioSalida>/<NombrePaquete>. Si ese directorio
// ya existe y no está vacío, se devuelve un error para no sobrescribir nada.
func Generar(cfg configuracion.Configuracion) (Resultado, error) {
	cfg.Normalizar()
	if err := cfg.Validar(); err != nil {
		return Resultado{}, err
	}

	dirProyecto := filepath.Clean(filepath.Join(cfg.DirectorioSalida, cfg.NombrePaquete))
	if err := comprobarDirectorioUtilizable(dirProyecto); err != nil {
		return Resultado{}, err
	}

	dirSrcTauri := filepath.Join(dirProyecto, "src-tauri")
	dirSrc := filepath.Join(dirSrcTauri, "src")
	dirIconos := filepath.Join(dirSrcTauri, "icons")
	for _, d := range []string{dirSrc, dirIconos} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return Resultado{}, fmt.Errorf("no se pudo crear el directorio %q: %w", d, err)
		}
	}

	res := Resultado{
		Directorio:    dirProyecto,
		Identificador: cfg.Identificador,
		NombrePaquete: cfg.NombrePaquete,
	}

	datos := datosPlantilla{
		Nombre:        cfg.Nombre,
		URL:           cfg.URL,
		Identificador: cfg.Identificador,
		NombrePaquete: cfg.NombrePaquete,
		Version:       cfg.Version,
		Ancho:         cfg.Ancho,
		Alto:          cfg.Alto,
	}

	// Archivos renderizados a partir de plantillas incrustadas.
	desdePlantilla := []struct {
		plantilla string
		destino   string
	}{
		{"tauri-v2/Cargo.toml.tmpl", filepath.Join(dirSrcTauri, "Cargo.toml")},
		{"tauri-v2/README.md.tmpl", filepath.Join(dirProyecto, "README.md")},
	}
	for _, a := range desdePlantilla {
		contenido, err := renderizar(a.plantilla, datos)
		if err != nil {
			return Resultado{}, err
		}
		if err := escribirArchivo(a.destino, contenido); err != nil {
			return Resultado{}, err
		}
		res.Archivos = append(res.Archivos, a.destino)
	}

	// Archivos estáticos copiados tal cual.
	desdeEstatico := []struct {
		origen  string
		destino string
	}{
		{"tauri-v2/build.rs", filepath.Join(dirSrcTauri, "build.rs")},
		{"tauri-v2/main.rs", filepath.Join(dirSrc, "main.rs")},
		{"tauri-v2/gitignore", filepath.Join(dirProyecto, ".gitignore")},
		{"tauri-v2/capabilities/default.json", filepath.Join(dirSrcTauri, "capabilities", "default.json")},
	}
	for _, a := range desdeEstatico {
		contenido, err := plantillas.TauriV2.ReadFile(a.origen)
		if err != nil {
			return Resultado{}, fmt.Errorf("no se pudo leer la plantilla %q: %w", a.origen, err)
		}
		if err := escribirArchivo(a.destino, contenido); err != nil {
			return Resultado{}, err
		}
		res.Archivos = append(res.Archivos, a.destino)
	}

	// Configuración de Tauri construida como estructura y serializada.
	conf, err := json.MarshalIndent(construirTauriConf(cfg), "", "  ")
	if err != nil {
		return Resultado{}, fmt.Errorf("no se pudo serializar tauri.conf.json: %w", err)
	}
	conf = append(conf, '\n')
	res.RutaTauriConf = filepath.Join(dirSrcTauri, "tauri.conf.json")
	if err := escribirArchivo(res.RutaTauriConf, conf); err != nil {
		return Resultado{}, err
	}
	res.Archivos = append(res.Archivos, res.RutaTauriConf)

	// Iconos por defecto (PNG, ICO e ICNS).
	iconos, err := generarIconos(dirIconos)
	if err != nil {
		return Resultado{}, err
	}
	res.Archivos = append(res.Archivos, iconos...)

	return res, nil
}

// comprobarDirectorioUtilizable garantiza que el destino no contiene archivos.
func comprobarDirectorioUtilizable(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("no se pudo comprobar %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("la ruta %q ya existe y no es un directorio", dir)
	}
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("no se pudo leer %q: %w", dir, err)
	}
	if len(entradas) > 0 {
		return fmt.Errorf("el directorio %q ya existe y no está vacío; usa otro --salida o --nombre", dir)
	}
	return nil
}

// renderizar procesa una plantilla de texto incrustada.
func renderizar(ruta string, datos any) ([]byte, error) {
	crudo, err := plantillas.TauriV2.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la plantilla %q: %w", ruta, err)
	}
	t, err := template.New(filepath.Base(ruta)).Parse(string(crudo))
	if err != nil {
		return nil, fmt.Errorf("plantilla %q no válida: %w", ruta, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, datos); err != nil {
		return nil, fmt.Errorf("no se pudo procesar la plantilla %q: %w", ruta, err)
	}
	return buf.Bytes(), nil
}

// escribirArchivo crea los directorios necesarios y escribe el contenido.
func escribirArchivo(ruta string, datos []byte) error {
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return fmt.Errorf("no se pudo crear el directorio de %q: %w", ruta, err)
	}
	if err := os.WriteFile(ruta, datos, 0o644); err != nil {
		return fmt.Errorf("no se pudo escribir %q: %w", ruta, err)
	}
	return nil
}

// Estructuras que representan tauri.conf.json según el esquema oficial de Tauri v2.
// Referencia: https://v2.tauri.app/reference/config/
type tauriConf struct {
	Schema      string      `json:"$schema"`
	ProductName string      `json:"productName"`
	Version     string      `json:"version"`
	Identifier  string      `json:"identifier"`
	Build       tauriBuild  `json:"build"`
	App         tauriApp    `json:"app"`
	Bundle      tauriBundle `json:"bundle"`
}

type tauriBuild struct {
	// FrontendDist admite una URL remota: Tauri la carga por defecto y no
	// incrusta activos locales.
	FrontendDist string `json:"frontendDist"`
}

type tauriApp struct {
	WithGlobalTauri bool          `json:"withGlobalTauri"`
	Windows         []tauriWindow `json:"windows"`
	Security        tauriSecurity `json:"security"`
}

type tauriWindow struct {
	Title     string `json:"title"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Resizable bool   `json:"resizable"`
}

type tauriSecurity struct {
	// CSP nil significa que Tauri no inyecta una política en activos locales.
	// No hay activos locales, así que no se debilita ninguna protección del
	// contenido remoto: la página aplica su propia CSP.
	CSP any `json:"csp"`
}

type tauriBundle struct {
	Active           bool     `json:"active"`
	Targets          string   `json:"targets"`
	Category         string   `json:"category"`
	ShortDescription string   `json:"shortDescription"`
	Icon             []string `json:"icon"`
}

// objetivosBundle elige el empaquetado más fiable para el sistema de
// compilación. Se evita "all" porque incluye objetivos que suelen fallar por
// falta de herramientas (rpmbuild, linuxdeploy…); si el empaquetado falla, el
// compilador reintenta automáticamente con --no-bundle.
func objetivosBundle(so string) string {
	switch so {
	case "darwin":
		return "app"
	case "windows":
		return "nsis"
	default:
		return "deb"
	}
}

// construirTauriConf genera la configuración de Tauri v2 para la aplicación.
func construirTauriConf(cfg configuracion.Configuracion) tauriConf {
	return tauriConf{
		Schema:      "https://schema.tauri.app/config/2",
		ProductName: cfg.Nombre,
		Version:     cfg.Version,
		Identifier:  cfg.Identificador,
		Build:       tauriBuild{FrontendDist: cfg.URL},
		App: tauriApp{
			// No se expone window.__TAURI__ al contenido remoto.
			WithGlobalTauri: false,
			Windows: []tauriWindow{{
				Title:     cfg.Nombre,
				Width:     cfg.Ancho,
				Height:    cfg.Alto,
				Resizable: cfg.Redimensionable,
			}},
			Security: tauriSecurity{CSP: nil},
		},
		Bundle: tauriBundle{
			Active:           true,
			Targets:          objetivosBundle(runtime.GOOS),
			Category:         "Utility",
			ShortDescription: "Aplicación de escritorio generada con Skynative",
			Icon: []string{
				"icons/32x32.png",
				"icons/128x128.png",
				"icons/128x128@2x.png",
				"icons/icon.icns",
				"icons/icon.ico",
			},
		},
	}
}
