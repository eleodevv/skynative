package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View dibuja la pantalla activa, centrada en la ventana.
func (m Modelo) View() string {
	switch m.estado {
	case estadoMenu:
		return m.vistaMenu()
	case estadoFormulario:
		return m.centrar(m.vistaFormulario())
	case estadoExplorador:
		return m.centrar(m.vistaExplorador())
	case estadoProgreso:
		return m.centrar(m.vistaProgreso())
	case estadoResultado:
		return m.centrar(m.vistaResultado())
	case estadoLista:
		return m.centrar(m.vistaLista())
	case estadoDiagnostico:
		return m.centrar(m.vistaDiagnostico())
	default:
		return ""
	}
}

// dimensiones devuelve un tamaño de ventana utilizable aunque todavía no haya
// llegado un mensaje de tamaño.
func (m Modelo) dimensiones() (int, int) {
	ancho, alto := m.ancho, m.alto
	if ancho <= 0 {
		ancho = 80
	}
	if alto <= 0 {
		alto = 24
	}
	return ancho, alto
}

// centrar coloca un bloque en el centro de la ventana.
func (m Modelo) centrar(contenido string) string {
	ancho, alto := m.dimensiones()
	return lipgloss.Place(ancho, alto, lipgloss.Center, lipgloss.Center, contenido)
}

// panel envuelve un cuerpo en un recuadro con un título.
func panel(titulo, cuerpo string) string {
	return estiloPanel.Render(estiloTitulo.Render(titulo) + "\n\n" + cuerpo)
}

// marca resalta el campo enfocado.
func (m Modelo) marca(indice int) string {
	if m.campo == indice {
		return estiloFoco.Render("❯")
	}
	return " "
}

// etiquetaCampo colorea la etiqueta del campo enfocado.
func (m Modelo) etiquetaCampo(indice int, texto string) string {
	if m.campo == indice {
		return estiloFoco.Render(texto)
	}
	return estiloEtiquetaCampo.Render(texto)
}

func (m Modelo) vistaMenu() string {
	ancho, _ := m.dimensiones()

	var cabecera string
	if ancho < anchoLogo("SKYNATIVE")+6 {
		cabecera = estiloMarca.Render("◆ SKYNATIVE")
	} else {
		cabecera = dibujarLogo("SKYNATIVE")
	}

	subtitulo := estiloSubtitulo.Render("Convierte una URL en una aplicación de escritorio · Tauri v2")

	const anchoFila = 34
	filas := make([]string, len(opcionesMenu))
	for i, opcion := range opcionesMenu {
		if i == m.opcionMenu {
			filas[i] = estiloSeleccion.Width(anchoFila).Render(" ❯ " + opcion)
		} else {
			filas[i] = lipgloss.NewStyle().Width(anchoFila).Render("   " + opcion)
		}
	}
	menu := estiloPanel.Render(strings.Join(filas, "\n"))

	contenido := lipgloss.JoinVertical(lipgloss.Center,
		cabecera,
		"",
		subtitulo,
		"",
		menu,
		"",
		estiloAyuda.Render("↑↓ moverse · Enter seleccionar · 1-5 atajos · q salir"),
	)
	return m.centrar(contenido)
}

func (m Modelo) vistaFormulario() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n   %s\n\n", m.marca(campoURL), m.etiquetaCampo(campoURL, "URL remota (http/https)"), m.url.View()))
	b.WriteString(fmt.Sprintf("%s %s\n   %s\n\n", m.marca(campoNombre), m.etiquetaCampo(campoNombre, "Nombre de la aplicación"), m.nombre.View()))

	dimensiones := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left,
			m.marca(campoAncho)+" "+m.etiquetaCampo(campoAncho, "Ancho"),
			"   "+m.anchoCampo.View()),
		"     ",
		lipgloss.JoinVertical(lipgloss.Left,
			m.marca(campoAlto)+" "+m.etiquetaCampo(campoAlto, "Alto"),
			"   "+m.altoCampo.View()),
	)
	b.WriteString(dimensiones + "\n\n")

	b.WriteString(fmt.Sprintf("%s %s\n   %s\n   %s\n\n",
		m.marca(campoCarpeta),
		m.etiquetaCampo(campoCarpeta, "Carpeta de salida"),
		estiloRuta.Render(m.carpeta),
		estiloAyuda.Render("Enter para abrir el explorador de carpetas")))
	b.WriteString(fmt.Sprintf("%s %s  %s\n\n",
		m.marca(campoRedimensionable),
		m.etiquetaCampo(campoRedimensionable, "Ventana redimensionable"),
		casilla(m.redimensionable)))

	if m.mensajeError != "" {
		b.WriteString(estiloError.Render("✖ "+m.mensajeError) + "\n\n")
	}

	if m.campo == campoCrear {
		b.WriteString(m.marca(campoCrear) + " " + estiloSeleccion.Render(" Crear aplicación ") + "\n")
	} else {
		b.WriteString("  " + estiloCorrecto.Render("Crear aplicación") + "\n")
	}

	cuerpo := b.String() + estiloAyuda.Render("\nTab/↑↓ moverse · Enter abrir o actuar · Espacio alternar · Ctrl+G crear · Esc volver")
	return panel("Nueva aplicación", cuerpo)
}

func (m Modelo) vistaExplorador() string {
	_, alto := m.dimensiones()

	var b strings.Builder
	if m.explorador.err != nil {
		b.WriteString(estiloError.Render("No se pudo leer la carpeta: "+m.explorador.err.Error()) + "\n\n")
	}

	visibles := alto - 14
	if visibles < 5 {
		visibles = 5
	}
	inicio := 0
	if m.explorador.indice >= visibles {
		inicio = m.explorador.indice - visibles + 1
	}
	fin := inicio + visibles
	if fin > len(m.explorador.entradas) {
		fin = len(m.explorador.entradas)
	}

	for i := inicio; i < fin; i++ {
		entrada := m.explorador.entradas[i]
		etiqueta := entrada.etiqueta
		if entrada.accion == accionEntrar {
			etiqueta = "▸ " + etiqueta
		}
		if i == m.explorador.indice {
			b.WriteString(estiloSeleccion.Render("❯ "+etiqueta) + "\n")
		} else {
			b.WriteString("  " + etiqueta + "\n")
		}
	}

	cuerpo := estiloRuta.Render(m.explorador.directorio) + "\n\n" + b.String() +
		estiloAyuda.Render("\n↑↓ moverse · Enter entrar/elegir · S usar esta carpeta · Retroceso subir · Inicio ~ · Esc cancelar")
	return panel("Elegir carpeta de salida", cuerpo)
}

func (m Modelo) vistaProgreso() string {
	accion := "Generando el proyecto…"
	switch {
	case m.compilandoAhora:
		accion = "Compilando con Tauri (puede tardar varios minutos)…"
	case m.instalandoAhora:
		accion = "Instalando la aplicación en el sistema…"
	}
	cabecera := m.girador.View() + " " + estiloTitulo.Render(accion)
	salida := estiloCaja.Render(m.registro.View())
	return lipgloss.JoinVertical(lipgloss.Center,
		cabecera,
		"",
		salida,
		"",
		estiloAyuda.Render("↑↓ desplazar la salida · Esc cancelar"),
	)
}

func (m Modelo) vistaResultado() string {
	var b strings.Builder
	if m.exito {
		b.WriteString(estiloCorrecto.Render("✔ "+m.resumen) + "\n")
	} else {
		b.WriteString(estiloError.Render("✖ "+m.resumen) + "\n")
	}

	if m.rutaProyecto != "" {
		b.WriteString("\n" + estiloEtiquetaCampo.Render("Proyecto  ") + estiloRuta.Render(m.rutaProyecto) + "\n")
	}

	if m.instalado {
		b.WriteString(estiloEtiquetaCampo.Render("Acceso    ") + "en el menú de aplicaciones (doble clic)\n")
		if m.entradaEscritorio != "" {
			b.WriteString(estiloEtiquetaCampo.Render("Lanzador  ") + estiloRuta.Render(m.entradaEscritorio) + "\n")
		}
	}

	if len(m.artefactos) > 0 {
		b.WriteString("\n" + estiloEtiquetaCampo.Render("Artefactos") + "\n")
		for _, artefacto := range m.artefactos {
			b.WriteString("  • " + artefacto + "\n")
		}
	}

	var ayudas []string
	if m.puedeCompilar && m.exito {
		ayudas = append(ayudas, "C compilar")
	}
	if m.puedeInstalar && m.exito {
		ayudas = append(ayudas, "I instalar en el sistema")
	}
	if m.rutaProyecto != "" {
		ayudas = append(ayudas, "O abrir carpeta")
	}
	ayudas = append(ayudas, "Enter crear otra", "Esc menú")

	titulo := "Resultado"
	if m.exito {
		titulo = "Listo"
	}
	return panel(titulo, b.String()+estiloAyuda.Render("\n"+strings.Join(ayudas, " · ")))
}

func (m Modelo) vistaLista() string {
	var b strings.Builder
	if len(m.proyectos) == 0 {
		b.WriteString("Todavía no hay proyectos.\nCrea uno desde el menú principal.")
	} else {
		for i, p := range m.proyectos {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(estiloMarca.Render(p.Nombre) + "\n")
			b.WriteString("  " + estiloEtiquetaCampo.Render("identificador  ") + p.Identificador + "\n")
			b.WriteString("  " + estiloEtiquetaCampo.Render("url            ") + p.URL + "\n")
			b.WriteString("  " + estiloEtiquetaCampo.Render("ruta           ") + estiloRuta.Render(p.Ruta) + "\n")
		}
	}
	cuerpo := b.String() + estiloAyuda.Render("\nEsc volver al menú")
	return panel("Proyectos registrados", cuerpo)
}

func (m Modelo) vistaDiagnostico() string {
	var b strings.Builder
	for _, e := range m.diagnosticos {
		if e.OK {
			b.WriteString(estiloCorrecto.Render("✔ ") + estiloEtiquetaCampo.Render(e.Nombre) + "  " + e.Detalle + "\n")
			continue
		}
		marca := "✖ "
		estilo := estiloError
		if !e.Critico {
			marca = "• "
			estilo = estiloAviso
		}
		b.WriteString(estilo.Render(marca) + estiloEtiquetaCampo.Render(e.Nombre) + "  " + e.Detalle + "\n")
		if e.Ayuda != "" {
			b.WriteString("    " + estiloAviso.Render("→ "+e.Ayuda) + "\n")
		}
	}
	cuerpo := b.String() + estiloAyuda.Render("\nEsc volver al menú")
	return panel("Diagnóstico del entorno", cuerpo)
}

// interfaz de compilación dinámica: evita que el compilador se queje de que
// Modelo no implementa tea.Model cuando se pasa por valor.
var _ tea.Model = Modelo{}
