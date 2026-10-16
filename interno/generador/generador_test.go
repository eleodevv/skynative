package generador

import (
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/skynative/skynative/interno/configuracion"
)

func TestGenerarProyectoCompleto(t *testing.T) {
	base := t.TempDir()
	cfg := configuracion.Nueva("https://ejemplo.com/inicio", "Mi Aplicación")
	cfg.DirectorioSalida = base
	cfg.Ancho = 900
	cfg.Alto = 700

	res, err := Generar(cfg)
	if err != nil {
		t.Fatalf("Generar devolvió error: %v", err)
	}

	// Archivos mínimos de un proyecto Tauri v2.
	esperados := []string{
		filepath.Join(res.Directorio, "src-tauri", "Cargo.toml"),
		filepath.Join(res.Directorio, "src-tauri", "build.rs"),
		filepath.Join(res.Directorio, "src-tauri", "src", "main.rs"),
		filepath.Join(res.Directorio, "src-tauri", "tauri.conf.json"),
		filepath.Join(res.Directorio, "src-tauri", "capabilities", "default.json"),
		filepath.Join(res.Directorio, "src-tauri", "icons", "32x32.png"),
		filepath.Join(res.Directorio, "src-tauri", "icons", "icon.ico"),
		filepath.Join(res.Directorio, "src-tauri", "icons", "icon.icns"),
	}
	for _, ruta := range esperados {
		if _, err := os.Stat(ruta); err != nil {
			t.Errorf("falta el archivo esperado %q: %v", ruta, err)
		}
	}

	// La configuración de Tauri debe ser JSON válido con los valores pedidos.
	datos, err := os.ReadFile(res.RutaTauriConf)
	if err != nil {
		t.Fatalf("no se pudo leer tauri.conf.json: %v", err)
	}
	var conf map[string]any
	if err := json.Unmarshal(datos, &conf); err != nil {
		t.Fatalf("tauri.conf.json no es JSON válido: %v", err)
	}
	if conf["productName"] != "Mi Aplicación" {
		t.Errorf("productName = %v", conf["productName"])
	}
	if conf["identifier"] != "com.skynative.mi-aplicacion" {
		t.Errorf("identifier = %v", conf["identifier"])
	}
	build, _ := conf["build"].(map[string]any)
	if build["frontendDist"] != "https://ejemplo.com/inicio" {
		t.Errorf("frontendDist = %v", build["frontendDist"])
	}

	app, _ := conf["app"].(map[string]any)
	ventanas, _ := app["windows"].([]any)
	if len(ventanas) != 1 {
		t.Fatalf("se esperaba una ventana, hay %d", len(ventanas))
	}
	ventana, _ := ventanas[0].(map[string]any)
	if ventana["title"] != "Mi Aplicación" {
		t.Errorf("title = %v", ventana["title"])
	}
	if ventana["width"].(float64) != 900 || ventana["height"].(float64) != 700 {
		t.Errorf("dimensiones de ventana incorrectas: %v x %v", ventana["width"], ventana["height"])
	}
	if ventana["resizable"] != true {
		t.Errorf("resizable = %v", ventana["resizable"])
	}
	if app["withGlobalTauri"] != false {
		t.Errorf("withGlobalTauri debería ser false, es %v", app["withGlobalTauri"])
	}
}

func TestGenerarEscapaValoresEnJSON(t *testing.T) {
	base := t.TempDir()
	cfg := configuracion.Nueva("https://ejemplo.com", "Café & Té")
	cfg.DirectorioSalida = base

	res, err := Generar(cfg)
	if err != nil {
		t.Fatalf("Generar devolvió error: %v", err)
	}
	datos, _ := os.ReadFile(res.RutaTauriConf)
	var conf map[string]any
	if err := json.Unmarshal(datos, &conf); err != nil {
		t.Fatalf("JSON inválido con caracteres especiales: %v", err)
	}
	if conf["productName"] != "Café & Té" {
		t.Errorf("productName escapado incorrectamente: %v", conf["productName"])
	}
}

func TestGenerarIconosValidos(t *testing.T) {
	base := t.TempDir()
	cfg := configuracion.Nueva("https://ejemplo.com", "Iconos")
	cfg.DirectorioSalida = base

	res, err := Generar(cfg)
	if err != nil {
		t.Fatalf("Generar devolvió error: %v", err)
	}
	dirIconos := filepath.Join(res.Directorio, "src-tauri", "icons")

	archivo, err := os.Open(filepath.Join(dirIconos, "32x32.png"))
	if err != nil {
		t.Fatalf("no se pudo abrir el PNG: %v", err)
	}
	defer archivo.Close()
	img, err := png.Decode(archivo)
	if err != nil {
		t.Fatalf("el PNG generado no es válido: %v", err)
	}
	if img.Bounds().Dx() != 32 || img.Bounds().Dy() != 32 {
		t.Errorf("tamaño del PNG inesperado: %v", img.Bounds())
	}

	// Tauri exige que los PNG tengan canal alfa (color type 6 = RGBA).
	for _, nombre := range []string{"32x32.png", "128x128.png", "128x128@2x.png", "icon.png"} {
		crudo, err := os.ReadFile(filepath.Join(dirIconos, nombre))
		if err != nil {
			t.Fatalf("no se pudo leer %q: %v", nombre, err)
		}
		if len(crudo) < 26 || crudo[25] != 6 {
			t.Errorf("%s debería ser RGBA (color type 6)", nombre)
		}
	}

	ico, err := os.ReadFile(filepath.Join(dirIconos, "icon.ico"))
	if err != nil {
		t.Fatalf("no se pudo leer el ICO: %v", err)
	}
	if len(ico) < 6 || ico[0] != 0 || ico[1] != 0 || ico[2] != 1 {
		t.Error("cabecera ICO no válida")
	}

	icns, err := os.ReadFile(filepath.Join(dirIconos, "icon.icns"))
	if err != nil {
		t.Fatalf("no se pudo leer el ICNS: %v", err)
	}
	if string(icns[:4]) != "icns" {
		t.Error("cabecera ICNS no válida")
	}
}

func TestGenerarRechazaDirectorioNoVacio(t *testing.T) {
	base := t.TempDir()
	cfg := configuracion.Nueva("https://ejemplo.com", "Ocupado")
	cfg.DirectorioSalida = base
	cfg.Normalizar()

	destino := filepath.Join(base, cfg.NombrePaquete)
	if err := os.MkdirAll(destino, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destino, "existente.txt"), []byte("hola"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Generar(cfg); err == nil {
		t.Error("Generar debería fallar si el directorio no está vacío")
	}
}

func TestGenerarRechazaURLInvalida(t *testing.T) {
	cfg := configuracion.Nueva("file:///etc/passwd", "Peligro")
	cfg.DirectorioSalida = t.TempDir()
	if _, err := Generar(cfg); err == nil {
		t.Error("Generar debería rechazar esquemas no http/https")
	}
}
