param(
    [string]$QtPrefix = "C:\Qt\6.11.1\mingw_64",
    [string]$MingwBin = "C:\Qt\Tools\mingw1310_64\bin",
    [string]$BuildDir = "build",
    [string]$PackageDir = "package-windows"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$client = Join-Path $root "client\desktop-qt"
$env:Path = "$MingwBin;$QtPrefix\bin;$env:Path"

& (Join-Path $PSScriptRoot "build-windows-client.ps1") -QtPrefix $QtPrefix -MingwBin $MingwBin -BuildDir $BuildDir

$out = Join-Path $client $PackageDir
if (Test-Path $out) {
    Remove-Item -LiteralPath $out -Recurse -Force
}
New-Item -ItemType Directory -Path $out | Out-Null

Copy-Item -LiteralPath (Join-Path $client "$BuildDir\NetCordDesktop.exe") -Destination $out
windeployqt --qmldir (Join-Path $client "qml") (Join-Path $out "NetCordDesktop.exe")

Write-Host "Packaged $out"
