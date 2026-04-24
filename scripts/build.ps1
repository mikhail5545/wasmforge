[CmdletBinding()]
param(
    [switch]$SkipNmpInstall
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Resolve-Path (Join-Path $ScriptDir "..")
$UiDir = Join-Path $RootDir "ui\admin-ui"
$UiOutDir = Join-Path $UiDir "out"
$EmbedDir = Join-Path $RootDir "pkg\ui\out"
$BinDir = Join-Path $RootDir "bin"
$BinPath = Join-Path $BinDir "wasmforge.exe"

function Require-Command([string]$Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' not found in PATH."
    }
}

Require-Command npm
Require-Command go

Write-Host "==> Building Admin UI"
Push-Location $UiDir
try{
    if (-not $SkipNmpInstall) {
        npm install
    }
    npm run build
}
finally{
    Pop-Location
}

Write-Host "==> Preparing embedded UI output"
if (Test-Path $EmbedDir) {
    Remove-Item -Path $EmbedDir -Recurse -Force
}
New-Item -ItemType Directory -Path $EmbedDir -Force | Out-Null
Copy-Item -Path (Join-Path $UiOutDir "*") -Destination $EmbedDir -Recurse -Force

Write-Host "==> Building Go gateway"
New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
go build -o $BinPath (Join-Path $RootDir "cmd\gateway\main.go")

Write-Host "Build complete: $BinPath"