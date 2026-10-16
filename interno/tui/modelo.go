// Package tui implementa la interfaz de terminal de Skynative con Bubble Tea.
//
// La TUI es la única interfaz de la herramienta: permite crear aplicaciones de
// escritorio, elegir la carpeta de salida con un explorador integrado, compilar
// y ver el diagnóstico del entorno, todo desde la terminal.
package tui

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/skynative/skynative/interno/configuracion"
	"github.com/skynative/skynative/interno/diagnostico"
	"github.com/skynative/skynative/interno/proyectos"
)

// estado identifica la pantalla activa.
type estado int

const (
	estadoMenu estado = iota
	estadoFormulario
	estadoExplorador
	estadoProgreso
	estadoResultado
	estadoLista
	estadoDiagnostico
)

// Campos del formulario de creación.
const (
	campoURL = iota
	campoNombre
	campoAncho
	campoAlto
	campoCarpeta
	campoRedimensionable
	campoEmpaquetar
	campoCrear
	numCampos
)

// opcionesMenu son las entradas del menú principal.
var opcionesMenu = []string{
	"Crear aplicación",
	"Instalar herramientas necesarias",
	"Listar proyectos",
	"Diagnóstico del entorno",
	"Ver versión",
	"Salir",
}

// Modelo es el estado completo de la interfaz.
type Modelo struct {
	ancho  int
	alto   int
	estado estado

	// Menú
	opcionMenu int

	// Formulario
	url             textinput.Model
	nombre          textinput.Model
	anchoCampo      textinput.Model
	altoCampo       textinput.Model
	campo           int
	redimensionable bool
	empaquetar      bool
	carpeta         string
	mensajeError    string

	// Explorador de carpetas integrado
	explorador exploradorModelo

	// Progreso
	girador         spinner.Model
	registro        viewport.Model
	lineas          []string
	trabajando      bool
	compilandoAhora bool
	cancelar        context.CancelFunc
	canalLog        chan string

	// Barra de progreso y tiempo estimado de la fase en curso.
	inicioProgreso   time.Time
	progresoEstimado time.Duration

	// Resultado
	exito         bool
	resumen       string
	detalle       []string
	notas         []string
	rutaProyecto  string
	artefactos    []string
	puedeCompilar bool

	// Instalación en el sistema (entrada de menú + icono)
	cfg               configuracion.Configuracion
	rutaEjecutable    string
	bundleApp         string
	puedeInstalar     bool
	instalandoAhora   bool
	instalado         bool
	entradaEscritorio string

	// Comprobación e instalación de herramientas externas
	herramientasListas     bool
	chequeoHecho           bool
	instalandoHerramientas bool

	// Listas
	diagnosticos            []diagnostico.Estado
	proyectos               []proyectos.Proyecto
	indiceProyecto          int
	confirmandoEliminar     bool
	confirmandoEliminarTodo bool
}

// NuevoModelo construye el modelo inicial.
func NuevoModelo() Modelo {
	url := textinput.New()
	url.Placeholder = "https://ejemplo.com"
	url.CharLimit = 500
	url.Width = 46
	url.Focus()

	nombre := textinput.New()
	nombre.Placeholder = "Mi Aplicación"
	nombre.CharLimit = configuracion.LongitudMaximaNombre
	nombre.Width = 46

	ancho := textinput.New()
	ancho.SetValue(strconv.Itoa(configuracion.AnchoPredeterminado))
	ancho.CharLimit = 5
	ancho.Width = 8

	alto := textinput.New()
	alto.SetValue(strconv.Itoa(configuracion.AltoPredeterminado))
	alto.CharLimit = 5
	alto.Width = 8

	girador := spinner.New()
	girador.Spinner = spinner.Dot
	girador.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	carpeta, err := os.Getwd()
	if err != nil {
		carpeta = "."
	}

	return Modelo{
		estado:          estadoMenu,
		url:             url,
		nombre:          nombre,
		anchoCampo:      ancho,
		altoCampo:       alto,
		campo:           campoURL,
		redimensionable: true,
		carpeta:         carpeta,
		girador:         girador,
		registro:        viewport.New(80, 14),
	}
}

// Init arranca el parpadeo del cursor y comprueba en segundo plano si están
// instaladas las herramientas necesarias.
func (m Modelo) Init() tea.Cmd {
	return tea.Batch(textinputBlink(), chequeoHerramientasCmd())
}

// textinputBlink devuelve el comando de parpadeo del cursor.
func textinputBlink() tea.Cmd {
	return textinput.Blink
}

// nuevoGirador crea un indicador de actividad nuevo.
func nuevoGirador() spinner.Model {
	g := spinner.New()
	g.Spinner = spinner.Dot
	g.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return g
}

// ubicacionInicio devuelve el directorio personal del usuario.
func ubicacionInicio() (string, error) {
	return os.UserHomeDir()
}

// enfocar pone el foco en el campo de texto correcto.
func (m *Modelo) enfocar() {
	m.url.Blur()
	m.nombre.Blur()
	m.anchoCampo.Blur()
	m.altoCampo.Blur()

	switch m.campo {
	case campoURL:
		m.url.Focus()
	case campoNombre:
		m.nombre.Focus()
	case campoAncho:
		m.anchoCampo.Focus()
	case campoAlto:
		m.altoCampo.Focus()
	}
}
