package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const zonaAnchuraCompleta = 1 << 20

// View dibuja la pantalla activa, centrada en la ventana.
func (m Modelo) View() string {
	contenido, _ := m.bloqueActual()
	return m.centrar(contenido)
}

// bloqueActual construye el contenido sin centrar y las regiones clicables en
// coordenadas relativas a ese contenido.
func (m Modelo) bloqueActual() (string, []zona) {
	switch m.estado {
	case estadoMenu:
		return m.bloqueMenu()
	case estadoFormulario:
		return m.bloqueFormulario()
	case estadoExplorador:
		return m.bloqueExplorador()
	case estadoProgreso:
		return m.bloqueProgreso(), nil
	case estadoResultado:
		return m.bloqueResultado()
	case estadoLista:
		return m.bloqueLista()
	case estadoDiagnostico:
		return m.bloqueDiagnostico(), nil
	default:
		return "", nil
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

// bloqueMenu construye el menú principal con filas-botón clicables.
func (m Modelo) bloqueMenu() (string, []zona) {
	ancho, _ := m.dimensiones()

	var lineas []string
	if ancho < anchoLogo("SKYNATIVE")+6 {
		lineas = append(lineas, estiloMarca.Render("◆ SKYNATIVE"))
	} else {
		lineas = append(lineas, strings.Split(dibujarLogo("SKYNATIVE"), "\n")...)
	}
	lineas = append(lineas,
		"",
		estiloSubtitulo.Render("☁  Convierte una URL en una aplicación de escritorio · Tauri v2"),
		"")

	const anchoFila = 50
	var zonas []zona
	for i, opcion := range opcionesMenu {
		texto := fmt.Sprintf("%d  %s", i+1, opcion)
		lineas = append(lineas, botonLinea(texto, i == m.opcionMenu, anchoFila))
		zonas = append(zonas, zona{
			fila0: len(lineas) - 1, fila1: len(lineas),
			col0: 0, col1: zonaAnchuraCompleta,
			accion: zonaOpcionMenu, indice: i,
		})
	}

	if m.chequeoHecho && !m.herramientasListas {
		lineas = append(lineas, "",
			estiloAviso.Render("⚠ Faltan herramientas para compilar · elige «Instalar herramientas necesarias» (2)"))
	}
	lineas = append(lineas, "",
		estiloAyuda.Render("↑↓ moverse · Enter seleccionar · 1-6 atajos · clic con el ratón · q salir"))
	return strings.Join(lineas, "\n"), zonas
}

// bloqueFormulario construye el formulario de creación.
func (m Modelo) bloqueFormulario() (string, []zona) {
	var cuerpo []string
	var zonas []zona
	add := func(texto string) int {
		cuerpo = append(cuerpo, texto)
		return len(cuerpo) - 1
	}
	campo := func(inicio, fin, indice int) {
		zonas = append(zonas, zona{fila0: inicio, fila1: fin, col0: 0, col1: zonaAnchuraCompleta, accion: zonaCampo, indice: indice})
	}

	i := add(m.marca(campoURL) + " " + m.etiquetaCampo(campoURL, "URL remota (http/https)"))
	add("   " + m.url.View())
	campo(i, i+2, campoURL)
	add("")

	i = add(m.marca(campoNombre) + " " + m.etiquetaCampo(campoNombre, "Nombre de la aplicación"))
	add("   " + m.nombre.View())
	campo(i, i+2, campoNombre)
	add("")

	etiquetasDim := m.marca(campoAncho) + " " + m.etiquetaCampo(campoAncho, "Ancho") +
		"      " + m.marca(campoAlto) + " " + m.etiquetaCampo(campoAlto, "Alto")
	filaDim := add(etiquetasDim)
	add("   " + m.anchoCampo.View() + "      " + "   " + m.altoCampo.View())
	zonas = append(zonas,
		zona{fila0: filaDim, fila1: filaDim + 2, col0: 0, col1: 28, accion: zonaCampo, indice: campoAncho},
		zona{fila0: filaDim, fila1: filaDim + 2, col0: 28, col1: zonaAnchuraCompleta, accion: zonaCampo, indice: campoAlto},
	)
	add("")

	i = add(m.marca(campoCarpeta) + " " + m.etiquetaCampo(campoCarpeta, "Carpeta de salida"))
	add("   " + estiloRuta.Render(m.carpeta))
	add("   " + estiloAyuda.Render("Enter para abrir el explorador de carpetas"))
	campo(i, i+3, campoCarpeta)
	add("")

	i = add(m.marca(campoRedimensionable) + " " + m.etiquetaCampo(campoRedimensionable, "Ventana redimensionable") + "  " + casilla(m.redimensionable))
	campo(i, i+1, campoRedimensionable)

	i = add(m.marca(campoEmpaquetar) + " " + m.etiquetaCampo(campoEmpaquetar, "Crear paquetes instalables (.deb/.app/.exe)") + "  " + casilla(m.empaquetar))
	campo(i, i+1, campoEmpaquetar)
	add("   " + estiloAyuda.Render("desactivado compila solo el binario (mucho más rápido)"))
	add("")

	if m.mensajeError != "" {
		add(estiloError.Render("✖ " + m.mensajeError))
		add("")
	}

	botonCrear := strings.Split(renderBoton("Crear aplicación", false), "\n")
	inicioBoton := len(cuerpo)
	cuerpo = append(cuerpo, botonCrear...)
	zonas = append(zonas, zona{fila0: inicioBoton, fila1: inicioBoton + len(botonCrear), col0: 0, col1: zonaAnchuraCompleta, accion: zonaBotonCrear})

	add("")
	add(estiloAyuda.Render("Tab/↑↓ moverse · Enter abrir o actuar · Espacio alternar · Ctrl+G crear · Esc volver"))

	return m.conMarco("Nueva aplicación", cuerpo, zonas)
}

// bloqueExplorador construye el explorador de carpetas.
func (m Modelo) bloqueExplorador() (string, []zona) {
	_, alto := m.dimensiones()

	var cuerpo []string
	var zonas []zona
	if m.explorador.err != nil {
		cuerpo = append(cuerpo, estiloError.Render("No se pudo leer la carpeta: "+m.explorador.err.Error()), "")
	}

	visibles := alto - 16
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

	cuerpo = append(cuerpo, estiloRuta.Render(m.explorador.directorio), "")
	for i := inicio; i < fin; i++ {
		entrada := m.explorador.entradas[i]
		etiqueta := entrada.etiqueta
		if entrada.accion == accionEntrar {
			etiqueta = "▸ " + etiqueta
		}
		if i == m.explorador.indice {
			cuerpo = append(cuerpo, estiloSeleccion.Render("❯ "+etiqueta))
		} else {
			cuerpo = append(cuerpo, "  "+etiqueta)
		}
		zonas = append(zonas, zona{fila0: len(cuerpo) - 1, fila1: len(cuerpo), col0: 0, col1: zonaAnchuraCompleta, accion: zonaEntradaExplorador, indice: i})
	}
	cuerpo = append(cuerpo, estiloAyuda.Render("↑↓ moverse · Enter entrar/elegir · S usar esta carpeta · Retroceso subir · Inicio ~ · Esc cancelar"))

	return m.conMarco("Elegir carpeta de salida", cuerpo, zonas)
}

// bloqueProgreso construye la pantalla de progreso.
func (m Modelo) bloqueProgreso() string {
	accion := "Generando el proyecto…"
	switch {
	case m.compilandoAhora:
		accion = "Compilando con Tauri (puede tardar varios minutos)…"
	case m.instalandoAhora:
		accion = "Instalando la aplicación en el sistema…"
	case m.instalandoHerramientas:
		accion = "Instalando las herramientas necesarias (puede tardar varios minutos)…"
	}
	cabecera := m.girador.View() + " " + estiloTitulo.Render(accion)
	salida := estiloCaja.Render(m.registro.View())
	return lipgloss.JoinVertical(lipgloss.Center,
		estiloSubtitulo.Render("☁  Skynative está trabajando"),
		"",
		cabecera,
		"",
		salida,
		"",
		estiloAyuda.Render("↑↓ desplazar · rueda del ratón · Esc cancelar"),
	)
}

// bloqueResultado construye la pantalla de resultado con botones de acción.
func (m Modelo) bloqueResultado() (string, []zona) {
	var cuerpo []string
	if m.exito {
		cuerpo = append(cuerpo, estiloCorrecto.Render("✔ "+m.resumen))
	} else {
		cuerpo = append(cuerpo, estiloError.Render("✖ "+m.resumen))
	}

	for _, nota := range m.notas {
		cuerpo = append(cuerpo, estiloAviso.Render("⚠ "+nota))
	}
	if m.exito && m.chequeoHecho && !m.herramientasListas {
		cuerpo = append(cuerpo, estiloAviso.Render("⚠ Faltan herramientas para compilar · menú → «Instalar herramientas necesarias»"))
	}
	if m.rutaProyecto != "" {
		cuerpo = append(cuerpo, estiloEtiquetaCampo.Render("Proyecto  ")+estiloRuta.Render(m.rutaProyecto))
	}
	if m.instalado {
		cuerpo = append(cuerpo, estiloEtiquetaCampo.Render("Acceso    ")+"en el menú de aplicaciones (doble clic)")
		if m.entradaEscritorio != "" {
			cuerpo = append(cuerpo, estiloEtiquetaCampo.Render("Lanzador  ")+estiloRuta.Render(m.entradaEscritorio))
		}
	}
	if len(m.artefactos) > 0 {
		cuerpo = append(cuerpo, "")
		cuerpo = append(cuerpo, estiloEtiquetaCampo.Render("Artefactos"))
		for _, artefacto := range m.artefactos {
			cuerpo = append(cuerpo, "  • "+artefacto)
		}
	}
	if len(m.detalle) > 0 {
		cuerpo = append(cuerpo, "")
		cuerpo = append(cuerpo, estiloEtiquetaCampo.Render("Últimas líneas de la salida"))
		for _, l := range m.detalle {
			cuerpo = append(cuerpo, "  "+l)
		}
	}
	cuerpo = append(cuerpo, "")

	var botones []defBoton
	if m.exito && m.puedeCompilar {
		botones = append(botones, defBoton{texto: "Compilar", accion: zonaBotonResultado, indice: 0})
	}
	if m.exito && m.puedeInstalar && !m.instalado {
		botones = append(botones, defBoton{texto: "Instalar en el sistema", accion: zonaBotonResultado, indice: 1})
	}
	if m.rutaProyecto != "" {
		botones = append(botones, defBoton{texto: "Abrir carpeta", accion: zonaBotonResultado, indice: 2})
	}
	botones = append(botones,
		defBoton{texto: "Crear otra", accion: zonaBotonResultado, indice: 3},
		defBoton{texto: "Menú", accion: zonaBotonResultado, indice: 4},
	)

	lineasBoton, zonasBoton := filaBotones(botones)
	inicio := len(cuerpo)
	cuerpo = append(cuerpo, lineasBoton...)
	desplazarZonas(zonasBoton, inicio, 0, 0)

	cuerpo = append(cuerpo, "", estiloAyuda.Render("C compilar · I instalar · O abrir carpeta · Enter crear otra · Esc menú"))

	titulo := "Resultado"
	if m.exito {
		titulo = "Listo"
	}
	return m.conMarco(titulo, cuerpo, zonasBoton)
}

// bloqueLista construye la lista de proyectos con acciones.
func (m Modelo) bloqueLista() (string, []zona) {
	var cuerpo []string
	var zonas []zona
	if len(m.proyectos) == 0 {
		cuerpo = append(cuerpo,
			"Todavía no hay proyectos.",
			"Crea uno desde el menú principal.",
			"",
			estiloAyuda.Render("Esc volver al menú"))
		return m.conMarco("Proyectos registrados", cuerpo, zonas)
	}

	for i, p := range m.proyectos {
		if i > 0 {
			cuerpo = append(cuerpo, "")
		}
		inicio := len(cuerpo)
		nombre := estiloMarca.Render(p.Nombre)
		if i == m.indiceProyecto {
			nombre = estiloSeleccion.Render("❯ " + p.Nombre)
		}
		cuerpo = append(cuerpo, nombre)
		cuerpo = append(cuerpo, "  "+estiloEtiquetaCampo.Render("identificador  ")+p.Identificador)
		cuerpo = append(cuerpo, "  "+estiloEtiquetaCampo.Render("url            ")+p.URL)
		cuerpo = append(cuerpo, "  "+estiloEtiquetaCampo.Render("ruta           ")+estiloRuta.Render(p.Ruta))
		zonas = append(zonas, zona{fila0: inicio, fila1: len(cuerpo), col0: 0, col1: zonaAnchuraCompleta, accion: zonaProyecto, indice: i})
	}
	cuerpo = append(cuerpo, "")

	var botones []defBoton
	switch {
	case m.confirmandoEliminar:
		cuerpo = append(cuerpo, estiloError.Render("¿Eliminar «"+m.proyectos[m.indiceProyecto].Nombre+"»? Se borrarán su acceso, su carpeta y el registro."))
		botones = []defBoton{
			{texto: "Sí, eliminar", accion: zonaBotonConfirmarSi, peligro: true},
			{texto: "Cancelar", accion: zonaBotonConfirmarNo},
		}
	case m.confirmandoEliminarTodo:
		botones = []defBoton{
			{texto: "Sí, eliminar todo", accion: zonaBotonConfirmarSi, peligro: true},
			{texto: "Cancelar", accion: zonaBotonConfirmarNo},
		}
	default:
		botones = []defBoton{
			{texto: "Eliminar seleccionado", accion: zonaBotonEliminar, peligro: true},
			{texto: "Eliminar todo", accion: zonaBotonEliminarTodo, peligro: true},
		}
	}

	lineasBoton, zonasBoton := filaBotones(botones)
	inicio := len(cuerpo)
	cuerpo = append(cuerpo, lineasBoton...)
	desplazarZonas(zonasBoton, inicio, 0, 0)
	zonas = append(zonas, zonasBoton...)

	cuerpo = append(cuerpo, "",
		estiloAyuda.Render("↑↓ elegir · D eliminar · A eliminar todo · Esc volver al menú"))

	return m.conMarco("Proyectos registrados", cuerpo, zonas)
}

// bloqueDiagnostico construye la pantalla de diagnóstico.
func (m Modelo) bloqueDiagnostico() string {
	var cuerpo []string
	for _, e := range m.diagnosticos {
		if e.OK {
			cuerpo = append(cuerpo, estiloCorrecto.Render("✔ ")+estiloEtiquetaCampo.Render(e.Nombre)+"  "+e.Detalle)
			continue
		}
		marca := "✖ "
		estilo := estiloError
		if !e.Critico {
			marca = "• "
			estilo = estiloAviso
		}
		cuerpo = append(cuerpo, estilo.Render(marca)+estiloEtiquetaCampo.Render(e.Nombre)+"  "+e.Detalle)
		if e.Ayuda != "" {
			cuerpo = append(cuerpo, "    "+estiloAviso.Render("→ "+e.Ayuda))
		}
	}
	cuerpo = append(cuerpo, "", estiloAyuda.Render("Esc volver al menú"))
	lineas, _, _ := marco("Diagnóstico del entorno", cuerpo)
	return strings.Join(lineas, "\n")
}

// conMarco envuelve el cuerpo en el recuadro y desplaza las zonas para que sus
// coordenadas sean relativas al contenido final.
func (m Modelo) conMarco(titulo string, cuerpo []string, zonas []zona) (string, []zona) {
	lineas, df, dc := marco(titulo, cuerpo)
	desplazarZonas(zonas, df, dc, 0)
	return strings.Join(lineas, "\n"), zonas
}

// interfaz de compilación dinámica: evita que el compilador se queje de que
// Modelo no implementa tea.Model cuando se pasa por valor.
var _ tea.Model = Modelo{}
