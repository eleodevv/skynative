package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/skynative/skynative/interno/compilador"
	"github.com/skynative/skynative/interno/configuracion"
	"github.com/skynative/skynative/interno/diagnostico"
	"github.com/skynative/skynative/interno/herramientas"
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

	case tea.MouseMsg:
		return m.actualizarRaton(msg)

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
	case msgChequeoHerramientas:
		return m.actualizarChequeoHerramientas(msg)
	case msgHerramientasInstaladas:
		return m.actualizarHerramientasInstaladas(msg)
	case msgSudoElevado:
		m.trabajando = false
		if msg.err != nil {
			m.exito = false
			m.resumen = "No se pudieron obtener permisos de administrador: " + msg.err.Error()
			m.estado = estadoResultado
			return m, nil
		}
		return m.iniciarInstalacionHerramientas()
	case msgDiagnostico:
		m.diagnosticos = msg
		m.estado = estadoDiagnostico
		return m, nil
	case msgProyectos:
		m.proyectos = msg
		if m.indiceProyecto >= len(m.proyectos) {
			m.indiceProyecto = max(0, len(m.proyectos)-1)
		}
		m.confirmandoEliminar = false
		m.estado = estadoLista
		return m, nil
	case msgProyectoEliminado:
		if msg.err != nil {
			m.exito = false
			m.resumen = "No se pudo eliminar: " + msg.err.Error()
			m.estado = estadoResultado
			return m, nil
		}
		m.confirmandoEliminarTodo = false
		return m, proyectosCmd()
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
	case estadoLista:
		return m.actualizarLista(msg)
	case estadoDiagnostico:
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" {
			m.estado = estadoMenu
		}
		return m, nil
	}
	return m, nil
}

// actualizarLista gestiona la pantalla de proyectos: navegación y borrado del
// registro.
func (m Modelo) actualizarLista(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmandoEliminar {
		switch msg.String() {
		case "s", "y", "enter":
			m.confirmandoEliminar = false
			if m.indiceProyecto < 0 || m.indiceProyecto >= len(m.proyectos) {
				return m, nil
			}
			return m, eliminarProyectoCmd(m.proyectos[m.indiceProyecto])
		case "n", "esc", "q":
			m.confirmandoEliminar = false
		}
		return m, nil
	}

	if m.confirmandoEliminarTodo {
		switch msg.String() {
		case "s", "y", "enter":
			m.confirmandoEliminarTodo = false
			if len(m.proyectos) == 0 {
				return m, nil
			}
			return m, eliminarTodoCmd(m.proyectos)
		case "n", "esc", "q":
			m.confirmandoEliminarTodo = false
		}
		return m, nil
	}

	switch msg.String() {
	case "esc", "q":
		m.estado = estadoMenu
	case "up", "k":
		if m.indiceProyecto > 0 {
			m.indiceProyecto--
		}
	case "down", "j":
		if m.indiceProyecto < len(m.proyectos)-1 {
			m.indiceProyecto++
		}
	case "d", "delete":
		if len(m.proyectos) > 0 {
			m.confirmandoEliminar = true
		}
	case "a", "A":
		if len(m.proyectos) > 0 {
			m.confirmandoEliminarTodo = true
		}
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
	case "1", "2", "3", "4", "5", "6":
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
		if herramientas.NecesitaSudo() {
			m.estado = estadoProgreso
			m.trabajando = true
			m.detalle = nil
			m.reiniciarProgreso(estimadoSudo)
			m.lineas = []string{"Solicitando permisos de administrador (sudo)…"}
			m.registro.SetContent(strings.Join(m.lineas, "\n"))
			return m, tea.ExecProcess(herramientas.ComandoElevarPrivilegios(), func(err error) tea.Msg {
				return msgSudoElevado{err: err}
			})
		}
		return m.iniciarInstalacionHerramientas()
	case 2:
		return m, proyectosCmd()
	case 3:
		return m, diagnosticoCmd()
	case 4:
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
		case campoEmpaquetar:
			m.empaquetar = !m.empaquetar
			return m, nil
		case campoCrear:
			return m.lanzarGeneracion()
		default:
			m.campo = (m.campo + 1) % numCampos
			m.enfocar()
			return m, nil
		}
	case " ":
		switch m.campo {
		case campoRedimensionable:
			m.redimensionable = !m.redimensionable
			return m, nil
		case campoEmpaquetar:
			m.empaquetar = !m.empaquetar
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
		return m.exploradorEnter()
	}
	return m, nil
}

// exploradorEnter entra en la carpeta seleccionada o la elige como salida.
func (m Modelo) exploradorEnter() (tea.Model, tea.Cmd) {
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
		if m.instalandoHerramientas {
			m.trabajando = false
			m.instalandoHerramientas = false
			m.exito = false
			m.resumen = "Instalación de herramientas cancelada por el usuario."
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
	m.notas = nil
	m.estado = estadoProgreso
	m.trabajando = true
	m.compilandoAhora = false
	m.reiniciarProgreso(estimadoGeneracion)
	m.lineas = []string{
		"Buscando el icono oficial del sitio…",
		"Generando el proyecto Tauri v2 en " + m.rutaProyecto + " ...",
	}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.girador = nuevoGirador()
	return m, tea.Batch(m.girador.Tick, generarCmd(cfg))
}

// iniciarCompilacion arranca la compilación de Tauri con salida en vivo.
func (m Modelo) iniciarCompilacion() (Modelo, tea.Cmd) {
	ctx, cancelar := context.WithCancel(context.Background())
	m.cancelar = cancelar
	m.canalLog = make(chan string, 4096)
	if m.empaquetar {
		m.lineas = []string{"Compilando " + m.rutaProyecto + " con Tauri (con empaquetado)…"}
	} else {
		m.lineas = []string{"Compilando " + m.rutaProyecto + " con Tauri (solo binario, más rápido)…"}
	}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.estado = estadoProgreso
	m.trabajando = true
	m.compilandoAhora = true
	m.instalandoAhora = false
	m.instalado = false
	m.entradaEscritorio = ""
	m.puedeInstalar = false
	if m.empaquetar {
		m.reiniciarProgreso(estimadoEmpaquetar)
	} else {
		m.reiniciarProgreso(estimadoCompilar)
	}
	m.girador = nuevoGirador()
	return m, tea.Batch(
		m.girador.Tick,
		escucharLog(m.canalLog),
		compilarCmd(ctx, m.rutaProyecto, compilador.Opciones{SinBundle: !m.empaquetar}, m.canalLog),
	)
}

// iniciarInstalacion registra la aplicación compilada en el sistema.
func (m Modelo) iniciarInstalacion() (Modelo, tea.Cmd) {
	m.estado = estadoProgreso
	m.trabajando = true
	m.instalandoAhora = true
	m.compilandoAhora = false
	m.reiniciarProgreso(estimadoInstalar)
	m.lineas = []string{"Instalando «" + m.cfg.Nombre + "» en el menú de aplicaciones…"}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.girador = nuevoGirador()

	op := instalador.Opciones{
		Nombre:      m.cfg.Nombre,
		Slug:        m.cfg.NombrePaquete,
		Comentario:  "Aplicación de escritorio creada con Skynative",
		Ejecutable:  m.rutaEjecutable,
		IconoOrigen: filepath.Join(m.rutaProyecto, "src-tauri", "icons", "128x128.png"),
		IconoICO:    filepath.Join(m.rutaProyecto, "src-tauri", "icons", "icon.ico"),
		BundleApp:   m.bundleApp,
	}
	return m, tea.Batch(m.girador.Tick, instalarCmd(op))
}

// iniciarInstalacionHerramientas instala las dependencias externas (Rust, CLI
// de Tauri y bibliotecas del sistema) con salida en vivo.
func (m Modelo) iniciarInstalacionHerramientas() (Modelo, tea.Cmd) {
	ctx, cancelar := context.WithCancel(context.Background())
	m.cancelar = cancelar
	m.canalLog = make(chan string, 4096)
	m.lineas = []string{"Instalando las herramientas necesarias (Rust, CLI de Tauri y dependencias)…"}
	m.registro.SetContent(strings.Join(m.lineas, "\n"))
	m.estado = estadoProgreso
	m.trabajando = true
	m.instalandoHerramientas = true
	m.compilandoAhora = false
	m.instalandoAhora = false
	m.reiniciarProgreso(estimadoHerramientas)
	m.girador = nuevoGirador()
	return m, tea.Batch(
		m.girador.Tick,
		escucharLog(m.canalLog),
		instalarHerramientasCmd(ctx, m.canalLog),
	)
}

// actualizarGenerado procesa el fin de la generación.
func (m Modelo) actualizarGenerado(msg msgGenerado) (tea.Model, tea.Cmd) {
	m.trabajando = false
	m.detalle = nil
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
	if msg.nota != "" {
		m.notas = append(m.notas, msg.nota)
	}
	// Si el entorno está listo, encadenamos la compilación automáticamente.
	if m.herramientasListas {
		m.resumen = fmt.Sprintf("Proyecto «%s» generado. Compilando automáticamente…", msg.res.NombrePaquete)
		return m.iniciarCompilacion()
	}
	m.resumen = fmt.Sprintf("Proyecto «%s» generado correctamente.", msg.res.NombrePaquete)
	m.estado = estadoResultado
	return m, nil
}

// actualizarChequeoHerramientas guarda el resultado de la comprobación en
// segundo plano sin cambiar de pantalla.
func (m Modelo) actualizarChequeoHerramientas(msg msgChequeoHerramientas) (tea.Model, tea.Cmd) {
	m.diagnosticos = []diagnostico.Estado(msg)
	m.chequeoHecho = true
	m.herramientasListas = !diagnostico.HayCriticosFallidos(m.diagnosticos)
	return m, nil
}

// actualizarHerramientasInstaladas procesa el fin de la instalación de las
// dependencias externas y refresca el estado del entorno.
func (m Modelo) actualizarHerramientasInstaladas(msg msgHerramientasInstaladas) (tea.Model, tea.Cmd) {
	m.trabajando = false
	m.instalandoHerramientas = false
	if m.cancelar != nil {
		m.cancelar()
		m.cancelar = nil
	}
	if msg.err != nil {
		m.exito = false
		m.resumen = "No se pudieron instalar todas las herramientas: " + msg.err.Error()
	} else {
		m.exito = true
		m.resumen = "Herramientas instaladas. Ya puedes crear y compilar aplicaciones."
	}
	m.estado = estadoResultado
	return m, chequeoHerramientasCmd()
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
		m.detalle = compilador.UltimasLineas(msg.res.Salida, 14)
		m.puedeCompilar = true
		m.estado = estadoResultado
		return m, nil
	}
	m.exito = true
	m.detalle = nil
	m.puedeCompilar = false
	m.artefactos = msg.res.Artefactos
	m.rutaEjecutable = detectarEjecutable(m.cfg, m.rutaProyecto, msg.res.Artefactos)
	m.bundleApp = detectarBundle(msg.res.Artefactos)
	m.puedeInstalar = m.rutaEjecutable != "" || m.bundleApp != ""
	if len(msg.res.Artefactos) == 0 {
		m.resumen = "Compilación terminada, pero no se encontraron artefactos."
	} else if msg.res.EmpaquetadoOmitido {
		m.resumen = fmt.Sprintf("Compilación terminada sin empaquetado (el empaquetado nativo falló). %d artefacto(s).", len(msg.res.Artefactos))
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
	m.detalle = nil
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
