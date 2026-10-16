package configuracion

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// reIdentificador valida un identificador en notación de dominio inverso.
var reIdentificador = regexp.MustCompile(`^[a-zA-Z0-9]+([.-][a-zA-Z0-9]+)*$`)

// reNombrePaquete valida el nombre del módulo Go del proyecto generado.
var reNombrePaquete = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// caracteresProhibidosNombre contiene los caracteres no válidos para el
// nombre visible de la aplicación.
const caracteresProhibidosNombre = `/\:*?"<>|`

// acentos asocia vocales acentuadas y otros caracteres frecuentes en español
// con su equivalente ASCII para construir identificadores legibles.
var acentos = map[rune]rune{
	'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a',
	'é': 'e', 'è': 'e', 'ë': 'e', 'ê': 'e',
	'í': 'i', 'ì': 'i', 'ï': 'i', 'î': 'i',
	'ó': 'o', 'ò': 'o', 'ö': 'o', 'ô': 'o',
	'ú': 'u', 'ù': 'u', 'ü': 'u', 'û': 'u',
	'ñ': 'n', 'ç': 'c',
}

// Validar comprueba todos los campos de la configuración y devuelve el primer
// error encontrado, con un mensaje en español listo para mostrar al usuario.
func (c Configuracion) Validar() error {
	if _, err := ValidarURL(c.URL); err != nil {
		return err
	}
	if _, err := ValidarNombre(c.Nombre); err != nil {
		return err
	}
	if err := ValidarDimensiones(c.Ancho, c.Alto); err != nil {
		return err
	}
	if !reIdentificador.MatchString(c.Identificador) {
		return fmt.Errorf("identificador de paquete no válido %q: usa letras, números, guiones y puntos", c.Identificador)
	}
	if !reNombrePaquete.MatchString(c.NombrePaquete) {
		return fmt.Errorf("nombre de paquete no válido %q: debe empezar por letra y contener solo minúsculas, números, guiones o guiones bajos", c.NombrePaquete)
	}
	if _, err := ValidarVersion(c.Version); err != nil {
		return err
	}
	return nil
}

// ValidarURL acepta únicamente URLs http o https con host. Devuelve la URL
// normalizada (recortada) cuando es válida.
func ValidarURL(bruta string) (string, error) {
	s := strings.TrimSpace(bruta)
	if s == "" {
		return "", errors.New("la URL no puede estar vacía")
	}
	if strings.ContainsAny(s, "\x00\r\n\t") {
		return "", errors.New("la URL no puede contener caracteres de control")
	}

	// Comprobación previa del esquema para dar mensajes claros a esquemas
	// peligrosos como javascript:, file: o data:.
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("URL no válida: %w", err)
	}

	esquema := strings.ToLower(u.Scheme)
	switch esquema {
	case "http", "https":
		// correcto
	case "javascript", "file", "data", "vbscript":
		return "", fmt.Errorf("esquema no permitido %q: solo se admiten http y https", u.Scheme)
	default:
		if esquema == "" {
			return "", errors.New("la URL debe incluir un esquema (http o https)")
		}
		return "", fmt.Errorf("esquema no permitido %q: solo se admiten http y https", u.Scheme)
	}

	if u.Host == "" {
		return "", errors.New("la URL debe incluir un host")
	}
	return s, nil
}

// ValidarNombre comprueba el nombre visible de la aplicación.
func ValidarNombre(bruto string) (string, error) {
	s := strings.TrimSpace(bruto)
	if s == "" {
		return "", errors.New("el nombre no puede estar vacío")
	}
	if len([]rune(s)) > LongitudMaximaNombre {
		return "", fmt.Errorf("el nombre no puede superar los %d caracteres", LongitudMaximaNombre)
	}
	if strings.ContainsAny(s, caracteresProhibidosNombre) {
		return "", fmt.Errorf("el nombre no puede contener ninguno de estos caracteres: %s", caracteresProhibidosNombre)
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return "", errors.New("el nombre no puede contener caracteres de control")
		}
	}
	return s, nil
}

// ValidarDimensiones comprueba los límites de la ventana.
func ValidarDimensiones(ancho, alto int) error {
	if ancho < AnchoMinimo || ancho > AnchoMaximo {
		return fmt.Errorf("el ancho debe estar entre %d y %d píxeles", AnchoMinimo, AnchoMaximo)
	}
	if alto < AltoMinimo || alto > AltoMaximo {
		return fmt.Errorf("el alto debe estar entre %d y %d píxeles", AltoMinimo, AltoMaximo)
	}
	return nil
}

// reVersion valida una versión semántica simplificada (MAYOR.MENOR.PARCHE).
var reVersion = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidarVersion comprueba una versión semántica simplificada.
func ValidarVersion(bruta string) (string, error) {
	s := strings.TrimSpace(bruta)
	if !reVersion.MatchString(s) {
		return "", fmt.Errorf("versión no válida %q: se espera el formato MAYOR.MENOR.PARCHE", s)
	}
	return s, nil
}

// GenerarIdentificador construye un identificador de dominio inverso estable
// a partir del nombre visible.
func GenerarIdentificador(nombre string) string {
	slug := GenerarSlug(nombre)
	if slug == "" {
		slug = "app"
	}
	return PrefijoIdentificador + "." + slug
}

// GenerarNombrePaquete construye un nombre de módulo válido (minúsculas).
func GenerarNombrePaquete(nombre string) string {
	slug := GenerarSlug(nombre)
	if slug == "" {
		slug = "app"
	}
	// Un crate no debe empezar por dígito.
	if slug[0] >= '0' && slug[0] <= '9' {
		slug = "app-" + slug
	}
	return slug
}

// GenerarSlug convierte un nombre en minúsculas con guiones, apto para
// identificadores, nombres de paquete y de directorio.
func GenerarSlug(nombre string) string {
	var b strings.Builder
	ultimoGuion := true // evita guiones iniciales
	for _, r := range strings.ToLower(strings.TrimSpace(nombre)) {
		if equivalente, ok := acentos[r]; ok {
			r = equivalente
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			ultimoGuion = false
		default:
			if !ultimoGuion {
				b.WriteByte('-')
				ultimoGuion = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
