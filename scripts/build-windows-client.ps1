param(
    [string]$QtPrefix = "C:\Qt\6.11.1\mingw_64",
    [string]$MingwBin = "C:\Qt\Tools\mingw1310_64\bin",
    [string]$BuildDir = "build"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$client = Join-Path $root "client\desktop-qt"
$env:Path = "$MingwBin;$QtPrefix\bin;$env:Path"

Push-Location $client
try {
    cmake -S . -B $BuildDir -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=$QtPrefix
    cmake --build $BuildDir --config Release
    Write-Host "Built $client\$BuildDir\NetCordDesktop.exe"
} finally {
    Pop-Location
}
