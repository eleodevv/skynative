// Package plantillas expone los archivos estáticos que Skynative copia al
// generar un proyecto Tauri v2.
//
// Los archivos se incrustan en el binario mediante go:embed para que los
// binarios de Skynative sean autocontenidos y no dependan de rutas externas.
package plantillas

import "embed"

// TauriV2 contiene la plantilla base de un proyecto Tauri v2.
//
// Rutas disponibles (relativas a este paquete):
//   - tauri-v2/Cargo.toml.tmpl
//   - tauri-v2/build.rs
//   - tauri-v2/main.rs
//   - tauri-v2/gitignore
//   - tauri-v2/README.md.tmpl
//   - tauri-v2/capabilities/default.json
//
//go:embed tauri-v2
var TauriV2 embed.FS
