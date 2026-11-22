#Requires -Version 5.1
<#+
.SYNOPSIS
  Instala Aray Minimap sem substituir o Elden Mod Loader nem o save do jogo.
.EXAMPLE
  .\install.ps1 -GamePath 'E:\Games\ELDEN RING' -DisableMapForGoblins
#>
[CmdletBinding()]
param(
 [string]$GamePath = '',
 [string]$MapSource = '',
 [switch]$DisableMapForGoblins,
 [switch]$NoPrompt
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
function Log([string]$text){Write-Host "[ARAY] $text";if($script:LogFile){Add-Content -LiteralPath $script:LogFile -Value "$(Get-Date -Format o) $text" -Encoding UTF8}}
$script:LogFile = $null
try {
 if (-not $GamePath) {
   if ($NoPrompt) {throw 'GamePath obrigatorio em modo nao interativo.'}
   $GamePath = Read-Host 'Diretorio raiz do Elden Ring ou pasta Game'
 }
 if (-not (Test-Path -LiteralPath $GamePath -PathType Container)){throw "Diretorio nao existe: $GamePath"}
 $GamePath = (Resolve-Path -LiteralPath $GamePath).ProviderPath
 $game = $GamePath
 if(-not (Test-Path -LiteralPath (Join-Path $game 'eldenring.exe'))){$game=Join-Path $GamePath 'Game'}
 if(-not (Test-Path -LiteralPath (Join-Path $game 'eldenring.exe'))){throw "eldenring.exe nao encontrado em: $game"}
 if(Get-Process -Name eldenring -ErrorAction SilentlyContinue){throw 'Feche o Elden Ring antes de instalar.'}
 $loader=Join-Path $game 'dinput8.dll'
 $config=Join-Path $game 'mod_loader_config.ini'
 if(-not(Test-Path -LiteralPath $loader) -or -not(Test-Path -LiteralPath $config)){
  throw 'Elden Mod Loader nao detectado. Instale primeiro a versao oficial de https://www.nexusmods.com/eldenring/mods/117 (dinput8.dll + config).'
 }
 $dll=Join-Path $PSScriptRoot 'dist\ArayMinimap.dll'
 $exe=Join-Path $PSScriptRoot 'dist\ArayMinimap.exe'
 $checker=Join-Path $PSScriptRoot 'dist\ArayMinimap-Check.exe'
 foreach($f in @($dll,$exe,$checker)){if(-not(Test-Path -LiteralPath $f)){throw "Binario ausente: $f"};if((Get-Item -LiteralPath $f).Length -lt 1024){throw "Arquivo invalido: $f"}}
 if($MapSource){
   $sourceCandidate=$MapSource
   if(Test-Path -LiteralPath $MapSource -PathType Container){$sourceCandidate=Join-Path $MapSource 'markers.js'}
   if(-not(Test-Path -LiteralPath $sourceCandidate)){throw "Arquivo de marcadores nao encontrado: $sourceCandidate"}
 }
 $root=Split-Path -Parent $game
 $backup=Join-Path $root ("_ArayMinimapBackups\{0}" -f (Get-Date -Format 'yyyyMMdd-HHmmss'))
 New-Item -ItemType Directory -Path $backup -Force|Out-Null
 $script:LogFile=Join-Path $backup 'install.log'
 Log "Jogo: $game"
 Log "Backup: $backup"
 $mods=Join-Path $game 'mods'
 New-Item -ItemType Directory -Path $mods -Force|Out-Null
 $targetDll=Join-Path $mods 'ArayMinimap.dll'
 $targetDir=Join-Path $mods 'ArayMinimap'
 $targetExe=Join-Path $targetDir 'ArayMinimap.exe'
 $targetChecker=Join-Path $targetDir 'ArayMinimap-Check.exe'
 $targetData=Join-Path $targetDir 'data'
 $copyBackup=Join-Path $backup 'anterior'
 New-Item -ItemType Directory -Path $copyBackup -Force|Out-Null
 foreach($existing in @($targetDll,$targetExe)){
   if(Test-Path -LiteralPath $existing){Copy-Item -LiteralPath $existing -Destination (Join-Path $copyBackup (Split-Path -Leaf $existing)) -Force;Log "Copia da versao anterior: $existing"}
 }
 if(-not(Test-Path -LiteralPath $targetDir)){New-Item -ItemType Directory -Path $targetDir -Force|Out-Null}
 if(-not(Test-Path -LiteralPath $targetData)){New-Item -ItemType Directory -Path $targetData -Force|Out-Null}
 if(-not $DisableMapForGoblins -and -not $NoPrompt -and (Test-Path -LiteralPath (Join-Path $mods 'MapForGoblins.dll'))){
   $confirm=Read-Host 'Desativar Map for Goblins para evitar conflito? [S/n]'
   if($confirm -eq '' -or $confirm -match '^[sSyY]'){$DisableMapForGoblins=$true}
 }
 $disabled=@()
 if($DisableMapForGoblins){
   foreach($name in @('MapForGoblins.dll','MapForGoblins.ini')){
     $from=Join-Path $mods $name
     if(Test-Path -LiteralPath $from){$to=Join-Path $backup $name;Move-Item -LiteralPath $from -Destination $to -Force;$disabled+=@{from=$from;to=$to};Log "Mod anterior desativado com backup: $name"}
   }
 }
 try {
   # Stage copies and avoid partial replacement of binaries.
   $stage=Join-Path $targetDir '.ArayMinimap.new.exe'
   Copy-Item -LiteralPath $exe -Destination $stage -Force
   if((Get-FileHash -Algorithm SHA256 -LiteralPath $stage).Hash -ne (Get-FileHash -Algorithm SHA256 -LiteralPath $exe).Hash){throw 'Hash do EXE diferente.'}
   Copy-Item -LiteralPath $dll -Destination ($targetDll+'.new') -Force
   if((Get-FileHash -Algorithm SHA256 -LiteralPath ($targetDll+'.new')).Hash -ne (Get-FileHash -Algorithm SHA256 -LiteralPath $dll).Hash){throw 'Hash da DLL diferente.'}
   Move-Item -LiteralPath $stage -Destination $targetExe -Force
   Copy-Item -LiteralPath $checker -Destination $targetChecker -Force
   Move-Item -LiteralPath ($targetDll+'.new') -Destination $targetDll -Force
   # The reference source is imported locally, never republished by this project.
   if($MapSource){
     $src=$MapSource
     if(Test-Path -LiteralPath $MapSource -PathType Container){$src=Join-Path $MapSource 'markers.js'}
     if(-not(Test-Path -LiteralPath $src)){throw "Arquivo de marcadores nao encontrado: $src"}
     Copy-Item -LiteralPath $src -Destination (Join-Path $targetData 'markers.js') -Force
     Log 'Marcadores de referencia importados localmente.'
   }
 } catch {
   foreach($entry in $disabled){if(Test-Path -LiteralPath $entry.to){Move-Item -LiteralPath $entry.to -Destination $entry.from -Force}}
   foreach($name in @('ArayMinimap.dll','ArayMinimap.exe')){
      $old=Join-Path $copyBackup $name
      $restore=if($name -eq 'ArayMinimap.dll'){$targetDll}else{$targetExe}
      if(Test-Path -LiteralPath $old){Copy-Item -LiteralPath $old -Destination $restore -Force}
      elseif(Test-Path -LiteralPath $restore){Remove-Item -LiteralPath $restore -Force}
   }
   foreach($temp in @((Join-Path $targetDir '.ArayMinimap.new.exe'),($targetDll+'.new'))){if(Test-Path -LiteralPath $temp){Remove-Item -LiteralPath $temp -Force}}
   throw
 }
 if(-not(Test-Path -LiteralPath $targetDll) -or -not(Test-Path -LiteralPath $targetExe)){throw 'Instalacao nao passou na validacao final.'}
 Log 'INSTALACAO CONCLUIDA. O jogo deve ser iniciado OFFLINE.'
 Log "DLL: $targetDll"
 Log "EXE: $targetExe"
 Log "Dados do mapa: $targetData"
 Write-Host 'ATENCAO: build experimental, leitura de posicao ainda nao validada no Elden Ring 1.16.' -ForegroundColor Yellow
 exit 0
} catch {
 Write-Host "[ARAY][ERRO] $($_.Exception.Message)" -ForegroundColor Red
 exit 1
}
