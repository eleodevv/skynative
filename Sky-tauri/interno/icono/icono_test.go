package icono

import (
	"bytes"
	"context"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// mustParse analiza una URL y falla la prueba si no es válida.
func mustParse(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("URL no válida %q: %v", s, err)
	}
	return u
}

// pngCuadrado devuelve un PNG de lado n.
func pngCuadrado(t *testing.T, n int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("no se pudo codificar el PNG: %v", err)
	}
	return buf.Bytes()
}

// icoConPNG envuelve un PNG en un archivo ICO de una sola entrada.
func icoConPNG(imagen []byte, n int) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	lado := byte(n)
	if n >= 256 {
		lado = 0
	}
	buf.WriteByte(lado)
	buf.WriteByte(lado)
	buf.WriteByte(0)
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(32))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(imagen)))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(22))
	buf.Write(imagen)
	return buf.Bytes()
}

func TestDescargarDesdeEnlaceIcono(t *testing.T) {
	png := pngCuadrado(t, 64)
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><link rel="icon" href="/assets/logo.png"></head></html>`))
		case "/assets/logo.png":
			w.Write(png)
		default:
			http.NotFound(w, r)
		}
	}))
	defer servidor.Close()

	img, err := Descargar(context.Background(), servidor.URL)
	if err != nil {
		t.Fatalf("Descargar devolvió error: %v", err)
	}
	if img.Bounds().Dx() != 64 {
		t.Errorf("se esperaba una imagen de 64px, es %d", img.Bounds().Dx())
	}
}

func TestDescargarFaviconPorDefecto(t *testing.T) {
	ico := icoConPNG(pngCuadrado(t, 32), 32)
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			w.Write(ico)
			return
		}
		http.NotFound(w, r)
	}))
	defer servidor.Close()

	img, err := Descargar(context.Background(), servidor.URL)
	if err != nil {
		t.Fatalf("Descargar devolvió error: %v", err)
	}
	if img.Bounds().Dx() != 32 {
		t.Errorf("se esperaba una imagen de 32px, es %d", img.Bounds().Dx())
	}
}

func TestDescargarSinIcono(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer servidor.Close()

	if _, err := Descargar(context.Background(), servidor.URL); err == nil {
		t.Error("se esperaba un error al no haber icono")
	}
}

func TestDescargarURLInvalida(t *testing.T) {
	if _, err := Descargar(context.Background(), "ftp://ejemplo.com"); err == nil {
		t.Error("se esperaba un error con un esquema no soportado")
	}
}

func TestEnlacesIconoPrefiereAppleTouch(t *testing.T) {
	base := mustParse(t, "https://ejemplo.com/pagina")
	html := `<link rel="icon" href="/a.png" sizes="16x16">` +
		`<link rel="apple-touch-icon" href="/grande.png" sizes="180x180">` +
		`<link rel="icon" href="/b.png" sizes="32x32">`
	enlaces := enlacesIcono(base, html)
	if len(enlaces) != 3 {
		t.Fatalf("se esperaban 3 enlaces, hay %d", len(enlaces))
	}
	if enlaces[0] != "https://ejemplo.com/grande.png" {
		t.Errorf("el primero debería ser el apple-touch-icon, es %q", enlaces[0])
	}
	if enlaces[1] != "https://ejemplo.com/b.png" {
		t.Errorf("el segundo debería ser el icono de 32px, es %q", enlaces[1])
	}
}

func TestEscalar(t *testing.T) {
	img := Escalar(image.NewRGBA(image.Rect(0, 0, 16, 16)), 256)
	if img.Bounds().Dx() != 256 || img.Bounds().Dy() != 256 {
		t.Errorf("tamaño inesperado: %v", img.Bounds())
	}
}
