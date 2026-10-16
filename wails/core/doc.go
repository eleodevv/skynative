// Package core agrupa la lógica pura de Skynative para el motor Wails:
// configuración y validación, generación desde plantillas embebidas,
// compilación, diagnóstico del entorno, registro de proyectos e instalación
// en el sistema.
//
// Regla del paquete: core NO importa nada de la TUI (cmd/skynative) ni del SDK
// de Wails. La dirección de dependencias es cmd/skynative → core → templates.
package core
