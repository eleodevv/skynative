package tui

import (
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Ejecutar arranca la interfaz de terminal en pantalla alternativa.
func Ejecutar() error {
	return ejecutarConEntradaSalida(os.Stdin, os.Stdout)
}

// ejecutarConEntradaSalida permite inyectar la entrada y la salida, lo que
// facilita las pruebas sin una terminal real.
func ejecutarConEntradaSalida(entrada io.Reader, salida io.Writer) error {
	programa := tea.NewProgram(
		NuevoModelo(),
		tea.WithAltScreen(),
		tea.WithInput(entrada),
		tea.WithOutput(salida),
	)
	_, err := programa.Run()
	return err
}
