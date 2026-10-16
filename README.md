# Skynative

Repositorio monolítico durante la migración de motor. Skynative convierte una
URL web en una aplicación de escritorio nativa y mantiene su **TUI en Go** como
única interfaz.

## Carpetas

| Carpeta | Estado | Motor | Descripción |
|---|---|---|---|
| [`Sky-tauri/`](Sky-tauri/) | congelado | [Tauri v2] | Versión original y funcionando, con la TUI completa. Respaldo: tag `v0-tauri-backup` y rama `backup/tauri-version`. |
| [`wails/`](wails/) | en migración | [Wails v2] | Nueva implementación: `core/` (lógica pura), `templates/` (Wails + Svelte), `ui/` (fase posterior). |
| [`docs/`](docs/) | activo | — | Landing y página de descarga (GitHub Pages). |

## Estructura nueva (wails/)

```text
wails/
├── core/           lógica pura en Go (validación, generación, build). Sin TUI ni Wails.
├── cmd/skynative/  CLI/TUI (usa core)
├── templates/      plantillas embebidas de apps Wails + Svelte
└── ui/             app Wails + Svelte (fase siguiente)
```

## Estado de la migración

Ver el historial de commits: la carpeta `wails/` incorporará la interfaz
`Engine` (`Scaffold`, `Build`) con `WailsEngine` como primera implementación.

[Tauri v2]: https://v2.tauri.app
[Wails v2]: https://wails.io
