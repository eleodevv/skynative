<#
    Desinstalador de Skynative para Windows (PowerShell).

    Uso:
      irm https://github.com/eleodevv/skynative/releases/latest/download/desinstalar.ps1 | iex

    Variables opcionales (o parámetros):
      -Dir  directorio donde se instaló
            (por defecto %LOCALAPPDATA%\Programs\Skynative)

    Elimina el ejecutable «skynative.exe», quita su directorio del PATH del
    usuario y borra la carpeta de instalación. No toca los proyectos ni las
    aplicaciones que hayas creado con la herramienta.
#>
param(
    [string]$Dir = $(if ($env:SKYNATIVE_DIR) { $env:SKYNATIVE_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\Skynative" })
)

$ErrorActionPreference = "Stop"

$binario = Join-Path $Dir "skynative.exe"

if (-not (Test-Path $binario)) {
    Write-Host "Skynative no está instalado en $binario"
    Write-Host ""
    Write-Host "Si lo instalaste en otro directorio, indícalo con -Dir:"
    Write-Host "  .\desinstalar.ps1 -Dir C:\ruta\alternativa"
    exit 0
}

# Quitar del PATH del usuario.
$rutaUsuario = [Environment]::GetEnvironmentVariable("Path", "User")
if ($rutaUsuario -and ($rutaUsuario -like "*$Dir*")) {
    $nuevas = ($rutaUsuario -split ';' | Where-Object { $_ -and $_ -ne $Dir }) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $nuevas, "User")
    Write-Host "Se quitó $Dir del PATH del usuario. Reinicia la terminal para aplicarlo."
}

Remove-Item -Recurse -Force $Dir -ErrorAction SilentlyContinue
Write-Host "Skynative eliminado de $Dir"
Write-Host ""
Write-Host "Hecho. Los proyectos creados con Skynative permanecen intactos."
