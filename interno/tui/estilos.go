package tui

import "github.com/charmbracelet/lipgloss"

// Paleta de estilos de la interfaz. Se usa lipgloss para el color y el formato.
var (
	estiloTitulo    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	estiloSubtitulo = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	estiloEtiqueta  = lipgloss.NewStyle().Foreground(lipgloss.Color("117"))
	estiloAyuda     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	estiloError     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	estiloCorrecto  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	estiloAviso     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	estiloFoco      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	estiloSeleccion = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("99"))
	estiloRuta      = lipgloss.NewStyle().Foreground(lipgloss.Color("150"))
	estiloCaja      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	estiloMarca     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	estiloPanel     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 3)
	estiloEtiquetaCampo = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
)

// casilla dibuja una casilla de verificación.
func casilla(activa bool) string {
	if activa {
		return estiloCorrecto.Render("[x]")
	}
	return estiloAyuda.Render("[ ]")
}
