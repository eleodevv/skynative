package configuracion

import "testing"

func TestValidarURLAceptaHTTPyHTTPS(t *testing.T) {
	validas := []string{
		"https://ejemplo.com",
		"http://localhost:3000",
		"https://sub.dominio.org/ruta?x=1#frag",
		"  https://espacios.example  ",
	}
	for _, u := range validas {
		if _, err := ValidarURL(u); err != nil {
			t.Errorf("ValidarURL(%q) devolvió error inesperado: %v", u, err)
		}
	}
}

func TestValidarURLRechazaEsquemasYVacios(t *testing.T) {
	invalidas := []string{
		"",
		"   ",
		"file:///etc/passwd",
		"javascript:alert(1)",
		"data:text/html,<h1>hola</h1>",
		"ftp://ejemplo.com",
		"ejemplo.com",
		"https://",
		"https://con\nsalto.com",
	}
	for _, u := range invalidas {
		if _, err := ValidarURL(u); err == nil {
			t.Errorf("ValidarURL(%q) debería haber devuelto error", u)
		}
	}
}

func TestValidarNombre(t *testing.T) {
	if _, err := ValidarNombre("  Mi Aplicación  "); err != nil {
		t.Errorf("nombre válido rechazado: %v", err)
	}
	invalidos := []string{"", "   ", "ruta/con/barra", "a|b", "c:d", "x*y", "muchos\nsaltos"}
	for _, n := range invalidos {
		if _, err := ValidarNombre(n); err == nil {
			t.Errorf("ValidarNombre(%q) debería haber devuelto error", n)
		}
	}
	if _, err := ValidarNombre(string(make([]rune, LongitudMaximaNombre+1))); err == nil {
		t.Error("un nombre demasiado largo debería fallar")
	}
}

func TestGenerarIdentificador(t *testing.T) {
	casos := map[string]string{
		"Mi Aplicación": "com.skynative.mi-aplicacion",
		"Ejemplo":       "com.skynative.ejemplo",
		"  ":            "com.skynative.app",
		"a--b":          "com.skynative.a-b",
	}
	for entrada, esperado := range casos {
		if obtenido := GenerarIdentificador(entrada); obtenido != esperado {
			t.Errorf("GenerarIdentificador(%q) = %q; se esperaba %q", entrada, obtenido, esperado)
		}
	}
}

func TestGenerarNombrePaquete(t *testing.T) {
	casos := map[string]string{
		"Mi Aplicación": "mi-aplicacion",
		"123app":        "app-123app",
		"Ñandú":         "nandu",
		"":              "app",
	}
	for entrada, esperado := range casos {
		if obtenido := GenerarNombrePaquete(entrada); obtenido != esperado {
			t.Errorf("GenerarNombrePaquete(%q) = %q; se esperaba %q", entrada, obtenido, esperado)
		}
	}
}

func TestValidarDimensiones(t *testing.T) {
	if err := ValidarDimensiones(1200, 800); err != nil {
		t.Errorf("dimensiones válidas rechazadas: %v", err)
	}
	if err := ValidarDimensiones(10, 800); err == nil {
		t.Error("ancho por debajo del mínimo debería fallar")
	}
	if err := ValidarDimensiones(1200, AnchoMaximo+1); err == nil {
		t.Error("alto por encima del máximo debería fallar")
	}
}

func TestNormalizarAplicaValoresPredeterminados(t *testing.T) {
	c := Nueva("https://ejemplo.com", "Ejemplo")
	c.Ancho = 0
	c.Alto = 0
	c.Version = ""
	c.Normalizar()

	if c.Ancho != AnchoPredeterminado || c.Alto != AltoPredeterminado {
		t.Errorf("dimensiones predeterminadas incorrectas: %dx%d", c.Ancho, c.Alto)
	}
	if c.Version != VersionPredeterminada {
		t.Errorf("versión predeterminada incorrecta: %q", c.Version)
	}
	if c.Identificador != "com.skynative.ejemplo" {
		t.Errorf("identificador inesperado: %q", c.Identificador)
	}
	if c.NombrePaquete != "ejemplo" {
		t.Errorf("nombre de paquete inesperado: %q", c.NombrePaquete)
	}
}

func TestValidarConfiguracionCompleta(t *testing.T) {
	c := Nueva("https://ejemplo.com", "Ejemplo")
	c.Normalizar()
	if err := c.Validar(); err != nil {
		t.Fatalf("configuración válida rechazada: %v", err)
	}

	c.URL = "javascript:alert(1)"
	if err := c.Validar(); err == nil {
		t.Error("configuración con URL peligrosa debería fallar")
	}
}

func TestValidarRechazaIdentificadorReservado(t *testing.T) {
	c := Nueva("https://ejemplo.com", "Ejemplo")
	c.Normalizar()
	c.Identificador = "com.tauri.dev"
	if err := c.Validar(); err == nil {
		t.Error("el identificador reservado com.tauri.dev debería fallar")
	}
}
