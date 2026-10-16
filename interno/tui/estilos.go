package tui

import "github.com/charmbracelet/lipgloss"

// Paleta «cielo» de la interfaz: azules profundos, nubes claras y acentos de
// cian y tormenta, en sintonía con el nombre de la herramienta.
var (
	colorCielo      = lipgloss.Color("#0b1b34")
	colorCieloClaro = lipgloss.Color("#12294a")
	colorAzul       = lipgloss.Color("#4aa8ff")
	colorAzulClaro  = lipgloss.Color("#a9d6ff")
	colorNube       = lipgloss.Color("#eaf4ff")
	colorCian       = lipgloss.Color("#38d6ff")
	colorTormenta   = lipgloss.Color("#7c8cff")
	colorExito      = lipgloss.Color("#7ef0c0")
	colorAviso      = lipgloss.Color("#ffd166")
	colorError      = lipgloss.Color("#ff6b81")
	colorTenue      = lipgloss.Color("#5b7aa6")
)

// Estilos de la interfaz. Se usa lipgloss para el color y el formato.
var (
	estiloTitulo    = lipgloss.NewStyle().Bold(true).Foreground(colorCian)
	estiloSubtitulo = lipgloss.NewStyle().Foreground(colorAzulClaro)
	estiloEtiqueta  = lipgloss.NewStyle().Foreground(colorAzul)
	estiloAyuda     = lipgloss.NewStyle().Foreground(colorTenue)
	estiloError     = lipgloss.NewStyle().Foreground(colorError)
	estiloCorrecto  = lipgloss.NewStyle().Foreground(colorExito)
	estiloAviso     = lipgloss.NewStyle().Foreground(colorAviso)
	estiloFoco      = lipgloss.NewStyle().Bold(true).Foreground(colorCian)
	estiloSeleccion = lipgloss.NewStyle().Bold(true).Foreground(colorNube).Background(colorAzul)
	estiloRuta      = lipgloss.NewStyle().Foreground(colorCian)
	estiloMarca     = lipgloss.NewStyle().Bold(true).Foreground(colorCian)
	estiloCaja      = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorTenue).
			Padding(0, 1)
	estiloPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAzul).
			Background(colorCielo).
			Padding(1, 3)
	estiloEtiquetaCampo = lipgloss.NewStyle().Foreground(colorAzulClaro)

	// Botones con borde (estilo «New Chat» de la captura): se usan en las filas
	// de acciones y en las confirmaciones.
	estiloBoton        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorTenue).Foreground(colorNube).Padding(0, 2)
	estiloBotonActivo  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorExito).Foreground(colorNube).Bold(true).Padding(0, 2)
	estiloBotonPeligro = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorError).Foreground(colorNube).Padding(0, 2)

	// Filas-botón del menú (una sola línea, con fondo).
	estiloFilaActiva   = lipgloss.NewStyle().Background(colorAzul).Foreground(colorNube).Bold(true)
	estiloFilaInactiva = lipgloss.NewStyle().Background(colorCieloClaro).Foreground(colorAzulClaro)
)

// casilla dibuja una casilla de verificación.
func casilla(activa bool) string {
	if activa {
		return estiloCorrecto.Render("[x]")
	}
	return estiloAyuda.Render("[ ]")
}

// botonLinea dibuja una fila-botón de ancho fijo para el menú.
func botonLinea(texto string, activo bool, ancho int) string {
	if activo {
		return estiloFilaActiva.Width(ancho).Render(" ❯ " + texto)
	}
	return estiloFilaInactiva.Width(ancho).Render("   " + texto)
}
