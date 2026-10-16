package tui

import "github.com/charmbracelet/lipgloss"

// Paleta nativa de la interfaz. Se usan colores estándar de 256 tonos y ningún
// fondo pintado, para respetar el tema (claro u oscuro) de la terminal y evitar
// que las zonas sin texto se vean deslavadas.
var (
	estiloTitulo    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	estiloSubtitulo = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	estiloEtiqueta  = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	estiloAyuda     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	estiloError     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	estiloCorrecto  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	estiloAviso     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	estiloFoco      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	estiloSeleccion = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	estiloRuta      = lipgloss.NewStyle().Foreground(lipgloss.Color("150"))
	estiloMarca     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	estiloCaja      = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)
	estiloPanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 3)
	estiloEtiquetaCampo = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))

	// Botones con borde para las filas de acciones y las confirmaciones.
	estiloBoton = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("252")).
			Padding(0, 2)
	estiloBotonActivo = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("42")).
				Foreground(lipgloss.Color("231")).
				Bold(true).
				Padding(0, 2)
	estiloBotonPeligro = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Foreground(lipgloss.Color("252")).
				Padding(0, 2)

	// Filas-botón del menú: sin fondo, solo color y marca de selección.
	estiloFilaActiva   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	estiloFilaInactiva = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

// casilla dibuja una casilla de verificación.
func casilla(activa bool) string {
	if activa {
		return estiloCorrecto.Render("[x]")
	}
	return estiloAyuda.Render("[ ]")
}

// botonLinea dibuja una fila-botón del menú, resaltando la opción activa.
func botonLinea(texto string, activo bool, ancho int) string {
	if activo {
		return estiloFilaActiva.Width(ancho).Render(" ❯ " + texto)
	}
	return estiloFilaInactiva.Width(ancho).Render("   " + texto)
}
