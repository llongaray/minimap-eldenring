#Requires -Version 5.1
# Instala a DLL direta do Aray Minimap; preserva saves, tiles e outros mods.
[CmdletBinding()]
param(
    [string]$GamePath = '',
    [string]$MapSource = '',
    [switch]$DisableMapForGoblins,
    [switch]$NoPrompt
)
$ErrorActionPreference = 'Stop'
$original = @()
$targetDll = $null
$staging = $null
try {
    if (-not $GamePath) {
        if ($NoPrompt) { throw 'GamePath obrigatorio em modo nao interativo.' }
        $GamePath = Read-Host 'Pasta raiz do Elden Ring ou Game'
    }
    $root = (Resolve-Path -LiteralPath $GamePath).ProviderPath
    $game = $root
    if (-not (Test-Path -LiteralPath (Join-Path $game 'eldenring.exe'))) { $game = Join-Path $root 'Game' }
    if (-not (Test-Path -LiteralPath (Join-Path $game 'eldenring.exe'))) { throw 'eldenring.exe nao encontrado.' }
    if (Get-Process -Name eldenring -ErrorAction SilentlyContinue) { throw 'Feche o Elden Ring antes de instalar.' }
    foreach ($required in @('dinput8.dll', 'mod_loader_config.ini')) {
        if (-not (Test-Path -LiteralPath (Join-Path $game $required))) { throw "Elden Mod Loader incompleto: $required" }
    }
    $dll = Join-Path $PSScriptRoot 'dist\ArayMinimap.dll'
    if (-not (Test-Path -LiteralPath $dll)) { throw 'DLL nativa nao compilada. Baixe o artefato de CI.' }
    if ((Get-Item -LiteralPath $dll).Length -lt 1024000) { throw 'DLL pequena demais: possivel bridge antigo. Baixe a DLL nativa.' }
    $mods = Join-Path $game 'mods'
    $folder = Join-Path $mods 'ArayMinimap'
    $data = Join-Path $folder 'data'
    New-Item -ItemType Directory -Force -Path $data | Out-Null
    $backup = Join-Path (Split-Path -Parent $game) ("_ArayMinimapBackups\native-{0}-{1}" -f (Get-Date -Format 'yyyyMMdd-HHmmss'),[Guid]::NewGuid().ToString('N').Substring(0,8))
    New-Item -ItemType Directory -Force -Path $backup | Out-Null
    Write-Host "[ARAY] Backup: $backup"
    Get-Process -Name 'ArayMinimap', 'ArayMinimap-Check' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    $targetDll = Join-Path $mods 'ArayMinimap.dll'
    $oldFiles = @(
        $targetDll,
        (Join-Path $mods 'ArayMinimap.dll.disabled'),
        (Join-Path $folder 'ArayMinimap.exe'),
        (Join-Path $folder 'ArayMinimap-Check.exe')
    )
    if ($DisableMapForGoblins) {
        $oldFiles += (Join-Path $mods 'MapForGoblins.dll')
        $oldFiles += (Join-Path $mods 'MapForGoblins.ini')
    } elseif (-not $NoPrompt -and (Test-Path -LiteralPath (Join-Path $mods 'MapForGoblins.dll'))) {
        $answer = Read-Host 'Desativar Map for Goblins com backup? [S/n]'
        if (-not $answer -or $answer -match '^[sSyY]') {
            $oldFiles += (Join-Path $mods 'MapForGoblins.dll')
            $oldFiles += (Join-Path $mods 'MapForGoblins.ini')
        }
    }
    if ($MapSource) {
        $markers = $MapSource
        if (Test-Path -LiteralPath $MapSource -PathType Container) { $markers = Join-Path $MapSource 'markers.js' }
        if (-not (Test-Path -LiteralPath $markers -PathType Leaf)) { throw "Marcadores nao encontrados: $markers" }
    }
    foreach ($old in $oldFiles) {
        if (Test-Path -LiteralPath $old) {
            $destination = Join-Path $backup (Split-Path -Leaf $old)
            Move-Item -LiteralPath $old -Destination $destination -ErrorAction Stop
            $original += [PSCustomObject]@{Old=$old;Backup=$destination}
            Write-Host "[ARAY] Backup: $(Split-Path -Leaf $old)"
        }
    }
    $staging = Join-Path $mods 'ArayMinimap.dll.new'
    Copy-Item -LiteralPath $dll -Destination $staging -Force
    if ((Get-FileHash -Algorithm SHA256 $staging).Hash -ne (Get-FileHash -Algorithm SHA256 $dll).Hash) {
        throw 'Hash da DLL apos copia nao confere'
    }
    Move-Item -LiteralPath $staging -Destination $targetDll
    if ($MapSource) {
        Copy-Item -LiteralPath $markers -Destination (Join-Path $data 'markers.js') -Force
    }
    Write-Host "[ARAY] DLL direta instalada: $targetDll"
    Write-Host "[ARAY] Nenhum executavel auxiliar requerido. Marcadores e tiles preservados em $data"
    Write-Host '[AVISO] Build experimental: abrir offline; nao validada em jogo real.' -ForegroundColor Yellow
    exit 0
} catch {
    Write-Host "[ARAY][ERRO] $($_.Exception.Message)" -ForegroundColor Red
    if ($staging -and (Test-Path -LiteralPath $staging)) { Remove-Item -LiteralPath $staging -Force -ErrorAction SilentlyContinue }
    if ($targetDll -and (Test-Path -LiteralPath $targetDll)) { Remove-Item -LiteralPath $targetDll -Force -ErrorAction SilentlyContinue }
    [array]::Reverse($original)
    foreach ($old in $original) {
        if (Test-Path -LiteralPath $old.Backup) { Move-Item -LiteralPath $old.Backup -Destination $old.Old -Force -ErrorAction SilentlyContinue }
    }
    exit 1
}
