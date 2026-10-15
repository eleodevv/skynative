package generador

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// Los iconos se generan por código para que el proyecto sea reproducible y no
// dependa de archivos binarios incrustados. Se producen PNG (Linux), ICO
// (Windows) e ICNS (macOS).

// entradaIcono asocia un tamaño con su imagen codificada en PNG.
type entradaIcono struct {
	tamano int
	img    *image.RGBA
}

// entradaIcns asocia un tipo de bloque ICNS con su imagen.
type entradaIcns struct {
	tipo string
	img  *image.RGBA
}

// generarIconos crea el conjunto de iconos por defecto en la carpeta indicada
// y devuelve las rutas escritas.
func generarIconos(carpeta string) ([]string, error) {
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		return nil, fmt.Errorf("no se pudo crear la carpeta de iconos %q: %w", carpeta, err)
	}

	tamanos := []int{16, 32, 48, 64, 128, 256, 512}
	imgs := make(map[int]*image.RGBA, len(tamanos))
	for _, t := range tamanos {
		imgs[t] = dibujarIcono(t)
	}

	var archivos []string

	// PNG requeridos por el empaquetador de Linux y referenciados en la config.
	pngs := []struct {
		nombre string
		tamano int
	}{
		{"32x32.png", 32},
		{"128x128.png", 128},
		{"128x128@2x.png", 256},
		{"icon.png", 512},
	}
	for _, p := range pngs {
		datos, err := codificarPNG(imgs[p.tamano])
		if err != nil {
			return nil, err
		}
		ruta := filepath.Join(carpeta, p.nombre)
		if err := os.WriteFile(ruta, datos, 0o644); err != nil {
			return nil, fmt.Errorf("no se pudo escribir %q: %w", ruta, err)
		}
		archivos = append(archivos, ruta)
	}

	// ICO para Windows.
	ico, err := construirICO([]entradaIcono{
		{16, imgs[16]},
		{32, imgs[32]},
		{48, imgs[48]},
		{64, imgs[64]},
	})
	if err != nil {
		return nil, err
	}
	rutaIco := filepath.Join(carpeta, "icon.ico")
	if err := os.WriteFile(rutaIco, ico, 0o644); err != nil {
		return nil, fmt.Errorf("no se pudo escribir %q: %w", rutaIco, err)
	}
	archivos = append(archivos, rutaIco)

	// ICNS para macOS.
	icns, err := construirICNS([]entradaIcns{
		{"ic07", imgs[128]},
		{"ic08", imgs[256]},
		{"ic09", imgs[512]},
	})
	if err != nil {
		return nil, err
	}
	rutaIcns := filepath.Join(carpeta, "icon.icns")
	if err := os.WriteFile(rutaIcns, icns, 0o644); err != nil {
		return nil, fmt.Errorf("no se pudo escribir %q: %w", rutaIcns, err)
	}
	archivos = append(archivos, rutaIcns)

	return archivos, nil
}

// dibujarIcono dibuja un icono cuadrado con degradado azul-violeta y un
// triángulo blanco.
func dibujarIcono(n int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	f := float64(n)
	if f <= 0 {
		f = 1
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			t := (float64(x) + float64(y)) / (2 * f)
			img.Set(x, y, color.RGBA{
				R: uint8(26 + t*(124-26)),
				G: uint8(115 + t*(58-115)),
				B: uint8(232 + t*(237-232)),
				A: 255,
			})
		}
	}
	dibujarTriangulo(img)
	return img
}

// dibujarTriangulo rellena un triángulo blanco apuntando hacia arriba.
func dibujarTriangulo(img *image.RGBA) {
	b := img.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())
	a := [2]float64{0.50 * w, 0.24 * h}
	p1 := [2]float64{0.30 * w, 0.72 * h}
	p2 := [2]float64{0.70 * w, 0.72 * h}
	blanco := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if dentroTriangulo(float64(x)+0.5, float64(y)+0.5, a, p1, p2) {
				img.Set(x, y, blanco)
			}
		}
	}
}

// dentroTriangulo indica si el punto (px, py) está dentro del triángulo.
func dentroTriangulo(px, py float64, a, b, c [2]float64) bool {
	d1 := signo(px, py, a, b)
	d2 := signo(px, py, b, c)
	d3 := signo(px, py, c, a)
	tieneNegativo := d1 < 0 || d2 < 0 || d3 < 0
	tienePositivo := d1 > 0 || d2 > 0 || d3 > 0
	return !(tieneNegativo && tienePositivo)
}

// signo calcula un producto vectorial usado en la prueba de punto en triángulo.
func signo(x, y float64, p1, p2 [2]float64) float64 {
	return (x-p2[0])*(p1[1]-p2[1]) - (p1[0]-p2[0])*(y-p2[1])
}

// codificarPNG serializa una imagen a PNG.
func codificarPNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("no se pudo codificar el icono PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// construirICO genera un archivo ICO con imágenes PNG incrustadas.
func construirICO(entradas []entradaIcono) ([]byte, error) {
	datos := make([][]byte, len(entradas))
	for i, e := range entradas {
		png, err := codificarPNG(e.img)
		if err != nil {
			return nil, err
		}
		datos[i] = png
	}

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0)) // reservado
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // tipo: icono
	_ = binary.Write(&buf, binary.LittleEndian, uint16(len(entradas)))

	desplazamiento := 6 + 16*len(entradas)
	for i, e := range entradas {
		lado := byte(e.tamano)
		if e.tamano >= 256 {
			lado = 0
		}
		buf.WriteByte(lado)                                     // ancho
		buf.WriteByte(lado)                                     // alto
		buf.WriteByte(0)                                        // colores de paleta
		buf.WriteByte(0)                                        // reservado
		_ = binary.Write(&buf, binary.LittleEndian, uint16(1))  // planos
		_ = binary.Write(&buf, binary.LittleEndian, uint16(32)) // bits por píxel
		_ = binary.Write(&buf, binary.LittleEndian, uint32(len(datos[i])))
		_ = binary.Write(&buf, binary.LittleEndian, uint32(desplazamiento))
		desplazamiento += len(datos[i])
	}
	for _, d := range datos {
		buf.Write(d)
	}
	return buf.Bytes(), nil
}

// construirICNS genera un archivo ICNS con imágenes PNG incrustadas.
func construirICNS(entradas []entradaIcns) ([]byte, error) {
	datos := make([][]byte, len(entradas))
	total := 8
	for i, e := range entradas {
		if len(e.tipo) != 4 {
			return nil, fmt.Errorf("tipo de bloque ICNS no válido %q", e.tipo)
		}
		png, err := codificarPNG(e.img)
		if err != nil {
			return nil, err
		}
		datos[i] = png
		total += 8 + len(png)
	}

	var buf bytes.Buffer
	buf.WriteString("icns")
	_ = binary.Write(&buf, binary.BigEndian, uint32(total))
	for i, e := range entradas {
		buf.WriteString(e.tipo)
		_ = binary.Write(&buf, binary.BigEndian, uint32(8+len(datos[i])))
		buf.Write(datos[i])
	}
	return buf.Bytes(), nil
}
