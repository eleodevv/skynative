package tui

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/skynative/sky-tauri/interno/configuracion"
	"github.com/skynative/sky-tauri/interno/generador"
	"github.com/skynative/sky-tauri/interno/proyectos"
)

// TestProgramaArrancaYSale demuestra que el bucle de Bubble Tea se inicia y
// termina al pulsar «q», sin necesidad de una terminal real.
func TestProgramaArrancaYSale(t *testing.T) {
	var salida bytes.Buffer
	if err := ejecutarConEntradaSalida(strings.NewReader("q"), &salida); err != nil {
		t.Fatalf("el programa no terminó correctamente: %v", err)
	}
	if !strings.Contains(salida.String(), "Crear aplicación") {
		t.Errorf("la salida no contenía el menú principal:\n%s", salida.String())
	}
}

// TestInstalarTrasCompilar comprueba que, con un ejecutable y un icono listos,
// la instalación crea la entrada de menú del sistema.
func TestInstalarTrasCompilar(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("la instalación directa a prueba se valida en Linux")
	}
	datos := t.TempDir()
	t.Setenv("XDG_DATA_HOME", datos)

	proyecto := t.TempDir()
	exe := filepath.Join(proyecto, "src-tauri", "target", "release", "prueba")
	if err := os.MkdirAll(filepath.Dir(exe), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	icono := filepath.Join(proyecto, "src-tauri", "icons", "128x128.png")
	if err := os.MkdirAll(filepath.Dir(icono), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(icono, []byte("PNG"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := NuevoModelo()
	m.cfg = configuracion.Nueva("https://ejemplo.com", "Prueba")
	m.cfg.Normalizar()
	m.rutaProyecto = proyecto
	m.rutaEjecutable = exe
	m.exito = true
	m.puedeInstalar = true

	m, cmd := m.iniciarInstalacion()
	if m.estado != estadoProgreso || !m.instalandoAhora {
		t.Fatalf("no se inició la instalación, estado = %v", m.estado)
	}
	var mensaje tea.Msg
	for _, candidato := range desglosar(cmd) {
		if _, ok := candidato.(msgInstalado); ok {
			mensaje = candidato
		}
	}
	if mensaje == nil {
		t.Fatal("no se generó el mensaje de instalación")
	}
	m = aplicar(m, mensaje)

	if !m.instalado || m.estado != estadoResultado {
		t.Fatalf("no se marcó como instalada: instalado=%v estado=%v (%s)", m.instalado, m.estado, m.resumen)
	}
	if _, err := os.Stat(m.entradaEscritorio); err != nil {
		t.Errorf("no se creó la entrada .desktop: %v", err)
	}
}

func runas(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func tecla(tipo tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: tipo}
}

func aplicar(m Modelo, msg tea.Msg) Modelo {
	actualizado, _ := m.Update(msg)
	return actualizado.(Modelo)
}

// desglosar ejecuta un comando y devuelve todos los mensajes que produce,
// aplanando los lotes que genera tea.Batch.
func desglosar(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if lote, ok := msg.(tea.BatchMsg); ok {
		var mensajes []tea.Msg
		for _, sub := range lote {
			mensajes = append(mensajes, desglosar(sub)...)
		}
		return mensajes
	}
	return []tea.Msg{msg}
}

func TestMenuNavegacion(t *testing.T) {
	m := NuevoModelo()
	if m.estado != estadoMenu {
		t.Fatalf("el estado inicial debería ser el menú, es %v", m.estado)
	}

	m = aplicar(m, tecla(tea.KeyDown))
	if m.opcionMenu != 1 {
		t.Errorf("bajar debería poner la opción en 1, es %d", m.opcionMenu)
	}

	m = aplicar(m, runas("3"))
	if m.opcionMenu != 2 {
		t.Errorf("la tecla 3 debería seleccionar la opción 2, es %d", m.opcionMenu)
	}

	m = aplicar(m, tecla(tea.KeyUp))
	if m.opcionMenu != 1 {
		t.Errorf("subir debería poner la opción en 1, es %d", m.opcionMenu)
	}
}

// TestOpcionInstalarHerramientas comprueba que la opción del menú prepara la
// instalación de dependencias externas (o la petición de sudo).
func TestOpcionInstalarHerramientas(t *testing.T) {
	m := NuevoModelo()
	m.opcionMenu = 1
	actualizado, cmd := m.ejecutarOpcionMenu()
	m = actualizado.(Modelo)
	if m.estado != estadoProgreso {
		t.Fatalf("debería pasar a progreso, estado = %v", m.estado)
	}
	if cmd == nil {
		t.Fatal("debería devolverse un comando")
	}
}

// TestIniciarInstalacionHerramientas comprueba el arranque del instalador.
func TestIniciarInstalacionHerramientas(t *testing.T) {
	m := NuevoModelo()
	m, _ = m.iniciarInstalacionHerramientas()
	if m.estado != estadoProgreso || !m.instalandoHerramientas {
		t.Fatalf("debería iniciar la instalación de herramientas, estado = %v", m.estado)
	}
}

// TestChequeoHerramientasMarcaEstado verifica la comprobación en segundo plano.
func TestChequeoHerramientasMarcaEstado(t *testing.T) {
	m := NuevoModelo()
	m = aplicar(m, msgChequeoHerramientas{})
	if !m.chequeoHecho {
		t.Error("el chequeo debería marcarse como hecho")
	}
	if !m.herramientasListas {
		t.Error("sin estados críticos fallidos las herramientas deberían estar listas")
	}
}

// TestHerramientasListasCompilaAutomaticamente comprueba que, si el entorno
// está preparado, tras generar se encadena la compilación.
func TestHerramientasListasCompilaAutomaticamente(t *testing.T) {
	m := NuevoModelo()
	m.herramientasListas = true
	m.cfg = configuracion.Nueva("https://ejemplo.com", "Prueba")
	m.cfg.Normalizar()
	m.estado = estadoProgreso
	m.rutaProyecto = t.TempDir()

	m = aplicar(m, msgGenerado{res: generador.Resultado{Directorio: m.rutaProyecto, NombrePaquete: "prueba"}})
	if m.estado != estadoProgreso || !m.compilandoAhora {
		t.Fatalf("debería encadenar la compilación automática, estado = %v, compilando = %v", m.estado, m.compilandoAhora)
	}
}

func TestOpcionVersionNoPermiteCompilar(t *testing.T) {
	m := NuevoModelo()
	m.opcionMenu = 4
	m = aplicar(m, tecla(tea.KeyEnter))

	if m.estado != estadoResultado {
		t.Fatalf("debería mostrarse el resultado, estado = %v", m.estado)
	}
	if !strings.Contains(m.resumen, "Skynative") {
		t.Errorf("el resumen debería mostrar la versión: %q", m.resumen)
	}
	if m.puedeCompilar {
		t.Error("no debería ofrecerse compilar en la pantalla de versión")
	}
}

func TestFormularioRechazaURLPeligrosa(t *testing.T) {
	m := NuevoModelo()
	m.estado = estadoFormulario
	m.url.SetValue("javascript:alert(1)")
	m.nombre.SetValue("Prueba")
	m.carpeta = t.TempDir()

	m, _ = m.lanzarGeneracion()

	if m.mensajeError == "" {
		t.Error("una URL con esquema peligroso debería producir un mensaje de error")
	}
	if m.estado != estadoFormulario {
		t.Errorf("debería permanecer en el formulario, estado = %v", m.estado)
	}
}

func TestFormularioValidaDimensiones(t *testing.T) {
	m := NuevoModelo()
	m.estado = estadoFormulario
	m.url.SetValue("https://ejemplo.com")
	m.nombre.SetValue("Prueba")
	m.anchoCampo.SetValue("10") // por debajo del mínimo
	m.carpeta = t.TempDir()

	m, _ = m.lanzarGeneracion()
	if m.mensajeError == "" {
		t.Error("unas dimensiones fuera de rango deberían producir un error")
	}
}

func TestFormularioEscVuelveAlMenu(t *testing.T) {
	m := NuevoModelo()
	m.estado = estadoFormulario
	m = aplicar(m, tecla(tea.KeyEsc))
	if m.estado != estadoMenu {
		t.Errorf("Esc debería volver al menú, estado = %v", m.estado)
	}
}

func TestFormularioTabRecorreCampos(t *testing.T) {
	m := NuevoModelo()
	m.estado = estadoFormulario
	if m.campo != campoURL {
		t.Fatalf("el campo inicial debería ser la URL, es %d", m.campo)
	}
	m = aplicar(m, tecla(tea.KeyTab))
	if m.campo != campoNombre {
		t.Errorf("Tab debería avanzar a Nombre, es %d", m.campo)
	}
	m = aplicar(m, tecla(tea.KeyShiftTab))
	if m.campo != campoURL {
		t.Errorf("Shift+Tab debería volver a la URL, es %d", m.campo)
	}
}

// TestGeneracionExtremoAExtremo comprueba el flujo real: el formulario arranca
// la generación y, al recibir el mensaje de resultado, el modelo pasa a la
// pantalla de resultado con el proyecto en disco.
func TestGeneracionExtremoAExtremo(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())

	carpeta := t.TempDir()
	m := NuevoModelo()
	m.estado = estadoFormulario
	m.url.SetValue("https://ejemplo.com")
	m.nombre.SetValue("Prueba TUI")
	m.anchoCampo.SetValue("820")
	m.altoCampo.SetValue("640")
	m.carpeta = carpeta

	m, _ = m.lanzarGeneracion()
	if m.estado != estadoProgreso {
		t.Fatalf("debería pasar a progreso, estado = %v (error: %s)", m.estado, m.mensajeError)
	}
	esperado := filepath.Join(carpeta, "prueba-tui")
	if m.rutaProyecto != esperado {
		t.Errorf("ruta = %q; se esperaba %q", m.rutaProyecto, esperado)
	}

	// Ejecución real del generador.
	cfg := configuracion.Nueva("https://ejemplo.com", "Prueba TUI")
	cfg.DirectorioSalida = carpeta
	cfg.Ancho, cfg.Alto = 820, 640
	mensaje := generarCmd(cfg)()
	m = aplicar(m, mensaje)

	if m.estado != estadoResultado {
		t.Fatalf("debería pasar a resultado, estado = %v", m.estado)
	}
	if !m.exito {
		t.Fatalf("la generación debería haber sido correcta: %s", m.resumen)
	}
	if _, err := os.Stat(filepath.Join(esperado, "src-tauri", "tauri.conf.json")); err != nil {
		t.Errorf("no se creó tauri.conf.json: %v", err)
	}
	if !m.puedeCompilar {
		t.Error("tras generar debería poderse compilar")
	}
}

func TestVistasNoVacias(t *testing.T) {
	estados := []estado{
		estadoMenu, estadoFormulario, estadoExplorador,
		estadoProgreso, estadoResultado, estadoLista, estadoDiagnostico,
	}
	for _, e := range estados {
		m := NuevoModelo()
		m.estado = e
		m.explorador = nuevoExplorador(t.TempDir())
		if strings.TrimSpace(m.View()) == "" {
			t.Errorf("la vista del estado %v está vacía", e)
		}
	}
}

// TestEliminarProyectoDesdeLista comprueba la confirmación y el borrado del
// registro desde la pantalla de proyectos.
func TestEliminarProyectoDesdeLista(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	m := NuevoModelo()
	m.proyectos = []proyectos.Proyecto{{Nombre: "Uno", Ruta: "/tmp/uno"}}
	m.estado = estadoLista

	m = aplicar(m, runas("d"))
	if !m.confirmandoEliminar {
		t.Fatal("D debería pedir confirmación")
	}
	m = aplicar(m, runas("n"))
	if m.confirmandoEliminar {
		t.Fatal("N debería cancelar la eliminación")
	}

	// Un segundo intento con confirmación devuelve el comando de borrado.
	m = aplicar(m, runas("d"))
	_, cmd := m.Update(runas("s"))
	if cmd == nil {
		t.Fatal("confirmar debería devolver un comando de borrado")
	}
}

// TestEliminarTodoDesdeLista comprueba el borrado completo del registro.
func TestEliminarTodoDesdeLista(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	m := NuevoModelo()
	m.proyectos = []proyectos.Proyecto{
		{Nombre: "Uno", Ruta: filepath.Join(t.TempDir(), "uno")},
		{Nombre: "Dos", Ruta: filepath.Join(t.TempDir(), "dos")},
	}
	m.estado = estadoLista

	m = aplicar(m, runas("A"))
	if !m.confirmandoEliminarTodo {
		t.Fatal("A debería pedir confirmación para eliminar todo")
	}
	_, cmd := m.Update(runas("s"))
	if cmd == nil {
		t.Fatal("confirmar debería devolver un comando de borrado total")
	}
}

// TestListaNavegacionNoSaleDeRango comprueba que la selección no se pasa.
func TestListaNavegacionNoSaleDeRango(t *testing.T) {
	m := NuevoModelo()
	m.proyectos = []proyectos.Proyecto{{Nombre: "Uno", Ruta: "/tmp/uno"}}
	m.estado = estadoLista
	m = aplicar(m, tecla(tea.KeyDown))
	if m.indiceProyecto != 0 {
		t.Errorf("el índice no debería pasar del final, es %d", m.indiceProyecto)
	}
	m = aplicar(m, tecla(tea.KeyUp))
	if m.indiceProyecto != 0 {
		t.Errorf("el índice no debería ser negativo, es %d", m.indiceProyecto)
	}
}

// TestZonasMenuAlineadas comprueba que cada zona del menú apunta a su fila.
func TestZonasMenuAlineadas(t *testing.T) {
	m := NuevoModelo()
	m.ancho, m.alto = 120, 40
	contenido, zonas := m.bloqueActual()
	lineas := strings.Split(contenido, "\n")
	for _, z := range zonas {
		if z.accion != zonaOpcionMenu {
			continue
		}
		if z.fila0 >= len(lineas) || !strings.Contains(lineas[z.fila0], opcionesMenu[z.indice]) {
			t.Errorf("la zona del menú %d no apunta a su opción", z.indice)
		}
	}
}

// TestZonasListaAlineadas comprueba que cada zona de proyecto apunta a su fila.
func TestZonasListaAlineadas(t *testing.T) {
	m := NuevoModelo()
	m.ancho, m.alto = 120, 40
	m.estado = estadoLista
	m.proyectos = []proyectos.Proyecto{{Nombre: "Alfa", Ruta: "/tmp/a"}, {Nombre: "Beta", Ruta: "/tmp/b"}}
	contenido, zonas := m.bloqueActual()
	lineas := strings.Split(contenido, "\n")
	for _, z := range zonas {
		if z.accion != zonaProyecto {
			continue
		}
		if z.fila0 >= len(lineas) || !strings.Contains(lineas[z.fila0], m.proyectos[z.indice].Nombre) {
			t.Errorf("la zona del proyecto %d no apunta a su fila", z.indice)
		}
	}
}

// clicEn simula un clic del ratón sobre la primera zona que coincide con la
// acción e índice indicados.
func clicEn(t *testing.T, m Modelo, accion accionZona, indice int) (tea.Model, tea.Cmd) {
	t.Helper()
	if m.ancho == 0 {
		m.ancho = 120
	}
	if m.alto == 0 {
		m.alto = 40
	}
	contenido, zonas := m.bloqueActual()
	x, y := desplazamientoCentrado(m.ancho, m.alto, contenido)
	for _, z := range zonas {
		if z.accion == accion && z.indice == indice {
			msg := tea.MouseMsg{X: x + z.col0 + 1, Y: y + z.fila0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
			return m.actualizarRaton(msg)
		}
	}
	t.Fatalf("no se encontró ninguna zona con acción %d e índice %d", accion, indice)
	return m, nil
}

// TestRatonMenuAbreFormulario comprueba que se puede crear una aplicación con
// un clic en el menú.
func TestRatonMenuAbreFormulario(t *testing.T) {
	m := NuevoModelo()
	actualizado, _ := clicEn(t, m, zonaOpcionMenu, 0)
	if actualizado.(Modelo).estado != estadoFormulario {
		t.Errorf("un clic en «Crear aplicación» debería abrir el formulario, estado = %v", actualizado.(Modelo).estado)
	}
}

// TestRatonMenuListaProyectos comprueba que la opción de listar responde al clic.
func TestRatonMenuListaProyectos(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	m := NuevoModelo()
	_, cmd := clicEn(t, m, zonaOpcionMenu, 2)
	if cmd == nil {
		t.Error("un clic en «Listar proyectos» debería cargar el registro")
	}
}

// TestRatonSeleccionaProyecto comprueba que un clic en una fila la selecciona.
func TestRatonSeleccionaProyecto(t *testing.T) {
	m := NuevoModelo()
	m.proyectos = []proyectos.Proyecto{
		{Nombre: "Uno", Ruta: "/tmp/uno"},
		{Nombre: "Dos", Ruta: "/tmp/dos"},
	}
	m.estado = estadoLista
	actualizado, _ := clicEn(t, m, zonaProyecto, 1)
	if actualizado.(Modelo).indiceProyecto != 1 {
		t.Errorf("debería seleccionar el segundo proyecto, índice = %d", actualizado.(Modelo).indiceProyecto)
	}
}

// TestRatonEliminarProyecto comprueba el flujo de borrado con el ratón.
func TestRatonEliminarProyecto(t *testing.T) {
	t.Setenv("SKYNATIVE_DIRECTORIO_CONFIG", t.TempDir())
	m := NuevoModelo()
	m.proyectos = []proyectos.Proyecto{{Nombre: "Uno", Ruta: "/tmp/uno"}}
	m.estado = estadoLista

	actualizado, _ := clicEn(t, m, zonaBotonEliminar, 0)
	m = actualizado.(Modelo)
	if !m.confirmandoEliminar {
		t.Fatal("el botón «Eliminar seleccionado» debería pedir confirmación")
	}
	_, cmd := clicEn(t, m, zonaBotonConfirmarSi, 0)
	if cmd == nil {
		t.Fatal("confirmar con el ratón debería devolver el comando de borrado")
	}
}

func TestEnteroPositivo(t *testing.T) {
	if v, err := enteroPositivo("", 800); err != nil || v != 800 {
		t.Errorf("valor vacío debería usar el predeterminado: %d, %v", v, err)
	}
	if v, err := enteroPositivo("1024", 800); err != nil || v != 1024 {
		t.Errorf("valor numérico mal interpretado: %d, %v", v, err)
	}
	if _, err := enteroPositivo("abc", 800); err == nil {
		t.Error("un valor no numérico debería fallar")
	}
}

func TestBarraProgreso(t *testing.T) {
	if got := barraProgreso(0, 20); !strings.Contains(got, strings.Repeat("░", 20)) {
		t.Errorf("barra al 0%% inesperada: %q", got)
	}
	if got := barraProgreso(1, 20); !strings.Contains(got, strings.Repeat("█", 20)) {
		t.Errorf("barra al 100%% inesperada: %q", got)
	}
	if got := barraProgreso(0.5, 20); !strings.Contains(got, strings.Repeat("█", 10)+strings.Repeat("░", 10)) {
		t.Errorf("barra al 50%% inesperada: %q", got)
	}
}

func TestFormatearDuracion(t *testing.T) {
	if got := formatearDuracion(45 * time.Second); got != "45s" {
		t.Errorf("45s formateado como %q", got)
	}
	if got := formatearDuracion(92 * time.Second); got != "1m32s" {
		t.Errorf("92s formateado como %q", got)
	}
}

func TestProgresoActualFraccion(t *testing.T) {
	m := NuevoModelo()
	if frac, _, _ := m.progresoActual(); frac != 0 {
		t.Errorf("sin estimación la fracción debe ser 0, fue %v", frac)
	}
	m.progresoEstimado = time.Minute
	m.inicioProgreso = time.Now().Add(-10 * time.Minute)
	if frac, _, _ := m.progresoActual(); frac != 0.95 {
		t.Errorf("la fracción debe toparse en 0.95, fue %v", frac)
	}
}

func TestLineaProgreso(t *testing.T) {
	m := NuevoModelo()
	m.inicioProgreso = time.Now().Add(-1 * time.Minute)
	m.progresoEstimado = 4 * time.Minute
	linea := m.lineaProgreso()
	if !strings.Contains(linea, "transcurridos") || !strings.Contains(linea, "quedan") {
		t.Errorf("la línea de progreso no incluye los tiempos: %q", linea)
	}
	m.inicioProgreso = time.Now().Add(-10 * time.Minute)
	if linea := m.lineaProgreso(); !strings.Contains(linea, "finalizando") {
		t.Errorf("esperaba «finalizando» al superar la estimación: %q", linea)
	}
}
