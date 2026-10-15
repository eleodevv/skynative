package tui

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/skynative/skynative/interno/compilador"
	"github.com/skynative/skynative/interno/configuracion"
	"github.com/skynative/skynative/interno/diagnostico"
	"github.com/skynative/skynative/interno/generador"
	"github.com/skynative/skynative/interno/herramientas"
	"github.com/skynative/skynative/interno/instalador"
	"github.com/skynative/skynative/interno/proyectos"
)

// Mensajes internos de Bubble Tea.
type (
	msgGenerado struct {
		res generador.Resultado
		err error
	}
	msgLineaLog   string
	msgLogCerrado struct{}
	msgCompilado  struct {
		res compilador.Resultado
		err error
	}
	msgInstalado struct {
		res instalador.Resultado
		err error
	}
	msgChequeoHerramientas    []diagnostico.Estado
	msgHerramientasInstaladas struct{ err error }
	msgDiagnostico            []diagnostico.Estado
	msgProyectos              []proyectos.Proyecto
	msgError                  struct{ err error }
)

// generarCmd crea el proyecto en disco y lo registra si todo va bien.
func generarCmd(cfg configuracion.Configuracion) tea.Cmd {
	return func() tea.Msg {
		res, err := generador.Generar(cfg)
		if err != nil {
			return msgGenerado{err: err}
		}
		// El registro es best-effort: no debe impedir el resultado principal.
		_ = proyectos.Agregar(proyectos.Proyecto{
			Nombre:        cfg.Nombre,
			Identificador: res.Identificador,
			URL:           cfg.URL,
			Ruta:          res.Directorio,
		})
		return msgGenerado{res: res}
	}
}

// compilarCmd lanza la compilación de Tauri y escribe la salida en el canal.
func compilarCmd(ctx context.Context, ruta string, canal chan string) tea.Cmd {
	return func() tea.Msg {
		escritor := nuevoEscritorCanal(canal)
		res, err := compilador.Compilar(ctx, ruta, compilador.Opciones{}, escritor)
		escritor.cerrar()
		close(canal)
		return msgCompilado{res: res, err: err}
	}
}

// instalarCmd registra la aplicación compilada en el menú del sistema para que
// pueda abrirse con doble clic.
func instalarCmd(op instalador.Opciones) tea.Cmd {
	return func() tea.Msg {
		res, err := instalador.Instalar(op)
		return msgInstalado{res: res, err: err}
	}
}

// detectarEjecutable localiza el binario principal compilado del proyecto.
func detectarEjecutable(cfg configuracion.Configuracion, rutaProyecto string, artefactos []string) string {
	if rutaProyecto == "" || cfg.NombrePaquete == "" {
		return ""
	}
	nombre := cfg.NombrePaquete
	if runtime.GOOS == "windows" {
		nombre += ".exe"
	}
	candidato := filepath.Join(rutaProyecto, "src-tauri", "target", "release", nombre)
	if info, err := os.Stat(candidato); err == nil && !info.IsDir() {
		return candidato
	}
	// Alternativa: un AppImage ya ejecutable.
	for _, a := range artefactos {
		if strings.EqualFold(filepath.Ext(a), ".appimage") {
			return a
		}
	}
	return ""
}

// detectarBundle localiza un paquete .app de macOS entre los artefactos.
func detectarBundle(artefactos []string) string {
	for _, a := range artefactos {
		if strings.EqualFold(filepath.Ext(a), ".app") {
			return a
		}
	}
	return ""
}

// escucharLog lee una línea del canal de la compilación.
func escucharLog(canal chan string) tea.Cmd {
	return func() tea.Msg {
		linea, abierto := <-canal
		if !abierto {
			return msgLogCerrado{}
		}
		return msgLineaLog(linea)
	}
}

// diagnosticoCmd ejecuta las comprobaciones del entorno.
func diagnosticoCmd() tea.Cmd {
	return func() tea.Msg {
		return msgDiagnostico(diagnostico.Ejecutar(context.Background()))
	}
}

// chequeoHerramientasCmd comprueba en segundo plano si el sistema tiene lo
// necesario para compilar, sin cambiar de pantalla.
func chequeoHerramientasCmd() tea.Cmd {
	return func() tea.Msg {
		return msgChequeoHerramientas(diagnostico.Ejecutar(context.Background()))
	}
}

// instalarHerramientasCmd instala Rust, la CLI de Tauri y las dependencias del
// sistema, escribiendo el progreso en el canal.
func instalarHerramientasCmd(ctx context.Context, canal chan string) tea.Cmd {
	return func() tea.Msg {
		escritor := nuevoEscritorCanal(canal)
		err := herramientas.Instalar(ctx, herramientas.PlanParaSistema(), escritor)
		escritor.cerrar()
		close(canal)
		return msgHerramientasInstaladas{err: err}
	}
}

// proyectosCmd carga el registro de proyectos.
func proyectosCmd() tea.Cmd {
	return func() tea.Msg {
		lista, err := proyectos.Listar()
		if err != nil {
			return msgError{err: err}
		}
		return msgProyectos(lista)
	}
}

// escritorCanal convierte las escrituras de un io.Writer en líneas de texto que
// se envían a un canal, para poder mostrar la compilación en vivo.
type escritorCanal struct {
	canal chan string
	resto string
}

func nuevoEscritorCanal(canal chan string) *escritorCanal {
	return &escritorCanal{canal: canal}
}

func (e *escritorCanal) Write(p []byte) (int, error) {
	e.resto += string(p)
	for {
		i := strings.IndexByte(e.resto, '\n')
		if i < 0 {
			break
		}
		e.enviar(strings.TrimRight(e.resto[:i], "\r"))
		e.resto = e.resto[i+1:]
	}
	return len(p), nil
}

// enviar no bloquea: si el canal está lleno se descarta la línea para no
// detener la compilación.
func (e *escritorCanal) enviar(linea string) {
	select {
	case e.canal <- linea:
	default:
	}
}

// cerrar envía la última línea incompleta.
func (e *escritorCanal) cerrar() {
	if e.resto != "" {
		e.enviar(e.resto)
		e.resto = ""
	}
}

// abrirEnExplorador abre la carpeta en el gestor de archivos del sistema.
func abrirEnExplorador(ruta string) {
	if ruta == "" {
		return
	}
	var nombre string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		nombre, args = "open", []string{ruta}
	case "windows":
		nombre, args = "explorer", []string{ruta}
	default:
		if _, err := exec.LookPath("xdg-open"); err != nil {
			return
		}
		nombre, args = "xdg-open", []string{ruta}
	}
	_ = exec.Command(nombre, args...).Start()
}
