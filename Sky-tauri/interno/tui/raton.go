package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// accionZona identifica qué representa una región clicable de la pantalla.
type accionZona int

const (
	zonaNinguna accionZona = iota
	zonaOpcionMenu
	zonaProyecto
	zonaEntradaExplorador
	zonaBotonEliminar
	zonaBotonEliminarTodo
	zonaBotonConfirmarSi
	zonaBotonConfirmarNo
	zonaCampo
	zonaBotonCrear
	zonaBotonResultado
)

// zona es un rectángulo clicable, en coordenadas relativas al contenido (antes
// de centrarlo en la ventana).
type zona struct {
	fila0, fila1 int
	col0, col1   int
	accion       accionZona
	indice       int
}

// desplazamientoCentrado replica el centrado de lipgloss.Place.
func desplazamientoCentrado(ancho, alto int, contenido string) (int, int) {
	if ancho <= 0 {
		ancho = 80
	}
	if alto <= 0 {
		alto = 24
	}
	x := (ancho - lipgloss.Width(contenido)) / 2
	y := (alto - lipgloss.Height(contenido)) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return x, y
}

// actualizarRaton procesa los eventos del ratón.
func (m Modelo) actualizarRaton(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// La rueda desplaza la salida de la compilación.
	if m.estado == estadoProgreso {
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
			m.registro, _ = m.registro.Update(msg)
		}
		return m, nil
	}
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	contenido, zonas := m.bloqueActual()
	x, y := desplazamientoCentrado(m.ancho, m.alto, contenido)
	fila := msg.Y - y
	col := msg.X - x
	for _, z := range zonas {
		if fila >= z.fila0 && fila < z.fila1 && col >= z.col0 && col < z.col1 {
			return m.ejecutarZona(z)
		}
	}
	return m, nil
}

// ejecutarZona ejecuta la acción asociada a una región clicable.
func (m Modelo) ejecutarZona(z zona) (tea.Model, tea.Cmd) {
	switch z.accion {
	case zonaOpcionMenu:
		m.opcionMenu = z.indice
		return m.ejecutarOpcionMenu()

	case zonaProyecto:
		m.indiceProyecto = z.indice
		m.confirmandoEliminar = false
		m.confirmandoEliminarTodo = false

	case zonaEntradaExplorador:
		m.explorador.indice = z.indice
		return m.exploradorEnter()

	case zonaBotonEliminar:
		if len(m.proyectos) > 0 {
			m.confirmandoEliminar = true
			m.confirmandoEliminarTodo = false
		}

	case zonaBotonEliminarTodo:
		if len(m.proyectos) > 0 {
			m.confirmandoEliminarTodo = true
			m.confirmandoEliminar = false
		}

	case zonaBotonConfirmarSi:
		if m.confirmandoEliminar {
			m.confirmandoEliminar = false
			if m.indiceProyecto >= 0 && m.indiceProyecto < len(m.proyectos) {
				return m, eliminarProyectoCmd(m.proyectos[m.indiceProyecto])
			}
			return m, nil
		}
		if m.confirmandoEliminarTodo {
			m.confirmandoEliminarTodo = false
			if len(m.proyectos) > 0 {
				return m, eliminarTodoCmd(m.proyectos)
			}
		}

	case zonaBotonConfirmarNo:
		m.confirmandoEliminar = false
		m.confirmandoEliminarTodo = false

	case zonaCampo:
		m.campo = z.indice
		m.mensajeError = ""
		m.enfocar()
		return m, textinputBlink()

	case zonaBotonCrear:
		return m.lanzarGeneracion()

	case zonaBotonResultado:
		switch z.indice {
		case 0:
			if m.exito && m.puedeCompilar && m.rutaProyecto != "" {
				return m.iniciarCompilacion()
			}
		case 1:
			if m.exito && m.puedeInstalar && !m.instalado && m.rutaProyecto != "" {
				return m.iniciarInstalacion()
			}
		case 2:
			abrirEnExplorador(m.rutaProyecto)
		case 3:
			m.estado = estadoFormulario
			m.mensajeError = ""
			m.enfocar()
			return m, textinputBlink()
		case 4:
			m.estado = estadoMenu
		}
	}
	return m, nil
}

// defBoton describe un botón de una fila de acciones.
type defBoton struct {
	texto   string
	accion  accionZona
	indice  int
	peligro bool
}

// renderBoton dibuja un botón con borde (tres líneas de alto).
func renderBoton(texto string, peligro bool) string {
	estilo := estiloBoton
	if peligro {
		estilo = estiloBotonPeligro
	}
	return estilo.Render(" " + texto + " ")
}

// filaBotones compone una fila horizontal de botones y sus zonas. Las zonas
// devueltas son relativas a la fila (fila 0 = primera línea de la fila).
func filaBotones(botones []defBoton) ([]string, []zona) {
	if len(botones) == 0 {
		return nil, nil
	}
	partes := make([][]string, len(botones))
	anchos := make([]int, len(botones))
	alto := 0
	for i, b := range botones {
		lineas := strings.Split(renderBoton(b.texto, b.peligro), "\n")
		partes[i] = lineas
		anchos[i] = lipgloss.Width(lineas[0])
		if len(lineas) > alto {
			alto = len(lineas)
		}
	}

	inicio := make([]int, len(botones))
	col := 0
	for i := range botones {
		inicio[i] = col
		col += anchos[i] + 3
	}

	lineas := make([]string, alto)
	for f := 0; f < alto; f++ {
		var sb strings.Builder
		for i := range botones {
			if f < len(partes[i]) {
				sb.WriteString(partes[i][f])
			} else {
				sb.WriteString(strings.Repeat(" ", anchos[i]))
			}
			if i < len(botones)-1 {
				sb.WriteString("   ")
			}
		}
		lineas[f] = sb.String()
	}

	var zonas []zona
	for i, b := range botones {
		zonas = append(zonas, zona{
			fila0: 0, fila1: alto,
			col0: inicio[i], col1: inicio[i] + anchos[i],
			accion: b.accion, indice: b.indice,
		})
	}
	return lineas, zonas
}

// marco envuelve un cuerpo en el recuadro con título de la interfaz. Devuelve
// las líneas finales y el desplazamiento (fila, columna) del cuerpo dentro del
// recuadro, para poder situar las zonas clicables.
func marco(titulo string, cuerpo []string) ([]string, int, int) {
	interior := make([]string, 0, len(cuerpo)+2)
	interior = append(interior, estiloTitulo.Render(titulo), "")
	interior = append(interior, cuerpo...)
	render := estiloPanel.Render(strings.Join(interior, "\n"))
	return strings.Split(render, "\n"), 4, 4
}

// desplazarZonas aplica el desplazamiento del marco a una lista de zonas.
func desplazarZonas(zonas []zona, fila, col int, base int) {
	for i := range zonas {
		zonas[i].fila0 += fila + base
		zonas[i].fila1 += fila + base
		zonas[i].col0 += col
		zonas[i].col1 += col
	}
}
