@echo off
setlocal EnableExtensions DisableDelayedExpansion
chcp 65001 >nul
set "SRC=%~dp0"
set "GAME=%~1"
if "%GAME%"=="" set /p "GAME=Diretorio raiz do Elden Ring ou Game: "
if "%GAME%"=="" (echo [ERRO] Diretorio obrigatorio.&exit /b 1)
if exist "%GAME%\Game\eldenring.exe" set "GAME=%GAME%\Game"
if not exist "%GAME%\eldenring.exe" (echo [ERRO] eldenring.exe ausente.&exit /b 1)
if not exist "%GAME%\dinput8.dll" (echo [ERRO] Elden Mod Loader nao detectado.&exit /b 1)
if not exist "%GAME%\mod_loader_config.ini" (echo [ERRO] Configuracao do loader ausente.&exit /b 1)
if not exist "%SRC%dist\ArayMinimap.dll" (echo [ERRO] DLL nao compilada.&exit /b 1)
if not exist "%SRC%dist\ArayMinimap.exe" (echo [ERRO] EXE nao compilado.&exit /b 1)
tasklist /FI "IMAGENAME eq eldenring.exe" | findstr /I /C:"eldenring.exe" >nul
if not errorlevel 1 (echo [ERRO] Feche o jogo antes de instalar.&exit /b 1)
set "MODS=%GAME%\mods"
if not exist "%MODS%" mkdir "%MODS%"
if errorlevel 1 (echo [ERRO] Sem permissao para criar pasta mods.&exit /b 1)
set "ARAY=%MODS%\ArayMinimap"
if not exist "%ARAY%" mkdir "%ARAY%"
if not exist "%ARAY%\data" mkdir "%ARAY%\data"
set "BACKUP=%ARAY%\backup-pre-instalacao"
if not exist "%BACKUP%" mkdir "%BACKUP%"
if exist "%MODS%\ArayMinimap.dll" copy /Y "%MODS%\ArayMinimap.dll" "%BACKUP%\ArayMinimap.dll" >nul
if exist "%ARAY%\ArayMinimap.exe" copy /Y "%ARAY%\ArayMinimap.exe" "%BACKUP%\ArayMinimap.exe" >nul
if exist "%MODS%\MapForGoblins.dll" (
 echo Map for Goblins presente. Recomenda-se desativar para evitar conflito.
 choice /C SN /N /M "Desativar com backup? [S/N] "
 if errorlevel 2 goto skip_old
 move /Y "%MODS%\MapForGoblins.dll" "%BACKUP%\MapForGoblins.dll" >nul
 if exist "%MODS%\MapForGoblins.ini" move /Y "%MODS%\MapForGoblins.ini" "%BACKUP%\MapForGoblins.ini" >nul
)
:skip_old
copy /Y "%SRC%dist\ArayMinimap.dll" "%MODS%\ArayMinimap.dll" >nul
if errorlevel 1 (echo [ERRO] Copia da DLL falhou.&exit /b 1)
copy /Y "%SRC%dist\ArayMinimap.exe" "%ARAY%\ArayMinimap.exe" >nul
if errorlevel 1 (echo [ERRO] Copia do EXE falhou.&exit /b 1)
if not exist "%ARAY%\ArayMinimap.exe" (echo [ERRO] Instalacao incompleta.&exit /b 1)
echo [ARAY] Concluido. Atalho F6, jogo OFFLINE. POSICAO 1.16 AINDA NAO VALIDADA.
echo [ARAY] Pasta de dados: "%ARAY%\data"
exit /b 0
