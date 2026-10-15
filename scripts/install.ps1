<#
    Instalador de Skynative para Windows (PowerShell).

    Uso:
      irm https://github.com/eleodevv/skynative/releases/latest/download/install.ps1 | iex

    Variables opcionales (o parámetros):
      -Version  versión a instalar (por defecto "latest"; acepta "v0.1.0")
      -Dir      directorio de instalación (por defecto %LOCALAPPDATA%\Programs\Skynative)
      SKYNATIVE_REPO  repositorio "usuario/proyecto" (por defecto "eleodevv/skynative")

    No ejecuta ningún binario antes de verificar su checksum SHA-256 y no
    requiere privilegios de administrador (instala en el perfil del usuario y
    modifica solo el PATH del usuario).

    Desinstalación:
      irm https://github.com/eleodevv/skynative/releases/latest/download/desinstalar.ps1 | iex
#>
param(
    [string]$Version = $(if ($env:SKYNATIVE_VERSION) { $env:SKYNATIVE_VERSION } else { "latest" }),
    [string]$Dir = $(if ($env:SKYNATIVE_DIR) { $env:SKYNATIVE_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\Skynative" })
)

$ErrorActionPreference = "Stop"
$Repo = if ($env:SKYNATIVE_REPO) { $env:SKYNATIVE_REPO } else { "eleodevv/skynative" }

# Detectar la arquitectura.
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "arquitectura no compatible: $env:PROCESSOR_ARCHITECTURE (solo se admiten amd64 y arm64)" }
}

$asset = "skynative-windows-$arch.exe"
if ($Version -eq "latest") {
    $base = "https://github.com/$Repo/releases/latest/download"
} else {
    $base = "https://github.com/$Repo/releases/download/$Version"
}
$url = "$base/$asset"

$tmp = Join-Path $env:TEMP ("skynative-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "Descargando $asset ($Version)..."
    Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmp $asset)
    Invoke-WebRequest -Uri "$url.sha256" -OutFile (Join-Path $tmp "$asset.sha256")

    # Verificar el checksum antes de instalar.
    $esperado = ((Get-Content (Join-Path $tmp "$asset.sha256")) -split '\s+')[0].ToLower()
    $obtenido = (Get-FileHash -Algorithm SHA256 (Join-Path $tmp $asset)).Hash.ToLower()
    if ($esperado -ne $obtenido) {
        throw "el checksum no coincide; se cancela la instalación"
    }
    Write-Host "Checksum verificado."

    New-Item -ItemType Directory -Force -Path $Dir | Out-Null
    $destino = Join-Path $Dir "skynative.exe"
    Copy-Item (Join-Path $tmp $asset) $destino -Force
    Write-Host "Instalado en $destino"

    $rutaUsuario = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($rutaUsuario -notlike "*$Dir*") {
        [Environment]::SetEnvironmentVariable("Path", "$rutaUsuario;$Dir", "User")
        Write-Host "Se añadió $Dir al PATH del usuario. Reinicia la terminal para aplicarlo."
    }

    Write-Host ""
    Write-Host "Comprueba la instalación con: skynative --version"
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
