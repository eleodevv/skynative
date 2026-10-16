package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// anchoBarra es el ancho de la barra de progreso, en columnas.
const anchoBarra = 40

// Estimaciones de duración por fase. Son aproximadas: la barra se apoya en el
// tiempo transcurrido y nunca llega al 100 % hasta que la fase termina.
const (
	estimadoGeneracion   = 10 * time.Second
	estimadoCompilar     = 4 * time.Minute
	estimadoEmpaquetar   = 5 * time.Minute
	estimadoInstalar     = 5 * time.Second
	estimadoHerramientas = 5 * time.Minute
	estimadoSudo         = 45 * time.Second
)

// Estilos de la barra: solo color de texto, sin fondos pintados.
var (
	estiloBarraLlena = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	estiloBarraVacia = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	estiloBarraDatos = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// reiniciarProgreso marca el comienzo de una fase y su duración estimada.
func (m *Modelo) reiniciarProgreso(estimado time.Duration) {
	m.inicioProgreso = time.Now()
	m.progresoEstimado = estimado
}

// progresoActual devuelve la fracción completada (0–1) y los tiempos
// transcurrido y restante estimado para la fase en curso.
func (m Modelo) progresoActual() (float64, time.Duration, time.Duration) {
	if m.inicioProgreso.IsZero() || m.progresoEstimado <= 0 {
		return 0, 0, 0
	}
	transcurrido := time.Since(m.inicioProgreso)
	if transcurrido < 0 {
		transcurrido = 0
	}
	restante := m.progresoEstimado - transcurrido
	if restante < 0 {
		restante = 0
	}
	fraccion := float64(transcurrido) / float64(m.progresoEstimado)
	if fraccion < 0 {
		fraccion = 0
	}
	if fraccion > 0.95 {
		// Nunca se llena del todo hasta que la fase termina de verdad.
		fraccion = 0.95
	}
	return fraccion, transcurrido, restante
}

// barraProgreso dibuja una barra determinada con bloques.
func barraProgreso(fraccion float64, ancho int) string {
	if ancho < 10 {
		ancho = 10
	}
	llenos := int(fraccion*float64(ancho) + 0.5)
	if llenos < 0 {
		llenos = 0
	}
	if llenos > ancho {
		llenos = ancho
	}
	return "[" + estiloBarraLlena.Render(strings.Repeat("█", llenos)) +
		estiloBarraVacia.Render(strings.Repeat("░", ancho-llenos)) + "]"
}

// formatearDuracion presenta una duración de forma compacta (p. ej. "1m32s").
func formatearDuracion(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	minutos := int(d.Minutes())
	segundos := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%02ds", minutos, segundos)
}

// lineaProgreso compone la barra con el porcentaje, el tiempo transcurrido y la
// estimación restante.
func (m Modelo) lineaProgreso() string {
	fraccion, transcurrido, restante := m.progresoActual()
	if m.inicioProgreso.IsZero() {
		return ""
	}
	porcentaje := int(fraccion*100 + 0.5)
	barra := barraProgreso(fraccion, anchoBarra)

	detalle := fmt.Sprintf("  %3d%%  ·  %s transcurridos", porcentaje, formatearDuracion(transcurrido))
	if transcurrido >= m.progresoEstimado {
		detalle += "  ·  finalizando…"
	} else {
		detalle += "  ·  quedan ~" + formatearDuracion(restante)
	}
	return barra + estiloBarraDatos.Render(detalle)
}
