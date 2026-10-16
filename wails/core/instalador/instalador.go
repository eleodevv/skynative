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
	// IconoOrigen es la ruta de un PNG que se usará como icono en Linux.
	IconoOrigen string
	// IconoICO es la ruta de un .ico para el acceso directo de Windows.
	IconoICO string
	// BundleApp es la ruta de un paquete .app (solo macOS).
	BundleApp string
	// DirectorioDatos permite sobrescribir la base XDG (útil en pruebas).
	DirectorioDatos string
	// DirectorioMenu permite sobrescribir el menú Inicio (útil en pruebas).
	DirectorioMenu string
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
	case "windows":
		return instalarWindows(op)
	default:
		return Resultado{}, fmt.Errorf("sistema operativo no soportado: %s", runtime.GOOS)
	}
}

// Desinstalar elimina la entrada creada por Instalar en el sistema. Es
// tolerante con lo que ya no exista y no devuelve error si no había nada que
// borrar. Devuelve las rutas que se eliminaron.
func Desinstalar(op Opciones) ([]string, error) {
	switch runtime.GOOS {
	case "linux":
		return desinstalarLinux(op)
	case "darwin":
		return desinstalarDarwin(op)
	case "windows":
		return desinstalarWindows(op)
	default:
		return nil, nil
	}
}

// desinstalarLinux borra la entrada .desktop y el icono instalados.
func desinstalarLinux(op Opciones) ([]string, error) {
	slug := limpiarSlug(op.Slug)
	if slug == "" {
		return nil, nil
	}
	base, err := directorioDatos(op.DirectorioDatos)
	if err != nil {
		return nil, err
	}
	var eliminados []string
	borrar := func(ruta string) {
		if err := os.Remove(ruta); err == nil {
			eliminados = append(eliminados, ruta)
		}
	}
	borrar(filepath.Join(base, "applications", slug+".desktop"))
	borrar(filepath.Join(base, "icons", "hicolor", "256x256", "apps", slug+".png"))
	octx := filepath.Join(base, "icons", "hicolor")
	if _, err := os.Stat(octx); err == nil {
		actualizarBaseDatos(filepath.Join(base, "applications"))
	}
	return eliminados, nil
}

// desinstalarDarwin borra el paquete .app copiado a ~/Applications.
func desinstalarDarwin(op Opciones) ([]string, error) {
	nombre := strings.TrimSpace(op.Nombre)
	if nombre == "" {
		return nil, nil
	}
	inicio, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	destino := filepath.Join(inicio, "Applications", filepath.Base(nombre)+".app")
	if _, err := os.Stat(destino); err != nil {
		return nil, nil
	}
	if err := os.RemoveAll(destino); err != nil {
		return nil, fmt.Errorf("no se pudo eliminar %q: %w", destino, err)
	}
	return []string{destino}, nil
}

// desinstalarWindows borra el acceso directo del menú Inicio.
func desinstalarWindows(op Opciones) ([]string, error) {
	dir, err := directorioMenuInicio(op.DirectorioMenu)
	if err != nil {
		return nil, nil
	}
	nombre := nombreArchivoWindows(op.Nombre)
	if nombre == "" {
		nombre = limpiarSlug(op.Slug)
	}
	if nombre == "" {
		return nil, nil
	}
	ruta := filepath.Join(dir, nombre+".lnk")
	if _, err := os.Stat(ruta); err != nil {
		return nil, nil
	}
	if err := os.Remove(ruta); err != nil {
		return nil, fmt.Errorf("no se pudo eliminar %q: %w", ruta, err)
	}
	return []string{ruta}, nil
}

// instalarWindows crea un acceso directo en el menú Inicio mediante PowerShell.
func instalarWindows(op Opciones) (Resultado, error) {
	if strings.TrimSpace(op.Ejecutable) == "" {
		return Resultado{}, fmt.Errorf("no se encontró el ejecutable compilado")
	}
	op.Slug = limpiarSlug(op.Slug)
	if op.Slug == "" {
		return Resultado{}, fmt.Errorf("el identificador de la aplicación está vacío")
	}

	dir, err := directorioMenuInicio(op.DirectorioMenu)
	if err != nil {
		return Resultado{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Resultado{}, fmt.Errorf("no se pudo crear %q: %w", dir, err)
	}

	nombre := nombreArchivoWindows(op.Nombre)
	if nombre == "" {
		nombre = op.Slug
	}
	res := Resultado{
		Entrada:    filepath.Join(dir, nombre+".lnk"),
		Ejecutable: op.Ejecutable,
		Mensaje:    "Aplicación instalada en el menú Inicio. Búscala y ábrela con doble clic.",
	}

	potencia, err := buscarPowerShell()
	if err != nil {
		return Resultado{}, err
	}
	script := scriptAccesoDirecto(op, res.Entrada)
	if salida, err := exec.Command(potencia, "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput(); err != nil {
		return Resultado{}, fmt.Errorf("no se pudo crear el acceso directo: %v: %s", err, strings.TrimSpace(string(salida)))
	}

	if _, err := os.Stat(res.Entrada); err != nil {
		return Resultado{}, fmt.Errorf("el acceso directo no se creó: %w", err)
	}
	return res, nil
}

// scriptAccesoDirecto construye el guion de PowerShell que crea el .lnk.
func scriptAccesoDirecto(op Opciones, destino string) string {
	citar := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

	var b strings.Builder
	b.WriteString("$ws = New-Object -ComObject WScript.Shell;")
	b.WriteString("$s = $ws.CreateShortcut(" + citar(destino) + ");")
	b.WriteString("$s.TargetPath = " + citar(op.Ejecutable) + ";")
	if dir := filepath.Dir(op.Ejecutable); dir != "" {
		b.WriteString("$s.WorkingDirectory = " + citar(dir) + ";")
	}
	if op.IconoICO != "" {
		b.WriteString("$s.IconLocation = " + citar(op.IconoICO) + ";")
	}
	if op.Nombre != "" {
		b.WriteString("$s.Description = " + citar(op.Nombre) + ";")
	}
	b.WriteString("$s.Save()")
	return b.String()
}

// buscarPowerShell localiza el intérprete de PowerShell disponible.
func buscarPowerShell() (string, error) {
	for _, nombre := range []string{"powershell", "powershell.exe", "pwsh", "pwsh.exe"} {
		if ruta, err := exec.LookPath(nombre); err == nil {
			return ruta, nil
		}
	}
	return "", fmt.Errorf("no se encontró PowerShell para crear el acceso directo")
}

// directorioMenuInicio resuelve la carpeta del menú Inicio del usuario.
func directorioMenuInicio(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return override, nil
	}
	if appdata := strings.TrimSpace(os.Getenv("APPDATA")); appdata != "" {
		return filepath.Join(appdata, "Microsoft", "Windows", "Start Menu", "Programs"), nil
	}
	return "", fmt.Errorf("no se pudo localizar el menú Inicio (variable APPDATA no definida)")
}

// nombreArchivoWindows elimina los caracteres no válidos en un nombre de archivo.
func nombreArchivoWindows(nombre string) string {
	limpiar := strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return -1
		}
		if r < 32 {
			return -1
		}
		return r
	}, nombre)
	return strings.TrimSpace(limpiar)
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
