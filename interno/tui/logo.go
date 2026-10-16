package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// El logotipo es una marca de texto dibujada con un tipo de letra propio de 5x5
// celdas, coloreada con un degradado vertical. Se usa en la pantalla de menú
// para presentar la herramienta como una gran cabecera centrada.

const (
	alturaLogo = 5
	anchoLetra = 5
)

// coloresLogo define el degradado vertical nativo (de rosa a violeta).
var coloresLogo = [alturaLogo]string{"212", "176", "135", "99", "63"}

// logoFuente dibuja cada carácter en una rejilla de 5x5.
var logoFuente = map[rune][alturaLogo]string{
	'S': {"█████", "█    ", "█████", "    █", "█████"},
	'K': {"█   █", "█  █ ", "███  ", "█  █ ", "█   █"},
	'Y': {"█   █", " █ █ ", "  █  ", "  █  ", "  █  "},
	'N': {"█   █", "██  █", "█ █ █", "█  ██", "█   █"},
	'A': {" ███ ", "█   █", "█████", "█   █", "█   █"},
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'I': {"█████", "  █  ", "  █  ", "  █  ", "█████"},
	'V': {"█   █", "█   █", "█   █", " █ █ ", "  █  "},
	'E': {"█████", "█    ", "████ ", "█    ", "█████"},
}

// glifoVacio se usa para el espacio y los caracteres desconocidos.
var glifoVacio = [alturaLogo]string{"     ", "     ", "     ", "     ", "     "}

// anchoLogo devuelve el ancho en columnas del logotipo para un texto dado.
func anchoLogo(texto string) int {
	letras := len([]rune(texto))
	if letras == 0 {
		return 0
	}
	return letras*anchoLetra + (letras - 1)
}

// dibujarLogo compone el logotipo con degradado vertical.
func dibujarLogo(texto string) string {
	filas := make([]string, alturaLogo)
	for f := 0; f < alturaLogo; f++ {
		var partes []string
		for _, r := range texto {
			glifo, ok := logoFuente[r]
			if !ok {
				glifo = glifoVacio
			}
			partes = append(partes, glifo[f])
		}
		filas[f] = strings.Join(partes, " ")
	}
	for f, fila := range filas {
		estilo := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(coloresLogo[f]))
		filas[f] = estilo.Render(fila)
	}
	return strings.Join(filas, "\n")
}
