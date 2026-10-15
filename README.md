# Skynative

Skynative convierte una URL web en una aplicación de escritorio usando
[Tauri v2](https://v2.tauri.app/). Es una herramienta en Go con una **interfaz
de terminal (TUI)** como único modo de uso, sin dependencias pesadas, pensada
para ser pequeña, reproducible, segura y fácil de mantener.

```bash
skynative
```

Al abrirla verás un menú con estas acciones:

```
1  Crear aplicación
2  Instalar herramientas necesarias
3  Listar proyectos
4  Diagnóstico del entorno
5  Ver versión
6  Salir
```

Si al arrancar falta Rust, la CLI de Tauri o las dependencias del sistema, el
menú lo avisa y la opción **Instalar herramientas necesarias** las instala. Con
el entorno listo, crear una aplicación **la genera y la compila automáticamente**.

El resultado es un proyecto Tauri v2 válido y, tras compilarlo, una aplicación
de escritorio que abre una ventana nativa cargando la página indicada.

Sitio del proyecto: <https://eleodevv.github.io/skynative/>

---

## Qué problemas resuelve

- Evita escribir a mano el andamiaje de un proyecto Tauri v2 (Cargo.toml,
  build.rs, main.rs, tauri.conf.json, iconos).
- Centraliza la conversión «URL → ventana de escritorio» en una interfaz
  repetible y guiada.
- Valida la URL y el nombre antes de generar nada, rechazando esquemas
  peligrosos (`file:`, `javascript:`, `data:`).
- Genera aplicaciones sin capacidades privilegiadas: la página remota no puede
  invocar comandos locales ni acceder al sistema de archivos.
- Aporta diagnóstico del entorno para saber qué falta antes de compilar.
- Incluye un explorador de carpetas integrado para elegir el destino sin salir
  de la terminal.

## Qué **no** hace

- No es un backend ni un servicio remoto: todo se ejecuta en local.
- No usa Docker.
- No expone una interfaz de línea de comandos con subcomandos: toda la
  interacción ocurre en la TUI (salvo `--help` y `--version`).
- No genera instaladores de Tauri sin el toolchain de Rust del sistema destino
  (ver limitaciones más abajo).

---

## Arquitectura

```text
skynative/
├── cmd/skynative/main.go        Punto de entrada (--help/--version + arranque de la TUI)
├── interno/
│   ├── tui/                     Interfaz de terminal (Bubble Tea): modelo, vistas, acciones, explorador
│   ├── configuracion/           Modelo de datos y validación de entradas
│   ├── generador/               Creación del proyecto Tauri v2 y de los iconos
│   ├── compilador/              Ejecución de la CLI de Tauri y descubrimiento de artefactos
│   ├── instalador/              Registro de la app compilada en el menú del sistema
│   ├── herramientas/            Instalación de Rust, la CLI de Tauri y dependencias
│   ├── diagnostico/             Comprobación de Go, Rust, Cargo, Tauri y dependencias
│   ├── proyectos/               Registro local de aplicaciones generadas
│   └── version/                 Versión y metadatos del binario
├── plantillas/tauri-v2/         Plantilla incrustada del proyecto Tauri v2
├── scripts/                     Compilación multiplataforma e instaladores
├── .github/workflows/           CI y publicación de releases
├── go.mod / go.sum
├── Makefile
└── README.md
```

### Flujo de ejecución

1. `main.go` atiende `--help` y `--version`; si no, comprueba que la entrada es
   una terminal y arranca `tui.Ejecutar`.
2. Al arrancar, la TUI comprueba el entorno en segundo plano (`diagnostico`) y
   avisa en el menú si faltan herramientas; la acción **Instalar herramientas
   necesarias** ejecuta `herramientas.Instalar` con la salida en vivo.
3. La TUI muestra el menú y, según la acción, pide datos en un formulario o
   recurre a los paquetes internos.
4. `configuracion` normaliza y valida: URL, nombre, dimensiones, versión,
   identificador y nombre de paquete.
5. `generador.Generar` escribe el proyecto: plantillas incrustadas
   (`embed`), `tauri.conf.json` serializado como JSON y los iconos PNG/ICO/ICNS
   generados por código. La TUI recibe la salida en vivo mientras trabaja.
6. `proyectos` registra la aplicación.
7. Si el entorno está listo, la compilación se **encadena automáticamente**
   tras generar. `compilador.Compilar` lanza `cargo tauri build` y busca los
   paquetes resultantes; la salida se muestra en la pantalla de progreso.
8. Tras compilar, `instalador.Instalar` registra la app en el sistema (entrada
   `.desktop` e icono en Linux, copia a `~/Applications` en macOS) para poder
   abrirla con doble clic y encontrarla al buscar.

### Decisiones de diseño

- **TUI como única interfaz.** [Bubble Tea](https://github.com/charmbracelet/bubbletea),
  [Bubbles](https://github.com/charmbracelet/bubbles) y
  [Lip Gloss](https://github.com/charmbracelet/lipgloss) ofrecen un menú, un
  formulario, un explorador de carpetas y pantallas de progreso/resultado sin
  multiplicar la superficie de la herramienta.
- **Lógica separada de la presentación.** Toda la validación, generación y
  compilación vive en paquetes independientes de la TUI; la interfaz solo los
  orquesta mediante `tea.Cmd` y mensajes, lo que mantiene el núcleo testeable.
- **Configuración como JSON tipado.** `tauri.conf.json` no se construye con
  plantillas de texto, sino a partir de estructuras de Go y `encoding/json`,
  lo que garantiza el escape correcto de los valores.
- **Sin `shell=true`.** Los procesos externos se lanzan con
  `exec.CommandContext` y argumentos separados.

---

## Requisitos por sistema operativo

Skynative (el binario de Go) es autocontenido. Para **compilar la aplicación
generada** (Tauri) necesitas el entorno de Tauri v2.

### Todos

- [Rust y Cargo](https://rustup.rs/) (Rust 1.77 o superior; recomendado estable).
- CLI de Tauri v2:
  ```bash
  cargo install tauri-cli --version "^2"
  # o bien
  npm install -g @tauri-apps/cli
  ```

### Linux (Debian/Ubuntu)

```bash
sudo apt update
sudo apt install libwebkit2gtk-4.1-dev libgtk-3-dev libsoup-3.0-dev \
  build-essential curl wget file libxdo-dev libssl-dev \
  libayatana-appindicator3-dev librsvg2-dev
```

Otras distribuciones: consulta la
[guía oficial de prerrequisitos](https://v2.tauri.app/start/prerequisites/).

### macOS

- Xcode Command Line Tools: `xcode-select --install`.

### Windows

- Microsoft Edge WebView2 Runtime (incluido en Windows 10/11).
- Las herramientas de compilación de Visual Studio (C++ Build Tools).

No hace falta instalarlos a mano: la opción **Instalar herramientas necesarias**
del menú los instala por ti (en Linux usa `apt-get`, `dnf` o `pacman` y, si hace
falta `sudo`, te muestra el comando exacto). Verifica el entorno en cualquier
momento desde **Diagnóstico del entorno**.

---

## Instalación desde releases

### Linux y macOS

```bash
curl -fsSL https://github.com/eleodevv/skynative/releases/latest/download/install.sh | sh
```

Opciones mediante variables de entorno:

```bash
SKYNATIVE_VERSION=v0.1.0 SKYNATIVE_DIR="$HOME/bin" \
  curl -fsSL https://github.com/eleodevv/skynative/releases/latest/download/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://github.com/eleodevv/skynative/releases/latest/download/install.ps1 | iex
```

Los instaladores detectan sistema y arquitectura, descargan el artefacto
correcto, **verifican su checksum SHA-256 antes de instalar** y no requieren
privilegios de administrador.

### Actualización

Vuelve a ejecutar el instalador (sin `SKYNATIVE_VERSION` instala la última
release).

### Desinstalación

```bash
# Linux y macOS
curl -fsSL https://github.com/eleodevv/skynative/releases/latest/download/desinstalar.sh | sh
```

```powershell
# Windows (PowerShell)
irm https://github.com/eleodevv/skynative/releases/latest/download/desinstalar.ps1 | iex
```

Los desinstaladores eliminan el binario (y, en Windows, su carpeta y la entrada
del PATH del usuario). No tocan los proyectos ni las aplicaciones creadas con
Skynative.

---

## Instalación desde el código fuente

Requiere Go 1.24 o superior.

```bash
git clone https://github.com/eleodevv/skynative.git
cd skynative
make compilar-local      # genera dist/skynative
./dist/skynative --version
```

O directamente:

```bash
go build -o skynative ./cmd/skynative
```

---

## Uso

Skynative es una interfaz de terminal. Ábrela sin argumentos:

```bash
skynative
```

Opciones disponibles fuera de la TUI:

```bash
skynative --help      # ayuda
skynative --version   # versión instalada
```

### Atajos de teclado

| Pantalla | Teclas |
| --- | --- |
| Menú | `↑`/`↓` o `1`-`6` para elegir · `Enter` abrir · `q` salir |
| Formulario | `Tab`/`Shift+Tab` o `↑`/`↓` moverse · `Enter` abrir el explorador en «Carpeta» o crear en «Crear aplicación» · `Espacio` alternar la casilla · `Ctrl+G` crear · `Esc` volver |
| Explorador | `↑`/`↓` moverse · `Enter` entrar o elegir · `S` usar la carpeta actual · `Retroceso` subir · `Inicio` ir al perfil · `Esc` cancelar |
| Progreso | `↑`/`↓` desplazar la salida · `Esc` cancelar |
| Resultado | `C` compilar · `I` instalar en el sistema · `O` abrir la carpeta · `Enter` crear otra · `Esc` volver al menú |

### Cómo crear, compilar e instalar una aplicación

1. Si el menú avisa de que faltan herramientas, elige **Instalar herramientas
   necesarias** y espera a que terminen; al volver, el entorno quedará listo.
2. Elige **Crear aplicación** en el menú.
3. Rellena la URL (solo `http`/`https`), el nombre, el ancho, el alto y la
   carpeta de salida (abre el explorador integrado con `Enter`).
4. Pulsa `Ctrl+G` o selecciona **Crear aplicación**.
5. Se crea `<carpeta>/<nombre>/` con `src-tauri/` (Cargo.toml, build.rs,
   main.rs, tauri.conf.json e iconos).
6. Si el entorno está listo, la compilación arranca **automáticamente**. Si no,
   pulsa `C` en la pantalla de resultado. Los artefactos aparecen en
   `<proyecto>/src-tauri/target/release/bundle/` (y el ejecutable en
   `<proyecto>/src-tauri/target/release/`).
7. Pulsa `I` para instalar la app en el sistema. Skynative copia el icono y
   crea el lanzador con el que podrás **abrirla con doble clic** y
   **encontrarla al buscar** en tus aplicaciones:
   - **Linux:** entrada `.desktop` en `~/.local/share/applications/` e icono en
     `~/.local/share/icons/hicolor/256x256/apps/`.
   - **macOS:** copia del paquete `.app` a `~/Applications`.
   - **Windows:** acceso directo `.lnk` en el menú Inicio (creado con
     PowerShell, con el icono de la app). También puedes usar el `.msi` o el
     `.exe` que genera Tauri.

La instalación no requiere privilegios de administrador: todo va al perfil del
usuario.

También puedes compilar sin Skynative:

```bash
cd <carpeta>/<nombre>/src-tauri
cargo tauri build
```

---

## Cómo ejecutar las pruebas

```bash
make probar      # go test ./...
make verificar   # gofmt + go vet + pruebas
```

Equivalente manual:

```bash
gofmt -l .
go vet ./...
go test ./...
go build ./cmd/skynative
```

Las pruebas cubren: validación de URL y nombre, generación de identificadores,
generación de configuración, manejo de errores de archivos, construcción segura
de comandos, detección de plataforma, selección de artefactos, versionado del
binario y las transiciones principales de la TUI (incluido un flujo completo de
generación y el arranque/cierre del programa Bubble Tea).

---

## Cómo generar releases

1. Asegúrate de que `make verificar` pasa.
2. Crea y publica una etiqueta de versión:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. El flujo `.github/workflows/release.yml` se activa automáticamente:
   - ejecuta pruebas y `go vet`;
   - compila los binarios de todas las plataformas y sus checksums;
   - publica una GitHub Release con los artefactos, `install.sh`, `install.ps1`,
     `desinstalar.sh` y `desinstalar.ps1`.

Pasos manuales equivalentes (local):

```bash
make compilar-todas-plataformas VERSION=0.1.0 COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo local)"
ls dist/
```

Artefactos generados:

| Plataforma            | Archivo                          |
| --------------------- | -------------------------------- |
| Linux x86-64          | `skynative-linux-amd64`          |
| Linux ARM64           | `skynative-linux-arm64`          |
| macOS Intel           | `skynative-darwin-amd64`         |
| macOS Apple Silicon   | `skynative-darwin-arm64`         |
| Windows x86-64        | `skynative-windows-amd64.exe`    |
| Windows ARM64         | `skynative-windows-arm64.exe`    |

Además se generan `*.sha256` por artefacto y un `SHA256SUMS` combinado.

> **Importante.** Compilar el binario de Go es distinto de compilar la
> aplicación Tauri. Un binario de Go se puede compilar para otra plataforma con
> `CGO_ENABLED=0`, pero un **instalador de Tauri** necesita el toolchain de Rust
> y las dependencias nativas de la plataforma destino. Genera los instaladores
> de Tauri en cada sistema (o con runners de CI por sistema operativo).

---

## Cómo añadir una nueva plataforma

1. En `scripts/compilar-binarios.sh`, añade una entrada al arreglo
   `PLATAFORMAS` con el formato `goos/goarch:nombre-del-artefacto`.
2. Si el instalador debe soportarla, amplía la detección de arquitectura en
   `scripts/install.sh` y/o `scripts/install.ps1`.
3. Añade la fila correspondiente a la tabla de artefactos de este README.
4. Ejecuta `make compilar-todas-plataformas` para comprobar que se genera.

No es necesario tocar el código Go: la TUI es independiente de la plataforma.

---

## Limitaciones de Tauri y de las webs remotas

La aplicación generada carga la URL remota en la WebView del sistema
(WebView2 en Windows, WKWebView en macOS, WebKitGTK en Linux). Por tanto:

- **Depende del motor del sistema.** El renderizado y las capacidades pueden
  variar entre sistemas operativos y versiones.
- **Sin acceso nativo.** Por diseño, la página remota **no** recibe comandos
  Tauri, acceso al sistema de archivos ni capacidades privilegiadas. No se
  declaran capacidades (`capabilities`) para contenido remoto.
- **Autenticación y cookies.** Funciones que dependen de extensiones del
  navegador, gestores de contraseñas, pagos con huella, DRM o APIs exclusivas
  de un navegador concreto pueden no funcionar.
- **Sitios con `X-Frame-Options` o CSP estrictas.** Normalmente no afectan
  porque se carga como ventana de nivel superior, pero algunos servicios
  detectan webviews y bloquean el acceso.
- **Descargas y ventanas emergentes.** El comportamiento de descargas y
  `window.open` puede diferir de un navegador de escritorio.
- **CSP.** Skynative no inyecta una CSP en contenido remoto; la página aplica la
  suya. No se desactiva ninguna protección para «hacer que una URL funcione».

---

## Solución de errores frecuentes

| Síntoma | Causa y solución |
| --- | --- |
| `Skynative necesita una terminal interactiva` | Se ejecutó a través de una tubería o sin TTY. Ábrelo directamente en una terminal. |
| `esquema no permitido` | Solo se admiten `http` y `https`. No uses `file:`, `javascript:` ni `data:`. |
| `el directorio ... ya existe y no está vacío` | Elige otra carpeta o nombre, o elimina el directorio. |
| `no se encontró la CLI de Tauri` | Instala `cargo install tauri-cli --version "^2"` o `npm install -g @tauri-apps/cli`. |
| `Dependencias de escritorio (Linux): faltan ...` | Instala los paquetes indicados por el diagnóstico del entorno. |
| La compilación de Tauri falla | Abre **Diagnóstico del entorno** y revisa la salida completa de Cargo en la pantalla de progreso. |
| `el checksum no coincide` al instalar | La descarga se corrompió o fue manipulada. Vuelve a intentarlo desde la release oficial. |
| En Windows se abre una consola negra | Es un binario de consola para interoperar con la terminal; la app Tauri generada no muestra consola. |
| La web no carga o se ve en blanco | Puede requerir autenticación o bloquear webviews. Revisa las limitaciones anteriores. |

---

## Apartado para desarrolladores

### Paquetes

- **`interno/version`** — versión, commit y fecha de compilación (sustituibles
  con `-ldflags -X`) y metadatos del repositorio.
- **`interno/configuracion`** — modelo `Configuracion`, valores predeterminados
  y toda la validación (`ValidarURL`, `ValidarNombre`, `ValidarDimensiones`,
  `GenerarIdentificador`, `GenerarNombrePaquete`, `GenerarSlug`).
- **`interno/generador`** — `Generar` escribe el proyecto; `iconos.go` produce
  los PNG, ICO e ICNS por código; `construirTauriConf` serializa la
  configuración de Tauri.
- **`interno/compilador`** — `Compilar` lanza la CLI de Tauri;
  `BuscarArtefactos` localiza los paquetes y el binario.
- **`interno/instalador`** — `Instalar` registra la app compilada en el menú del
  sistema: entrada `.desktop` e icono (Linux) o copia del `.app` (macOS).
- **`interno/herramientas`** — `Plan` describe los pasos para instalar Rust, la
  CLI de Tauri y las dependencias nativas; `Instalar` los ejecuta con la salida
  en vivo (sin `shell=true`; rustup se descarga con `net/http` y se ejecuta el
  fichero descargado).
- **`interno/diagnostico`** — `Ejecutar` comprueba Go, Rust, Cargo, la CLI de
  Tauri y las dependencias por plataforma.
- **`interno/proyectos`** — registro JSON en el directorio de configuración del
  usuario (`SKYNATIVE_DIRECTORIO_CONFIG` lo sobrescribe).
- **`interno/tui`** — interfaz de terminal: `modelo.go` (estado), `vistas.go`
  (render con Lip Gloss), `logo.go` (logotipo centrado), `actualizar.go`
  (`Update`), `acciones.go` (comandos y mensajes), `explorador.go` (selector de
  carpetas) y `estilos.go`.
- **`plantillas`** — archivos de Tauri incrustados con `go:embed`.

### Convenciones

- Código, comentarios, mensajes y documentación en español.
- Identificadores de Go en español cuando no chocan con las herramientas
  (`NombrePaquete`, `DirectorioSalida`); los términos que las herramientas
  externas exigen (por ejemplo `Cargo.toml`, `frontendDist`, `build.rs`) se
  mantienen tal cual.
- Sin comentarios innecesarios ni `shell=true`.
- Toda la salida de error devuelve un código de salida distinto de cero.

### Añadir una pantalla o acción a la TUI

1. Define el estado en `interno/tui/modelo.go`.
2. Gestiona las teclas en `interno/tui/actualizar.go` (`Update`).
3. Dibuja la pantalla en `interno/tui/vistas.go` (`View`) con los estilos de
   `estilos.go`.
4. Encapsula cualquier proceso o acceso a disco como un `tea.Cmd` en
   `acciones.go`, devolviendo un mensaje que `Update` consuma.
5. Añade pruebas en `interno/tui/*_test.go`.

Consulta `CONTRIBUTING.md` para las pautas de contribución.

---

## Licencia

MIT. Ver [LICENSE](LICENSE).
