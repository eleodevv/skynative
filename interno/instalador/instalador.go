// Package instalador registra una aplicación ya compilada en el menú de
// aplicaciones del sistema, de modo que el usuario pueda encontrarla al buscar
// y abrirla con doble clic.
//
// En Linux escribe una entrada .desktop y el icono en las rutas XDG del
// usuario. En macOS copia el paquete .app a ~/Applications. No requiere
// privilegios de administrador: todo se instala en el perfil del usuario.
package instalador

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Opciones describe la aplicación que se va a registrar.
type Opciones struct {
	// Nombre es el nombre visible de la aplicación.
	Nombre string
	// Slug es el identificador de archivo (sin espacios ni barras).
	Slug string
	// Comentario es la descripción breve de la entrada.
	Comentario string
	// Ejecutable es la ruta absoluta del binario compilado (Linux/Windows).
	Ejecutable string
	// IconoOrigen es la ruta de un PNG que se usará como icono.
	IconoOrigen string
	// BundleApp es la ruta de un paquete .app (solo macOS).
	BundleApp string
	// DirectorioDatos permite sobrescribir la base XDG (útil en pruebas).
	DirectorioDatos string
}

// Resultado describe lo que se instaló.
type Resultado struct {
	// Entrada es la ruta del lanzador instalado (.desktop o .app).
	Entrada string
	// Icono es la ruta del icono instalado, si lo hubo.
	Icono string
	// Ejecutable es el binario enlazado por el lanzador.
	Ejecutable string
	// Mensaje es un texto legible para mostrar al usuario.
	Mensaje string
}

// Instalar registra la aplicación en el sistema según la plataforma actual.
func Instalar(op Opciones) (Resultado, error) {
	switch runtime.GOOS {
	case "linux":
		return instalarLinux(op)
	case "darwin":
		return instalarDarwin(op)
	default:
		return Resultado{}, fmt.Errorf(
			"la instalación automática aún no está disponible en %s; "+
				"usa el instalador .msi o .exe que genera Tauri", runtime.GOOS)
	}
}

// instalarLinux crea una entrada .desktop y copia el icono a las rutas XDG.
func instalarLinux(op Opciones) (Resultado, error) {
	if strings.TrimSpace(op.Ejecutable) == "" {
		return Resultado{}, fmt.Errorf("no se encontró el ejecutable compilado")
	}
	op.Slug = limpiarSlug(op.Slug)
	if op.Slug == "" {
		return Resultado{}, fmt.Errorf("el identificador de la aplicación está vacío")
	}

	base, err := directorioDatos(op.DirectorioDatos)
	if err != nil {
		return Resultado{}, err
	}
	dirApps := filepath.Join(base, "applications")
	dirIconos := filepath.Join(base, "icons", "hicolor", "256x256", "apps")
	if err := os.MkdirAll(dirApps, 0o755); err != nil {
		return Resultado{}, fmt.Errorf("no se pudo crear %q: %w", dirApps, err)
	}

	res := Resultado{Ejecutable: op.Ejecutable, Mensaje: "Aplicación instalada. Búscala en tus aplicaciones y ábrela con doble clic."}

	if op.IconoOrigen != "" {
		if err := os.MkdirAll(dirIconos, 0o755); err != nil {
			return Resultado{}, fmt.Errorf("no se pudo crear %q: %w", dirIconos, err)
		}
		res.Icono = filepath.Join(dirIconos, op.Slug+".png")
		if err := copiarArchivo(op.IconoOrigen, res.Icono); err != nil {
			return Resultado{}, err
		}
	}

	res.Entrada = filepath.Join(dirApps, op.Slug+".desktop")
	if err := os.WriteFile(res.Entrada, []byte(contenidoEscritorio(op)), 0o755); err != nil {
		return Resultado{}, fmt.Errorf("no se pudo escribir %q: %w", res.Entrada, err)
	}

	actualizarBaseDatos(dirApps)
	return res, nil
}

// instalarDarwin copia el paquete .app a ~/Applications.
func instalarDarwin(op Opciones) (Resultado, error) {
	if strings.TrimSpace(op.BundleApp) == "" {
		return Resultado{}, fmt.Errorf("no se encontró el paquete .app compilado")
	}
	inicio, err := os.UserHomeDir()
	if err != nil {
		return Resultado{}, err
	}
	dirApps := filepath.Join(inicio, "Applications")
	if err := os.MkdirAll(dirApps, 0o755); err != nil {
		return Resultado{}, fmt.Errorf("no se pudo crear %q: %w", dirApps, err)
	}
	destino := filepath.Join(dirApps, filepath.Base(op.BundleApp))
	_ = os.RemoveAll(destino)
	if err := copiarArbol(op.BundleApp, destino); err != nil {
		return Resultado{}, err
	}
	return Resultado{
		Entrada:    destino,
		Ejecutable: op.Ejecutable,
		Mensaje:    "Aplicación instalada en Aplicaciones. Ábrela con doble clic o desde Launchpad.",
	}, nil
}

// contenidoEscritorio construye el texto de la entrada .desktop.
func contenidoEscritorio(op Opciones) string {
	nombre := strings.TrimSpace(op.Nombre)
	if nombre == "" {
		nombre = op.Slug
	}
	comentario := strings.TrimSpace(op.Comentario)
	if comentario == "" {
		comentario = "Aplicación de escritorio generada con Skynative"
	}
	return fmt.Sprintf(`[Desktop Entry]
Type=Application
Version=1.0
Name=%s
Comment=%s
Exec=%s
Icon=%s
Terminal=false
Categories=Network;Utility;
StartupNotify=true
`, nombre, comentario, comandoEscritorio(op.Ejecutable), op.Slug)
}

// comandoEscritorio cita una ruta para el campo Exec de la especificación
// Desktop Entry, escapando los caracteres reservados.
func comandoEscritorio(ruta string) string {
	reemplazador := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"`", "\\`",
		`$`, `\$`,
	)
	return `"` + reemplazador.Replace(ruta) + `"`
}

// limpiarSlug convierte un nombre en un identificador de archivo seguro.
func limpiarSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	var b strings.Builder
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-.")
}

// directorioDatos resuelve la base de datos XDG del usuario.
func directorioDatos(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return override, nil
	}
	if x := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); x != "" {
		return x, nil
	}
	inicio, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(inicio, ".local", "share"), nil
}

// actualizarBaseDatos refresca la caché de aplicaciones de forma best-effort.
func actualizarBaseDatos(dirApps string) {
	if ruta, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(ruta, dirApps).Run()
	}
}

// copiarArchivo copia un archivo conservando permisos legibles.
func copiarArchivo(origen, destino string) error {
	entrada, err := os.Open(origen)
	if err != nil {
		return fmt.Errorf("no se pudo leer el icono %q: %w", origen, err)
	}
	defer entrada.Close()

	salida, err := os.Create(destino)
	if err != nil {
		return fmt.Errorf("no se pudo crear %q: %w", destino, err)
	}
	defer salida.Close()

	if _, err := io.Copy(salida, entrada); err != nil {
		return fmt.Errorf("no se pudo copiar a %q: %w", destino, err)
	}
	return nil
}

// copiarArbol copia un directorio completo de forma recursiva.
func copiarArbol(origen, destino string) error {
	return filepath.WalkDir(origen, func(ruta string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(origen, ruta)
		if err != nil {
			return err
		}
		objetivo := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(objetivo, 0o755)
		}
		if d.Type()&os.ModeSymlink != 0 {
			enlace, err := os.Readlink(ruta)
			if err != nil {
				return err
			}
			return os.Symlink(enlace, objetivo)
		}
		if err := os.MkdirAll(filepath.Dir(objetivo), 0o755); err != nil {
			return err
		}
		return copiarArchivo(ruta, objetivo)
	})
}
