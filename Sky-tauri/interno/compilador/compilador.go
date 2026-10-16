// Package compilador ejecuta la CLI oficial de Tauri sobre un proyecto
// generado y localiza los artefactos resultantes.
//
// Los procesos se lanzan con exec.CommandContext y argumentos separados: no se
// usa intérprete de comandos ni se concatenan cadenas con datos del usuario.
package compilador

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Opciones controla la compilación.
type Opciones struct {
	// Depuracion compila en modo de depuración (cargo tauri build --debug).
	Depuracion bool
	// SinBundle compila solo el binario, sin generar paquetes de instalación.
	SinBundle bool
	// ComandoTauri permite forzar la ruta del ejecutable de la CLI de Tauri.
	ComandoTauri string
}

// Resultado describe el resultado de la compilación.
type Resultado struct {
	// CodigoSalida es el código devuelto por la CLI de Tauri (0 si todo fue bien).
	CodigoSalida int
	// Artefactos son las rutas de los paquetes y binarios generados.
	Artefactos []string
	// Salida contiene la salida combinada de la compilación.
	Salida string
	// EmpaquetadoOmitido indica que se generó solo el binario porque el
	// empaquetado falló (por ejemplo, por falta de rpmbuild o linuxdeploy).
	EmpaquetadoOmitido bool
}

// Compilar lanza la compilación de Tauri del proyecto situado en rutaProyecto.
//
// rutaProyecto puede ser la raíz del proyecto (que contiene src-tauri) o el
// propio directorio src-tauri. Si log no es nil, la salida se escribe además en
// ese destino para mostrar el progreso en tiempo real.
//
// Primero intenta la compilación completa (con empaquetado). Si falla y no se
// pidió «SinBundle», reintenta con --no-bundle: así se obtiene siempre el
// binario aunque el empaquetado nativo no esté disponible en el sistema.
func Compilar(ctx context.Context, rutaProyecto string, op Opciones, log io.Writer) (Resultado, error) {
	dirSrcTauri, err := localizarSrcTauri(rutaProyecto)
	if err != nil {
		return Resultado{}, err
	}
	if log == nil {
		log = io.Discard
	}

	res, err := ejecutar(ctx, dirSrcTauri, op, log)
	if err == nil || op.SinBundle {
		return res, err
	}

	fmt.Fprintln(log, "")
	fmt.Fprintln(log, "La compilación con empaquetado falló; reintentando sin empaquetado (--no-bundle)…")
	op2 := op
	op2.SinBundle = true
	res2, err2 := ejecutar(ctx, dirSrcTauri, op2, log)
	if err2 != nil {
		// Devolvemos el error original (el del empaquetado) por ser más informativo.
		return res, err
	}
	res2.EmpaquetadoOmitido = true
	return res2, nil
}

// ejecutar lanza una única compilación con las opciones indicadas.
func ejecutar(ctx context.Context, dirSrcTauri string, op Opciones, log io.Writer) (Resultado, error) {
	nombre, argsBase, err := comandoTauri(op.ComandoTauri)
	if err != nil {
		return Resultado{}, err
	}
	args := append(append([]string{}, argsBase...), "build")
	if op.Depuracion {
		args = append(args, "--debug")
	}
	if op.SinBundle {
		args = append(args, "--no-bundle")
	}

	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, nombre, args...)
	cmd.Dir = dirSrcTauri
	cmd.Stdout = io.MultiWriter(log, &buf)
	cmd.Stderr = io.MultiWriter(log, &buf)

	if err := cmd.Run(); err != nil {
		codigo := codigoDeError(err)
		return Resultado{CodigoSalida: codigo, Salida: buf.String()},
			fmt.Errorf("la compilación de Tauri falló (código %d): %w", codigo, err)
	}

	artefactos, err := BuscarArtefactos(dirSrcTauri)
	if err != nil {
		return Resultado{Salida: buf.String()}, err
	}
	return Resultado{CodigoSalida: 0, Artefactos: artefactos, Salida: buf.String()}, nil
}

// UltimasLineas devuelve las últimas n líneas no vacías de una salida.
func UltimasLineas(salida string, n int) []string {
	lineas := strings.Split(strings.ReplaceAll(salida, "\r\n", "\n"), "\n")
	var limpias []string
	for _, l := range lineas {
		l = strings.TrimRight(l, "\r ")
		if strings.TrimSpace(l) != "" {
			limpias = append(limpias, l)
		}
	}
	if n > 0 && len(limpias) > n {
		limpias = limpias[len(limpias)-n:]
	}
	return limpias
}

// comandoTauri decide cómo invocar la CLI de Tauri.
func comandoTauri(forzado string) (string, []string, error) {
	if forzado != "" {
		return forzado, nil, nil
	}
	if _, err := exec.LookPath("cargo"); err == nil {
		// cargo tauri solo existe si cargo-tauri está instalado.
		return "cargo", []string{"tauri"}, nil
	}
	if _, err := exec.LookPath("tauri"); err == nil {
		return "tauri", nil, nil
	}
	return "", nil, errors.New(
		"no se encontró la CLI de Tauri; instala Rust y luego " +
			"`cargo install tauri-cli --version \"^2\"` o `npm install -g @tauri-apps/cli`")
}

// localizarSrcTauri encuentra el directorio src-tauri de un proyecto.
func localizarSrcTauri(ruta string) (string, error) {
	candidatos := []string{
		filepath.Join(ruta, "src-tauri"),
		ruta,
	}
	for _, c := range candidatos {
		conf := filepath.Join(c, "tauri.conf.json")
		if info, err := os.Stat(conf); err == nil && !info.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("no se encontró tauri.conf.json en %q: ¿es un proyecto de Tauri v2?", ruta)
}

// extensionesBundle son las extensiones de los paquetes de instalación.
var extensionesBundle = map[string]bool{
	".deb":      true,
	".rpm":      true,
	".appimage": true,
	".dmg":      true,
	".msi":      true,
	".exe":      true,
}

// BuscarArtefactos localiza los paquetes y el binario principal generados.
func BuscarArtefactos(dirSrcTauri string) ([]string, error) {
	dirRelease := filepath.Join(dirSrcTauri, "target", "release")
	if _, err := os.Stat(dirRelease); err != nil {
		return nil, fmt.Errorf("no se encontró el directorio de compilación %q: %w", dirRelease, err)
	}

	var artefactos []string

	// Paquetes de instalación.
	dirBundle := filepath.Join(dirRelease, "bundle")
	if _, err := os.Stat(dirBundle); err == nil {
		err := filepath.WalkDir(dirBundle, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if strings.HasSuffix(p, ".app") {
					artefactos = append(artefactos, p)
					return fs.SkipDir
				}
				return nil
			}
			if extensionesBundle[strings.ToLower(filepath.Ext(p))] {
				artefactos = append(artefactos, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Binario principal (ejecutable suelto en target/release).
	entradas, err := os.ReadDir(dirRelease)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer %q: %w", dirRelease, err)
	}
	for _, e := range entradas {
		if e.IsDir() {
			continue
		}
		ruta := filepath.Join(dirRelease, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}
		esEjecutable := info.Mode()&0o111 != 0
		if strings.EqualFold(filepath.Ext(e.Name()), ".exe") || esEjecutable {
			artefactos = append(artefactos, ruta)
		}
	}

	sort.Strings(artefactos)
	return artefactos, nil
}

// codigoDeError extrae el código de salida de un error de proceso.
func codigoDeError(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 1
}
