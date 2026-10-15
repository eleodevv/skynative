# Contribuir a Skynative

Gracias por tu interés. Estas pautas mantienen el proyecto pequeño, seguro y
mantenible.

## Requisitos

- Go 1.23 o superior.
- Para probar la compilación de Tauri: Rust, Cargo y la CLI de Tauri v2, además
  de las dependencias de escritorio de tu sistema.

## Puesta en marcha

```bash
git clone https://github.com/eleodevv/skynative.git
cd skynative
go build ./cmd/skynative
```

## Antes de enviar cambios

Ejecuta siempre:

```bash
make verificar
```

que equivale a:

```bash
gofmt -l .      # no debe listar archivos
go vet ./...
go test ./...
```

Corrige cualquier problema antes de abrir una *pull request*.

## Estilo de código

- Escribe **todo en español**: código, comentarios, mensajes de error, logs y
  documentación.
- Usa nombres de variables, funciones y paquetes en español cuando las
  convenciones y herramientas de Go lo permitan. Mantén en inglés los
  identificadores que exijan herramientas externas (por ejemplo campos de
  `tauri.conf.json`).
- No añadas comentarios innecesarios. Los comentarios explican el *por qué*.
- Evita dependencias nuevas salvo que aporten un valor claro. Justifícalas en
  la *pull request*.
- Nunca uses `shell=true` ni construyas comandos concatenando datos del usuario.
  Usa `exec.CommandContext` con argumentos separados.
- Considera la entrada del usuario como no confiable.

## Pruebas

- Añade pruebas para cualquier función nueva o modificada.
- No simules resultados: las pruebas deben ejecutarse de verdad.
- Los archivos de prueba usan el sufijo `_test.go` dentro del mismo paquete.

## Añadir una pantalla o acción a la TUI

La interfaz completa vive en `interno/tui`:

1. Define el nuevo estado en `interno/tui/modelo.go` y añádelo a la constante
   de estados.
2. Gestiona las teclas del estado en `interno/tui/actualizar.go` (`Update`).
3. Dibuja la pantalla en `interno/tui/vistas.go` (`View`) usando los estilos de
   `interno/tui/estilos.go`.
4. Si la acción lanza un proceso o toca el disco, encapsúlala como un
   `tea.Cmd` en `interno/tui/acciones.go`, devolviendo un mensaje que `Update`
   consuma.
5. Añade pruebas de la transición en `interno/tui/*_test.go`; puedes usar
   `aplicar(modelo, mensaje)` para avanzar el modelo sin terminal real.

## Añadir una plataforma

Sigue la sección «Cómo añadir una nueva plataforma» del `README.md`.

## Seguridad

- Acepta únicamente URLs `http` o `https`.
- No añadas capacidades privilegiadas de Tauri para contenido remoto.
- No desactives protecciones (TLS, CSP, validaciones) para que algo funcione.
- Reporta vulnerabilidades de forma privada al mantenedor en lugar de abrir un
  *issue* público.

## Commits

- Mensajes concisos y en español, en modo imperativo (por ejemplo
  «Añade validación de dimensiones»).
- Un cambio coherente por commit.

## Licencia

Al contribuir aceptas que tu aportación se distribuya bajo la licencia MIT del
proyecto.
