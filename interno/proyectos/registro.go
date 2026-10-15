// Package proyectos mantiene un registro local de las aplicaciones generadas
// por Skynative para poder listarlas y recompilarlas.
//
// El registro se guarda en un archivo JSON dentro del directorio de
// configuración del usuario. La variable de entorno SKYNATIVE_DIRECTORIO_CONFIG
// permite sobrescribir ese directorio (útil en pruebas).
package proyectos

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Proyecto es una entrada del registro.
type Proyecto struct {
	Nombre        string `json:"nombre"`
	Identificador string `json:"identificador"`
	URL           string `json:"url"`
	Ruta          string `json:"ruta"`
	Creado        string `json:"creado"`
}

// rutaRegistro devuelve la ruta del archivo de registro.
func rutaRegistro() (string, error) {
	if override := os.Getenv("SKYNATIVE_DIRECTORIO_CONFIG"); override != "" {
		return filepath.Join(override, "proyectos.json"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("no se pudo determinar el directorio de configuración: %w", err)
	}
	return filepath.Join(base, "skynative", "proyectos.json"), nil
}

// Cargar lee el registro. Si no existe, devuelve una lista vacía.
func Cargar() ([]Proyecto, error) {
	ruta, err := rutaRegistro()
	if err != nil {
		return nil, err
	}
	datos, err := os.ReadFile(ruta)
	if errors.Is(err, os.ErrNotExist) {
		return []Proyecto{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el registro %q: %w", ruta, err)
	}
	var lista []Proyecto
	if err := json.Unmarshal(datos, &lista); err != nil {
		return nil, fmt.Errorf("el registro %q está dañado: %w", ruta, err)
	}
	return lista, nil
}

// Agregar añade o actualiza un proyecto en el registro.
func Agregar(p Proyecto) error {
	if p.Creado == "" {
		p.Creado = time.Now().UTC().Format(time.RFC3339)
	}
	ruta, err := rutaRegistro()
	if err != nil {
		return err
	}
	lista, err := Cargar()
	if err != nil {
		return err
	}

	actualizado := false
	for i := range lista {
		if lista[i].Ruta == p.Ruta {
			lista[i] = p
			actualizado = true
			break
		}
	}
	if !actualizado {
		lista = append(lista, p)
	}

	datos, err := json.MarshalIndent(lista, "", "  ")
	if err != nil {
		return fmt.Errorf("no se pudo serializar el registro: %w", err)
	}
	datos = append(datos, '\n')

	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return fmt.Errorf("no se pudo crear el directorio del registro: %w", err)
	}
	if err := os.WriteFile(ruta, datos, 0o644); err != nil {
		return fmt.Errorf("no se pudo escribir el registro %q: %w", ruta, err)
	}
	return nil
}

// Listar devuelve los proyectos ordenados por nombre.
func Listar() ([]Proyecto, error) {
	lista, err := Cargar()
	if err != nil {
		return nil, err
	}
	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Nombre < lista[j].Nombre
	})
	return lista, nil
}
