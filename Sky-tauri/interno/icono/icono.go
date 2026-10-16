// Package icono obtiene el icono «oficial» de un sitio web para usarlo como
// icono de la aplicación generada.
//
// La estrategia es la habitual de los navegadores: se busca un
// <link rel="icon"> (prefiriendo apple-touch-icon y los tamaños grandes) en el
// HTML de la página y, si no aparece, se prueba con /favicon.ico. Se admiten
// los formatos PNG, JPEG, GIF e ICO (incluidos los ICO con PNG incrustado o
// mapa de bits).
package icono

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	xdraw "golang.org/x/image/draw"
)

// maximoPagina limita cuánto HTML se lee para no descargar páginas enormes.
const maximoPagina = 512 << 10

// maximoIcono limita el tamaño del icono descargado.
const maximoIcono = 4 << 20

// Descargar busca y descarga el icono del sitio indicado. Devuelve un error si
// no encuentra ningún icono descargable.
func Descargar(ctx context.Context, paginaURL string) (image.Image, error) {
	base, err := url.Parse(strings.TrimSpace(paginaURL))
	if err != nil {
		return nil, fmt.Errorf("URL no válida: %w", err)
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return nil, fmt.Errorf("solo se admiten URLs http o https")
	}

	cliente := &http.Client{Timeout: 15 * time.Second}
	if ctx == nil {
		ctx = context.Background()
	}

	var candidatos []string
	if html, err := descargarTexto(ctx, cliente, base.String()); err == nil {
		candidatos = enlacesIcono(base, string(html))
	}
	// Reserva: el favicon clásico de la raíz del sitio.
	candidatos = append(candidatos, (&url.URL{Scheme: base.Scheme, Host: base.Host, Path: "/favicon.ico"}).String())

	var ultimoErr error
	for _, c := range deduplicar(candidatos) {
		datos, err := descargarBytes(ctx, cliente, c, maximoIcono)
		if err != nil {
			ultimoErr = err
			continue
		}
		img, err := imagenDesdeBytes(datos)
		if err != nil {
			ultimoErr = err
			continue
		}
		return img, nil
	}
	if ultimoErr == nil {
		ultimoErr = fmt.Errorf("no se encontró ningún icono")
	}
	return nil, ultimoErr
}

// Escalar redimensiona una imagen a un cuadrado de lado n con buena calidad.
func Escalar(img image.Image, n int) *image.RGBA {
	if n <= 0 {
		n = 1
	}
	destino := image.NewRGBA(image.Rect(0, 0, n, n))
	xdraw.CatmullRom.Scale(destino, destino.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return destino
}

// enlacesIcono extrae los enlaces de icono del HTML, ordenados por preferencia.
func enlacesIcono(base *url.URL, html string) []string {
	tipo := regexp.MustCompile(`(?i)<link\b[^>]*>`)
	atributo := func(tag, nombre string) string {
		re := regexp.MustCompile(`(?i)\b` + nombre + `\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`)
		if m := re.FindStringSubmatch(tag); m != nil {
			return primeroNoVacio(m[2], m[3], m[4])
		}
		return ""
	}

	type candidato struct {
		href      string
		prioridad int
		tamano    int
	}
	var encontrados []candidato
	for _, tag := range tipo.FindAllString(html, -1) {
		rel := strings.ToLower(atributo(tag, "rel"))
		if !strings.Contains(rel, "icon") {
			continue
		}
		href := strings.TrimSpace(atributo(tag, "href"))
		if href == "" || strings.HasPrefix(href, "data:") {
			continue
		}
		prioridad := 1
		switch {
		case strings.Contains(rel, "apple-touch-icon"):
			prioridad = 3
		case strings.Contains(rel, "mask-icon"):
			prioridad = 0
		}
		encontrados = append(encontrados, candidato{
			href:      resolver(base, href),
			prioridad: prioridad,
			tamano:    mayorTamano(atributo(tag, "sizes")),
		})
	}

	sort.SliceStable(encontrados, func(i, j int) bool {
		if encontrados[i].prioridad != encontrados[j].prioridad {
			return encontrados[i].prioridad > encontrados[j].prioridad
		}
		return encontrados[i].tamano > encontrados[j].tamano
	})

	var salida []string
	for _, c := range encontrados {
		salida = append(salida, c.href)
	}
	return salida
}

// mayorTamano interpreta el atributo sizes («32x32 16x16») y devuelve el mayor.
func mayorTamano(sizes string) int {
	mayor := 0
	for _, parte := range strings.Fields(sizes) {
		if i := strings.IndexByte(parte, 'x'); i > 0 {
			if n, err := strconv.Atoi(parte[:i]); err == nil && n > mayor {
				mayor = n
			}
		}
	}
	return mayor
}

// resolver convierte un href relativo en absoluto usando la URL base.
func resolver(base *url.URL, href string) string {
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

// deduplicar elimina cadenas repetidas conservando el orden.
func deduplicar(valores []string) []string {
	vistos := make(map[string]bool, len(valores))
	salida := make([]string, 0, len(valores))
	for _, v := range valores {
		if v == "" || vistos[v] {
			continue
		}
		vistos[v] = true
		salida = append(salida, v)
	}
	return salida
}

// primeroNoVacio devuelve el primer valor no vacío.
func primeroNoVacio(valores ...string) string {
	for _, v := range valores {
		if v != "" {
			return v
		}
	}
	return ""
}

// descargarTexto descarga el HTML de una página (limitado).
func descargarTexto(ctx context.Context, cliente *http.Client, destino string) ([]byte, error) {
	return descargarBytes(ctx, cliente, destino, maximoPagina)
}

// descargarBytes hace una petición GET y devuelve como mucho limite bytes.
func descargarBytes(ctx context.Context, cliente *http.Client, destino string, limite int64) ([]byte, error) {
	peticion, err := http.NewRequestWithContext(ctx, http.MethodGet, destino, nil)
	if err != nil {
		return nil, err
	}
	peticion.Header.Set("User-Agent", "Skynative/1.0 (+https://github.com/eleodevv/skynative)")
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return nil, err
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode < 200 || respuesta.StatusCode >= 300 {
		return nil, fmt.Errorf("respuesta HTTP %d", respuesta.StatusCode)
	}
	return io.ReadAll(io.LimitReader(respuesta.Body, limite))
}

// imagenDesdeBytes decodifica PNG, JPEG, GIF o ICO.
func imagenDesdeBytes(datos []byte) (image.Image, error) {
	if len(datos) >= 8 && bytes.Equal(datos[:8], []byte("\x89PNG\r\n\x1a\n")) {
		return png.Decode(bytes.NewReader(datos))
	}
	if len(datos) >= 4 && datos[0] == 0 && datos[1] == 0 && datos[2] == 1 && datos[3] == 0 {
		return decodificarICO(datos)
	}
	img, _, err := image.Decode(bytes.NewReader(datos))
	if err != nil {
		return nil, fmt.Errorf("formato de imagen no reconocido: %w", err)
	}
	return img, nil
}

// entradaICO describe una imagen dentro de un archivo ICO.
type entradaICO struct {
	ancho, alto int
	offset      int
	tamano      int
}

// decodificarICO lee un archivo ICO y devuelve su imagen de mayor resolución.
func decodificarICO(datos []byte) (image.Image, error) {
	if len(datos) < 6 {
		return nil, fmt.Errorf("ICO truncado")
	}
	cantidad := int(binary.LittleEndian.Uint16(datos[4:6]))
	if cantidad <= 0 || len(datos) < 6+16*cantidad {
		return nil, fmt.Errorf("ICO sin entradas válidas")
	}

	entradas := make([]entradaICO, 0, cantidad)
	for i := 0; i < cantidad; i++ {
		base := 6 + 16*i
		ancho := int(datos[base])
		alto := int(datos[base+1])
		if ancho == 0 {
			ancho = 256
		}
		if alto == 0 {
			alto = 256
		}
		tamano := int(binary.LittleEndian.Uint32(datos[base+8 : base+12]))
		offset := int(binary.LittleEndian.Uint32(datos[base+12 : base+16]))
		if offset < 0 || tamano <= 0 || offset+tamano > len(datos) {
			continue
		}
		entradas = append(entradas, entradaICO{ancho: ancho, alto: alto, offset: offset, tamano: tamano})
	}
	if len(entradas) == 0 {
		return nil, fmt.Errorf("ICO sin imágenes legibles")
	}
	sort.SliceStable(entradas, func(i, j int) bool {
		return entradas[i].ancho*entradas[i].alto > entradas[j].ancho*entradas[j].alto
	})

	var ultimoErr error
	for _, e := range entradas {
		bloque := datos[e.offset : e.offset+e.tamano]
		if len(bloque) >= 8 && bytes.Equal(bloque[:8], []byte("\x89PNG\r\n\x1a\n")) {
			if img, err := png.Decode(bytes.NewReader(bloque)); err == nil {
				return img, nil
			}
			ultimoErr = fmt.Errorf("PNG incrustado no válido")
			continue
		}
		if img, err := decodificarDIB(bloque); err == nil {
			return img, nil
		} else {
			ultimoErr = err
		}
	}
	if ultimoErr == nil {
		ultimoErr = fmt.Errorf("no se pudo decodificar el ICO")
	}
	return nil, ultimoErr
}

// decodificarDIB decodifica el mapa de bits de una entrada ICO (BITMAPINFOHEADER).
func decodificarDIB(b []byte) (*image.RGBA, error) {
	if len(b) < 40 {
		return nil, fmt.Errorf("mapa de bits truncado")
	}
	tamanoCabecera := int(binary.LittleEndian.Uint32(b[0:4]))
	ancho := int(int32(binary.LittleEndian.Uint32(b[4:8])))
	altoCompleto := int(int32(binary.LittleEndian.Uint32(b[8:12])))
	bits := int(binary.LittleEndian.Uint16(b[14:16]))
	coloresUsados := int(binary.LittleEndian.Uint32(b[32:36]))

	if ancho <= 0 || altoCompleto <= 0 || tamanoCabecera < 40 || tamanoCabecera > len(b) {
		return nil, fmt.Errorf("cabecera DIB no válida")
	}
	alto := altoCompleto / 2 // la altura incluye el mapa AND.
	if alto <= 0 {
		alto = altoCompleto
	}
	if ancho > 4096 || alto > 4096 {
		return nil, fmt.Errorf("mapa de bits demasiado grande")
	}

	paletaCantidad := 0
	paleta := make([]color.RGBA, 0)
	switch bits {
	case 32, 24:
		// sin paleta
	case 8, 4, 1:
		if coloresUsados == 0 {
			coloresUsados = 1 << uint(bits)
		}
		paletaCantidad = coloresUsados
	default:
		return nil, fmt.Errorf("profundidad de bits no soportada: %d", bits)
	}

	pos := tamanoCabecera
	for i := 0; i < paletaCantidad; i++ {
		if pos+4 > len(b) {
			return nil, fmt.Errorf("paleta DIB truncada")
		}
		paleta = append(paleta, color.RGBA{R: b[pos+2], G: b[pos+1], B: b[pos], A: 255})
		pos += 4
	}

	fila := ((ancho*bs(bits) + 31) / 32) * 4
	mascaraFila := ((ancho + 31) / 32) * 4
	posMascara := pos + fila*alto
	if pos+fila*alto > len(b) {
		return nil, fmt.Errorf("datos DIB truncados")
	}

	img := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	for y := 0; y < alto; y++ {
		origen := pos + (alto-1-y)*fila
		for x := 0; x < ancho; x++ {
			var c color.RGBA
			switch bits {
			case 32:
				i := origen + x*4
				c = color.RGBA{R: b[i+2], G: b[i+1], B: b[i], A: b[i+3]}
			case 24:
				i := origen + x*3
				c = color.RGBA{R: b[i+2], G: b[i+1], B: b[i], A: 255}
			case 8:
				idx := int(b[origen+x])
				if idx < len(paleta) {
					c = paleta[idx]
				}
			case 4:
				val := b[origen+x/2]
				idx := int(val >> 4)
				if x%2 == 1 {
					idx = int(val & 0x0f)
				}
				if idx < len(paleta) {
					c = paleta[idx]
				}
			case 1:
				idx := int((b[origen+x/8] >> uint(7-x%8)) & 1)
				if idx < len(paleta) {
					c = paleta[idx]
				}
			}
			img.SetRGBA(x, y, c)
		}
	}

	// El mapa AND marca la transparencia (1 bit por píxel, alineado a 4 bytes).
	// En los ICO de 32 bits el alfa suele venir en el propio canal; si está
	// vacío, se considera opaco y se usa el mapa AND.
	usarMascara := bits != 32
	if bits == 32 && mascaraAlphaVacia(img) {
		for i := 3; i < len(img.Pix); i += 4 {
			img.Pix[i] = 255
		}
		usarMascara = true
	}
	if usarMascara {
		aplicarMascaraAnd(img, b, posMascara, mascaraFila)
	}
	return img, nil
}

// bs devuelve los bytes por píxel de una profundidad de bits.
func bs(bits int) int {
	return (bits + 7) / 8
}

// mascaraAlphaVacia indica si el canal alfa de 32 bits es todo cero (iconos que
// guardan la transparencia solo en el mapa AND).
func mascaraAlphaVacia(img *image.RGBA) bool {
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 {
			return false
		}
	}
	return true
}

// aplicarMascaraAnd pone alfa 0 donde el mapa AND tenga un bit activo.
func aplicarMascaraAnd(img *image.RGBA, b []byte, base, filaMascara int) {
	alto := img.Bounds().Dy()
	ancho := img.Bounds().Dx()
	for y := 0; y < alto; y++ {
		inicio := base + (alto-1-y)*filaMascara
		if inicio < 0 || inicio+filaMascara > len(b) {
			return
		}
		for x := 0; x < ancho; x++ {
			if (b[inicio+x/8]>>uint(7-x%8))&1 == 1 {
				img.Pix[img.PixOffset(x, y)+3] = 0
			}
		}
	}
}
