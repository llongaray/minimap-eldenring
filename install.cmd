@echo off
setlocal EnableExtensions DisableDelayedExpansion
chcp 65001 >nul
set "GAME=%~1"
if "%GAME%"=="" set /p "GAME=Diretorio raiz do Elden Ring ou Game: "
if "%GAME%"=="" (echo [ERRO] Informe a pasta do jogo.&exit /b 1)
if exist "%GAME%\Game\eldenring.exe" set "GAME=%GAME%\Game"
if not exist "%GAME%\eldenring.exe" (echo [ERRO] eldenring.exe nao encontrado.&exit /b 1)
if not exist "%GAME%\dinput8.dll" (echo [ERRO] Elden Mod Loader ausente.&exit /b 1)
if not exist "%GAME%\mod_loader_config.ini" (echo [ERRO] mod_loader_config.ini ausente.&exit /b 1)
if not exist "%~dp0dist\ArayMinimap.dll" (echo [ERRO] DLL nativa ausente em dist. Baixe o artefato do GitHub Actions.&exit /b 1)
tasklist /FI "IMAGENAME eq eldenring.exe" | findstr /I /C:"eldenring.exe" >nul
if not errorlevel 1 (echo [ERRO] Feche o jogo antes de instalar.&exit /b 1)
set "MODS=%GAME%\mods"
set "DATA=%MODS%\ArayMinimap\data"
set "BACKUP=%MODS%\ArayMinimap\backup-native-%RANDOM%%RANDOM%"
if not exist "%MODS%" mkdir "%MODS%"
if not exist "%MODS%\ArayMinimap" mkdir "%MODS%\ArayMinimap"
if not exist "%DATA%" mkdir "%DATA%"
mkdir "%BACKUP%"
if errorlevel 1 (echo [ERRO] Sem permissao de escrita.&exit /b 1)
echo [ARAY] Backup de binarios antigos: "%BACKUP%"
taskkill /F /IM ArayMinimap.exe >nul 2>nul
taskkill /F /IM ArayMinimap-Check.exe >nul 2>nul
if exist "%MODS%\ArayMinimap.dll" move /Y "%MODS%\ArayMinimap.dll" "%BACKUP%\ArayMinimap.dll" >nul
if exist "%MODS%\ArayMinimap.dll.disabled" move /Y "%MODS%\ArayMinimap.dll.disabled" "%BACKUP%\ArayMinimap.dll.disabled" >nul
if exist "%MODS%\ArayMinimap\ArayMinimap.exe" move /Y "%MODS%\ArayMinimap\ArayMinimap.exe" "%BACKUP%\ArayMinimap.exe" >nul
if exist "%MODS%\ArayMinimap\ArayMinimap-Check.exe" move /Y "%MODS%\ArayMinimap\ArayMinimap-Check.exe" "%BACKUP%\ArayMinimap-Check.exe" >nul
if exist "%MODS%\MapForGoblins.dll" (
 echo [ARAY] Map for Goblins encontrado. Recomenda-se desativar para evitar conflito.
 choice /C SN /N /M "Desativar com backup? [S/N] "
 if errorlevel 2 goto skip_goblin
 move /Y "%MODS%\MapForGoblins.dll" "%BACKUP%\MapForGoblins.dll" >nul
 if exist "%MODS%\MapForGoblins.ini" move /Y "%MODS%\MapForGoblins.ini" "%BACKUP%\MapForGoblins.ini" >nul
)
:skip_goblin
copy /Y "%~dp0dist\ArayMinimap.dll" "%MODS%\ArayMinimap.dll" >nul
if errorlevel 1 (echo [ERRO] Falha ao copiar DLL. Recupere do backup acima.&exit /b 1)
if not exist "%MODS%\ArayMinimap.dll" (echo [ERRO] DLL nao encontrada apos copiar.&exit /b 1)
echo.
echo [ARAY] DLL nativa instalada em: "%MODS%\ArayMinimap.dll"
echo [ARAY] Nenhum ArayMinimap.exe e necessario.
echo [ARAY] Marcadores e tiles preservados: "%DATA%"
echo [ARAY] Execute o jogo offline. F6 abre o mapa experimental.
echo [AVISO] Build preview: comportamento no jogo ainda exige validacao.
exit /b 0
