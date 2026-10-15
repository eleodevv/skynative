package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// View dibuja la pantalla activa.
func (m Modelo) View() string {
	switch m.estado {
	case estadoMenu:
		return m.vistaMenu()
	case estadoFormulario:
		return m.vistaFormulario()
	case estadoExplorador:
		return m.vistaExplorador()
	case estadoProgreso:
		return m.vistaProgreso()
	case estadoResultado:
		return m.vistaResultado()
	case estadoLista:
		return m.vistaLista()
	case estadoDiagnostico:
		return m.vistaDiagnostico()
	default:
		return ""
	}
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
	return estiloEtiqueta.Render(texto)
}

func (m Modelo) vistaMenu() string {
	var b strings.Builder
	b.WriteString(estiloTitulo.Render("Skynative") + "\n")
	b.WriteString(estiloSubtitulo.Render("Convierte una URL en una aplicación de escritorio (Tauri v2)") + "\n\n")
	for i, opcion := range opcionesMenu {
		if i == m.opcionMenu {
			b.WriteString(estiloSeleccion.Render("❯ "+opcion) + "\n")
		} else {
			b.WriteString("  " + opcion + "\n")
		}
	}
	b.WriteString(estiloAyuda.Render("\n↑↓ moverse · Enter seleccionar · 1-5 acceso directo · q salir"))
	return b.String()
}

func (m Modelo) vistaFormulario() string {
	var b strings.Builder
	b.WriteString(estiloTitulo.Render("Nueva aplicación") + "\n\n")

	b.WriteString(fmt.Sprintf("%s %s\n   %s\n\n", m.marca(campoURL), m.etiquetaCampo(campoURL, "URL remota (http/https)"), m.url.View()))
	b.WriteString(fmt.Sprintf("%s %s\n   %s\n\n", m.marca(campoNombre), m.etiquetaCampo(campoNombre, "Nombre de la aplicación"), m.nombre.View()))
	b.WriteString(fmt.Sprintf("%s %s\n   %s\n\n", m.marca(campoAncho), m.etiquetaCampo(campoAncho, "Ancho (px)"), m.anchoCampo.View()))
	b.WriteString(fmt.Sprintf("%s %s\n   %s\n\n", m.marca(campoAlto), m.etiquetaCampo(campoAlto, "Alto (px)"), m.altoCampo.View()))
	b.WriteString(fmt.Sprintf("%s %s\n   %s\n   %s\n\n",
		m.marca(campoCarpeta),
		m.etiquetaCampo(campoCarpeta, "Carpeta de salida"),
		estiloRuta.Render(m.carpeta),
		estiloAyuda.Render("Pulsa Enter para abrir el explorador de carpetas")))
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
	b.WriteString(estiloAyuda.Render("\nTab/↑↓ moverse · Enter abrir o actuar · Espacio alternar · Ctrl+G crear · Esc volver"))
	return b.String()
}

func (m Modelo) vistaExplorador() string {
	var b strings.Builder
	b.WriteString(estiloTitulo.Render("Elegir carpeta de salida") + "\n")
	b.WriteString(estiloRuta.Render(m.explorador.directorio) + "\n\n")

	if m.explorador.err != nil {
		b.WriteString(estiloError.Render("No se pudo leer la carpeta: "+m.explorador.err.Error()) + "\n\n")
	}

	visibles := m.alto - 9
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

	b.WriteString(estiloAyuda.Render("\n↑↓ moverse · Enter entrar/seleccionar · S usar esta carpeta · Retroceso subir · Inicio ~ · Esc cancelar"))
	return b.String()
}

func (m Modelo) vistaProgreso() string {
	accion := "Generando el proyecto..."
	if m.compilandoAhora {
		accion = "Compilando con Tauri (puede tardar varios minutos)..."
	}
	cabecera := m.girador.View() + " " + estiloTitulo.Render(accion)
	return cabecera + "\n\n" + estiloCaja.Render(m.registro.View()) + "\n" + estiloAyuda.Render("↑↓ desplazar la salida · Esc cancelar")
}

func (m Modelo) vistaResultado() string {
	var b strings.Builder
	if m.exito {
		b.WriteString(estiloCorrecto.Render("✔ "+m.resumen) + "\n\n")
	} else {
		b.WriteString(estiloError.Render("✖ "+m.resumen) + "\n\n")
	}

	if m.rutaProyecto != "" {
		b.WriteString(estiloEtiqueta.Render("Proyecto: ") + estiloRuta.Render(m.rutaProyecto) + "\n")
	}

	if len(m.artefactos) > 0 {
		b.WriteString("\n" + estiloEtiqueta.Render("Artefactos generados:") + "\n")
		for _, artefacto := range m.artefactos {
			b.WriteString("  • " + artefacto + "\n")
		}
	}

	var ayudas []string
	if m.puedeCompilar && m.exito {
		ayudas = append(ayudas, "C compilar")
	}
	if m.rutaProyecto != "" {
		ayudas = append(ayudas, "O abrir carpeta")
	}
	ayudas = append(ayudas, "Enter crear otra", "Esc menú")
	b.WriteString(estiloAyuda.Render("\n" + strings.Join(ayudas, " · ")))
	return b.String()
}

func (m Modelo) vistaLista() string {
	var b strings.Builder
	b.WriteString(estiloTitulo.Render("Proyectos registrados") + "\n\n")
	if len(m.proyectos) == 0 {
		b.WriteString("Todavía no hay proyectos. Crea uno desde el menú principal.\n")
	} else {
		for _, p := range m.proyectos {
			b.WriteString(estiloFoco.Render(p.Nombre) + "\n")
			b.WriteString("  " + estiloEtiqueta.Render("identificador: ") + p.Identificador + "\n")
			b.WriteString("  " + estiloEtiqueta.Render("url: ") + p.URL + "\n")
			b.WriteString("  " + estiloEtiqueta.Render("ruta: ") + estiloRuta.Render(p.Ruta) + "\n\n")
		}
	}
	b.WriteString(estiloAyuda.Render("Esc volver al menú"))
	return b.String()
}

func (m Modelo) vistaDiagnostico() string {
	var b strings.Builder
	b.WriteString(estiloTitulo.Render("Diagnóstico del entorno") + "\n\n")
	for _, e := range m.diagnosticos {
		if e.OK {
			b.WriteString(estiloCorrecto.Render("✔ ") + estiloEtiqueta.Render(e.Nombre) + ": " + e.Detalle + "\n")
			continue
		}
		marca := "✖ "
		estilo := estiloError
		if !e.Critico {
			marca = "• "
			estilo = estiloAviso
		}
		b.WriteString(estilo.Render(marca) + estiloEtiqueta.Render(e.Nombre) + ": " + e.Detalle + "\n")
		if e.Ayuda != "" {
			b.WriteString("    " + estiloAviso.Render("→ "+e.Ayuda) + "\n")
		}
	}
	b.WriteString(estiloAyuda.Render("\nEsc volver al menú"))
	return b.String()
}

// interfaz de compilación dinámica: evita que el compilador se queje de que
// Modelo no implementa tea.Model cuando se pasa por valor.
var _ tea.Model = Modelo{}
