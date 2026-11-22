#Requires -Version 5.1
[CmdletBinding()]
param([string]$GamePath = '',[switch]$RemoveUserData,[switch]$NoPrompt)
$ErrorActionPreference='Stop'
try {
 if(-not $GamePath){if($NoPrompt){throw 'GamePath obrigatorio.'};$GamePath=Read-Host 'Pasta raiz Elden Ring ou Game'}
 $game=(Resolve-Path -LiteralPath $GamePath).ProviderPath
 if(-not(Test-Path (Join-Path $game 'eldenring.exe'))){$game=Join-Path $game 'Game'}
 if(-not(Test-Path (Join-Path $game 'eldenring.exe'))){throw 'Jogo nao encontrado'}
 if(Get-Process -Name eldenring -ErrorAction SilentlyContinue){throw 'Feche o jogo antes de desinstalar'}
 $mods=Join-Path $game 'mods'
 $dll=Join-Path $mods 'ArayMinimap.dll'
 $folder=Join-Path $mods 'ArayMinimap'
 $exe=Join-Path $folder 'ArayMinimap.exe'
 if(Test-Path $dll){Remove-Item -LiteralPath $dll -Force;Write-Host "Removida DLL: $dll"}
 if(Test-Path $exe){Remove-Item -LiteralPath $exe -Force;Write-Host "Removido EXE: $exe"}
 $data=Join-Path $folder 'data'
 if($RemoveUserData -and (Test-Path $data)){
  $safe=Join-Path (Split-Path $game -Parent) ("_ArayMinimapBackups\dados-removidos-"+(Get-Date -Format 'yyyyMMdd-HHmmss'))
  New-Item -ItemType Directory -Path (Split-Path $safe -Parent) -Force|Out-Null
  Move-Item -LiteralPath $data -Destination $safe -Force
  Write-Host "Dados preservados no backup: $safe"
 }
 if(Test-Path $folder){$remaining=Get-ChildItem -LiteralPath $folder -Force;if(@($remaining).Count -eq 0){Remove-Item -LiteralPath $folder -Force}}
 Write-Host 'Desinstalacao concluida. Elden Mod Loader e demais mods preservados.'
 exit 0
}catch{Write-Host "[ARAY][ERRO] $($_.Exception.Message)" -ForegroundColor Red;exit 1}
