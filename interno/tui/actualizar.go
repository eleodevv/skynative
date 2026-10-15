package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/skynative/skynative/interno/configuracion"
	"github.com/skynative/skynative/interno/instalador"
	"github.com/skynative/skynative/interno/version"
)

// maxLineasLog limita la memoria dedicada al registro de compilación.
const maxLineasLog = 4000

// Update procesa los mensajes de Bubble Tea.
func (m Modelo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ancho, m.alto = msg.Width, msg.Height
		m.registro.Width = max(20, msg.Width-4)
		m.registro.Height = max(5, msg.Height-8)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.cancelar != nil {
				m.cancelar()
			}
			return m, tea.Quit
		}
		return m.actualizarTecla(msg)

	case msgGenerado:
		return m.actualizarGenerado(msg)
	case msgLineaLog:
		m.lineas = append(m.lineas, string(msg))
		if len(m.lineas) > maxLineasLog {
			m.lineas = m.lineas[len(m.lineas)-maxLineasLog:]
		}
		m.registro.SetContent(strings.Join(m.lineas, "\n"))
		m.registro.GotoBottom()
		return m, escucharLog(m.canalLog)
	case msgLogCerrado:
		return m, nil
	case msgCompilado:
		return m.actualizarCompilado(msg)
	case msgInstalado:
		return m.actualizarInstalado(msg)
	case msgDiagnostico:
		m.diagnosticos = msg
		m.estado = estadoDiagnostico
		return m, nil
	case msgProyectos:
		m.proyectos = msg
		m.estado = estadoLista
		return m, nil
	case msgError:
		m.exito = false
		m.resumen = msg.err.Error()
		m.puedeCompilar = false
		m.estado = estadoResultado
		return m, nil
	}

	// Mensajes genéricos (tics del girador, etc.).
	var cmd tea.Cmd
	m.girador, cmd = m.girador.Update(msg)
	m.registro, _ = m.registro.Update(msg)
	return m, cmd
}

// actualizarTecla enruta las teclas según la pantalla activa.
func (m Modelo) actualizarTecla(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.estado {
	case estadoMenu:
		return m.actualizarMenu(msg)
	case estadoFormulario:
		return m.actualizarFormulario(msg)
	case estadoExplorador:
		return m.actualizarExplorador(msg)
	case estadoProgreso:
		return m.actualizarProgreso(msg)
	case estadoResultado:
		return m.actualizarResultado(msg)
	case estadoLista, estadoDiagnostico:
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" {
			m.estado = estadoMenu
		}
		return m, nil
	}
	return m, nil
}

// actualizarMenu gestiona el menú principal.
func (m Modelo) actualizarMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.opcionMenu = (m.opcionMenu - 1 + len(opcionesMenu)) % len(opcionesMenu)
	case "down", "j":
		m.opcionMenu = (m.opcionMenu + 1) % len(opcionesMenu)
	case "q":
		return m, tea.Quit
	case "1", "2", "3", "4", "5":
		m.opcionMenu = int(msg.String()[0] - '1')
		return m.ejecutarOpcionMenu()
	case "enter":
		return m.ejecutarOpcionMenu()
	}
	return m, nil
}

// ejecutarOpcionMenu abre la pantalla de la opción seleccionada.
func (m Modelo) ejecutarOpcionMenu() (tea.Model, tea.Cmd) {
	switch m.opcionMenu {
	case 0:
		m.estado = estadoFormulario
		m.mensajeError = ""
		m.enfocar()
		return m, textinputBlink()
	case 1:
		return m, proyectosCmd()
	case 2:
		return m, diagnosticoCmd()
	case 3:
		m.exito = true
		m.puedeCompilar = false
		m.rutaProyecto = ""
		m.artefactos = nil
		m.resumen = "Skynative " + version.Cadena()
		m.estado = estadoResultado
		return m, nil
	default:
		return m, tea.Quit
	}
}

// actualizarFormulario gestiona el formulario de creación.
func (m Modelo) actualizarFormulario(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.estado = estadoMenu
		return m, nil
	case "tab", "down":
		m.campo = (m.campo + 1) % numCampos
		m.mensajeError = ""
		m.enfocar()
		return m, nil
	case "shift+tab", "up":
		m.campo = (m.campo - 1 + numCampos) % numCampos
		m.mensajeError = ""
		m.enfocar()
		return m, nil
	case "enter":
		switch m.campo {
		case campoCarpeta:
			m.iniciarExplorador()
			return m, nil
		case campoRedimensionable:
			m.redimensionable = !m.redimensionable
			return m, nil
		case campoCrear:
			return m.lanzarGeneracion()
		default:
			m.campo = (m.campo + 1) % numCampos
			m.enfocar()
			return m, nil
		}
	case " ":
		if m.campo == campoRedimensionable {
			m.redimensionable = !m.redimensionable
			return m, nil
		}
	case "ctrl+g":
		return m.lanzarGeneracion()
	}

	// El resto de teclas van al campo de texto enfocado.
	var cmd tea.Cmd
	switch m.campo {
	case campoURL:
		m.url, cmd = m.url.Update(msg)
	case campoNombre:
		m.nombre, cmd = m.nombre.Update(msg)
	case campoAncho:
		m.anchoCampo, cmd = m.anchoCampo.Update(msg)
	case campoAlto:
		m.altoCampo, cmd = m.altoCampo.Update(msg)
	}
	return m, cmd
}

// iniciarExplorador abre el explorador de carpetas integrado.
func (m *Modelo) iniciarExplorador() {
	m.explorador = nuevoExplorador(m.carpeta)
	m.estado = estadoExplorador
}

// actualizarExplorador gestiona el explorador de carpetas.
func (m Modelo) actualizarExplorador(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.estado = estadoFormulario
		m.enfocar()
	case "up", "k":
		m.explorador.mover(-1)
	case "down", "j":
		m.explorador.mover(1)
	case "backspace", "h", "left":
		m.explorador.subir()
	case "right", "l":
		if entrada, ok := m.explorador.seleccionado(); ok && entrada.accion == accionEntrar {
			m.explorador.entrar(entrada.ruta)
		}
	case "home":
		if inicio, err := ubicacionInicio(); err == nil {
			m.explorador.entrar(inicio)
		}
	case "s":
		m.carpeta = m.explorador.directorio
		m.estado = estadoFormulario
		m.enfocar()
	case "enter":
		entrada, ok := m.explorador.seleccionado()
		if !ok {
			return m, nil
		}
		switch entrada.accion {
		case accionSeleccionar:
			m.carpeta = entrada.ruta
			m.estado = estadoFormulario
			m.enfocar()
		case accionSubir:
			m.explorador.subir()
		default:
			m.explorador.entrar(entrada.ruta)
		}
	}
	return m, nil
}

// actualizarProgreso gestiona la pantalla de progreso (generación/compilación).
func (m Modelo) actualizarProgreso(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if m.cancelar != nil {
			m.cancelar()
			m.cancelar = nil
		}
		if m.compilandoAhora {
			m.exito = false
			m.resumen = "Compilación cancelada por el usuario."
			m.puedeCompilar = false
			m.estado = estadoResultado
			m.trabajando = false
			m.compilandoAhora = false
			return m, nil
		}
		if m.instalandoAhora {
			m.trabajando = false
			m.instalandoAhora = false
			m.estado = estadoResultado
		}
		return m, nil
	}
	// Permitir desplazar el registro.
	m.registro, _ = m.registro.Update(msg)
	return m, nil
}

// actualizarResultado gestiona la pantalla de resultado.
func (m Modelo) actualizarResultado(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.estado = estadoMenu
	case "enter", "m":
		m.estado = estadoFormulario
		m.mensajeError = ""
		m.enfocar()
		return m, textinputBlink()
	case "o":
		abrirEnExplorador(m.rutaProyecto)
	case "c":
		if m.exito && m.puedeCompilar && m.rutaProyecto != "" {
			return m.iniciarCompilacion()
		}
	case "i":
		if m.exito && m.puedeInstalar && !m.instalado && m.rutaProyecto != "" {
			return m.iniciarInstalacion()
		}
	}
	return m, nil
}

// lanzarGeneracion valida el formulario y arranca la creación del proyecto.
func (m Modelo) lanzarGeneracion() (Modelo, tea.Cmd) {
	cfg := configuracion.Nueva(m.url.Value(), m.nombre.Value())
	cfg.DirectorioSalida = m.carpeta
	cfg.Redimensionable = m.redimensionable

	ancho, err := enteroPositivo(m.anchoCampo.Value(), configuracion.AnchoPredeterminado)
	if err != nil {
		m.mensajeError = "El ancho no es válido: " + err.Error()
		m.campo = campoAncho
		m.enfocar()
		return m, nil
	}
	alto, err := enteroPositivo(m.altoCampo.Value(), configuracion.AltoPredeterminado)
	if err != nil {
		m.mensajeError = "El alto no es válido: " + err.Error()
		m.campo = campoAlto
		m.enfocar()
		return m, nil
	}
	cfg.Ancho = ancho
	cfg.Alto = alto

	cfg.Normalizar()
	if err := cfg.Validar(); err != nil {
		m.mensajeError = err.Error()
		return m, nil
	}

	m.mensajeError = ""
	m.cfg = cfg
	m.rutaEjecutable = ""
	m.bundleApp = ""
	m.puedeInstalar = false
	m.instalado = false
	m.entradaEscritorio = ""
	m.rutaProyecto = filepath.Join(cfg.DirectorioSalida, cfg.NombrePaquete)
	m.estado = estadoProgreso
	m.trabajando = true
	m.compilandoAhora = false
	m.lineas = []string{"Generando el proyecto Tauri v2 en " + m.rutaProyecto + " ..."}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.girador = nuevoGirador()
	return m, tea.Batch(m.girador.Tick, generarCmd(cfg))
}

// iniciarCompilacion arranca la compilación de Tauri con salida en vivo.
func (m Modelo) iniciarCompilacion() (Modelo, tea.Cmd) {
	ctx, cancelar := context.WithCancel(context.Background())
	m.cancelar = cancelar
	m.canalLog = make(chan string, 4096)
	m.lineas = []string{"Compilando " + m.rutaProyecto + " con Tauri..."}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.estado = estadoProgreso
	m.trabajando = true
	m.compilandoAhora = true
	m.instalandoAhora = false
	m.instalado = false
	m.entradaEscritorio = ""
	m.puedeInstalar = false
	m.girador = nuevoGirador()
	return m, tea.Batch(
		m.girador.Tick,
		escucharLog(m.canalLog),
		compilarCmd(ctx, m.rutaProyecto, m.canalLog),
	)
}

// iniciarInstalacion registra la aplicación compilada en el sistema.
func (m Modelo) iniciarInstalacion() (Modelo, tea.Cmd) {
	m.estado = estadoProgreso
	m.trabajando = true
	m.instalandoAhora = true
	m.compilandoAhora = false
	m.lineas = []string{"Instalando «" + m.cfg.Nombre + "» en el menú de aplicaciones…"}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.girador = nuevoGirador()

	op := instalador.Opciones{
		Nombre:      m.cfg.Nombre,
		Slug:        m.cfg.NombrePaquete,
		Comentario:  "Aplicación de escritorio creada con Skynative",
		Ejecutable:  m.rutaEjecutable,
		IconoOrigen: filepath.Join(m.rutaProyecto, "src-tauri", "icons", "128x128.png"),
		BundleApp:   m.bundleApp,
	}
	return m, tea.Batch(m.girador.Tick, instalarCmd(op))
}

// actualizarGenerado procesa el fin de la generación.
func (m Modelo) actualizarGenerado(msg msgGenerado) (tea.Model, tea.Cmd) {
	m.trabajando = false
	if msg.err != nil {
		m.exito = false
		m.puedeCompilar = false
		m.resumen = "No se pudo generar el proyecto: " + msg.err.Error()
		m.estado = estadoResultado
		return m, nil
	}
	m.exito = true
	m.puedeCompilar = true
	m.rutaProyecto = msg.res.Directorio
	m.artefactos = nil
	m.resumen = fmt.Sprintf("Proyecto «%s» generado correctamente.", msg.res.NombrePaquete)
	m.estado = estadoResultado
	return m, nil
}

// actualizarCompilado procesa el fin de la compilación.
func (m Modelo) actualizarCompilado(msg msgCompilado) (tea.Model, tea.Cmd) {
	m.trabajando = false
	m.compilandoAhora = false
	if m.cancelar != nil {
		m.cancelar()
		m.cancelar = nil
	}
	if msg.err != nil {
		m.exito = false
		m.resumen = "La compilación falló: " + msg.err.Error()
		m.puedeCompilar = true
		m.estado = estadoResultado
		return m, nil
	}
	m.exito = true
	m.puedeCompilar = false
	m.artefactos = msg.res.Artefactos
	m.rutaEjecutable = detectarEjecutable(m.cfg, m.rutaProyecto, msg.res.Artefactos)
	m.bundleApp = detectarBundle(msg.res.Artefactos)
	m.puedeInstalar = m.rutaEjecutable != "" || m.bundleApp != ""
	if len(msg.res.Artefactos) == 0 {
		m.resumen = "Compilación terminada, pero no se encontraron artefactos."
	} else {
		m.resumen = fmt.Sprintf("Compilación terminada. %d artefacto(s) generados.", len(msg.res.Artefactos))
	}
	m.estado = estadoResultado
	return m, nil
}

// actualizarInstalado procesa el fin de la instalación en el sistema.
func (m Modelo) actualizarInstalado(msg msgInstalado) (tea.Model, tea.Cmd) {
	m.trabajando = false
	m.instalandoAhora = false
	if msg.err != nil {
		m.exito = false
		m.puedeInstalar = m.rutaEjecutable != "" || m.bundleApp != ""
		m.resumen = "No se pudo instalar la aplicación: " + msg.err.Error()
		m.estado = estadoResultado
		return m, nil
	}
	m.exito = true
	m.instalado = true
	m.puedeInstalar = false
	m.entradaEscritorio = msg.res.Entrada
	if msg.res.Mensaje != "" {
		m.resumen = msg.res.Mensaje
	} else {
		m.resumen = "Aplicación instalada. Búscala en tus aplicaciones y ábrela con doble clic."
	}
	m.estado = estadoResultado
	return m, nil
}

// enteroPositivo interpreta un campo numérico opcional.
func enteroPositivo(valor string, porDefecto int) (int, error) {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return porDefecto, nil
	}
	n, err := strconv.Atoi(valor)
	if err != nil {
		return 0, fmt.Errorf("debe ser un número entero")
	}
	return n, nil
}
