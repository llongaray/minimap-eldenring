@echo off
setlocal EnableExtensions DisableDelayedExpansion
set "GAME=%~1"
if "%GAME%"=="" set /p "GAME=Diretorio raiz do Elden Ring ou Game: "
if exist "%GAME%\Game\eldenring.exe" set "GAME=%GAME%\Game"
if not exist "%GAME%\eldenring.exe" (echo [ERRO] Jogo nao encontrado.&exit /b 1)
tasklist /FI "IMAGENAME eq eldenring.exe" | findstr /I /C:"eldenring.exe" >nul
if not errorlevel 1 (echo [ERRO] Feche o jogo.&exit /b 1)
if exist "%GAME%\mods\ArayMinimap.dll" del /F /Q "%GAME%\mods\ArayMinimap.dll"
if exist "%GAME%\mods\ArayMinimap\ArayMinimap.exe" del /F /Q "%GAME%\mods\ArayMinimap\ArayMinimap.exe"
if exist "%GAME%\mods\ArayMinimap\ArayMinimap-Check.exe" del /F /Q "%GAME%\mods\ArayMinimap\ArayMinimap-Check.exe"
echo [ARAY] Binarios removidos. Dados do usuario preservados em mods\ArayMinimap\data.
exit /b 0
