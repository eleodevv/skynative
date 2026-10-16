// Package configuracion define el modelo de datos de una aplicación que
// Skynative va a generar y agrupa toda la validación de las entradas.
//
// La URL de entrada se considera información no confiable: aquí se comprueba
// el esquema, el host y que no contenga caracteres de control.
package configuracion

import (
	"image"
	"strings"
)

// Valores predeterminados y límites de validación.
const (
	AnchoPredeterminado  = 1200
	AltoPredeterminado   = 800
	AnchoMinimo          = 200
	AltoMinimo           = 200
	AnchoMaximo          = 20000
	AltoMaximo           = 20000
	LongitudMaximaNombre = 80

	// PrefijoIdentificador se antepone al identificador derivado del nombre.
	PrefijoIdentificador = "com.skynative"

	// VersionPredeterminada es la versión inicial de la aplicación generada.
	VersionPredeterminada = "0.1.0"
)

// Configuracion describe una aplicación de escritorio que Skynative genera.
type Configuracion struct {
	// URL remota que cargará la ventana (solo http o https).
	URL string
	// Nombre visible de la aplicación.
	Nombre string
	// Identificador de paquete en notación de dominio inverso.
	Identificador string
	// NombrePaquete es el nombre del crate de Cargo (minúsculas, sin espacios).
	NombrePaquete string
	// Version es la versión semántica de la aplicación generada.
	Version string
	// Ancho y Alto son las dimensiones iniciales de la ventana en píxeles.
	Ancho int
	Alto  int
	// Redimensionable indica si el usuario puede cambiar el tamaño de la ventana.
	Redimensionable bool
	// DirectorioSalida es el directorio base donde se creará el proyecto.
	DirectorioSalida string
	// Icono es la imagen base para los iconos de la aplicación. Si es nil se
	// usa el icono por defecto que dibuja el generador. No se serializa.
	Icono image.Image
}

// Nueva devuelve una configuración con los valores predeterminados aplicados.
// No valida: para ello usa Normalizar y Validar.
func Nueva(url, nombre string) Configuracion {
	return Configuracion{
		URL:              strings.TrimSpace(url),
		Nombre:           strings.TrimSpace(nombre),
		Version:          VersionPredeterminada,
		Ancho:            AnchoPredeterminado,
		Alto:             AltoPredeterminado,
		Redimensionable:  true,
		DirectorioSalida: ".",
	}
}

// Normalizar completa los campos derivados o vacíos sin devolver error. Está
// pensada para ejecutarse antes que Validar.
func (c *Configuracion) Normalizar() {
	c.URL = strings.TrimSpace(c.URL)
	c.Nombre = strings.TrimSpace(c.Nombre)
	c.Version = strings.TrimSpace(c.Version)

	if c.Version == "" {
		c.Version = VersionPredeterminada
	}
	if c.Ancho <= 0 {
		c.Ancho = AnchoPredeterminado
	}
	if c.Alto <= 0 {
		c.Alto = AltoPredeterminado
	}
	if strings.TrimSpace(c.DirectorioSalida) == "" {
		c.DirectorioSalida = "."
	}

	if strings.TrimSpace(c.NombrePaquete) == "" {
		c.NombrePaquete = GenerarNombrePaquete(c.Nombre)
	}
	if strings.TrimSpace(c.Identificador) == "" {
		c.Identificador = GenerarIdentificador(c.Nombre)
	}
}
